package logic

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func drawBBoxes(img image.Image, results []nexa_sdk.CVResult) *image.RGBA {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	const bboxLineWidth = 2
	d := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(color.White),
		Face: basicfont.Face7x13,
	}
	colors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},   // Red
		{R: 0, G: 255, B: 255, A: 255}, // Cyan
		{R: 0, G: 255, B: 0, A: 255},   // Green
		{R: 0, G: 0, B: 255, A: 255},   // Blue
		{R: 255, G: 255, B: 0, A: 255}, // Yellow
		{R: 255, G: 0, B: 255, A: 255}, // Magenta
		{R: 255, G: 165, B: 0, A: 255}, // Orange
		{R: 128, G: 0, B: 128, A: 255}, // Purple
	}
	for _, r := range results {
		if r.BBox.Width <= 0 || r.BBox.Height <= 0 {
			continue
		}
		bboxColor := colors[r.ClassID%int32(len(colors))]
		x, y, w, h := int(r.BBox.X), int(r.BBox.Y), int(r.BBox.Width), int(r.BBox.Height)
		if x < 0 {
			x, w = 0, w+x
		}
		if y < 0 {
			y, h = 0, h+y
		}
		if x+w > bounds.Dx() {
			w = bounds.Dx() - x
		}
		if y+h > bounds.Dy() {
			h = bounds.Dy() - y
		}

		for i := range bboxLineWidth {
			for j := x; j < x+w && j < bounds.Dx(); j++ {
				if y+i < bounds.Dy() {
					rgba.Set(j, y+i, bboxColor)
				}
				if y+h-1-i >= 0 {
					rgba.Set(j, y+h-1-i, bboxColor)
				}
			}
			for j := y; j < y+h && j < bounds.Dy(); j++ {
				if x+i < bounds.Dx() {
					rgba.Set(x+i, j, bboxColor)
				}
				if x+w-1-i >= 0 {
					rgba.Set(x+w-1-i, j, bboxColor)
				}
			}
		}

		textLabel := r.Text
		label := fmt.Sprintf("%s %.2f", textLabel, r.Confidence)
		labelWidth := d.MeasureString(label).Ceil()
		labelHeight := 12
		padding := 4

		labelY := y - labelHeight - padding*2
		if labelY < 0 {
			labelY = y + h + padding
		}

		bgRect := image.Rect(x, labelY, x+labelWidth+padding*2, labelY+labelHeight+padding*2)
		if bgRect.Max.X > bounds.Dx() {
			bgRect.Max.X = bounds.Dx()
		}
		if bgRect.Max.Y > bounds.Dy() {
			bgRect.Max.Y = bounds.Dy()
		}
		if bgRect.Min.Y < 0 {
			bgRect.Min.Y = 0
		}
		draw.Draw(rgba, bgRect, image.NewUniform(bboxColor), image.Point{}, draw.Over)
		d.Dot = fixed.P(x+padding, labelY+labelHeight+padding)
		d.DrawString(label)
	}

	return rgba
}

func drawMask(img image.Image, mask []float32) *image.RGBA {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	if len(mask) != bounds.Dx()*bounds.Dy() {
		slog.Error("Mask size does not match image size", "mask_size", len(mask), "image_size", bounds.Dx()*bounds.Dy())
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
		return rgba
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			idx := y*bounds.Dx() + x
			color := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			color.A = uint8(mask[idx] * 255)
			rgba.Set(x, y, color)
		}
	}

	return rgba
}

func CVPostProcess(input string, results []nexa_sdk.CVResult) (string, error) {
	file, err := os.Open(input)
	if err != nil {
		slog.Error("Failed to open image", "error", err)
		return "", err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		slog.Error("Failed to decode image", "error", err)
		return "", err
	}

	var rgba *image.RGBA
	if len(results) == 1 && len(results[0].Mask) > 0 &&
		results[0].BBox.X == 0 && results[0].BBox.Y == 0 &&
		results[0].BBox.Width == 0 && results[0].BBox.Height == 0 {
		rgba = drawMask(img, results[0].Mask)
	} else {
		rgba = drawBBoxes(img, results)
	}

	// save output image
	baseName := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
	outputPath := filepath.Join(".", baseName+"_output.png")
	outFile, err := os.Create(outputPath)
	if err != nil {
		slog.Error("Failed to create output image file", "error", err)
		return "", err
	}
	defer outFile.Close()

	err = png.Encode(outFile, rgba)
	if err != nil {
		slog.Error("Failed to encode output image", "error", err)
		return "", err
	}

	return outputPath, nil
}
