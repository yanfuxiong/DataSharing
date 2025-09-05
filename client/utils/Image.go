package utils

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	rtkMisc "rtk-cross-share/misc"
	"time"
)

const (
	kImgQuality = 100
)

func GetByteImageInfo(data []byte) (format string, width, height int) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		log.Printf("[%s] decode config err: %v", rtkMisc.GetFuncInfo(), err)
		return "", 0, 0
	}

	return format, cfg.Width, cfg.Height
}

func ImageToJpeg(format string, data []byte) ([]byte, error) {
	if format == "jpeg" {
		return data, nil
	}

	log.Printf("(SRC) Start to convert %s to jpg Qaulity(%d)", format, kImgQuality)
	startTime := time.Now().UnixMilli()

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Printf("[%s] decode img err: %v", rtkMisc.GetFuncInfo(), err)
		return nil, err
	}
	buf := &bytes.Buffer{}
	err = jpeg.Encode(buf, img, &jpeg.Options{Quality: kImgQuality})
	if err != nil {
		log.Printf("[%s] encode to jpeg err: %v", rtkMisc.GetFuncInfo(), err)
		return nil, err
	}

	log.Printf("(SRC) Convert %s to jpg. Input size:[%d] Output size:[%d] use [%d] ms...", format, len(data), buf.Len(), time.Now().UnixMilli()-startTime)

	return buf.Bytes(), nil
}

func BmpToJpg(data []byte, width, height, bitCount int) ([]byte, error) {
	log.Printf("(SRC) Start to convert bmp to jpg Qaulity(%d)", kImgQuality)
	startTime := time.Now().UnixMilli()

	expectedSize := width * height * 4
	if len(data) != expectedSize {
		log.Printf("invalid data size: expected %d, got %d", expectedSize, len(data))
		return data, errors.New("invalid data size")
	}

	if bitCount != 32 {
		log.Printf("[WARNING] unsupport bit count: %d", bitCount)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bytesPerPixel := 4
	stride := width * bytesPerPixel

	for y := 0; y < height; y++ {
		srcOffset := y * stride
		dstOffset := (height - 1 - y) * stride
		for x := 0; x < width; x++ {
			b := data[srcOffset]
			g := data[srcOffset+1]
			r := data[srcOffset+2]
			a := data[srcOffset+3]

			img.Pix[dstOffset+0] = r
			img.Pix[dstOffset+1] = g
			img.Pix[dstOffset+2] = b
			img.Pix[dstOffset+3] = a

			srcOffset += 4
			dstOffset += 4
		}
	}

	var buffer bytes.Buffer
	options := &jpeg.Options{Quality: kImgQuality}
	err := jpeg.Encode(&buffer, img, options)
	if err != nil {
		return nil, err
	}

	log.Printf("(SRC) Convert bmp to jpg. Input size:[%d]. Ouput size:[%d] use [%d] ms...", len(data), buffer.Len(), time.Now().UnixMilli()-startTime)
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

	bmpData := make([]byte, width*height*4)
	rowStride := width * 4

	rgbaImg, ok := img.(*image.RGBA)
	if !ok {
		rgbaImg = image.NewRGBA(bounds)
		draw.Draw(rgbaImg, bounds, img, bounds.Min, draw.Src)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcOffset := y*rgbaImg.Stride + x*4
			dstOffset := (height-1-y)*rowStride + x*4

			r := rgbaImg.Pix[srcOffset+0]
			g := rgbaImg.Pix[srcOffset+1]
			b := rgbaImg.Pix[srcOffset+2]
			a := rgbaImg.Pix[srcOffset+3]

			bmpData[dstOffset+0] = b // B
			bmpData[dstOffset+1] = g // G
			bmpData[dstOffset+2] = r // R
			bmpData[dstOffset+3] = a // A
		}
	}

	log.Printf("[JpgToBmp] successfully. Width:%d, Height:%d, BitCount:%d", width, height, bitCount)
	return bmpData, nil
}
