package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
	"github.com/jsdey/fractal"
)

const (
	FPS        = 30
	MovLength  = 0.333
	Iterations = 200
	PX         = 3840
	PY         = 2160
	OUTDIR     = "./mov"
)

var (
	Frames = int(math.Round(float64(MovLength)*60)) * FPS
)

func main() {
	directoryPath := "../manExplore/pic"
	filePath, err := getFileName(directoryPath)

	err = setDirectory(filePath)
	if err != nil {
		log.Fatal(err)
	}

	err = calcFrames(filePath)
	if err != nil {
		log.Fatal(err)
	}
}

func calcFrames(filePath string) error {

	mOrig, err := getMetadata(filePath)
	if err != nil {
		return err
	}

	s := mOrig.Scale
	e := float64(1) / float64(Frames)
	scaleFactor := math.Pow(s, e)

	m := &fractal.MandelData{}
	m.Scale = 1

	m.X = mOrig.X
	m.Y = mOrig.Y

	// fmt.Println("calcFrames:", m)

	f := &fractal.Fractal{}
	f.SetFractal(m, Iterations)

	// fmt.Println("calcFrames:", f)

	// Set the file for the mp4 movie
	rootName, err := getFileElements(filePath)
	if err != nil {
		return err
	}
	rootName += ".mp4"
	outFile := OUTDIR + "/" + rootName
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

	for i := 0; i < Frames; i++ {
		img, err := calcMandel(m, i)
		if err != nil {
			return err
		}

		// Stream img to output
		err = jpeg.Encode(stdin, img, &jpeg.Options{Quality: 90})
		if err != nil {
			panic(err)
		}

		m.Scale *= scaleFactor
		f.SetFractal(m, Iterations)

		if i%25 == 0 {
			fmt.Println()
			fmt.Printf("%04d   ", i)
		} else {
			fmt.Print(".")
		}
		// fmt.Println("calcFrames:", "ENDING FOR TESTING.")
		// os.Exit(2)
		// fmt.Println("calcFrames: Frame Ending")
	}
	fmt.Println()
	return nil
}

func getMetadata(filePath string) (*fractal.MandelData, error) {
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
			m := &fractal.MandelData{}
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

func calcMandel(meta *fractal.MandelData, frameNo int) (*image.RGBA, error) {
	// fmt.Println("calcMandel")
	f := &fractal.Fractal{}
	f.SetFractal(meta, Iterations)

	img := image.NewRGBA(image.Rect(0, 0, PX, PY))
	// fmt.Println("calcMandel:", f)
	for py := 0; py < PY; py++ {
		for px := 0; px < PX; px++ {
			c := f.Mandelbrot(px, py, PX, PY)
			img.Set(px, py, c)
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
