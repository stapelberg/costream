// +build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func sendToPeer(ctx context.Context) error {
	var (
		peer = flag.String(
			"peer",
			"10.0.0.66",
			"TODO")

		listenAddr = flag.String(
			"listen",
			"midna.zekjur.net",
			"TODO")

		v4l2srcdevice = flag.String(
			"v4l2src_device",
			"/dev/video10",
			"V4L2 video device to send to peer")

		pulsesrcdevice = flag.String(
			"pulsesrc_device",
			// cannot open ALSA subdevice ,1 via pulse, so we just use pulseaudio’s monitor
			// mic: alsa_input.usb-RODE_MICROPHONESj_Rode_Podcaster-00.mono-fallback
			"alsa_output.platform-snd_aloop.0.analog-stereo.monitor",
			"PulseAudio source (see pactl list sources) to send to peer")
	)
	flag.Parse()

	// pipeline: source → filter → sink
	// - all of these are called gstreamer elements
	// - elements communicate with each other through pads
	// - source elements have a src pad
	// - filter elements have a src and a sink
	// - sink elements have a sink
	// - demuxer element has one sink pad (through which data arrives) and multiple source pads, one for each stream found in the container

	// TODO: is there a better way to describe gstreamer pipelines programmatically?
	// - could make the gstreamer pipeline description a bit nicer by using go types
	// - check pkg.go.dev if there already is an API for gstreamer in Go

	gst := exec.Command("gst-launch-1.0",
		"-v",
		// RTP session setup
		"rtpbin", "name=rtpbin", "latency=0", // default 200ms jit
		// TODO: max-rtcp-rtp-time-diff=50?

		// Video setup
		"v4l2src", "device="+*v4l2srcdevice,
		"!", "video/x-raw,width=1920,height=1080,framerate=30/1",
		"!", "videoscale",
		"!", "videoconvert",
		// TODO: queue here? or elsewhere in the video stage?
		"!", "x264enc",
		"tune=zerolatency",
		"bitrate=1000",
		"speed-preset=ultrafast",
		// TODO: intra-refresh=true?
		// TODO: quantizer=30?
		// TODO: pass=5?
		"!", "rtph264pay",
		"!", "rtpbin.send_rtp_sink_0",

		// Send video RTP to <peer>:5000
		"rtpbin.send_rtp_src_0",
		"!", "udpsink", "host="+*peer, "port=5000",
		// https://gstreamer.freedesktop.org/data/doc/gstreamer/head/gst-plugins-good/html/gst-plugins-good-plugins-rtpbin.html#id-1.2.136.9.12
		// Since RTCP packets from the sender should be sent as soon as possible
		// and do not participate in preroll, sync=false and async=false is
		// configured on udpsink
		"rtpbin.send_rtcp_src_0",
		"!", "udpsink", "host="+*peer, "port=5001", "sync=false", "async=false",
		// Receive RTCP packets for video stream on :5005
		"udpsrc", "address="+*listenAddr, "port=5005",
		"!", "rtpbin.recv_rtcp_sink_0",

		// Audio setup
		// TODO: for lower latency, could capture the microphone directly,
		// but have to set up microphone filters on the remote OBS
		"pulsesrc", "device="+*pulsesrcdevice,
		"!", "audio/x-raw,rate=44100",
		"!", "audioresample",
		"!", "opusenc", "audio-type=voice",
		"!", "rtpopuspay",
		"!", "rtpbin.send_rtp_sink_1",

		// Send audio RTP to <peer>:5002
		"rtpbin.send_rtp_src_1", "!", "udpsink", "host="+*peer, "port=5002",
		"rtpbin.send_rtcp_src_1", "!", "udpsink", "host="+*peer, "port=5003", "sync=false", "async=false",
		"udpsrc", "address="+*listenAddr, "port=5007", "!", "rtpbin.recv_rtcp_sink_1")
	gst.Stdout = os.Stdout
	gst.Stderr = os.Stderr
	log.Println(gst.Args)
	// TODO(later): read rtpsession stats line by line and print an interactive progress like mpv
	if err := gst.Run(); err != nil {
		return fmt.Errorf("%v: %v", gst.Args, err)
	}
	return nil
}

func main() {
	if err := sendToPeer(context.Background()); err != nil {
		log.Fatal(err)
	}
}
