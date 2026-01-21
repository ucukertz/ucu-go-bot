package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"

	"golang.org/x/image/webp"
)

const (
	PIC_RESO_1k  = 1024
	PIC_RESO_1k2 = 1280
	PIC_RESO_1k5 = 1536
	PIC_RESO_2k  = 2048
)

func PicByte2ImgImg(b []byte) (image.Image, error) {
	imgimg, _, err := image.Decode(bytes.NewReader(b))
	if err == nil {
		return imgimg, nil
	}
	pngimg, err := png.Decode(bytes.NewReader(b))
	if err == nil {
		return pngimg, nil
	}
	jpegimg, err := jpeg.Decode(bytes.NewReader(b))
	if err == nil {
		return jpegimg, nil
	}
	webpimg, err := webp.Decode(bytes.NewReader(b))
	if err == nil {
		return webpimg, nil
	}

	return nil, fmt.Errorf("unsupported image format")
}

// Helper to round an integer to the nearest multiple of 16 (AI friendly)
func PicSnap16(n int) int {
	if n <= 0 {
		return 16
	}
	remainder := n % 16
	if remainder == 0 {
		return n
	}
	if remainder >= 8 {
		return n + (16 - remainder)
	}
	return n - remainder
}

func Pic2DSnap16(w, h int) (int, int) {
	return PicSnap16(w), PicSnap16(h)
}

// PicAdjustReso calculates new dimensions (w, h) such that their total pixel count
// is close to target*target, while preserving aspect ratio.
// All inputs and outputs are snapped to the nearest multiple of 16 (AI friendly).
func PicAdjustReso(w, h, target int) (int, int) {
	targetSnapped := PicSnap16(target)
	targetArea := float64(targetSnapped * targetSnapped)

	aspectRatio := float64(w) / float64(h)

	newHFloat := math.Sqrt(targetArea / aspectRatio)
	newWFloat := newHFloat * aspectRatio

	finalW := PicSnap16(int(math.Round(newWFloat)))
	finalH := PicSnap16(int(math.Round(newHFloat)))

	return finalW, finalH
}

// PicExpandLow expands the smaller dimension (w or h) to target,
// preserving aspect ratio. If both dimensions are >= target, no change is made.
func PicExpandLow(w, h, target int) (int, int) {
	if w < target || h < target {
		if w < h {
			w = target
			h = int(float64(h) * (float64(target) / float64(w)))
		} else {
			h = target
			w = int(float64(w) * (float64(target) / float64(h)))
		}
	}

	return w, h
}
