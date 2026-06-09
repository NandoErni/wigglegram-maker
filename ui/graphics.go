package ui

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"wigglegram-maker/models"
	"wigglegram-maker/processor"
	"wigglegram-maker/store"
)

func UpdateCrosshairGraphic(
	crosshairX, crosshairY *canvas.Line,
	state *store.AppState,
) {
	crosshairX.Position1 = fyne.NewPos(state.CrosshairX-20, state.CrosshairY)
	crosshairX.Position2 = fyne.NewPos(state.CrosshairX+20, state.CrosshairY)

	crosshairY.Position1 = fyne.NewPos(state.CrosshairX, state.CrosshairY-20)
	crosshairY.Position2 = fyne.NewPos(state.CrosshairX, state.CrosshairY+20)

	crosshairX.Refresh()
	crosshairY.Refresh()
}

func UpdateCropBoxGraphic(
	cropBoxOutline *canvas.Rectangle,
	handles [4]*canvas.Rectangle,
	state *store.AppState,
) {
	cropBoxOutline.Move(fyne.NewPos(state.CropMinX, state.CropMinY))
	cropBoxOutline.Resize(fyne.NewSize(state.CropMaxX-state.CropMinX, state.CropMaxY-state.CropMinY))

	handles[0].Move(fyne.NewPos(state.CropMinX, state.CropMinY))
	handles[0].Resize(fyne.NewSize(models.HandleSize, models.HandleSize))

	handles[1].Move(fyne.NewPos(state.CropMaxX-models.HandleSize, state.CropMinY))
	handles[1].Resize(fyne.NewSize(models.HandleSize, models.HandleSize))

	handles[2].Move(fyne.NewPos(state.CropMinX, state.CropMaxY-models.HandleSize))
	handles[2].Resize(fyne.NewSize(models.HandleSize, models.HandleSize))

	handles[3].Move(fyne.NewPos(state.CropMaxX-models.HandleSize, state.CropMaxY-models.HandleSize))
	handles[3].Resize(fyne.NewSize(models.HandleSize, models.HandleSize))

	cropBoxOutline.Refresh()
	for _, h := range handles {
		h.Refresh()
	}
}

func GetCropHandle(pos fyne.Position, state *store.AppState) string {
	const radius = 60

	if abs(pos.X-(state.CropMinX-6)) < radius && abs(pos.Y-(state.CropMinY-6)) < radius {
		return "tl"
	}
	if abs(pos.X-(state.CropMaxX-6)) < radius && abs(pos.Y-(state.CropMinY-6)) < radius {
		return "tr"
	}
	if abs(pos.X-(state.CropMinX-6)) < radius && abs(pos.Y-(state.CropMaxY-6)) < radius {
		return "bl"
	}
	if abs(pos.X-(state.CropMaxX-6)) < radius && abs(pos.Y-(state.CropMaxY-6)) < radius {
		return "br"
	}

	return ""
}

func RenderLayers(
	canvasImages []*canvas.Image,
	cropBoxOutline *canvas.Rectangle,
	handles [4]*canvas.Rectangle,
	crosshairX, crosshairY *canvas.Line,
	state *store.AppState,
) {
	if !state.HasFrames() {
		return
	}

	for _, img := range canvasImages {
		if img != nil {
			img.Hide()
		}
	}

	if len(canvasImages) > 0 && canvasImages[0] != nil {
		canvasImages[0].Move(fyne.NewPos(state.Offsets[0].X, state.Offsets[0].Y))

		if state.LockTransparency && state.CurrentActiveFrame != 0 {
			canvasImages[0].Translucency = 0.9
		} else {
			canvasImages[0].Translucency = 0.0
		}

		canvasImages[0].Show()
		canvasImages[0].Refresh()
	}

	if state.CurrentActiveFrame > 0 &&
		state.CurrentActiveFrame < len(canvasImages) &&
		state.CurrentActiveFrame < len(state.Offsets) &&
		canvasImages[state.CurrentActiveFrame] != nil {
		canvasImages[state.CurrentActiveFrame].Move(fyne.NewPos(
			state.Offsets[state.CurrentActiveFrame].X,
			state.Offsets[state.CurrentActiveFrame].Y,
		))

		canvasImages[state.CurrentActiveFrame].Translucency = 0.35
		canvasImages[state.CurrentActiveFrame].Show()
		canvasImages[state.CurrentActiveFrame].Refresh()
	}

	UpdateCropBoxGraphic(cropBoxOutline, handles, state)
	UpdateCrosshairGraphic(crosshairX, crosshairY, state)
}

