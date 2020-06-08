// +build ignore

package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
)

func must(name string, args ...string) {
	c := exec.Command(name, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	log.Println(c.Args)
	if err := c.Run(); err != nil {
		log.Fatalf("%v: %v", c.Args, err)
	}
}

func setup() error {
	var (
		restart = flag.Bool("restart", false, "whether to try to remove the kernel modules first")
	)
	flag.Parse()

	// TODO(safety): check if snd-aloop or v4l2loopback are already loaded and
	// ask for confirmation by way of passing -restart (or manually unloading
	// the modules)

	if *restart {
		must("pulseaudio", "-k")
		must("sudo", "rmmod", "snd-aloop", "v4l2loopback")
	}

	must("sudo", "modprobe", "v4l2loopback", "video_nr=10,11", "card_label=v4l2-topeer,v4l2-frompeer", "exclusive_caps=1")

	// TODO: to get rid of the snd-aloop kernel module altogether,
	// I tried using a PulseAudio null sink, but OBS would not list
	// it as an option for the audio monitor device.

	// TODO: consider using pcm_substreams=1. will that effectively get rid of
	// the un-intuitive cross-connection?
	// - tried it and it behaved weirdly
	must("sudo", "modprobe", "snd-aloop", "enable=1", "index=10", "id=aloop-topeer")

	must("pacmd", "update-source-proplist alsa_input.platform-snd_aloop.0.analog-stereo device.description=\"aloop-topeer source\"")
	must("pacmd", "update-sink-proplist alsa_output.platform-snd_aloop.0.analog-stereo device.description=\"aloop-topeer sink\"")

	return nil
}

func main() {
	if err := setup(); err != nil {
		log.Fatal(err)
	}
}
