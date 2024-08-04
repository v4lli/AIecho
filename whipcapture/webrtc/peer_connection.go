package webrtc

import (
	"errors"
	"fmt"
	"github.com/pion/rtcp"
	"gitlab.lrz.de/cm/nms-whep-exercise/server/vp8decoder"
	"log/slog"
	"time"

	"github.com/pion/ice/v2"
	"github.com/pion/interceptor"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"gitlab.lrz.de/cm/nms-whep-exercise/server"
)

type PeerConnectionFactory struct {
	api *webrtc.API
}

func NewPeerConnectionFactory(nat1To1IPs []string) (*PeerConnectionFactory, error) {
	se := webrtc.SettingEngine{}
	if len(nat1To1IPs) > 0 {
		se.SetNAT1To1IPs(nat1To1IPs, webrtc.ICECandidateTypeHost)
		um, err := ice.NewMultiUDPMuxFromPort(5000)
		if err != nil {
			return nil, err
		}
		se.SetICEUDPMux(um)
	}
	me := &webrtc.MediaEngine{}
	if err := me.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}
	ir := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(me, ir); err != nil {
		return nil, err
	}
	api := webrtc.NewAPI(
		webrtc.WithMediaEngine(me),
		webrtc.WithInterceptorRegistry(ir),
		webrtc.WithSettingEngine(se),
	)
	return &PeerConnectionFactory{
		api: api,
	}, nil
}

func (f *PeerConnectionFactory) Create(location, resource, pipeline string, p server.Publisher, fc *vp8decoder.FrameContainer) (server.Peer, error) {
	pc, err := newPeerConnection(f.api, location, resource, pipeline, p, fc)
	if err != nil {
		return nil, err
	}
	return pc, nil
}

type PeerConnection struct {
	resource      string
	location      string
	pc            *webrtc.PeerConnection
	publisher     server.Publisher
	subscriptions []server.Publication
	logger        *slog.Logger
	fc            *vp8decoder.FrameContainer
	frameChannel  chan *rtp.Packet
	pipelineName  string
}

func newPeerConnection(api *webrtc.API, location, resource, pipeline string, publisher server.Publisher, fc *vp8decoder.FrameContainer) (*PeerConnection, error) {
	pc, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, err
	}
	p := &PeerConnection{
		resource:     resource,
		location:     location,
		publisher:    publisher,
		pc:           pc,
		fc:           fc,
		logger:       slog.Default().WithGroup("PEER_CONNECTION"),
		frameChannel: make(chan *rtp.Packet, 100),
		pipelineName: pipeline,
	}
	pc.OnTrack(p.onTrack)
	pc.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		if pcs == webrtc.PeerConnectionStateClosed || pcs == webrtc.PeerConnectionStateDisconnected || pcs == webrtc.PeerConnectionStateFailed {
			publisher.Unpublish(p.resource)
			for _, s := range p.subscriptions {
				s.Unsubscribe(p.location)
			}
			if p.frameChannel != nil && !IsClosed(p.frameChannel) {
				close(p.frameChannel)
			}
			fc.RemoveResource(p.resource)
		}
	})
	pc.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			dc := fc.GetDataChannel(p.resource)
			go func() {
				for {
					v := <-dc
					if string(v) == "" {
						continue
					}
					fmt.Printf("DataChannel '%s': '%s'\n", d.Label(), string(v))
					if sendErr := d.SendText(string(v)); sendErr != nil {
						fmt.Printf("data channel error, closing: %s", sendErr)
						return
					}
				}
			}()
		})
	})
	return p, nil
}

func (p *PeerConnection) Signal(sdpOffer string) (string, error) {
	if err := p.pc.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdpOffer,
	}); err != nil {
		return "", err
	}
	answer, err := p.pc.CreateAnswer(nil)
	if err != nil {
		return "", err
	}
	promise := webrtc.GatheringCompletePromise(p.pc)
	if err := p.pc.SetLocalDescription(answer); err != nil {
		return "", err
	}
	select {
	case <-promise:
	case <-time.After(time.Minute):
		return "", errors.New("candidate gathering timeout")
	}
	return p.pc.LocalDescription().SDP, nil
}

func (p *PeerConnection) AcceptTracks() error {
	if _, err := p.pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
		Direction:     webrtc.RTPTransceiverDirectionRecvonly,
		SendEncodings: []webrtc.RTPEncodingParameters{},
	}); err != nil {
		return err
	}
	if _, err := p.pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
		Direction:     webrtc.RTPTransceiverDirectionRecvonly,
		SendEncodings: []webrtc.RTPEncodingParameters{},
	}); err != nil {
		return err
	}
	return nil
}

func (p *PeerConnection) AddTrack(publication server.Publication) error {
	var trackRemote *webrtc.TrackLocalStaticSample
	var err error

	switch publication.Kind() {
	case "video":
		trackRemote, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video", "pion")
	case "audio":
		trackRemote, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
	default:
		return fmt.Errorf("unknown track kind: %v", publication.Kind())
	}
	if err != nil {
		return err
	}
	_, err = p.pc.AddTrack(trackRemote)
	if err != nil {
		return err
	}
	return nil
}

func IsClosed(ch <-chan *rtp.Packet) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

func (p *PeerConnection) onTrack(trackRemote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	// TODO: Remove periodic PLI and only request one when a new subscriber
	// joins
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			if writeErr := p.pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(trackRemote.SSRC())}}); writeErr != nil {
				fmt.Println(writeErr)
			}
		}
	}()

	vd := vp8decoder.NewVDecoder(vp8decoder.CodecVP8, p.frameChannel, p.pipelineName)
	// XXX I think this stays alive forever and eventually runs out of memory
	go vd.SaveToFramecontainer(p.fc, p.resource)

	for {
		rtpPacket, _, err := trackRemote.ReadRTP()
		if err != nil {
			p.logger.Error("failed to read RTP packet, exiting track write loop", "error", err)
			return
		}
		// if !IsClosed(p.frameChannel) {
		p.frameChannel <- rtpPacket
		// }
	}
}
