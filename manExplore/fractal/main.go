package fractal

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

type Fractal struct {
	currIterations          uint
	currScale, currX, currY float64

	startIterations            uint
	startScale, startX, startY float64

	window fyne.Window
	canvas fyne.CanvasObject
}

func (f *Fractal) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	f.canvas.Resize(size)
}

func (f *Fractal) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(320, 240)
}

//lint:ignore U1000  See TODO inside the .Show() method.
func (f *Fractal) refresh() {
	if f.currScale >= 1.0 {
		f.currIterations = 100
	} else {
		f.currIterations = uint(100 * (1 + math.Pow((math.Log10(1/f.currScale)), 1.25)))
	}

	f.window.Canvas().Refresh(f.canvas)
}

func (f *Fractal) scaleChannel(c float64, start, end uint32) uint8 {
	if end >= start {
		return (uint8)(c*float64(uint8(end-start))) + uint8(start)
	}

	return (uint8)((1-c)*float64(uint8(start-end))) + uint8(end)
}

func (f *Fractal) scaleColor(c float64, start, end color.Color) color.Color {
	r1, g1, b1, _ := start.RGBA()
	r2, g2, b2, _ := end.RGBA()
	return color.RGBA{f.scaleChannel(c, r1, r2), f.scaleChannel(c, g1, g2), f.scaleChannel(c, b1, b2), 0xff}
}

func (f *Fractal) mandelbrot(px, py, w, h int) color.Color {
	drawScale := 3.5 * f.currScale
	aspect := (float64(h) / float64(w))
	cRe := ((float64(px)/float64(w))-0.5)*drawScale + f.currX
	cIm := ((float64(py)/float64(w))-(0.5*aspect))*drawScale - f.currY

	var i uint
	var x, y, xsq, ysq float64

	for i = 0; i < f.currIterations && (xsq+ysq <= 4); i++ {
		xNew := float64(xsq-ysq) + cRe
		y = 2*x*y + cIm
		x = xNew

		xsq = x * x
		ysq = y * y
	}

	if i == f.currIterations {
		return theme.BackgroundColor()
	}

	mu := (float64(i) / float64(f.currIterations))
	c := math.Sin((mu / 2) * math.Pi)

	return f.scaleColor(c, theme.PrimaryColor(), theme.ForegroundColor())
}

//lint:ignore U1000 See TODO inside the .Show() method.
func (f *Fractal) fractalRune(r rune) {
	if r == '+' {
		f.currScale /= 1.1
	} else if r == '-' {
		f.currScale *= 1.1
	} else if r == 's' {
		f.reset()
	} else if r == 'p' {
		CreateJPG(f)
	} else {
		return
	}

	f.refresh()
}

//lint:ignore U1000 See TODO inside the .Show() method.
func (f *Fractal) fractalKey(ev *fyne.KeyEvent) {
	delta := f.currScale * 0.2
	if ev.Name == fyne.KeyUp {
		f.currY -= delta
	} else if ev.Name == fyne.KeyDown {
		f.currY += delta
	} else if ev.Name == fyne.KeyLeft {
		f.currX += delta
	} else if ev.Name == fyne.KeyRight {
		f.currX -= delta
	} else {
		return
	}

	f.refresh()
}

func (f *Fractal) reset() {
	f.currIterations = f.startIterations
	f.currScale = f.startScale
	f.currX = f.startX
	f.currY = f.startY

	f.refresh()
}

// Show loads a Mandelbrot fractal example window for the specified app context
func Show(win fyne.Window) fyne.CanvasObject {
	fractal := &Fractal{window: win}
	fractal.canvas = canvas.NewRasterWithPixels(fractal.mandelbrot)

	fractal.currIterations = 100
	fractal.currScale = 1.0
	fractal.currX = -0.75
	fractal.currY = 0.0
	// TODO: Register, and unregister, these keys:
	win.Canvas().SetOnTypedRune(fractal.fractalRune)
	win.Canvas().SetOnTypedKey(fractal.fractalKey)

	a := theme.PrimaryColor
	b := theme.ForegroundColor
	c := theme.BackgroundColor
	fmt.Printf("%v %v %v\n", a(), b(), c())
	return container.New(fractal, fractal.canvas)
}
