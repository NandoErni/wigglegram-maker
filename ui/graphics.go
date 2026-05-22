package ui

import (
	"image"
	"image/draw"

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

func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
