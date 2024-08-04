package server

const (
	WHIP = "whip"
)

type SessionDescription struct {
	ResourceName string
	SDP          string
	Method       string
	PipelineName string
}
