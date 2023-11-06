//go:generate fyne bundle -o data.go Icon.png

package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"jsdey.com/fractal"
)

type appInfo struct {
	name string
	canv bool
	run  func(fyne.Window) fyne.CanvasObject
}

var apps = []appInfo{
	{"Fractal", true, fractal.Show},
}

func main() {
	a := app.New()

	content := container.NewMax()
	w := a.NewWindow("Mandelbrot")

	content.Objects = []fyne.CanvasObject{apps[0].run(w)}

	w.SetContent(content)
	w.Resize(fyne.NewSize(480, 270))
	w.ShowAndRun()
}
