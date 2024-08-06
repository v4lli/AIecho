package vp8decoder

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media/ivfwriter"
	"github.com/xlab/libvpx-go/vpx"
	"gocv.io/x/gocv"
)

// based on https://stackoverflow.com/questions/23454940/getting-bytes-buffer-does-not-implement-io-writer-error-message

type VDecoder struct {
	src      <-chan *rtp.Packet
	ctx      *vpx.CodecCtx
	iface    *vpx.CodecIface
	pipeline string
}

type VCodec string

const (
	CodecVP8 VCodec = "V_VP8"
	CodecVP9 VCodec = "V_VP9"
)

func NewVDecoder(codec VCodec, src <-chan *rtp.Packet, pipeline string) *VDecoder {
	dec := &VDecoder{
		src:      src,
		ctx:      vpx.NewCodecCtx(),
		pipeline: pipeline,
	}
	switch codec {
	case CodecVP8:
		dec.iface = vpx.DecoderIfaceVP8()
	case CodecVP9:
		dec.iface = vpx.DecoderIfaceVP9()
	default: // others are currently disabled
		log.Println("[WARN] unsupported VPX codec:", codec)
		return dec
	}
	err := vpx.Error(vpx.CodecDecInitVer(dec.ctx, dec.iface, nil, 0, vpx.DecoderABIVersion))
	if err != nil {
		log.Println("[WARN]", err)
		return dec
	}
	return dec
}

const (
	frameX    = 1440
	frameY    = 720
	frameSize = frameX * frameY * 3
)

func (v *VDecoder) SaveToFramecontainer(fc *FrameContainer, resource string) {
	ffmpeg := exec.Command("ffmpeg", "-i", "pipe:0", "-filter:v", "fps=fps=1", "-r", "1", "-pix_fmt", "bgr24", "-s", strconv.Itoa(frameX)+"x"+strconv.Itoa(frameY), "-f", "rawvideo", "pipe:1") //nolint
	ffmpegIn, _ := ffmpeg.StdinPipe()
	ffmpegOut, _ := ffmpeg.StdoutPipe()
	ffmpegErr, _ := ffmpeg.StderrPipe()

	go func() {
		scanner := bufio.NewScanner(ffmpegErr)
		for scanner.Scan() {
			fmt.Printf("🎥 | %s\n", scanner.Text())
		}
	}()

	ivfWriter, err := ivfwriter.NewWith(ffmpegIn)
	if err != nil {
		panic(err)
	}

	if err := ffmpeg.Start(); err != nil {
		panic(err)
	}

	go func() {
		for {
			buf := make([]byte, frameSize)
			if _, err := io.ReadFull(ffmpegOut, buf); err != nil {
				fc.RemoveResource(resource)
				break
			}
			img, _ := gocv.NewMatFromBytes(frameY, frameX, gocv.MatTypeCV8UC3, buf)
			if img.Empty() {
				println("empty image")
				continue
			}
			jpeg_buf, _ := gocv.IMEncode(".jpg", img)
			fc.AddFrame(resource, v.pipeline, jpeg_buf.GetBytes(), FrameMeta{Timestamp: 0})
		}
		ffmpeg.Wait()
	}()

	for pkt := range v.src {
		if ivfWriterErr := ivfWriter.WriteRTP(pkt); ivfWriterErr != nil {
			println("ivfWriterErr", ivfWriterErr)
			break
		}
	}
	ffmpegIn.Close()
}
