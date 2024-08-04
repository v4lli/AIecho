package http

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"gitlab.lrz.de/cm/nms-whep-exercise/server"
	"gitlab.lrz.de/cm/nms-whep-exercise/server/vp8decoder"
	"io"
	"log/slog"
	"net/http"
	"slices"
)

type WishHandler struct {
	handler SessionHandler
	logger  *slog.Logger
}

type FrameHandler struct {
	handler SessionHandler
	logger  *slog.Logger
	fc      *vp8decoder.FrameContainer
}

func (h *WishHandler) post(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("id")
	if id == "" {
		h.logger.Info("got request with invalid id parameter", "id", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	wishMethodString := params.ByName("wish")
	if wishMethodString != server.WHIP {
		h.logger.Info("got invalid wish request", "wish", wishMethodString)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	pipelineString := params.ByName("pipeline")
	if pipelineString != "normal" && pipelineString != "fast" {
		h.logger.Info("got invalid pipeloine name request", "pipeline", pipelineString)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := uuid.Validate(id); wishMethodString == server.WHIP && err != nil {
		h.logger.Info("got invalid stream ID", "id", id)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid stream ID, please use a valid UUID"))
		return
	}
	if !slices.Contains(r.Header["Content-Type"], "application/sdp") {
		h.logger.Info("got request with invalid content-type header", "Content-Type", r.Header["Content-Type"])
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sd := server.SessionDescription{
		ResourceName: id,
		SDP:          string(buf),
		Method:       wishMethodString,
		PipelineName: pipelineString,
	}
	res, err := h.handler.Handle(sd)
	if err != nil {
		h.logger.Error("failed to setup new peer", "error", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Add("Content-Type", "application/sdp")
	w.Header().Add("Location", fmt.Sprintf("http://%v%v/%v", r.Host, r.URL.String(), res.ResourceName))
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(res.SDP))
	if err != nil {
		h.logger.Info("failed to write response body", "error", err)
	}
}

// GetFrames returns the frames for a given resource
//
//		@Summary      Get frames for a particular resource
//		@Param        resource    path     string  true  "resource ID"
//		@Param        frame_index    path     int  true  "relative frame index"
//	    @Success      200              {string}  file    "JPEG frame"
//		@Router       /internal/frame/{resource}/{frame_index} [get]
func (h *FrameHandler) get(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	resource := params.ByName("resource")
	if resource == "" {
		h.logger.Info("got request with invalid id parameter", "resource", resource)
		w.Write([]byte("no frames found"))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := uuid.Validate(resource); err != nil {
		h.logger.Info("got invalid stream ID", "resource", resource)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid stream ID, please use a valid UUID"))
		return
	}
	if !h.fc.HasResource(resource) {
		h.logger.Info("resource not found", "resource", resource)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("no frames found"))
		return
	}

	frames, _ := h.fc.GetFrames(resource)
	if len(frames) == 0 {
		h.logger.Info("no frames found", "resource", resource)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("no frames found"))
		return
	}
	w.Header().Add("Content-Type", "image/jpeg")
	if len(frames) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)

	frame_index := params.ByName("resource")
	if frame_index == "" {
		h.logger.Info("got request with invalid frame_index parameter")
		w.Write([]byte("no frames found"))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// convert frame_index to int:
	// i, err := strconv.Atoi(frame_index)
	// if err != nil {
	// 	// ... handle error
	// 	panic(err)
	// }
	_, err := w.Write(frames[0])

	if err != nil {
		h.logger.Info("failed to write response body", "error", err)
	}
}

type Peer struct {
	UUID     string `json:"uuid"`
	Pipeline string `json:"pipeline"`
}

type PeerList struct {
	Peers []Peer `json:"peers"`
}

// Get peers returns the list of peers
//
//		@Summary Get all currently connected peers
//	    @Success 200 {object} PeerList
//		@Router  /internal/peers/ [get]
func (h *FrameHandler) get_peers(w http.ResponseWriter, r *http.Request) {
	peers, pipelines := h.fc.Resources()
	w.WriteHeader(http.StatusOK)
	peerList := make([]Peer, len(peers))
	for i, peer := range peers {
		peerList[i] = Peer{
			UUID:     peer,
			Pipeline: pipelines[i],
		}
	}
	pl := &PeerList{
		Peers: peerList,
	}
	b, err := json.Marshal(pl)
	if err != nil {
		h.logger.Info("failed to write response body", "error", err)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		h.logger.Info("failed to write response body", "error", err)
		return
	}
}

type TranscriptContainer struct {
	Transcript string `json:"transcript"`
}

// PostTranscripts accepts a transcribed frame as text and forwards it to the correct client through a datachannel
//
//			@Summary      Forward transcribed frame contents to client
//			@Param        resource    path     string  true  "resource ID"
//			@Param        container    body     TranscriptContainer  true  "transcript container"
//	        @Success      200              {string}  string    "ok"
//			@Router       /internal/transcripts/{resource} [post]
func (h *FrameHandler) post_transcripts(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	resource := params.ByName("resource")
	if resource == "" {
		h.logger.Info("got request with invalid id parameter", "resource", resource)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := uuid.Validate(resource); err != nil {
		h.logger.Info("got invalid stream ID", "resource", resource)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid stream ID, please use a valid UUID"))
		return
	}
	if !slices.Contains(r.Header["Content-Type"], "application/json") {
		h.logger.Info("got request with invalid content-type header", "Content-Type", r.Header["Content-Type"])
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// deserialize the body:
	var tc TranscriptContainer
	if err := json.Unmarshal(body, &tc); err != nil {
		h.logger.Error("failed to unmarshal request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.fc.GetDataChannel(resource) <- []byte(tc.Transcript)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
