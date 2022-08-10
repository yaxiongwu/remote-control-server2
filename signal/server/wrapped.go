package server

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	rtc "github.com/yaxiongwu/remote-control-server/pkg/proto/rtc"
	"github.com/yaxiongwu/remote-control-server/pkg/stun"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	log "github.com/pion/ion-log"
	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

/*
type StatsHandler struct {
}

var connsMutex sync.Mutex
var conns map[*stats.ConnTagInfo]string = make(map[*stats.ConnTagInfo]string)

type connCtxKey struct{}

func getConnTagFromContext(ctx context.Context) (*stats.ConnTagInfo, bool) {
	tag, ok := ctx.Value(connCtxKey{}).(*stats.ConnTagInfo)
	return tag, ok
}

//TagConn可以将一些信息附加到给定的上下文。
// TagConn 用来给连接打个标签，以此来标识连接(实在是找不出还有什么办法来标识连接).
// 这个标签是个指针，可保证每个连接唯一。
// 将该指针添加到上下文中去，键为 connCtxKey{}.
func (h *StatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	fmt.Println("TagConn info:%v,ctx:%v", info, ctx)
	return context.WithValue(ctx, connCtxKey{}, info)
}

// 会在连接开始和结束时被调用，分别会输入不同的状态.
// HandleConn 会在连接开始和结束时被调用，分别会输入不同的状态.
func (h *StatsHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	tag, ok := getConnTagFromContext(ctx)
	if !ok {
		fmt.Errorf("can not get conn tag")
	}
	fmt.Println("HanadleConn ctx:", ctx, s)
	connsMutex.Lock()
	defer connsMutex.Unlock()

	switch s.(type) {
	case *stats.ConnBegin:
		conns[tag] = ""
		fmt.Printf("begin conn, tag = (%p)%#v, now connections = %d\n", tag, tag, len(conns))
	case *stats.ConnEnd:
		delete(conns, tag)
		fmt.Printf("end conn, tag = (%p)%#v, now connections = %d\n", tag, tag, len(conns))
	default:
		fmt.Printf("illegal ConnStats type\n")
	}
}

// TagRPC可以将一些信息附加到给定的上下文

func (h *StatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	fmt.Println("tagrpc...@" + info.FullMethodName)
	return ctx
}

func (h *StatsHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	switch s.(type) {
	case *stats.Begin:
		fmt.Println("handlerRPC begin...", s)
	case *stats.End:
		fmt.Println("handlerRPC End...", s)
	case *stats.InHeader:
		fmt.Println("handlerRPC InHeader...")
	case *stats.InPayload:
		fmt.Println("handlerRPC InPayload...")
	case *stats.InTrailer:
		fmt.Println("handlerRPC InTrailer...")
	case *stats.OutHeader:
		fmt.Println("handlerRPC OutHeader...")
	case *stats.OutPayload:
		fmt.Println("handlerRPC OutPayload...")
	default:
		fmt.Println("handleRPC...")
	}
}
*/
type WrapperedServerOptions struct {
	Addr                  string
	Cert                  string
	Key                   string
	AllowAllOrigins       bool
	AllowedOrigins        *[]string
	AllowedHeaders        *[]string
	UseWebSocket          bool
	WebsocketPingInterval time.Duration
}

func DefaultWrapperedServerOptions() WrapperedServerOptions {
	return WrapperedServerOptions{
		Addr:                  ":9090",
		Cert:                  "",
		Key:                   "",
		AllowAllOrigins:       true,
		AllowedHeaders:        &[]string{},
		AllowedOrigins:        &[]string{},
		UseWebSocket:          true,
		WebsocketPingInterval: 0,
	}
}

func NewWrapperedServerOptions(addr, cert, key string, websocket bool) WrapperedServerOptions {
	return WrapperedServerOptions{
		Addr:                  addr,
		Cert:                  cert,
		Key:                   key,
		AllowAllOrigins:       true,
		AllowedHeaders:        &[]string{},
		AllowedOrigins:        &[]string{},
		UseWebSocket:          true,
		WebsocketPingInterval: 0,
	}
}

type WrapperedGRPCWebServer struct {
	options    WrapperedServerOptions
	GRPCServer *grpc.Server
}

func NewWrapperedGRPCWebServer(options WrapperedServerOptions, s *grpc.Server) *WrapperedGRPCWebServer {
	return &WrapperedGRPCWebServer{
		options:    options,
		GRPCServer: s,
	}
}

type allowedOrigins struct {
	origins map[string]struct{}
}

func (a *allowedOrigins) IsAllowed(origin string) bool {
	_, ok := a.origins[origin]
	return ok
}

