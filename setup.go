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
	// TODO: consider using pcm_substreams=1. will that effectively get rid of
	// the un-intuitive cross-connection?
	must("sudo", "modprobe", "snd-aloop", "enable=1,1", "index=10,11", "id=aloop-topeer,aloop-frompeer")

	must("pacmd", "update-source-proplist alsa_input.platform-snd_aloop.0.analog-stereo device.description=\"aloop-topeer source\"")
	must("pacmd", "update-sink-proplist alsa_output.platform-snd_aloop.0.analog-stereo device.description=\"aloop-topeer sink\"")

	must("pacmd", "update-source-proplist alsa_input.platform-snd_aloop.1.analog-stereo device.description=\"aloop-frompeer source\"")
	must("pacmd", "update-sink-proplist alsa_output.platform-snd_aloop.1.analog-stereo device.description=\"aloop-frompeer sink\"")

	return nil
}

func main() {
	if err := setup(); err != nil {
		log.Fatal(err)
	}
}
