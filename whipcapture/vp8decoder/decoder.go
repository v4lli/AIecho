package vp8decoder

import (
	"bufio"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media/ivfwriter"
	"gocv.io/x/gocv"
	"io"
	"os/exec"
	"strconv"
)

// based on https://stackoverflow.com/questions/23454940/getting-bytes-buffer-does-not-implement-io-writer-error-message

import (
	"fmt"
	"log"

	"github.com/xlab/libvpx-go/vpx"
)

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

// func startFFmpeg(width, height int) (io.ReadCloser, webm.BlockWriteCloser) {
// 	ffmpeg := exec.Command("ffmpeg", "-i", "pipe:0", "-crf", "22", "-force_key_frames", "expr:gte(t,n_forced*0.5)", "-g", "1", "-loglevel", "debug", "pipe:1")
// 	ffmpegIn, _ := ffmpeg.StdinPipe()
// 	ffmpegOut, _ := ffmpeg.StdoutPipe()
// 	ffmpegErr, _ := ffmpeg.StderrPipe()
// 	if err := ffmpeg.Start(); err != nil {
// 		panic(err)
// 	}
//
//
// 	ws, err := webm.NewSimpleBlockWriter(ffmpegIn,
// 		[]webm.TrackEntry{
// 			{
// 				Name:            "Video",
// 				TrackNumber:     1,
// 				TrackUID:        67890,
// 				CodecID:         "V_VP8",
// 				TrackType:       1,
// 				DefaultDuration: 33333333,
// 				Video: &webm.Video{
// 					PixelWidth:  uint64(width),
// 					PixelHeight: uint64(height),
// 				},
// 			},
// 		})
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("WebM saver has started with video width=%d, height=%d\n", width, height)
// 	videoWriter = ws[0]
// 	return ffmpegOut, ws[0]
// }
//
// var (
// 	videoWriter    webm.BlockWriteCloser
// 	videoBuilder   *samplebuilder.SampleBuilder
// 	videoTimestamp time.Duration
// )

// func pushVP8(rtpPacket *rtp.Packet) {
// 	videoBuilder.Push(rtpPacket)
//
// 	for {
// 		sample := videoBuilder.Pop()
// 		if sample == nil {
// 			return
// 		}
// 		// Read VP8 header.
// 		videoKeyframe := (sample.Data[0]&0x1 == 0)
// 		if videoKeyframe {
// 			fmt.Println("RECEIVED KEY FRAME")
// 			// Keyframe has frame information.
// 			raw := uint(sample.Data[6]) | uint(sample.Data[7])<<8 | uint(sample.Data[8])<<16 | uint(sample.Data[9])<<24
// 			width := int(raw & 0x3FFF)
// 			height := int((raw >> 16) & 0x3FFF)
//
// 			if videoWriter == nil {
// 				// Initialize WebM saver using received frame size.
// 				startFFmpeg(width, height)
// 			}
// 		}
// 		if videoWriter != nil {
// 			videoTimestamp += sample.Duration
// 			fmt.Println("Writing frame timestamp: ", videoTimestamp)
// 			if _, err := videoWriter.Write(videoKeyframe, int64(videoTimestamp/time.Millisecond), sample.Data); err != nil {
// 				panic(err)
// 			}
// 		}
// 	}
// }

func (v *VDecoder) SaveToFramecontainer(fc *FrameContainer, resource string) {
	// videoBuilder = samplebuilder.New(100, &codecs.VP8Packet{}, 90000)
	// sampleBuilder := samplebuilder.New(20000, &codecs.VP8Packet{}, 90000)

	ffmpeg := exec.Command("ffmpeg", "-i", "pipe:0", "-filter:v", "fps=fps=1", "-r", "1", "-pix_fmt", "bgr24", "-s", strconv.Itoa(frameX)+"x"+strconv.Itoa(frameY), "-loglevel", "debug", "-f", "rawvideo", "pipe:1") //nolint
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
	println("subprocess start returned")

	go func() {
		for {
			buf := make([]byte, frameSize)
			println("reading frame")
			if _, err := io.ReadFull(ffmpegOut, buf); err != nil {
				fmt.Printf("ReadFull %s\n", err)
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
	}()

	for pkt := range v.src {
		if ivfWriterErr := ivfWriter.WriteRTP(pkt); ivfWriterErr != nil {
			continue
		}
		print(".")

		// sampleBuilder.Push(pkt)
		// sample := sampleBuilder.Pop()
		// if sample == nil {
		// 	continue
		// }
		// dataSize := uint32(len(sample.Data))
		// // println("isKeyFrame")
		// // println(sample.Data[0] & 0x1)
		// // // write out raw sample:
		// // ioutil.WriteFile("sample.raw", sample.Data, 0644)

		// err := vpx.Error(vpx.CodecDecode(v.ctx, string(sample.Data), dataSize, nil, 0))
		// if err != nil {
		// 	log.Println("[WARN]", err)
		// 	continue
		// }

		// var iter vpx.CodecIter
		// img := vpx.CodecGetFrame(v.ctx, &iter)
		// if img != nil {
		// 	img.Deref()

		// 	buffer := new(bytes.Buffer)
		// 	if err = jpeg.Encode(buffer, img.ImageYCbCr(), nil); err != nil {
		// 		fmt.Printf("jpeg Encode Error: %s\r\n", err)
		// 		continue
		// 	}

		// 	fc.AddFrame(resource, buffer.Bytes(), FrameMeta{Timestamp: pkt.Timestamp})
		// }
		// if IsClosed(v.src) {
		// 	println("channel closed")
		// 	break
		// }
	}

}
