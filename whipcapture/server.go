package server

import (
	"errors"
	"gitlab.lrz.de/cm/nms-whep-exercise/server/vp8decoder"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync/v3"
)

type Sample struct {
	Payload   []byte
	Timestamp time.Time
	Duration  time.Duration
}

type SampleWriter interface {
	WriteSample(Sample) error
}

type Publisher interface {
	Publish(publication, track, kind string) (SampleWriter, error)
	Unpublish(publication string)
}

type Publication interface {
	Subscribe(string, SampleWriter)
	Unsubscribe(string)
	Kind() string
}

type Peer interface {
	Signal(string) (string, error)
	AcceptTracks() error
	AddTrack(Publication) error
}

type PeerFactory interface {
	Create(name, resource, pipeline string, p Publisher, fc *vp8decoder.FrameContainer) (peer Peer, err error)
}

type OnDemandSource interface {
	Play() error
}

type OnDemandFactory interface {
	Open(p Publisher, location, resource string) (OnDemandSource, error)
}

type publicationName string
type trackName string

type Server struct {
	publishers *xsync.MapOf[publicationName, *xsync.MapOf[trackName, Publication]]
	pf         PeerFactory
	of         OnDemandFactory
	rn         *randomNameGenerator
	logger     *slog.Logger
	fc         *vp8decoder.FrameContainer
}

func New(pf PeerFactory, fc *vp8decoder.FrameContainer) *Server {
	return &Server{
		publishers: xsync.NewMapOf[publicationName, *xsync.MapOf[trackName, Publication]](),
		pf:         pf,
		rn:         newRandomNameGenerator(),
		logger:     slog.Default().WithGroup("SERVER"),
		fc:         fc,
	}
}

func (s *Server) Publish(resource, track, kind string) (SampleWriter, error) {
	s.logger.Info("saving track for publisher", "publisher", resource, "track", track)
	pub, ok := s.publishers.Load(publicationName(resource))
	if !ok {
		s.logger.Warn("publisher not found", "publisher", resource)
		return nil, errors.New("publisher not found")
	}
	writer := newTrack(kind)
	_, ok = pub.LoadOrStore(trackName(track), writer)
	if ok {
		s.logger.Warn("track name already in use", "track", track)
		return nil, errors.New("track name already in use")
	}
	return writer, nil
}

func (s *Server) Unpublish(resource string) {
	s.publishers.Delete(publicationName(resource))
}

func (s *Server) Handle(sd SessionDescription) (SessionDescription, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return SessionDescription{}, err
	}
	location := uuid.String()
	p, err := s.pf.Create(location, sd.ResourceName, sd.PipelineName, s, s.fc)
	if err != nil {
		return SessionDescription{}, err
	}
	if sd.Method == WHIP {
		s.logger.Info("storing whip peer", "location", location, "resource", publicationName(sd.ResourceName))
		s.publishers.LoadOrStore(publicationName(sd.ResourceName), xsync.NewMapOf[trackName, Publication]())
		if err = p.AcceptTracks(); err != nil {
			return SessionDescription{}, err
		}
	}
	answer, err := p.Signal(sd.SDP)
	if err != nil {
		return SessionDescription{}, err
	}
	return SessionDescription{
		ResourceName: location,
		SDP:          answer,
		Method:       sd.Method,
	}, nil
}

type randomNameGenerator struct {
	alphabet []rune
	random   *rand.Rand
}

func newRandomNameGenerator() *randomNameGenerator {
	return &randomNameGenerator{
		alphabet: []rune("abcdefghijklmnopqrstuvwxyz"),
		random:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *randomNameGenerator) Generate() string {
	size := len(g.alphabet)
	var sb strings.Builder
	for i := 0; i < 6; i++ {
		ch := g.alphabet[g.random.Intn(size)]
		sb.WriteRune(ch)
	}
	return sb.String()
}
