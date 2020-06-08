Conceptually, we want to transport the OBS output (video and audio) to another
computer with low latency.

Having OBS stream into a custom RTMP server turned out to have too much latency.

TODO: insert mega-diagram.jpg

TODO: is netclock something we want? https://github.com/NHGmaniac/gst-videowall/blob/master/lib/netclock.py

dumping a dot graph out of gstreamer: https://developer.ridgerun.com/wiki/index.php/How_to_generate_a_Gstreamer_pipeline_diagram_(graph)

TODO: audio has some distortion. could be related to the clock rate difference?
- issue disappears when using the default input (RODE podcaster) instead of the loop device

## OBS video setup

TODO: describe how to set up the v4l2sink in OBS

## OBS audio setup

First, point OBS’s monitoring feature to the `snd-loop` ALSA device:

* File
* → Settings
* → Audio
* → Advanced
* → Monitoring device: select the loopback device (name cannot be changed?)

Then, make OBS route your microphone into the monitoring device:

* In the mixer at the bottom of the OBS window, right-click your mic
* → Advanced Audio Properties
* → Audio Monitoring
* → Monitor and Output

## Dependencies

### Debian

| package | version number |
|---------|----------------|
| [v4l2loopback-dkms](https://packages.debian.org/bullseye/v4l2loopback-dkms) | 0.12.5-1 |
| [obs-v4l2sink.deb](https://github.com/CatxFish/obs-v4l2sink/releases/download/0.1.0/obs-v4l2sink.deb) | 0.1.0 |
| gstreamer1.0-alsa | TODO

### Arch Linux

| package | version number |
|---------|----------------|
| [community/v4l2loopback-dkms](https://www.archlinux.org/packages/community/any/v4l2loopback-dkms/) | 0.12.5-1 |
| [AUR:obs-v4l2sink](https://aur.archlinux.org/packages/obs-v4l2sink/) | 0.1.0 |
| [extra/gst-plugins-ugly](TODO) | 1.16.2-3

## Setup

```shell
go run setup.go
```

TODO: install setup.go such that it will be run as root after boot

## Running

UDP port 5000 to 5007 need to be open.

stapelberg runs:
```
go run send-to-peer.go   -peer=rtp6.servnerr.com -listen=rtp6.zekjur.net
go run recv-from-peer.go -peer=rtp6.servnerr.com -listen=rtp6.zekjur.net
```

mdlayher runs:
```
go run send-to-peer.go   -peer=rtp6.zekjur.net -listen=rtp6.servnerr.com
go run recv-from-peer.go -peer=rtp6.zekjur.net -listen=rtp6.servnerr.com
```

TODO: currently, the receiver needs to be restarted when the sender is restarted

TODO: check if the converse is also true

TODO: make it so that no restarts are necessary either way
