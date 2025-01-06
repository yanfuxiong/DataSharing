package utils

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
)

func mirrorHorizontal(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	mirrored := image.NewRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			mirrored.Set(width-1-x, y, img.At(x, y))
		}
	}

	return mirrored
}

func rotateImage(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			nx := width - 1 - x
			ny := height - 1 - y
			newImg.Set(nx, ny, img.At(x, y))
		}
	}

	return newImg
}

// PC>android
func BitmapToImage(bitmapData []byte, w, h int) []byte {
	//w, h, _ := GetByteImageInfo(bitmapData)

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (y*w + x) * 4

			/*
				img.Pix[4*(x+y*w)] = bitmapData[i+2]   //B
				img.Pix[4*(x+y*w)+1] = bitmapData[i+1] //g
				img.Pix[4*(x+y*w)+2] = bitmapData[i]   //r
				img.Pix[4*(x+y*w)+3] = 255             //A
			*/

			offset := 4 * ((w - 1 - x) + y*w)
			img.Pix[offset] = bitmapData[i+2]
			img.Pix[offset+1] = bitmapData[i+1]
			img.Pix[offset+2] = bitmapData[i]
			img.Pix[offset+3] = 255
		}
	}

	newImage := rotateImage(img)

	var buffer bytes.Buffer
	err := png.Encode(&buffer, newImage)
	if err != nil {
		log.Println(err)
		return nil
	}

	return buffer.Bytes()
}

func GetByteImageInfo(data []byte) (wight, height, size int) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		img, err = jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			log.Println("jpeg decode err:", err)
			return 0, 0, 0
		}
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()

	return w, h, (4 * w * h)

}

// android>PC
func ImageToBitmap(imgData []byte) []byte {
	img, err := png.Decode(bytes.NewReader(imgData))
	if err != nil {
		log.Println("png decode err, try jpeg...")
		img, err = jpeg.Decode(bytes.NewReader(imgData))
		if err != nil {
			log.Println("jpeg decode err:", err)
			return nil
		}
	}

	rgba := mirrorHorizontal(rotateImage(img))

	bitmapData := make([]byte, (rgba.Bounds().Dx())*(rgba.Bounds().Dy())*4)
	for y := 0; y < int(rgba.Bounds().Dy()); y++ {
		for x := 0; x < int(rgba.Bounds().Dx()); x++ {
			c := rgba.At(x, y)

			offset := (y*int(rgba.Bounds().Dx()) + x) * 4

			r, g, b, _ := c.RGBA()
			bitmapData[offset+2] = uint8(r)
			bitmapData[offset+1] = uint8(g)
			bitmapData[offset] = uint8(b)
			bitmapData[offset+3] = 0
		}
	}
	return bitmapData
}

func BmpToJpg(data []byte, width, height, bitCount int) ([]byte, error) {
	expectedSize := width * height * 4
	if len(data) != expectedSize {
		log.Printf("invalid data size: expected %d, got %d", expectedSize, len(data))
		return data, errors.New("invalid data size")
	}

	if bitCount != 32 {
		return data, errors.New("invalid bit count")
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	bytesPerPixel := 4
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := (y*width + x) * bytesPerPixel
			if offset+3 >= len(data) {
				log.Printf("unexpected end of data at offset %d", offset)
				return data, errors.New("unexpected end of data offset")
			}
			b := data[offset]
			g := data[offset+1]
			r := data[offset+2]
			a := data[offset+3]
			img.Set(x, height-y-1, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}

	var buffer bytes.Buffer
	options := &jpeg.Options{Quality: 90}
	err := jpeg.Encode(&buffer, img, options)
	if err != nil {
		return nil, err
	}

	log.Printf("[BmpToJpg] successfully. Width:%d, Height:%d, BitCount:%d", width, height, bitCount)
	return buffer.Bytes(), nil
}

func JpgToBmp(jpegData []byte) ([]byte, error) {
	img, err := jpeg.Decode(bytes.NewReader(jpegData))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	bitCount := 32
	bmpData := make([]byte, width*height*(bitCount/8)) // for 32bits
	rowStride := width * (bitCount/8) // for 32bits

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			offset := (height-y-1)*rowStride + x*(bitCount/8)

			bmpData[offset] = byte(b >> 8)
			bmpData[offset+1] = byte(g >> 8)
			bmpData[offset+2] = byte(r >> 8)
			bmpData[offset+3] = byte(a >> 8)
		}
	}

	log.Printf("[JpgToBmp] successfully. Width:%d, Height:%d, BitCount:%d", width, height, bitCount)
	return bmpData, nil
}
