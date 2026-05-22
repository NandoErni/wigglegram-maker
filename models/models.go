package models

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type FrameOffset struct {
	X, Y float32
}

type InteractionLayer struct {
	widget.BaseWidget

	OnDrag      func(delta fyne.Delta)
	OnRight     func(pos fyne.Position)
	OnMouseDown func(pos fyne.Position)
	OnMouseUp   func()
}

func (i *InteractionLayer) CreateRenderer() fyne.WidgetRenderer {
	r := canvas.NewRectangle(color.Transparent)
	return widget.NewSimpleRenderer(r)
}

func (i *InteractionLayer) Dragged(e *fyne.DragEvent) {
	if i.OnDrag != nil {
		i.OnDrag(e.Dragged)
	}
}

func (i *InteractionLayer) DragEnd() {}

func (i *InteractionLayer) Tapped(*fyne.PointEvent) {}

func (i *InteractionLayer) TappedSecondary(e *fyne.PointEvent) {
	if i.OnRight != nil {
		i.OnRight(e.Position)
	}
}

func (i *InteractionLayer) MouseDown(e *desktop.MouseEvent) {
	if i.OnMouseDown != nil {
		i.OnMouseDown(e.Position)
	}
}

func (i *InteractionLayer) MouseUp(e *desktop.MouseEvent) {
	if i.OnMouseUp != nil {
		i.OnMouseUp()
	}
}

func NewInteractionLayer() *InteractionLayer {
	i := &InteractionLayer{}
	i.ExtendBaseWidget(i)
	return i
}

const (
	CropHandleRadius float32 = 60
	HandleSize       float32 = 20
	CanvasSize       float32 = 600

	DefaultCropMinX float32 = 50.0
	DefaultCropMinY float32 = 50.0
	DefaultCropMaxX float32 = 550.0
	DefaultCropMaxY float32 = 550.0
)

const (
	PreviewFrameDelayNormal int = 110
	GIFFrameDelayNormal     int = 11

	PreviewFrameDelaySlow int = 180
	GIFFrameDelaySlow     int = 18

	PreviewFrameDelayFast int = 70
	GIFFrameDelayFast     int = 7

	PreviewFrameDelayHyper int = 40
	GIFFrameDelayHyper     int = 4
)

const (
	ThumbnailSize       int  = 128
	ThumbnailUISize     int  = 90
	PreviewSize         int  = 280
	DefaultGIFLoopCount int  = 0
	DefaultBounceMode   bool = true
	DefaultPauseFrames  int  = 1
)
