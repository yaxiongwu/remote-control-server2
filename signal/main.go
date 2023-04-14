// Package cmd contains an entrypoint for running an ion-sfu instance.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/yaxiongwu/remote-control-server2/signal/server"

	"github.com/yaxiongwu/remote-control-server2/pkg/stun"

	log "github.com/yaxiongwu/remote-control-server2/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"github.com/unrolled/secure"
	"github.com/yaxiongwu/remote-control-server2/pkg/proto/rtc"
)

type grpcConfig struct {
	Port string `mapstructure:"port"`
}

// Config defines parameters for configuring the sfu instance
type Config struct {
	//sfu.Config `mapstructure:",squash"`
	GRPC      grpcConfig       `mapstructure:"grpc"`
	LogConfig log.GlobalConfig `mapstructure:"log"`
}

var (
	conf           = Config{}
	file           string
	addr           string
	metricsAddr    string
	verbosityLevel int
	paddr          string

	enableTLS bool
	cert      string
	key       string

	logger = log.New()
)

const (
	portRangeLimit = 100
)

func showHelp() {
	fmt.Printf("Usage:%s {params}\n", os.Args[0])
	fmt.Println("      -c {config file}")
	fmt.Println("      -a {listen addr}")
	fmt.Println("      -tls (enable tls)")
	fmt.Println("      -cert {cert file}")
	fmt.Println("      -key {key file}")
	fmt.Println("      -h (show help info)")
	fmt.Println("      -v {0-10} (verbosity level, default 0)")
	fmt.Println("      -paddr {pprof listen addr}")

}

func load() bool {
	_, err := os.Stat(file)
	if err != nil {
		return false
	}

	viper.SetConfigFile(file)
	viper.SetConfigType("toml")

	err = viper.ReadInConfig()
	if err != nil {
		logger.Error(err, "config file read failed", "file", file)
		return false
	}
	err = viper.GetViper().Unmarshal(&conf)
	if err != nil {
		logger.Error(err, "sfu config file loaded failed", "file", file)
		return false
	}

	// if len(conf.WebRTC.ICEPortRange) > 2 {
	// 	logger.Error(nil, "config file loaded failed. webrtc port must be [min,max]", "file", file)
	// 	return false
	// }

	// if len(conf.WebRTC.ICEPortRange) != 0 && conf.WebRTC.ICEPortRange[1]-conf.WebRTC.ICEPortRange[0] < portRangeLimit {
	// 	logger.Error(nil, "config file loaded failed. webrtc port must be [min, max] and max - min >= portRangeLimit", "file", file, "portRangeLimit", portRangeLimit)
	// 	return false
	// }

	// if len(conf.Turn.PortRange) > 2 {
	// 	logger.Error(nil, "config file loaded failed. turn port must be [min,max]", "file", file)
	// 	return false
	// }

	logger.V(0).Info("Config file loaded", "file", file)
	return true
}

func parse() bool {
	flag.StringVar(&file, "c", "config.toml", "config file")
	flag.StringVar(&addr, "a", "0.0.0.0:5551", "address to use")
	//flag.StringVar(&addr, "a", "192.168.1.199:5551", "address to use")
	flag.BoolVar(&enableTLS, "tls", false, "enable tls")
	//flag.StringVar(&cert, "cert", "https-tls/cert.pem", "cert file")
	//flag.StringVar(&key, "key", "https-tls/key.pem", "key file")
	flag.StringVar(&key, "key", "", "key file")
	flag.StringVar(&cert, "cert", "", "cert file")
	flag.StringVar(&metricsAddr, "m", ":8100", "merics to use")
	flag.IntVar(&verbosityLevel, "v", -1, "verbosity level, higher value - more logs")
	flag.StringVar(&paddr, "paddr", "", "pprof listening address")
	help := flag.Bool("h", false, "help info")
	flag.Parse()

	if paddr == "" {
		paddr = getEnv("paddr")
	}

	if !load() {
		return false
	}

	if *help {
		return false
	}
	return true
}

func getEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return ""
}

func startMetrics(addr string) {
	// start metrics server
	m := http.NewServeMux()
	m.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{
		Handler: m,
	}

	metricsLis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error(err, "cannot bind to metrics endpoint", "addr", addr)
		os.Exit(1)
	}
	logger.Info("Metrics Listening starter", "addr", addr)

	err = srv.Serve(metricsLis)
	if err != nil {
		logger.Error(err, "debug server stopped. got err: %s")
	}
}

func main() {

	if !parse() {
		showHelp()
		os.Exit(-1)
	}

	// Check that the -v is not set (default -1)
	if verbosityLevel < 0 {
		verbosityLevel = conf.LogConfig.V
	}

	log.SetGlobalOptions(log.GlobalConfig{V: verbosityLevel})
	logger := log.New()

	logger.Info("--- Starting SFU Node ---")

	if paddr != "" {
		go func() {
			logger.Info("PProf Listening", "addr", paddr)
			_ = http.ListenAndServe(paddr, http.DefaultServeMux)
		}()
	}
	go startMetrics(metricsAddr)

	// SFU instance needs to be created with logr implementation
	//sfu.Logger = logger

	// nsfu := sfu.NewSFU(conf.Config)
	// dc := nsfu.NewDatachannel(sfu.APIChannelLabel)
	// dc.Use(datachannel.SubscriberAPI)
	nstun := stun.NewSTUN()
	go jsonGin(nstun)
	//err := server.WrapperedGRPCWebServe(nstun, addr, cert, key)
	err := server.WrapperedGRPCWebServe(nstun, addr, "bxzryd.pem", "bxzryd.key")
	if err != nil {
		logger.Error(err, "failed to serve SFU")
		os.Exit(1)
	}
}

type sourceInfo struct {
	id   string
	name string
}

func getSourceList(s *stun.STUN, sourceType rtc.SourceType) []stun.ClientInfo {
	var sourceList []stun.ClientInfo
	sources := s.GetSessions()
	for _, value := range sources {
		if value.GetSourceType() == sourceType {
			sourceList = append(sourceList, value.GetSourceClient().GetInfo())
		}
		//logger.Info("source:", "source client info ", value.GetSourceClient().GetInfo())
	}
	return sourceList
}

func jsonGin(s *stun.STUN) bool {
	r := gin.Default()

	r.GET("/sourcesList", func(c *gin.Context) {
		logger.Info("c.Query", "c.Query(sourceType) ", c.Query("sourceType"))
		list, _ := json.Marshal(getSourceList(s, rtc.SourceType(rtc.SourceType_value[c.Query("sourceType")])))
		c.JSON(200, gin.H{
			"list": string(list), //[]byte会自动转换成base64传输
		})
	})
	r.LoadHTMLGlob("web/*")
	r.GET("/vedio", func(c *gin.Context) {
		logger.Info("c.Query", "c.Query(id)", c.Query("id"))
		c.HTML(http.StatusOK, "index.html", gin.H{
			"id": c.Query("id"),
		})
	})
	r.GET("/index/:destination", func(context *gin.Context) {
		context.HTML(http.StatusOK, "index_phone.html", gin.H{
			"destination": context.Param("destination"),
		})
	})
	r.GET("/index_pc/:destination", func(context *gin.Context) {
		context.HTML(http.StatusOK, "index_pc.html", gin.H{
			"destination": context.Param("destination"),
		})
	})
	r.Use(LoadTls())
	r.RunTLS(":8080", "bxzryd.pem", "bxzryd.key")
	//r.Run(":8080")
	return true
}

func LoadTls() gin.HandlerFunc {
	return func(c *gin.Context) {
		middleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     "localhost:8080",
		})
		err := middleware.Process(c.Writer, c.Request)
		if err != nil {
			logger.Error(err, "error")
			return
		}
		c.Next()
	}
}
