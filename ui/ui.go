package ui

import (
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"wigglegram-maker/models"
	"wigglegram-maker/processor"
	"wigglegram-maker/store"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func LoadImagesFromFolderPath(state *store.AppState, folderPath string) error {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return err
	}

	var paths []string
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if !f.IsDir() && (ext == ".jpg" || ext == ".jpeg" || ext == ".png") {
			paths = append(paths, filepath.Join(folderPath, f.Name()))
		}
	}
	sort.Strings(paths)

	if len(paths) == 0 {
		return nil
	}

	state.RawImages = make([]image.Image, 0, len(paths))
	state.ThumbnailImagesSrc = make([]image.Image, 0)
	state.FrameNames = make([]string, 0, len(paths))
	state.FrameOrder = make([]int, 0, len(paths))
	state.Offsets = make([]models.FrameOffset, len(paths))
	state.CurrentActiveFrame = 0

	for _, path := range paths {
		img, err := processor.LoadImageFile(path)
		if err != nil {
			println("Error loading image:", err)
			continue
		}
		state.RawImages = append(state.RawImages, img)
		state.ThumbnailImagesSrc = append(
			state.ThumbnailImagesSrc,
			processor.MakeThumbnail(img, models.ThumbnailSize),
		)
		state.FrameNames = append(state.FrameNames, filepath.Base(path))
		state.FrameOrder = append(state.FrameOrder, len(state.RawImages)-1)
	}

	state.ResetCropBox()
	return nil
}

func LoadImagesFromFolder(state *store.AppState, uri fyne.ListableURI) error {
	return LoadImagesFromFolderPath(state, uri.Path())
}

func CreateLockTransparencyCheck(state *store.AppState, onStatusUpdate func(string)) *widget.Check {
	lockTransparencyCheck := widget.NewCheck("Lock Transparency (Glint Mode)", func(checked bool) {
		state.LockTransparency = checked
		if checked {
			onStatusUpdate("Glint Layering Mode: Active layer is solid; others are ghosted backdrops.")
		} else {
			onStatusUpdate("Alpha Layering Mode: Layers use transparency order.")
		}
	})
	return lockTransparencyCheck
}

func CreateSpeedSelect(state *store.AppState) *widget.Select {
	speedSelect := widget.NewSelect(
		[]string{"Slow", "Normal", "Fast", "Hyper"},
		func(v string) {
			switch v {
			case "Slow":
				state.PreviewFrameDelay = models.PreviewFrameDelaySlow
				state.GIFFrameDelay = models.GIFFrameDelaySlow
			case "Normal":
				state.PreviewFrameDelay = models.PreviewFrameDelayNormal
				state.GIFFrameDelay = models.GIFFrameDelayNormal
			case "Fast":
				state.PreviewFrameDelay = models.PreviewFrameDelayFast
				state.GIFFrameDelay = models.GIFFrameDelayFast
			case "Hyper":
				state.PreviewFrameDelay = models.PreviewFrameDelayHyper
				state.GIFFrameDelay = models.GIFFrameDelayHyper
			}
		},
	)
	speedSelect.SetSelected("Normal")
	return speedSelect
}

func CreateLoopCountSelect(state *store.AppState) *widget.Select {
	loopCountSelect := widget.NewSelect(
		[]string{"Infinite", "1", "2", "3", "5", "10"},
		func(v string) {
			switch v {
			case "Infinite":
				state.GIFLoopCount = 0
			case "1":
				state.GIFLoopCount = 1
			case "2":
				state.GIFLoopCount = 2
			case "3":
				state.GIFLoopCount = 3
			case "5":
				state.GIFLoopCount = 5
			case "10":
				state.GIFLoopCount = 10
			}
		},
	)
	loopCountSelect.SetSelected("Infinite")
	return loopCountSelect
}

func CreatePauseSelect(state *store.AppState) *widget.Select {
	pauseSelect := widget.NewSelect(
		[]string{"None", "Short", "Medium", "Long"},
		func(v string) {
			switch v {
			case "None":
				state.PauseFrames = 0
			case "Short":
				state.PauseFrames = 1
			case "Medium":
				state.PauseFrames = 2
			case "Long":
				state.PauseFrames = 4
			}
		},
	)
	pauseSelect.SetSelected("Short")
	return pauseSelect
}

func CreateBounceCheck(state *store.AppState) *widget.Check {
	bounceCheck := widget.NewCheck("Bounce Animation", func(v bool) {
		state.BounceMode = v
	})
	bounceCheck.SetChecked(true)
	return bounceCheck
}
