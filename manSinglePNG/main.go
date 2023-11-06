package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/cmplx"
	"os"
)

type Mandelbrot struct {
	W, H                  int        // with and height in pixals
	S, E                  complex128 // start and end of plot window
	I                     int        // Max interations
	Scale, XShift, YShift float64
	FileName              string
}

// Generate a Mandelbrot set and write a png to disk
func (m *Mandelbrot) createMandelbrotImage() (*image.RGBA, error) {

	img := image.NewRGBA(image.Rect(0, 0, m.W, m.H))

	var cmPt complex128
	var col color.RGBA
	var err error

	for px := 0; px < m.W; px++ {
		for py := 0; py < m.H; py++ {
			cmPt = m.trans(px, py)
			col, err = m.mandelbrot(cmPt)
			if err != nil {
				fmt.Println(err)
				return img, err
			}
			img.Set(px, py, col)
		}
	}
	return img, nil
}

func (m *Mandelbrot) mandelbrot(c complex128) (color.RGBA, error) {

	z := complex(0, 0)

	var n int
	for n = 0; n < m.I; n++ {
		if cmplx.Abs(z) > 2 {
			break
		}
		z = z*z + c
	}

	if n == m.I {
		return color.RGBA{255, 99, 0, 255}, nil
	}

	adj := float64(n) + 1. - math.Log(math.Log2(cmplx.Abs(z)))
	mu := (float64(adj) / float64(m.I))
	hue := 360 * mu
	sat := float64(1)
	val := float64(0)
	if n < m.I {
		val = float64(1)
	}

	return hsvToRGB(hue, sat, val)
}

// Transform pixal space to mandelbrot space
// expressed as a complex number.
func (m *Mandelbrot) trans(x, y int) complex128 {
	drawScale := 3.5 * m.Scale
	aspect := float64(m.H) / float64(m.W)
	cRe := ((float64(x)/float64(m.W))-0.5)*drawScale + m.XShift
	cIm := ((float64(y)/float64(m.W))-(0.5*aspect))*drawScale - m.YShift
	return complex(cRe, cIm)
}

// Saves an image to a file
func (m *Mandelbrot) SavePNG(img *image.RGBA) error {
	// Create a new file
	file, err := os.Create(m.FileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode the image as a PNG and write it to the file
	err = png.Encode(file, img)
	if err != nil {
		return err
	}
	return nil
}

//type Point struct {
//	x, y int
//}

func main() {

	m := Mandelbrot{
		W:        3840,
		H:        2160,
		S:        complex(-2, -1),
		E:        complex(1, 1),
		I:        1000,
		FileName: "./newOut.png",
		Scale:    1,
		XShift:   -0.7,
		YShift:   0,
		// Scale: 0.0064001136585278761
		// XShift: -1.2411110166880112704
		// YShift: 0.0868955541831085976
		// Scale:  .0048085001191043395,
		// XShift: -0.06772061171736621,
		// YShift: 0.6670099929922334,
	}

	img, err := m.createMandelbrotImage()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = m.SavePNG(img)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// hslToRGB converts an HSV triple to an RGB triple.
func hsvToRGB(h, s, v float64) (color.RGBA, error) {
	if h < 0 || h > 360 ||
		s < 0 || s > 1 ||
		v < 0 || v > 1 {
		fmt.Println(h, s, v)
		return color.RGBA{}, errors.New("hsvToRGB: inputs out of range")
	}
	// When 0 ≤ h < 360, 0 ≤ s ≤ 1 and 0 ≤ v ≤ 1:
	C := v * s
	X := C * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - C
	var Rnot, Gnot, Bnot float64
	switch {
	case 0 <= h && h < 60:
		Rnot, Gnot, Bnot = C, X, 0
	case 60 <= h && h < 120:
		Rnot, Gnot, Bnot = X, C, 0
	case 120 <= h && h < 180:
		Rnot, Gnot, Bnot = 0, C, X
	case 180 <= h && h < 240:
		Rnot, Gnot, Bnot = 0, X, C
	case 240 <= h && h < 300:
		Rnot, Gnot, Bnot = X, 0, C
	case 300 <= h && h < 360:
		Rnot, Gnot, Bnot = C, 0, X
	}
	r := uint8(math.Round((Rnot + m) * 255))
	g := uint8(math.Round((Gnot + m) * 255))
	b := uint8(math.Round((Bnot + m) * 255))

	return color.RGBA{r, g, b, 255}, nil
}
