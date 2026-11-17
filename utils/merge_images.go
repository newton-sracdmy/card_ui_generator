package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"

	xdraw "golang.org/x/image/draw"
)

const (
	maxMergedWidth   = 2400
	maxMergedHeight  = 2400
	maxFileSizeBytes = 3500000 // 3.5 MB safety threshold (OneID limit is 4MB)
)

func MergeImages(img1Bytes, img2Bytes []byte) ([]byte, error) {
	img1, _, err := image.Decode(bytes.NewReader(img1Bytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode first image: %w", err)
	}

	img2, _, err := image.Decode(bytes.NewReader(img2Bytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode second image: %w", err)
	}

	gap := 40 
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	mergedWidth := bounds1.Dx() + gap + bounds2.Dx()
	mergedHeight := bounds1.Dy()
	if bounds2.Dy() > mergedHeight {
		mergedHeight = bounds2.Dy()
	}

	mergedImg := image.NewRGBA(image.Rect(0, 0, mergedWidth, mergedHeight))

	draw.Draw(mergedImg, image.Rect(0, 0, bounds1.Dx(), bounds1.Dy()), img1, image.Point{}, draw.Src)

	draw.Draw(mergedImg, image.Rect(bounds1.Dx()+gap, 0, mergedWidth, bounds2.Dy()), img2, image.Point{}, draw.Src)

	var finalImg image.Image = mergedImg
	if mergedWidth > maxMergedWidth || mergedHeight > maxMergedHeight {
		scale := float64(maxMergedWidth) / float64(mergedWidth)
		if float64(maxMergedHeight)/float64(mergedHeight) < scale {
			scale = float64(maxMergedHeight) / float64(mergedHeight)
		}
		newW := int(float64(mergedWidth) * scale)
		newH := int(float64(mergedHeight) * scale)

		dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
		xdraw.CatmullRom.Scale(dst, dst.Bounds(), mergedImg, mergedImg.Bounds(), draw.Over, nil)
		finalImg = dst
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, finalImg); err != nil {
		return nil, fmt.Errorf("failed to encode merged image: %w", err)
	}

	resBytes := buf.Bytes()
	for len(resBytes) > maxFileSizeBytes {
		currW := finalImg.Bounds().Dx()
		currH := finalImg.Bounds().Dy()
		newW := int(float64(currW) * 0.8)
		newH := int(float64(currH) * 0.8)
		if newW < 300 || newH < 300 {
			break
		}
		dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
		xdraw.CatmullRom.Scale(dst, dst.Bounds(), finalImg, finalImg.Bounds(), draw.Over, nil)
		finalImg = dst

		buf.Reset()
		if err := png.Encode(&buf, finalImg); err != nil {
			return nil, fmt.Errorf("failed to encode resized merged image: %w", err)
		}
		resBytes = buf.Bytes()
	}

	return resBytes, nil
}
