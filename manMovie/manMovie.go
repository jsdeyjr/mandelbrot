package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"math/cmplx"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
)

type Mandelbrot struct {
	W, H        int        // with and height in pixals
	S, E        complex128 // start and end of plot window
	I           int        // Max interations
	Scale, X, Y float64
	Author      string
	FileName    string
}

type Movie struct {
	Mandelbrot
	FPS         int
	Frames      int
	ScaleFactor float64
	MovLength   float64 // In minutes
	InDir       string
	OutDir      string
}

func main() {

	m := &Movie{}
	m.FPS = 30
	m.MovLength = 0.333
	m.Frames = int(math.Round(float64(m.MovLength)*60)) * m.FPS
	m.InDir = "../manExplore/pic"
	m.OutDir = "./mov"
	m.W = 960 // 3840
	m.H = 560 // 2160
	m.S = complex(-2, -1)
	m.E = complex(1, 1)
	m.I = 300
	// m.Scale = 1
	// m.X = -0.7
	// m.Y = 0
	m.Scale = 0.0064001136585278761
	m.X = -1.2411110166880112704
	m.Y = 0.0868955541831085976

	filePath, err := getFileName(m.InDir)

	err = setDirectory(filePath)
	if err != nil {
		log.Fatal(err)
	}

	err = m.calcFrames(filePath)
	if err != nil {
		log.Fatal(err)
	}
}

func (m *Movie) calcFrames(filePath string) error {

	mOrig, err := getMetadata(filePath)
	if err != nil {
		return err
	}

	s := mOrig.Scale
	e := float64(1) / float64(m.Frames)
	m.ScaleFactor = math.Pow(s, e)

	m.Scale = 1
	m.X = mOrig.X
	m.Y = mOrig.Y

	// Set the file for the mp4 movie
	rootName, err := getFileElements(filePath)
	if err != nil {
		return err
	}
	outFile := m.OutDir + "/" + rootName + ".mp4"
	fmt.Println("calcFrames:", outFile)

	// Command to run FFmpeg
	cmd := exec.Command("ffmpeg", "-f", "image2pipe", "-i",
		"pipe:0", "-r", "30", "-pix_fmt", "yuv420p", "-vcodec",
		"libx264", outFile)

	// Get FFmpeg's standard input pipe
	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	defer stdin.Close()

	// Start the FFmpeg command
	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	for i := 0; i < m.Frames; i++ {
		img, err := m.createMandelbrotImage()
		if err != nil {
			return err
		}

		// Stream img to output
		err = png.Encode(stdin, img)
		if err != nil {
			panic(err)
		}

		m.Scale *= m.ScaleFactor

		if i%50 == 0 {
			fmt.Println()
			fmt.Printf("%04d   ", i)
		} else {
			fmt.Print(".")
		}
	}
	fmt.Println()
	return nil
}

func getMetadata(filePath string) (*Mandelbrot, error) {
	intfc, err := jis.NewJpegMediaParser().ParseFile(filePath)
	if err != nil {
		return nil, err
	}
	sl := intfc.(*jis.SegmentList)

	_, _, exifTags, err := sl.DumpExif()
	if err != nil {
		return nil, err
	}

	for _, et := range exifTags {
		if et.TagName == "DocumentName" {
			m := &Mandelbrot{}
			s := et.FormattedFirst
			json.Unmarshal([]byte(s), m)
			fmt.Println("getMetadata\n", m)
			return m, nil
		}
	}
	return nil, nil
}

func fileOpen(dir string, frame int) (*os.File, error) {

	s := fmt.Sprintf("%04d", frame+1)
	fileAbs := dir + "/" + s + ".jpg"
	// fmt.Println(fileAbs)

	outputFile, err := os.Create(fileAbs)
	if err != nil {
		return nil, err
	}
	return outputFile, nil
}

// Generate a Mandelbrot image
func (m *Movie) createMandelbrotImage() (*image.RGBA, error) {

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

func getFileElements(filePath string) (string, error) {
	_, file := filepath.Split(filePath)
	ex := filepath.Ext(file)
	fmt.Println("getFileElements:", filePath, file, ex)
	if ex != ".JPG" && ex != ".jpg" {
		return "", errors.New("getFileElements: File must be a jpeg type.")
	}
	rootName := file[:len(file)-len(filepath.Ext(file))]
	fmt.Println("getFileElements:", rootName)

	return rootName, nil
}

func setDirectory(filePath string) error {

	rootName, err := getFileElements(filePath)
	if err != nil {
		return err
	}
	// Create a directory with the root name if it doesn't exist
	if _, err := os.Stat(rootName); os.IsNotExist(err) {
		err := os.Mkdir(rootName, 0755)
		if err != nil {
			return err
		}
	} else {
		// If the directory already exists, clear out all files in the directory
		files, err := filepath.Glob(filepath.Join(rootName, "*"))
		if err != nil {
			return err
		}
		for _, file := range files {
			err := os.Remove(file)
			if err != nil {
				return err
			}
		}
	}
	fmt.Println("setDirectory:", "Successfull!")
	return nil
}

func getFileName(directoryPath string) (string, error) {
	absPath, err := filepath.Abs(directoryPath)
	if err != nil {
		return "", err
	}

	files, err := ioutil.ReadDir(absPath)
	if err != nil {
		fmt.Println("err: ioutil.ReadDir(absPath)")
		return "", err
	}

	// Slice to store JPEG file names
	var jpegFiles []string

	// Iterate through the files and store the names of JPEG files
	for _, file := range files {
		if file.IsDir() {
			continue // Skip directories
		}
		// Check if the file has a .jpg or .jpeg extension
		if strings.HasSuffix(strings.ToLower(file.Name()),
			".jpg") || strings.HasSuffix(strings.ToLower(file.Name()),
			".jpeg") {
			jpegFiles = append(jpegFiles, file.Name())
		}
	}

	// Print the list of JPEG files
	fmt.Println("List of JPEG files:")
	for i, fileName := range jpegFiles {
		fmt.Printf("%d. %s\n", i+1, fileName)
	}

	// Ask the user to choose a file
	fmt.Print("Enter the number of the file you want to print: ")
	var choice int
	_, err = fmt.Scanln(&choice)
	if err != nil || choice < 1 || choice > len(jpegFiles) {
		return "", errors.New("err: Invalid choice.")
	}

	// Print the chosen JPEG file name
	selectedFile := jpegFiles[choice-1]
	file := absPath + "/" + selectedFile
	return file, nil
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

// Transform pixal space to mandelbrot space
// expressed as a complex number.
func (m *Movie) trans(x, y int) complex128 {
	drawScale := 3.5 * m.Scale
	aspect := float64(m.H) / float64(m.W)
	cRe := ((float64(x)/float64(m.W))-0.5)*drawScale + m.X
	cIm := ((float64(y)/float64(m.W))-(0.5*aspect))*drawScale - m.Y
	return complex(cRe, cIm)
}
