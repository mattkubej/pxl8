package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"os"
)

// Pixel -> color space representation
type Pixel struct {
	R int
	G int
	B int
	A int
}

func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	var pixel Pixel

	pixel.R = int(r / 257)
	pixel.G = int(g / 257)
	pixel.B = int(b / 257)
	pixel.A = int(a / 257)

	return pixel
}

// https://stackoverflow.com/a/41185404
func getPixels(file io.Reader) ([][]Pixel, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	var pixels [][]Pixel
	for x := 0; x < width; x++ {
		var row []Pixel
		for y := 0; y < height; y++ {
			color := img.At(x, y)
			row = append(row, rgbaToPixel(color.RGBA()))
		}
		pixels = append(pixels, row)
	}

	return pixels, nil
}

func averagePixels(a Pixel, b Pixel) Pixel {
	var result Pixel

	result.R = int((a.R + b.R) / 2)
	result.G = int((a.G + b.G) / 2)
	result.B = int((a.B + b.B) / 2)
	result.A = int((a.A + b.A) / 2)

	return result
}

func pixelate(pixels [][]Pixel, blockSize int) [][]Pixel {
	width, height := len(pixels), len(pixels[0])
	result := pixels

	averages := make([][]Pixel, int(width/blockSize)+1)
	for i := range averages {
		averages[i] = make([]Pixel, int(height/blockSize)+1)
	}

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			avgX := int(x / blockSize)
			avgY := int(y / blockSize)
			averages[avgX][avgY] = averagePixels(averages[avgX][avgY], pixels[x][y])
		}
	}

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			result[x][y] = averages[int(x/blockSize)][int(y/blockSize)]
		}
	}

	return result
}

func pixelToRGBA(pixel Pixel) color.RGBA {
	var rgba color.RGBA

	rgba.R = uint8(pixel.R)
	rgba.G = uint8(pixel.G)
	rgba.B = uint8(pixel.B)
	rgba.A = uint8(pixel.A)

	return rgba
}

func outputImg(pixels [][]Pixel) {
	width, height := len(pixels), len(pixels[0])

	img := image.NewRGBA64(image.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			pixel := pixels[x][y]
			img.Set(x, y, pixelToRGBA(pixel))
		}
	}

	f, _ := os.OpenFile("out.png", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()

	png.Encode(f, img)
}

func main() {
	in := flag.String("i", "", "image to pixelate")
	blockSize := flag.Int("bs", 8, "block size, defaults to 8")
	flag.Parse()

	if *in == "" {
		fmt.Fprintln(os.Stderr, "input image requred, use the -h flag for help")
		os.Exit(1)
	}

	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

	file, err := os.Open(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to open image")
		os.Exit(1)
	}
	defer file.Close()

	pixels, err := getPixels(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get pixels from image")
		os.Exit(1)
	}

	result := pixelate(pixels, *blockSize)

	outputImg(result)
}
