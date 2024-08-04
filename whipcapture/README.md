# WHIP Capture

Decodes WHIP RTP frames, converts them to JPEG and serves them over HTTP.

# Local development

all in this directory (except for frontend stuff):

* install ffmpeg with libvpx support
* `go run gitlab.lrz.de/cm/nms-whep-exercise/server/cmd`
* `npm run dev` in the frontend, then run chrome with websecurity disabled
  (on mac: `open -na Google\ Chrome --args --user-data-dir=/tmp/temporary-chrome-profile-dir --disable-web-security
  --disable-site-isolation-trials` -> probably similar on Linux)
* go to localhost:3000 and start streaming, check output of go process