func makeAllowedOrigins(origins []string) *allowedOrigins {
	o := map[string]struct{}{}
	for _, allowedOrigin := range origins {
		o[allowedOrigin] = struct{}{}
	}
	return &allowedOrigins{
		origins: o,
	}
}

func (s *WrapperedGRPCWebServer) makeHTTPOriginFunc(allowedOrigins *allowedOrigins) func(origin string) bool {
	if s.options.AllowAllOrigins {
		return func(origin string) bool {
			return true
		}
	}
	return allowedOrigins.IsAllowed
}

func (s *WrapperedGRPCWebServer) makeWebsocketOriginFunc(allowedOrigins *allowedOrigins) func(req *http.Request) bool {
	if s.options.AllowAllOrigins {
		return func(req *http.Request) bool {
			return true
		}
	}
	return func(req *http.Request) bool {
		origin, err := grpcweb.WebsocketRequestOrigin(req)
		if err != nil {
			log.Warnf("%v", err)
			return false
		}
		return allowedOrigins.IsAllowed(origin)
	}
}

func (s *WrapperedGRPCWebServer) Serve() error {
	addr := s.options.Addr

	if s.options.AllowAllOrigins && s.options.AllowedOrigins != nil && len(*s.options.AllowedOrigins) != 0 {
		log.Errorf("Ambiguous --allow_all_origins and --allow_origins configuration. Either set --allow_all_origins=true OR specify one or more origins to whitelist with --allow_origins, not both.")
	}

	allowedOrigins := makeAllowedOrigins(*s.options.AllowedOrigins)

	options := []grpcweb.Option{
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithOriginFunc(s.makeHTTPOriginFunc(allowedOrigins)),
	}

	if s.options.UseWebSocket {
		log.Infof("Using websockets")
		options = append(
			options,
			grpcweb.WithWebsockets(true),
			grpcweb.WithWebsocketOriginFunc(s.makeWebsocketOriginFunc(allowedOrigins)),
		)

		if s.options.WebsocketPingInterval >= time.Second {
			log.Infof("websocket keepalive pinging enabled, the timeout interval is %s", s.options.WebsocketPingInterval.String())
			options = append(
				options,
				grpcweb.WithWebsocketPingInterval(s.options.WebsocketPingInterval),
			)
		}
	}

	if s.options.AllowedHeaders != nil && len(*s.options.AllowedHeaders) > 0 {
		options = append(
			options,
			grpcweb.WithAllowedRequestHeaders(*s.options.AllowedHeaders),
		)
	}

	wrappedServer := grpcweb.WrapServer(s.GRPCServer, options...)
	handler := func(resp http.ResponseWriter, req *http.Request) {
		wrappedServer.ServeHTTP(resp, req)
	}

	httpServer := http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(handler),
	}

	var listener net.Listener

	enableTLS := s.options.Cert != "" && s.options.Key != ""

	if enableTLS {
		cer, err := tls.LoadX509KeyPair(s.options.Cert, s.options.Key)
		if err != nil {
			log.Panicf("failed to load x509 key pair: %v", err)
			return err
		}
		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		tls, err := tls.Listen("tcp", addr, config)
		if err != nil {
			log.Panicf("failed to listen: tls %v", err)
			return err
		}
		listener = tls
	} else {
		tcp, err := net.Listen("tcp", addr)
		if err != nil {
			log.Panicf("failed to listen: tcp %v", err)
			return err
		}
		listener = tcp
	}

	log.Infof("Starting gRPC/gRPC-Web combo server, bind: %s, with TLS: %v", addr, enableTLS)

	m := cmux.New(listener)
	grpcListener := m.Match(cmux.HTTP2())
	httpListener := m.Match(cmux.HTTP1Fast())
	g := new(errgroup.Group)
	g.Go(func() error { return s.GRPCServer.Serve(grpcListener) })
	g.Go(func() error { return httpServer.Serve(httpListener) })
	g.Go(m.Serve)
	log.Infof("Run server: %v", g.Wait())
	return nil
}

func WrapperedGRPCWebServe(stun *stun.STUN, addr, cert, key string) error {
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		//grpc.StatsHandler(&StatsHandler{}),
	)

	rtc.RegisterRTCServer(grpcServer, &STUNServer{STUN: stun})
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())
	grpc_prometheus.Register(grpcServer)

	log.Infof("wrappered grpc listening %v", addr)
	options := NewWrapperedServerOptions(addr, cert, key, true)
	wrapperedSrv := NewWrapperedGRPCWebServer(options, grpcServer)
	if err := wrapperedSrv.Serve(); err != nil {
		log.Errorf("wrappered grpc listening error: %v", err)
		return err
	}
	return nil
}
