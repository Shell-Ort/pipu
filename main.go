package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
	"gonum.org/v1/gonum/dsp/fourier"
)

var (
	longitudLinea = 100
)

const relleno = "□"

func main() {
	args := os.Args[1:]
	n := len(args)
	if n == 0 {
		_, _ = fmt.Fprintf(
			os.Stderr,
			"Wrong number of arguments (%d), it must be at least 1.\n",
			n,
		)
		os.Exit(1)
	}

	if n > 2 {
		_, _ = fmt.Fprintf(
			os.Stderr,
			"Incorrect number of arguments (%d), max 2. Incorrect arguments: [%s]",
			n,
			strings.Join(args[2:], ", "),
		)
		os.Exit(1)
	}

	fileName := args[0]
	if _, err := os.Stat(fileName); err != nil {
		_, _ = fmt.Fprintf(
			os.Stderr,
			"File %s doesn't exists.\n",
			fileName,
		)
		os.Exit(1)
	}
	if n == 2 {
		lon, err := strconv.Atoi(args[1])
		if err != nil || lon <= 0 {
			_, _ = fmt.Fprintln(
				os.Stderr,
				"Length of the line must be an int of at least 1.",
			)
			os.Exit(1)
		}
		longitudLinea = lon
	}

	stat := NewState()

	fileBytes, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	decoded, err := mp3.NewDecoder(bytes.NewReader(fileBytes))
	if err != nil {
		panic(err)
	}

	sr := decoded.SampleRate()
	const channels = 2
	var pcmInt16 []int16
	buf := make([]byte, 4096)
	for {
		n, err := decoded.Read(buf)
		if n > 0 {
			for i := 0; i < n-1; i += 2 {
				sample := int16(binary.LittleEndian.Uint16(buf[i:]))
				pcmInt16 = append(pcmInt16, sample)
			}
		}
		if err != nil {
			break
		}
	}

	numFrames := len(pcmInt16) / channels
	pcmFloat := make([]float64, numFrames)
	for i := range numFrames {
		pcmFloat[i] = float64(pcmInt16[i*channels]) / 32768.0
	}

	op := new(oto.NewContextOptions{
		SampleRate:   sr,
		ChannelCount: channels,
		Format:       oto.FormatSignedInt16LE,
	})

	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		panic(err)
	}
	<-readyChan

	pcmBytes := make([]byte, len(pcmInt16)*2)
	for i, s := range pcmInt16 {
		binary.LittleEndian.PutUint16(pcmBytes[i*2:], uint16(s))
	}

	player := otoCtx.NewPlayer(bytes.NewReader(pcmBytes))

	const fftSize = 2048
	fft := fourier.NewFFT(fftSize)
	nyquist := float64(sr) / 2.0

	// No existe sincronización real, esperemos que se tenga un pc normal y contamos más o menos sincronización
	go func() {
		m := len(pcmFloat)
		for pos := 0; pos+fftSize <= m; pos += fftSize / 2 {
			windowed := make([]float64, fftSize)
			for i := range fftSize {
				hann := (1 - math.Cos(2*math.Pi*float64(i)/float64(fftSize-1))) / 2
				windowed[i] = pcmFloat[pos+i] * hann
			}

			coeffs := fft.Coefficients(nil, windowed)
			printBands(coeffs, nyquist, fftSize, stat)

			time.Sleep(time.Second / 8)
		}
	}()

	player.Play()
	for player.IsPlaying() {
		time.Sleep(100 * time.Millisecond)
	}
}
