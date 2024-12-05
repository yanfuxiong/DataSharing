package utils

import (
	"bytes"
	"image"
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
