//go:build !js
// +build !js

package main

import (
	"fmt"
	"net"

	"github.com/hajimehoshi/oto/v2"
	sdk "github.com/yaxiongwu/remote-control-client-go2"

	//ilog "github.com/pion/ion-log"

	// Note: If you don't have a camera or microphone or your adapters are not supported,
	//       you can always swap your adapters with our dummy adapters below.
	// _ "github.com/pion/mediadevices/pkg/driver/videotest"
	// _ "github.com/pion/mediadevices/pkg/driver/audiotest"
	//"github.com/pion/mediadevices/pkg/codec/mmal"

	// This is required to use opus audio encoder

	//"github.com/pion/mediadevices/pkg/codec/mmal"
	//"github.com/pion/mediadevices/pkg/codec/vpx"
	"github.com/BurntSushi/toml"
	log "github.com/pion/ion-log"
	_ "github.com/pion/mediadevices/pkg/driver/camera"     // This is required to register camera adapter
	_ "github.com/pion/mediadevices/pkg/driver/microphone" // This is required to register microphone adapter
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	opusdecoder "github.com/yaxiongwu/remote-control-client-go2/pkg/opus/decoder"
	rtcproto "github.com/yaxiongwu/remote-control-client-go2/pkg/proto/rtc"
)

type udpConn struct {
	conn        *net.UDPConn
	port        int
	payloadType uint8
}
type Config struct {
	MaxTimeControl int
	MaxTimeView    int
	MaxClientsNum  int
	LogLevel       int8
	Address        string
}

var (
	// log = ilog.NewLoggerWithFields(ilog.DebugLevel, "main", nil)
	config Config
)

func main() {

	configFilePath := "./config.toml"
	if _, err := toml.DecodeFile(configFilePath, &config); err != nil {
		fmt.Println("load config file error!", err)
		return
	}

	// gst.CreatePipeline("h264_omx", []*webrtc.TrackLocalStaticSample{videoTrack}, videoSrc, rtmpudp.GetConn()).Start()
	// gst.CreatePipeline("opus", []*webrtc.TrackLocalStaticSample{audioTrack}, audioSrc, rtmpudp.GetConn()).Start()

	connector := sdk.NewConnector(config.Address)
	rtc, err := sdk.NewRTC(connector)
	rtc.MaxTimeControl = config.MaxTimeControl
	rtc.MaxTimeView = config.MaxTimeView
	rtc.MaxClientsNum = config.MaxClientsNum
	if err != nil {
		panic(err)
	}

	err = rtc.RegisterNewVideoSource("PiVideoSource", "远程视频控制小车", rtcproto.SourceType_Car)
	rtc.OnDataChannelMessage = func(msg webrtc.DataChannelMessage) {
		log.Infof("rtc.OnDataChannelMessage msg:%s", msg)

	}

	rtc.OnDataChannel = func(dc *webrtc.DataChannel) {
		log.Infof("rtc.OnDatachannel:%v", dc.Label())
		dc.OnOpen(func() {
			log.Infof("%v,dc.onopen,dc.ReadyState:%v", dc.Label(), dc.ReadyState())
			//	dc.SendText("wuyaxiong nbcl")
		})
	}
	rtc.OnDataChannel = func(dc *webrtc.DataChannel) {
		log.Infof("rtc.OnDataChannel dc:%s", dc)
	}

	rtc.OnTrack = func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		codec := track.Codec()
		log.Infof("track.Codec():%v", codec)
		if codec.MimeType == "audio/opus" {
			samplingRate := 48000

			// Number of channels (aka locations) to play sounds from. Either 1 or 2.
			// 1 is mono sound, and 2 is stereo (most speakers are stereo).
			numOfChannels := 1

			// Bytes used by a channel to represent one sample. Either 1 or 2 (usually 2).
			audioBitDepth := 2

			otoCtx, readyChan, err := oto.NewContext(samplingRate, numOfChannels, audioBitDepth)
			if err != nil {
				panic("oto.NewContext failed: " + err.Error())
			}
			// It might take a bit for the hardware audio devices to be ready, so we wait on the channel.
			<-readyChan

			decoder, err := opusdecoder.NewOpusDecoder(samplingRate, numOfChannels)
			if err != nil {
				fmt.Printf("Error creating")
			}
			player := otoCtx.NewPlayer(decoder)
			defer player.Close()
			//player.Play()
			//pipeReader, pipeWriter := io.Pipe()

			b := make([]byte, 1500)
			rtpPacket := &rtp.Packet{}
			for {

				// Read
				n, _, readErr := track.Read(b)
				if readErr != nil {
					log.Errorf("OnTrack read error: %v", readErr)
					return
					//panic(readErr)
				}

				// Unmarshal the packet and update the PayloadType
				if err = rtpPacket.Unmarshal(b[:n]); err != nil {
					log.Errorf("OnTrack UnMarshal error: %v", err)
					return
					//panic(err)
				}

				//复制一份，以防覆盖
				temp := make([]byte, len(rtpPacket.Payload))
				copy(temp, rtpPacket.Payload)
				//decoder.SetOpusData(rtpPacket.Payload)
				decoder.Write(temp)
				player.Play()
			}
		}
	}
	select {}
}