package main

import (
	"fmt"
	"math"
	"strings"
)

type freqBand struct {
	Name   string
	Lo, Hi float64
}

var bands = []freqBand{
	{"Sub-bass", 20, 60},
	{"Bass", 60, 250},
	{"Low Mid", 250, 500},
	{"Mid", 500, 2000},
	{"Upper Mid", 2000, 4000},
	{"Presence", 4000, 6000},
	{"Brilliance", 6000, 20000},
}

func printBands(coeffs []complex128, nyquist float64, fftSize int, stat *State) {
	stat.GoTo(0)

	type bandResult struct {
		avg    float64
		initxt string
	}
	results := make([]bandResult, len(bands))
	maxAvg := 0.0

	for idx, b := range bands {
		loBin := int(b.Lo / nyquist * float64(fftSize/2))
		hiBin := int(b.Hi / nyquist * float64(fftSize/2))
		if loBin < 1 {
			loBin = 1
		}
		if hiBin >= len(coeffs) {
			hiBin = len(coeffs) - 1
		}

		sum := 0.0
		count := 0
		for i := loBin; i <= hiBin; i++ {
			sum += math.Hypot(real(coeffs[i]), imag(coeffs[i]))
			count++
		}
		avg := sum / float64(count)
		if avg > maxAvg {
			maxAvg = avg
		}
		initxt := fmt.Sprintf("%-12s | %6.0f-%-7.0f Hz |", b.Name, b.Lo, b.Hi)

		results[idx] = bandResult{avg, initxt}
	}

	for _, r := range results {
		bars := 0
		if maxAvg > 0 {
			normalized := r.avg / maxAvg
			bars = int(math.Sqrt(normalized) * float64(longitudLinea))
		}
		if bars > longitudLinea {
			bars = longitudLinea
		}

		printWithFill(r.initxt, bars, stat)
		stat.Down(1)
	}
}

func printWithFill(ini string, n int, stat *State) {
	if n > longitudLinea {
		panic("Error in printing.")
	}

	var sal strings.Builder
	sal.WriteString(ini)
	sal.WriteByte(' ')

	nFill := n
	sal.WriteString(strings.Repeat(relleno, nFill))

	empty := longitudLinea - nFill
	sal.WriteString(strings.Repeat(" ", empty))

	stat.Write(sal.String())
}
