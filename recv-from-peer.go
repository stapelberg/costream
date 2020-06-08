// +build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// interruptibleContext returns a context which is canceled when the program is
// interrupted (i.e. receiving SIGINT or SIGTERM).
func interruptibleContext() (context.Context, context.CancelFunc) {
	ctx, canc := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		// Subsequent signals will result in immediate termination, which is
		// useful in case cleanup hangs:
		signal.Stop(sig)
		canc()
	}()
	return ctx, canc
}

func recvFromPeer(ctx context.Context) error {
	var (
		peer           = flag.String("peer", "10.0.0.76", "TODO")
		v4l2sinkdevice = flag.String("v4l2sink_device", "/dev/video11", "device to send to peer")
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

	// TODO: ensure no other gst-launch processes are running based on a magic string in our args (e.g. rtpbin name)
	// - is this still necessary after switching to ctx?

	gst := exec.CommandContext(ctx, "gst-launch-1.0",
		"-v",
		// RTP session setup
		"rtpbin", "name=rtpbin", "latency=0", // default 200ms jitter buffer

		// UDP network setup
		"udpsrc", "caps=application/x-rtp,media=(string)video,clock-rate=(int)90000,encoding-name=(string)H264",
		"port=5000", "!", "rtpbin.recv_rtp_sink_0",

		// Video setup
		"rtpbin.", "!", "rtph264depay", "!", "decodebin", "!", "videoconvert", "!", "v4l2sink", "device="+*v4l2sinkdevice,

		// More UDP network setup
		"udpsrc", "port=5001", "!", "rtpbin.recv_rtcp_sink_0",
		"rtpbin.send_rtcp_src_0", "!", "udpsink", "host="+*peer, "port=5005", "sync=false", "async=false",
		"udpsrc", "caps=application/x-rtp,media=(string)audio,clock-rate=(int)48000,encoding-name=(string)OPUS,encoding-params=(string)1,octet-align=(string)1",
		"port=5002", "!", "rtpbin.recv_rtp_sink_1",

		// audio setup
		"rtpbin.", "!", "rtpopusdepay", "!", "opusdec", "!", "audioconvert", "!", "audioresample", "!", "alsasink", // to default ALSA sink

		// Even more UDP network setup
		"udpsrc", "port=5003", "!", "rtpbin.recv_rtcp_sink_1",
		"rtpbin.send_rtcp_src_1", "!", "udpsink", "host="+*peer, "port=5007", "sync=false", "async=false")

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
	ctx, canc := interruptibleContext()
	defer canc()
	if err := recvFromPeer(ctx); err != nil {
		log.Fatal(err)
	}
}
