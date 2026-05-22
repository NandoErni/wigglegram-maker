package processor

import (
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"math"
	"os"
	"sync"

	xdraw "golang.org/x/image/draw"
	"wigglegram-maker/models"
)

func MakeThumbnail(src image.Image, size int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	xdraw.CatmullRom.Scale(
		dst,
		dst.Bounds(),
		src,
		src.Bounds(),
		xdraw.Over,
		nil,
	)
	return dst
}

func BuildLoopOrder(order []int, bounce bool, pauseFrames int) []int {
	if len(order) == 0 {
		return nil
	}

	if !bounce {
		return append([]int(nil), order...)
	}

	result := append([]int(nil), order...)

	for i := 0; i < pauseFrames; i++ {
		result = append(result, order[len(order)-1])
	}

	for i := len(order) - 2; i > 0; i-- {
		result = append(result, order[i])
	}

	for i := 0; i < pauseFrames; i++ {
		result = append(result, order[0])
	}

	return result
}

func LoadImageFile(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	file.Close()
	return img, err
}

func CanvasImageRect(img image.Image, offset models.FrameOffset) (float32, float32, float32, float32) {
	bounds := img.Bounds()
	imgW, imgH := float32(bounds.Dx()), float32(bounds.Dy())

	scale := modelsCanvasSize / imgW
	if imgH*scale > modelsCanvasSize {
		scale = modelsCanvasSize / imgH
	}

	displayW := imgW * scale
	displayH := imgH * scale
	padX := (modelsCanvasSize - displayW) / 2
	padY := (modelsCanvasSize - displayH) / 2

	return padX + offset.X,
		padY + offset.Y,
		padX + offset.X + displayW,
		padY + offset.Y + displayH
}

// MapCanvasCropToSource maps the 600x600 editor crop box onto the source image
// using the same contain-fit geometry used by the on-screen canvas image.
func MapCanvasCropToSource(
	img image.Image,
	offsetX, offsetY float32,
	cropMinX, cropMinY, cropMaxX, cropMaxY float32,
) (image.Rectangle, image.Point) {
	bounds := img.Bounds()
	minX, minY, maxX, _ := CanvasImageRect(img, models.FrameOffset{})
	imgW := float32(bounds.Dx())
	scale := (maxX - minX) / imgW

	srcX := (cropMinX - offsetX - minX) / scale
	srcY := (cropMinY - offsetY - minY) / scale
	srcW := (cropMaxX - cropMinX) / scale
	srcH := (cropMaxY - cropMinY) / scale

	width := int(math.Round(float64(srcW)))
	height := int(math.Round(float64(srcH)))
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	return image.Rect(0, 0, width, height), image.Point{
		X: bounds.Min.X + int(math.Round(float64(srcX))),
		Y: bounds.Min.Y + int(math.Round(float64(srcY))),
	}
}

const modelsCanvasSize = 600.0

func GenerateGIF(
	rawImages []image.Image,
	loopOrder []int,
	offsets []models.FrameOffset,
	cropMinX, cropMinY, cropMaxX, cropMaxY float32,
	gifLoopCount, gifFrameDelay int,
) (*gif.GIF, error) {
	if len(rawImages) == 0 || len(loopOrder) == 0 {
		return nil, nil
	}

	outGif := &gif.GIF{
		LoopCount: gifLoopCount,
	}

	uniqueFrames := make(map[int]*image.Paletted)
	for _, index := range loopOrder {
		if index >= 0 && index < len(rawImages) && index < len(offsets) {
			uniqueFrames[index] = nil
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for index := range uniqueFrames {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			img := rawImages[index]
			offset := offsets[index]
			cropRect, sourcePoint := MapCanvasCropToSource(
				img,
				offset.X,
				offset.Y,
				cropMinX,
				cropMinY,
				cropMaxX,
				cropMaxY,
			)

			croppedFrame := image.NewRGBA(cropRect)
			draw.Draw(croppedFrame, cropRect, img, sourcePoint, draw.Src)

			palettedFrame := image.NewPaletted(cropRect, palette.Plan9)
			draw.FloydSteinberg.Draw(palettedFrame, cropRect, croppedFrame, image.Point{})

			mu.Lock()
			uniqueFrames[index] = palettedFrame
			mu.Unlock()
		}(index)
	}

	wg.Wait()

	for _, index := range loopOrder {
		frame := uniqueFrames[index]
		if frame == nil {
			continue
		}
		outGif.Image = append(outGif.Image, frame)
		outGif.Delay = append(outGif.Delay, gifFrameDelay)
	}

	return outGif, nil
}
