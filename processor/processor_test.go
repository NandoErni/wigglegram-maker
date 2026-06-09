package processor

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"testing"

	"wigglegram-maker/models"
)

func TestGenerateGIFNormalizesMismatchedFrameSizes(t *testing.T) {
	rawImages := []image.Image{
		solidImage(2480, 3088, color.RGBA{255, 0, 0, 255}),
		solidImage(2526, 3089, color.RGBA{0, 255, 0, 255}),
		solidImage(2521, 3101, color.RGBA{0, 0, 255, 255}),
		solidImage(2461, 3098, color.RGBA{255, 255, 0, 255}),
	}
	offsets := make([]models.FrameOffset, len(rawImages))
	loopOrder := []int{0, 1, 2, 3, 2, 1}

	out, err := GenerateGIF(rawImages, loopOrder, offsets, 0, 0, 600, 600, 0, 11, 1)
	if err != nil {
		t.Fatal(err)
	}
	if out == nil {
		t.Fatal("expected GIF output")
	}
	if len(out.Image) != len(loopOrder) {
		t.Fatalf("expected %d frames, got %d", len(loopOrder), len(out.Image))
	}

	firstBounds := out.Image[0].Bounds()
	for i, frame := range out.Image {
		if frame.Bounds() != firstBounds {
			t.Fatalf("frame %d bounds = %v, want %v", i, frame.Bounds(), firstBounds)
		}
	}

	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, out); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateGIFAppliesExportScale(t *testing.T) {
	rawImages := []image.Image{
		solidImage(1000, 1000, color.RGBA{255, 0, 0, 255}),
	}
	offsets := []models.FrameOffset{{}}

	full, err := GenerateGIF(rawImages, []int{0}, offsets, 0, 0, 600, 600, 0, 11, 1)
	if err != nil {
		t.Fatal(err)
	}
	half, err := GenerateGIF(rawImages, []int{0}, offsets, 0, 0, 600, 600, 0, 11, 0.5)
	if err != nil {
		t.Fatal(err)
	}

	if half.Image[0].Bounds().Dx() >= full.Image[0].Bounds().Dx() {
		t.Fatalf("scaled export width = %d, full width = %d", half.Image[0].Bounds().Dx(), full.Image[0].Bounds().Dx())
	}
}

func solidImage(width, height int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)
	return img
}
