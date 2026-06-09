package store

import (
	"image"

	"wigglegram-maker/models"
)

type AppState struct {
	RawImages          []image.Image
	ThumbnailImagesSrc []image.Image
	FrameNames         []string
	FrameOrder         []int
	Offsets            []models.FrameOffset
	SourceFolder       string

	CurrentActiveFrame int
	IsPlayingPreview   bool

	CropMinX, CropMinY float32
	CropMaxX, CropMaxY float32

	PreviewFrameDelay int
	GIFFrameDelay     int
	GIFLoopCount      int
	BounceMode        bool
	PauseFrames       int
	LockTransparency  bool
	ExportScale       float32

	CrosshairX float32
	CrosshairY float32

	StatusMessage string
}

func NewAppState() *AppState {
	return &AppState{
		RawImages:          make([]image.Image, 0),
		ThumbnailImagesSrc: make([]image.Image, 0),
		FrameNames:         make([]string, 0),
		FrameOrder:         make([]int, 0),
		Offsets:            make([]models.FrameOffset, 0),
		SourceFolder:       "",
		CurrentActiveFrame: 0,
		IsPlayingPreview:   false,
		CropMinX:           models.DefaultCropMinX,
		CropMinY:           models.DefaultCropMinY,
		CropMaxX:           models.DefaultCropMaxX,
		CropMaxY:           models.DefaultCropMaxY,
		PreviewFrameDelay:  models.PreviewFrameDelayNormal,
		GIFFrameDelay:      models.GIFFrameDelayNormal,
		GIFLoopCount:       models.DefaultGIFLoopCount,
		BounceMode:         models.DefaultBounceMode,
		PauseFrames:        models.DefaultPauseFrames,
		LockTransparency:   false,
		ExportScale:        0.75,
		CrosshairX:         300,
		CrosshairY:         300,
		StatusMessage:      "Welcome. Import folder sequence files to begin editing adjustments.",
	}
}

func (s *AppState) SwapFrameOrder(a, b int) {
	if a < 0 || b < 0 || a >= len(s.FrameOrder) || b >= len(s.FrameOrder) {
		return
	}
	s.FrameOrder[a], s.FrameOrder[b] = s.FrameOrder[b], s.FrameOrder[a]
}

func (s *AppState) ResetCropBox() {
	s.CropMinX = 0
	s.CropMinY = 0
	s.CropMaxX = models.CanvasSize
	s.CropMaxY = models.CanvasSize
}

func (s *AppState) UpdateOffsetX(delta float32) {
	if s.CurrentActiveFrame < 0 || s.CurrentActiveFrame >= len(s.Offsets) {
		return
	}
	s.Offsets[s.CurrentActiveFrame].X += delta
}

func (s *AppState) UpdateOffsetY(delta float32) {
	if s.CurrentActiveFrame < 0 || s.CurrentActiveFrame >= len(s.Offsets) {
		return
	}
	s.Offsets[s.CurrentActiveFrame].Y += delta
}

func (s *AppState) SetCropMinX(val float32) {
	if val < s.CropMaxX {
		s.CropMinX = val
	}
}

func (s *AppState) SetCropMinY(val float32) {
	if val < s.CropMaxY {
		s.CropMinY = val
	}
}

func (s *AppState) SetCropMaxX(val float32) {
	if val > s.CropMinX {
		s.CropMaxX = val
	}
}

func (s *AppState) SetCropMaxY(val float32) {
	if val > s.CropMinY {
		s.CropMaxY = val
	}
}

func (s *AppState) GetOffsetForFrame(index int) models.FrameOffset {
	if index < 0 || index >= len(s.Offsets) {
		return models.FrameOffset{X: 0, Y: 0}
	}
	return s.Offsets[index]
}

func (s *AppState) GetFrameCount() int {
	return len(s.RawImages)
}

func (s *AppState) HasFrames() bool {
	return len(s.RawImages) > 0
}
