package http

import (
	"gitlab.lrz.de/cm/nms-whep-exercise/server/vp8decoder"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"gitlab.lrz.de/cm/nms-whep-exercise/server"
)

// @title           AIEcho Ingestion Backend
// @version         1.0
// @description     This backend exposes an API for ingesting video streams and providing them to internal processing APIs.

// @license.name  MIT License
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /internal/

type SessionHandler interface {
	Handle(server.SessionDescription) (server.SessionDescription, error)
}

type Server struct {
	router     *httprouter.Router
	httpServer *http.Server
	handler    SessionHandler
	cert, key  string
	fc         *vp8decoder.FrameContainer
}

func New(addr, cert, key string, sh SessionHandler, fc *vp8decoder.FrameContainer) *Server {
	r := httprouter.New()
	return &Server{
		router: r,
		httpServer: &http.Server{
			Addr:                         addr,
			DisableGeneralOptionsHandler: true,
			TLSConfig:                    nil,
			ReadTimeout:                  20 * time.Second,
			ReadHeaderTimeout:            20 * time.Second,
			WriteTimeout:                 20 * time.Second,
			IdleTimeout:                  60 * time.Second,
			MaxHeaderBytes:               0,
			TLSNextProto:                 nil,
			ConnState:                    nil,
			ErrorLog:                     &log.Logger{},
			BaseContext:                  nil,
			ConnContext:                  nil,
		},
		handler: sh,
		cert:    cert,
		key:     key,
		fc:      fc,
	}
}

func handleCors(mux http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodOptions,
			http.MethodPost,
			http.MethodPatch,
		},
		AllowedHeaders:      []string{"*"},
		ExposedHeaders:      []string{"Content-Type", "Location"},
		AllowCredentials:    false,
		MaxAge:              0,
		OptionsPassthrough:  false,
		Debug:               false,
		AllowPrivateNetwork: true,
	}).Handler(mux)
}

func logRequest(mux http.Handler) http.Handler {
	requestLogger := slog.Default().WithGroup("HTTP_REQUEST_LOGGER")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger.Info("New Request", "request", r)
		mux.ServeHTTP(w, r)
	})
}

func (s *Server) setupRoutes() {
	wish := WishHandler{
		handler: s.handler,
		logger:  slog.Default().WithGroup("HTTP_SERVER"),
	}
	s.router.Handler("POST", "/wish/:wish/:id/:pipeline", http.HandlerFunc(wish.post))
	frame := FrameHandler{
		handler: s.handler,
		logger:  slog.Default().WithGroup("HTTP_SERVER"),
		fc:      s.fc,
	}
	s.router.Handler("GET", "/internal/frame/:resource/:index", http.HandlerFunc(frame.get))
	s.router.Handler("GET", "/internal/peers/", http.HandlerFunc(frame.get_peers))
	s.router.Handler("POST", "/internal/transcripts/:resource", http.HandlerFunc(frame.post_transcripts))
}

func (s *Server) ListenAndServe() error {
	s.setupRoutes()
	s.httpServer.Handler = logRequest(handleCors(s.router))
	if len(s.cert) > 0 && len(s.key) > 0 {
		return s.httpServer.ListenAndServeTLS(s.cert, s.key)
	}
	return s.httpServer.ListenAndServe()
}
