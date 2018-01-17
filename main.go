package main

import (
	"flag"
	"log"
	"math"
	"os"
	"time"
)

const sampleHz = 44100

func main() {
	out := flag.String("out", "out.wav", "output location")
	gain := flag.Float64("gain", 0.05, "volume gain")
	octavesHz := flag.Float64("octavesHz", 0.1, "octaves per second rise in pitch")
	durationSec := flag.Float64("durationSec", 60, "duration in seconds of generated wav")
	chordName := flag.String("chordName", "maj", "chord to play")
	volumeCenterHz := flag.Float64("volumeCenterHz", 1000, "center frequency for volume modulation")
	volumeStdDevHz := flag.Float64("volumeStdDevHz", 800, "std dev frequency for volume modulation")

	flag.Parse()
	cfg := config{
		octavesHz:      *octavesHz,
		duration:       time.Duration(*durationSec * float64(time.Second)),
		volumeCenterHz: *volumeCenterHz,
		volumeStdDevHz: *volumeStdDevHz,
		gain:           *gain,
	}
	chord, ok := chordForName(*chordName)
	if !ok {
		log.Fatal("unknown chord: ", *chordName)
	}

	f, err := os.Create(*out)
	fatal(err)
	up := uprise{cfg: cfg, chord: chord}
	samples := up.generateSamples()
	fatal(WriteWav(f, samples))
	fatal(f.Close())
}

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type config struct {
	octavesHz float64
	duration  time.Duration

	volumeCenterHz float64
	volumeStdDevHz float64

	gain float64
}

type uprise struct {
	cfg      config
	angles   []float64
	chord    []float64
	chordIdx int
	octave   int
}

func (u *uprise) generateSamples() []int16 {
	u.angles = make([]float64, 100)

	totalSamples := int(sampleHz * u.cfg.duration.Seconds())
	samples := make([]int16, totalSamples)
	for i := range samples {
		t := time.Duration(float64(i) * float64(time.Second) / sampleHz)
		u.addNewNotes(t)
		samples[i] = u.calculateSample(t)
	}
	return samples
}

func (u *uprise) addNewNotes(t time.Duration) {
	for {
		newOctave := u.octave
		newChordIdx := u.chordIdx
		newChordIdx--
		if newChordIdx < 0 {
			newChordIdx += len(u.chord)
			newOctave--
		}
		hz := u.hz(newOctave, newChordIdx, t)
		if hz > 20 {
			u.angles = append([]float64{0}, u.angles...)
			u.octave = newOctave
			u.chordIdx = newChordIdx
		} else {
			return
		}
	}
}

func (u *uprise) calculateSample(t time.Duration) int16 {
	// Sample each note, and track where we can remove note(s) that are too
	// high pitched.
	discardAt := len(u.angles)
	var s float64
	oct := u.octave
	chIdx := u.chordIdx
	for i := range u.angles {
		chIdx++
		if chIdx >= len(u.chord) {
			chIdx = 0
			oct++
		}
		hz := u.hz(oct, chIdx, t)
		if hz > 20e3 {
			discardAt = i
			break
		}
		s += u.sample(i, hz) * u.bellCurve(hz)
	}
	u.angles = u.angles[:discardAt]

	// Fade in/out.
	if t < time.Second {
		s *= float64(t) / float64(time.Second)
	} else if t > u.cfg.duration-time.Second {
		s *= float64(u.cfg.duration-t) / float64(time.Second)
	}

	// Scale down to reduce clipping.
	s *= u.cfg.gain
	return quantize(s)
}

func (u *uprise) hz(octave, chordIdx int, t time.Duration) float64 {
	exp := float64(t.Seconds()*u.cfg.octavesHz) + float64(octave)
	return u.chord[chordIdx] * math.Pow(2, exp)
}

func (u *uprise) bellCurve(hz float64) float64 {
	exp := (hz - u.cfg.volumeCenterHz) / u.cfg.volumeStdDevHz
	exp *= exp
	return math.Exp(-0.5 * exp)
}

func (u *uprise) sample(i int, hz float64) float64 {
	angle := u.angles[i]
	angle += hz / sampleHz * 2 * math.Pi
	angle = math.Mod(angle, 2*math.Pi)
	u.angles[i] = angle
	return math.Sin(angle)
}

func quantize(f float64) int16 {
	f = math.Max(-1.0, math.Min(1.0, f))
	return int16(f * math.MaxInt16)
}

func chordForName(name string) ([]float64, bool) {
	const (
		root  = 1
		min3  = 5.0 / 4.0
		maj3  = 4.0 / 3.0
		dim   = 36.0 / 25.0
		perf5 = 3.0 / 2.0
		aug5  = 8.0 / 5.0
		maj6  = 5.0 / 3.0
		min7  = 9.0 / 5.0
		maj7  = 15.0 / 8.0
	)
	ch, ok := map[string][]float64{
		"maj":  []float64{root, maj3, perf5},
		"min":  []float64{root, min3, perf5},
		"aug":  []float64{root, maj3, aug5},
		"dim":  []float64{root, min3, dim},
		"maj6": []float64{root, maj3, perf5, maj6},
		"min6": []float64{root, min3, perf5, maj6},
		"dom7": []float64{root, maj3, perf5, min7},
		"maj7": []float64{root, maj3, perf5, maj7},
		"min7": []float64{root, min3, perf5, min7},
		"aug7": []float64{root, maj3, aug5, min7},
	}[name]
	return ch, ok
}
