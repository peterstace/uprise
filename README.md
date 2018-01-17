# Uprise

Uprise generates [Shepard tones](https://en.wikipedia.org/wiki/Shepard_tone).
These are sounds that always increase in pitch locally, but maintain constant
average pitch globally (over a time domain).

An example of its output can be found
[here](https://soundcloud.com/user-889310662/shepard-tone-aug7-chord).

## Features

- Outputs Shepard tones in `wav` format (16bit, 44100Hz, mono).

- Support for many chords: major, minor, augmented, diminished, major 6th,
  minor 6th, dominant 7th, major 7th, minor 7th, augmented 7th.

- Customisable pitch increase speed.

- Customisable volume modulation.

- Adjustable output volume gain.

## To Run

```
go install github.com/peterstace/uprise
$GOPATH/bin/uprise --out out.wav
aplay out.wav
```
