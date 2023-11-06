package fractal

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"strconv"
	"time"

	exif "github.com/dsoprea/go-exif/v3"
	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
)

const (
	PX = 3840
	PY = 2160
)

func CreateJPG(f *Fractal) {

	img := image.NewRGBA(image.Rect(0, 0, PX, PY))

	for py := 0; py < PY; py++ {
		for px := 0; px < PX; px++ {
			c := f.Mandelbrot(px, py, PX, PY)
			img.Set(px, py, c)
		}
	}

	fileName := "./pic/" + Time2str() + ".jpg"

	err := SaveJPG(f, img, fileName)
	if err != nil {
		log.Fatalln("SaveJPG:", err)
	}

	err = AddMetadata(f, fileName)
	if err != nil {
		log.Fatalln("AddMetadata:", err)
	}
}

func SaveJPG(f *Fractal, img *image.RGBA, fileName string) error {
	// Create a new file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode the image as a JPEG and write it to the file
	err = jpeg.Encode(file, img, nil)
	if err != nil {
		return err
	}
	return nil
}

func (m *MandelData) Encode() ([]byte, error) {
	scale := json.Number(strconv.FormatFloat(m.Scale, 'g', -1, 64))
	x := json.Number(strconv.FormatFloat(m.X, 'g', -1, 64))
	y := json.Number(strconv.FormatFloat(m.Y, 'g', -1, 64))

	return json.Marshal(&struct {
		Author   string      `json:"Author"`
		FileName string      `json:"FileName"`
		Scale    json.Number `json:"Scale"`
		X        json.Number `json:"X"`
		Y        json.Number `json:"Y"`
	}{
		Author:   m.Author,
		FileName: m.FileName,
		Scale:    scale,
		X:        x,
		Y:        y,
	})
}

func AddMetadata(f *Fractal, fileName string) error {

	intfc, err := jis.NewJpegMediaParser().ParseFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	sl := intfc.(*jis.SegmentList)

	ib, err := sl.ConstructExifBuilder()
	if err != nil {
		log.Fatal(err)
	}

	ifd0Ib, err := exif.GetOrCreateIbFromRootIb(ib, "IFD0")
	if err != nil {
		log.Fatal(err)
	}

	mandel := &MandelData{"John S. Dey Jr.", fileName,
		f.currScale, f.currX, f.currY}

	b, err := json.Marshal(mandel)
	if err != nil {
		log.Fatal(err)
	}

	jsonData := string(b)
	fmt.Println(jsonData)

	err = ifd0Ib.SetStandardWithName("DocumentName", jsonData)
	if err != nil {
		log.Fatal(err)
	}

	err = ifd0Ib.SetStandardWithName("Artist", mandel.Author)
	if err != nil {
		log.Fatal(err)
	}

	err = sl.SetExif(ib)
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = sl.Write(file)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func Time2str() string {
	now := time.Now()

	// Format the time as "YYMMDD@HHMMSS"
	formattedTime := fmt.Sprintf("%02d%02d%02d@%02d%02d%02d",
		now.Year()%100,
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second())

	return formattedTime
}
