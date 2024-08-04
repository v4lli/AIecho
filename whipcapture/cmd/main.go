package main

import (
	"flag"
	"gitlab.lrz.de/cm/nms-whep-exercise/server/vp8decoder"
	"log"
	"log/slog"
	"os"

	"gitlab.lrz.de/cm/nms-whep-exercise/server"
	"gitlab.lrz.de/cm/nms-whep-exercise/server/http"
	"gitlab.lrz.de/cm/nms-whep-exercise/server/webrtc"
)

func main() {
	addr := flag.String("addr", "0.0.0.0:9091", "address to listen on")
	natip := flag.String("natip", "", "1 to 1 NAT IP")
	certFile := flag.String("cert", "", "TLS certificate file")
	keyFile := flag.String("key", "", "TLS key file")
	flag.Parse()

	ll := new(slog.LevelVar)
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: ll})
	slog.SetDefault(slog.New(h))
	ll.Set(slog.LevelDebug)

	log.Printf("XXX Listening on %s\n", *addr)

	if err := serve(*addr, *certFile, *keyFile, *natip); err != nil {
		log.Fatal(err)
	}
}

func serve(addr, cert, key, natip string) error {
	natMappings := []string{}
	if len(natip) > 0 {
		natMappings = append(natMappings, natip)
	}
	pf, err := webrtc.NewPeerConnectionFactory(natMappings)
	if err != nil {
		return err
	}
	fc := vp8decoder.NewFrameContainer()
	sfu := server.New(pf, fc)
	httpServer := http.New(addr, cert, key, sfu, fc)
	return httpServer.ListenAndServe()
}
