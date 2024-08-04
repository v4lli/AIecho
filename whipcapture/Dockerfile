FROM golang:1.22.1-alpine as builder

WORKDIR /build

COPY go.mod go.mod
COPY go.sum go.sum

RUN apk add ffmpeg g++ libvpx-dev libvpx libogg libvorbis opus portaudio gcc libc++-dev libc-dev opencv-dev && go mod download

COPY . ./

RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/main.go

# FROM alpine
# 
# COPY --from=builder /build/server /server

ENTRYPOINT ["./server", "-natip", "162.55.186.91"]