func GeneratePreviewFrame(
	rawImages []image.Image,
	frameIdx int,
	offsets []models.FrameOffset,
	cropMinX, cropMinY, cropMaxX, cropMaxY float32,
) image.Image {
	if frameIdx < 0 || frameIdx >= len(rawImages) || frameIdx >= len(offsets) {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	img := rawImages[frameIdx]
	offset := offsets[frameIdx]
	cropRect, sourcePoint := processor.MapCanvasCropToSource(
		img,
		offset.X,
		offset.Y,
		cropMinX,
		cropMinY,
		cropMaxX,
		cropMaxY,
	)

	previewCanvasObj := image.NewRGBA(cropRect)
	draw.Draw(
		previewCanvasObj,
		cropRect,
		img,
		sourcePoint,
		draw.Src,
	)

	return previewCanvasObj
}

func GenerateLoupeFrame(
	rawImages []image.Image,
	offsets []models.FrameOffset,
	activeFrame int,
	center fyne.Position,
	sampleSize int,
	outputSize int,
) image.Image {
	if len(rawImages) == 0 || len(offsets) == 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	dst := image.NewRGBA(image.Rect(0, 0, outputSize, outputSize))
	halfSample := float32(sampleSize) / 2
	scale := float32(sampleSize) / float32(outputSize)

	for y := 0; y < outputSize; y++ {
		for x := 0; x < outputSize; x++ {
			wx := center.X - halfSample + (float32(x)+0.5)*scale
			wy := center.Y - halfSample + (float32(y)+0.5)*scale

			c := sampleCanvasColor(rawImages, offsets, 0, wx, wy)
			if activeFrame > 0 && activeFrame < len(rawImages) {
				active := sampleCanvasColor(rawImages, offsets, activeFrame, wx, wy)
				c = blendRGBA(c, active, 0.65)
			}

			dst.SetRGBA(x, y, c)
		}
	}

	drawLoupeCrosshair(dst)
	return dst
}

func sampleCanvasColor(
	rawImages []image.Image,
	offsets []models.FrameOffset,
	frameIndex int,
	wx float32,
	wy float32,
) color.RGBA {
	if frameIndex < 0 || frameIndex >= len(rawImages) || frameIndex >= len(offsets) {
		return color.RGBA{20, 20, 20, 255}
	}

	img := rawImages[frameIndex]
	minX, minY, maxX, maxY := processor.CanvasImageRect(img, offsets[frameIndex])
	if wx < minX || wx >= maxX || wy < minY || wy >= maxY {
		return color.RGBA{20, 20, 20, 255}
	}

	bounds := img.Bounds()
	sx := bounds.Min.X + int(math.Floor(float64((wx-minX)/(maxX-minX)*float32(bounds.Dx()))))
	sy := bounds.Min.Y + int(math.Floor(float64((wy-minY)/(maxY-minY)*float32(bounds.Dy()))))
	if sx < bounds.Min.X {
		sx = bounds.Min.X
	}
	if sy < bounds.Min.Y {
		sy = bounds.Min.Y
	}
	if sx >= bounds.Max.X {
		sx = bounds.Max.X - 1
	}
	if sy >= bounds.Max.Y {
		sy = bounds.Max.Y - 1
	}

	return color.RGBAModel.Convert(img.At(sx, sy)).(color.RGBA)
}

func blendRGBA(base color.RGBA, overlay color.RGBA, alpha float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(base.R)*(1-alpha) + float64(overlay.R)*alpha),
		G: uint8(float64(base.G)*(1-alpha) + float64(overlay.G)*alpha),
		B: uint8(float64(base.B)*(1-alpha) + float64(overlay.B)*alpha),
		A: 255,
	}
}

func drawLoupeCrosshair(img *image.RGBA) {
	bounds := img.Bounds()
	cx := bounds.Dx() / 2
	cy := bounds.Dy() / 2
	c := color.RGBA{255, 60, 60, 230}

	for x := cx - 10; x <= cx+10; x++ {
		if x >= bounds.Min.X && x < bounds.Max.X {
			img.SetRGBA(x, cy, c)
		}
	}
	for y := cy - 10; y <= cy+10; y++ {
		if y >= bounds.Min.Y && y < bounds.Max.Y {
			img.SetRGBA(cx, y, c)
		}
	}
}

func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
