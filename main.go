package main

import (
	"crypto/rand"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	fynedialog "fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ncruces/zenity"

	"wigglegram-maker/models"
	"wigglegram-maker/processor"
	"wigglegram-maker/store"
	"wigglegram-maker/ui"
)

func main() {
	myApp := app.New()
	myApp.SetIcon(appIcon)
	myWindow := myApp.NewWindow("Wigglegram Maker")
	myWindow.Resize(fyne.NewSize(1300, 850))

	state := store.NewAppState()
	var canvasImages []*canvas.Image

	crosshairX := canvas.NewLine(color.RGBA{255, 0, 0, 255})
	crosshairY := canvas.NewLine(color.RGBA{255, 0, 0, 255})
	crosshairX.StrokeWidth = 2
	crosshairY.StrokeWidth = 2

	cropBoxOutline := canvas.NewRectangle(color.Transparent)
	cropBoxOutline.StrokeColor = color.RGBA{255, 255, 255, 230}
	cropBoxOutline.StrokeWidth = 2.5

	cropHandles := [4]*canvas.Rectangle{
		canvas.NewRectangle(theme.PrimaryColor()),
		canvas.NewRectangle(theme.PrimaryColor()),
		canvas.NewRectangle(theme.PrimaryColor()),
		canvas.NewRectangle(theme.PrimaryColor()),
	}
	loupeImage := canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	loupeImage.FillMode = canvas.ImageFillStretch
	loupeImage.Resize(fyne.NewSize(180, 180))
	loupeImage.Hide()

	loupeBorder := canvas.NewRectangle(color.Transparent)
	loupeBorder.StrokeColor = color.RGBA{255, 255, 255, 240}
	loupeBorder.StrokeWidth = 2
	loupeBorder.Resize(fyne.NewSize(180, 180))
	loupeBorder.Hide()

	imageWorkspace := container.NewWithoutLayout()
	imageWorkspace.Add(crosshairX)
	imageWorkspace.Add(crosshairY)
	imageWorkspace.Add(cropBoxOutline)
	for _, h := range cropHandles {
		imageWorkspace.Add(h)
	}
	imageWorkspace.Add(loupeImage)
	imageWorkspace.Add(loupeBorder)

	interactionLayer := models.NewInteractionLayer()
	interactionLayer.Resize(fyne.NewSize(models.CanvasSize, models.CanvasSize))
	imageWorkspace.Add(interactionLayer)

	var activeCropHandle string
	statusLabel := widget.NewLabel(state.StatusMessage)
	statusLabel.Wrapping = fyne.TextWrapWord
	updateLoupe := func(pos fyne.Position, show bool) {
		if !show || !state.HasFrames() || activeCropHandle != "" {
			loupeImage.Hide()
			loupeBorder.Hide()
			return
		}

		loupeFrame := ui.GenerateLoupeFrame(
			state.RawImages,
			state.Offsets,
			state.CurrentActiveFrame,
			pos,
			72,
			180,
		)
		loupeImage.Image = loupeFrame

		x := pos.X + 24
		y := pos.Y + 24
		if x+180 > models.CanvasSize {
			x = pos.X - 204
		}
		if y+180 > models.CanvasSize {
			y = pos.Y - 204
		}
		if x < 0 {
			x = 0
		}
		if y < 0 {
			y = 0
		}

		loupePos := fyne.NewPos(x, y)
		loupeImage.Move(loupePos)
		loupeBorder.Move(loupePos)
		loupeImage.Show()
		loupeBorder.Show()
		loupeImage.Refresh()
		loupeBorder.Refresh()
	}

	interactionLayer.OnDrag = func(delta fyne.Delta, pos fyne.Position) {
		if activeCropHandle != "" {
			switch activeCropHandle {
			case "tl":
				state.CropMinX += delta.DX
				state.CropMinY += delta.DY
			case "tr":
				state.CropMaxX += delta.DX
				state.CropMinY += delta.DY
			case "bl":
				state.CropMinX += delta.DX
				state.CropMaxY += delta.DY
			case "br":
				state.CropMaxX += delta.DX
				state.CropMaxY += delta.DY
			}
			ui.UpdateCropBoxGraphic(cropBoxOutline, cropHandles, state)
			updateLoupe(pos, false)
			return
		}

		if !state.HasFrames() {
			return
		}
		state.UpdateOffsetX(delta.DX)
		state.UpdateOffsetY(delta.DY)

		if state.CurrentActiveFrame >= 0 &&
			state.CurrentActiveFrame < len(canvasImages) &&
			canvasImages[state.CurrentActiveFrame] != nil {
			canvasImages[state.CurrentActiveFrame].Move(fyne.NewPos(
				state.Offsets[state.CurrentActiveFrame].X,
				state.Offsets[state.CurrentActiveFrame].Y,
			))
			canvasImages[state.CurrentActiveFrame].Refresh()
		}
		updateLoupe(pos, true)
	}

	interactionLayer.OnRight = func(position fyne.Position) {
		state.CrosshairX = position.X
		state.CrosshairY = position.Y
		ui.UpdateCrosshairGraphic(crosshairX, crosshairY, state)
		ui.RenderLayers(canvasImages, cropBoxOutline, cropHandles, crosshairX, crosshairY, state)
		statusLabel.SetText("Reference point moved.")
	}

	interactionLayer.OnMouseDown = func(pos fyne.Position) {
		activeCropHandle = ui.GetCropHandle(pos, state)
	}
	interactionLayer.OnMouseUp = func() {
		activeCropHandle = ""
		updateLoupe(fyne.Position{}, false)
	}
	interactionLayer.OnMouseOut = func() {
		updateLoupe(fyne.Position{}, false)
	}

	myWindow.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		if !state.HasFrames() {
			return
		}
		var speed float32 = 1.0
		switch k.Name {
		case fyne.KeyLeft:
			state.UpdateOffsetX(-speed)
		case fyne.KeyRight:
			state.UpdateOffsetX(speed)
		case fyne.KeyUp:
			state.UpdateOffsetY(-speed)
		case fyne.KeyDown:
			state.UpdateOffsetY(speed)
		case fyne.KeyR:
			state.ResetCropBox()
			ui.UpdateCropBoxGraphic(cropBoxOutline, cropHandles, state)
		default:
			return
		}
		ui.RenderLayers(canvasImages, cropBoxOutline, cropHandles, crosshairX, crosshairY, state)
	})

	previewImageRef := canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	previewImageRef.FillMode = canvas.ImageFillContain
	previewImageRef.SetMinSize(fyne.NewSize(float32(models.PreviewSize), float32(models.PreviewSize)))

	lockTransparencyCheck := ui.CreateLockTransparencyCheck(state, func(msg string) {
		statusLabel.SetText(msg)
	})

	speedSelect := ui.CreateSpeedSelect(state)
	loopCountSelect := ui.CreateLoopCountSelect(state)
	bounceCheck := ui.CreateBounceCheck(state)
	pauseSelect := ui.CreatePauseSelect(state)
	exportQualitySelect := ui.CreateExportQualitySelect(state)

	maxSafeCropButton := widget.NewButton("Max Safe Crop", func() {
		if !state.HasFrames() {
			return
		}

		minX, minY := float32(0), float32(0)
		maxX, maxY := models.CanvasSize, models.CanvasSize

		for i, img := range state.RawImages {
			if i >= len(state.Offsets) {
				continue
			}

			imageMinX, imageMinY, imageMaxX, imageMaxY :=
				processor.CanvasImageRect(img, state.Offsets[i])

			if imageMinX > minX {
				minX = imageMinX
			}
			if imageMinY > minY {
				minY = imageMinY
			}
			if imageMaxX < maxX {
				maxX = imageMaxX
			}
			if imageMaxY < maxY {
				maxY = imageMaxY
			}
		}

		if maxX <= minX || maxY <= minY {
			statusLabel.SetText("No shared image area remains after alignment.")
			return
		}

		state.CropMinX = minX
		state.CropMinY = minY
		state.CropMaxX = maxX
		state.CropMaxY = maxY
		ui.RenderLayers(canvasImages, cropBoxOutline, cropHandles, crosshairX, crosshairY, state)
		statusLabel.SetText("Crop fitted to shared image area.")
	})

	thumbnailStrip := container.NewHBox()
	var refreshThumbnails func()
	refreshThumbnails = func() {
		thumbnailStrip.Objects = nil

		for i := 0; i < len(state.FrameOrder); i++ {
			position := i
			actualIndex := state.FrameOrder[position]
			if actualIndex < 0 ||
				actualIndex >= len(state.RawImages) ||
				actualIndex >= len(state.ThumbnailImagesSrc) {
				continue
			}

			thumb := canvas.NewImageFromImage(state.ThumbnailImagesSrc[actualIndex])
			thumb.FillMode = canvas.ImageFillContain
			thumb.SetMinSize(fyne.NewSize(float32(models.ThumbnailUISize), float32(models.ThumbnailUISize)))

			border := canvas.NewRectangle(color.Transparent)
			if state.CurrentActiveFrame == actualIndex {
				border.StrokeColor = color.RGBA{0, 255, 0, 255}
				border.StrokeWidth = 3
			}

			leftBtn := widget.NewButton("<", func() {
				if position == 0 {
					return
				}
				state.SwapFrameOrder(position, position-1)
				refreshThumbnails()
			})

			rightBtn := widget.NewButton(">", func() {
				if position >= len(state.FrameOrder)-1 {
					return
				}
				state.SwapFrameOrder(position, position+1)
				refreshThumbnails()
			})

			name := fmt.Sprintf("Frame %d", actualIndex+1)
			if actualIndex < len(state.FrameNames) {
				name = state.FrameNames[actualIndex]
			}

			labelBtn := widget.NewButton(name, func() {
				state.CurrentActiveFrame = actualIndex
				refreshThumbnails()
				ui.RenderLayers(canvasImages, cropBoxOutline, cropHandles, crosshairX, crosshairY, state)
			})

			item := container.NewStack(
				border,
				container.NewVBox(
					thumb,
					labelBtn,
					container.NewHBox(leftBtn, rightBtn),
				),
			)
			thumbnailStrip.Add(item)
		}

		thumbnailStrip.Refresh()
	}

	imageFilters := zenity.FileFilters{
		{Name: "Images", Patterns: []string{"*.jpg", "*.jpeg", "*.png"}, CaseFold: true},
	}

	reloadWorkspace := func() {
		imageWorkspace.Objects = nil
		canvasImages = make([]*canvas.Image, 0, len(state.RawImages))
		for _, img := range state.RawImages {
			canvasImage := canvas.NewImageFromImage(img)
			canvasImage.FillMode = canvas.ImageFillContain
			canvasImage.Resize(fyne.NewSize(models.CanvasSize, models.CanvasSize))
			canvasImages = append(canvasImages, canvasImage)
			imageWorkspace.Add(canvasImage)
		}
		imageWorkspace.Add(crosshairX)
		imageWorkspace.Add(crosshairY)
		imageWorkspace.Add(cropBoxOutline)
		for _, h := range cropHandles {
			imageWorkspace.Add(h)
		}
		imageWorkspace.Add(loupeImage)
		imageWorkspace.Add(loupeBorder)
		imageWorkspace.Add(interactionLayer)
		imageWorkspace.Refresh()

		ui.RenderLayers(canvasImages, cropBoxOutline, cropHandles, crosshairX, crosshairY, state)
		refreshThumbnails()
	}

	loadButton := widget.NewButton("Open Images", func() {
		options := []zenity.Option{
			zenity.Title("Open wigglegram frames"),
			imageFilters,
		}
		if state.SourceFolder != "" {
			options = append(options, zenity.Filename(state.SourceFolder))
		}

		paths, err := zenity.SelectFileMultiple(options...)
		if err != nil {
			if err != zenity.ErrCanceled {
				fynedialog.ShowError(err, myWindow)
			}
			return
		}

		go func() {
			err = ui.LoadImagesFromPaths(state, paths)
			if err != nil {
				fyne.Do(func() {
					fynedialog.ShowError(err, myWindow)
				})
				return
			}

			if !state.HasFrames() {
				fyne.Do(func() {
					statusLabel.SetText("No supported images selected.")
				})
				return
			}

			fyne.Do(func() {
				reloadWorkspace()
				statusLabel.SetText(fmt.Sprintf("%d frame(s) loaded successfully", len(state.RawImages)))
			})
		}()
	})

	compileAndSaveGIF := func(filename string) {
		if !state.HasFrames() {
			return
		}
		if !strings.HasSuffix(strings.ToLower(filename), ".gif") {
			filename += ".gif"
		}

		loopOrder := processor.BuildLoopOrder(state.FrameOrder, state.BounceMode, state.PauseFrames)
		progress := fynedialog.NewProgressInfinite("Compiling GIF", "Encoding photographic frames...", myWindow)
		progress.Show()

		go func() {
			outGif, err := processor.GenerateGIF(
				state.RawImages,
				loopOrder,
				state.Offsets,
				state.CropMinX, state.CropMinY, state.CropMaxX, state.CropMaxY,
				state.GIFLoopCount, state.GIFFrameDelay,
				state.ExportScale,
			)

			fyne.Do(func() {
				progress.Hide()
			})

			if err != nil {
				fyne.Do(func() {
					fynedialog.ShowError(err, myWindow)
				})
				return
			}

			file, err := os.Create(filename)
			if err != nil {
				fyne.Do(func() {
					fynedialog.ShowError(err, myWindow)
				})
				return
			}
			defer file.Close()

			if err := gif.EncodeAll(file, outGif); err != nil {
				fyne.Do(func() {
					fynedialog.ShowError(err, myWindow)
				})
				return
			}

			fyne.Do(func() {
				statusLabel.SetText(fmt.Sprintf("GIF saved: %s", filepath.Base(filename)))
			})
		}()
	}

	saveButton := widget.NewButton("Save", func() {
		if !state.HasFrames() {
			return
		}
		if state.SourceFolder == "" {
			fynedialog.ShowError(fmt.Errorf("no source folder available"), myWindow)
			return
		}

		filename, err := randomGIFPath(state.SourceFolder)
		if err != nil {
			fynedialog.ShowError(err, myWindow)
			return
		}

		compileAndSaveGIF(filename)
	})

	saveAsButton := widget.NewButton("Save As", func() {
		if !state.HasFrames() {
			return
		}

		defaultPath := state.SourceFolder
		if defaultPath != "" {
			defaultPath = filepath.Join(defaultPath, randomGIFName())
		}

		options := []zenity.Option{
			zenity.Title("Save wigglegram GIF"),
			zenity.FileFilters{
				{Name: "GIF image", Patterns: []string{"*.gif"}, CaseFold: true},
			},
			zenity.ConfirmOverwrite(),
		}
		if defaultPath != "" {
			options = append(options, zenity.Filename(defaultPath))
		}

		filename, err := zenity.SelectFileSave(options...)
		if err != nil {
			if err != zenity.ErrCanceled {
				fynedialog.ShowError(err, myWindow)
			}
			return
		}

		compileAndSaveGIF(filename)
	})

	animationSettings := container.NewVBox(
		widget.NewLabel("Speed"),
		speedSelect,
		widget.NewLabel("GIF Repetitions"),
		loopCountSelect,
		widget.NewLabel("Export Quality"),
		exportQualitySelect,
		bounceCheck,
		widget.NewLabel("Pause At Ends"),
		pauseSelect,
	)

	animationAccordion := widget.NewAccordion(
		widget.NewAccordionItem("Animation Settings", animationSettings),
	)

	go func() {
		step := 0
		for {
			if state.IsPlayingPreview && state.HasFrames() {
				loopOrder := processor.BuildLoopOrder(state.FrameOrder, state.BounceMode, state.PauseFrames)
				if len(loopOrder) > 0 {
					frameIdx := loopOrder[step%len(loopOrder)]
					previewFrame := ui.GeneratePreviewFrame(state.RawImages, frameIdx, state.Offsets,
						state.CropMinX, state.CropMinY, state.CropMaxX, state.CropMaxY)

					fyne.Do(func() {
						previewImageRef.Image = previewFrame
						previewImageRef.Refresh()
					})
					step++
				}
			}
			time.Sleep(time.Duration(state.PreviewFrameDelay) * time.Millisecond)
		}
	}()

	controls := container.NewVBox(
		widget.NewLabel("Wigglegram Maker"),
		loadButton,
		widget.NewSeparator(),
		lockTransparencyCheck,
		maxSafeCropButton,
		widget.NewLabel("Tip: Use arrow keys to align the active frame. [R] resets the crop."),
		widget.NewSeparator(),
		widget.NewButton("Toggle Live Animation Preview", func() { state.IsPlayingPreview = !state.IsPlayingPreview }),
		container.NewCenter(previewImageRef),
		widget.NewSeparator(),
		widget.NewLabel("Animation Settings"),
		animationAccordion,
		widget.NewSeparator(),
		container.NewGridWithColumns(2, saveButton, saveAsButton),
		statusLabel,
	)

	workspaceCanvas := container.NewWithoutLayout()
	workspaceCanvas.Add(imageWorkspace)
	workspaceCanvas.Add(interactionLayer)

	thumbnailScroller := container.NewHScroll(thumbnailStrip)
	thumbnailScroller.SetMinSize(fyne.NewSize(0, 155))

	workspace := container.NewBorder(
		nil,
		thumbnailScroller,
		nil,
		nil,
		workspaceCanvas,
	)

	split := container.NewHSplit(controls, workspace)
	split.SetOffset(0.3)

	myWindow.SetContent(split)
	myWindow.ShowAndRun()
}

func randomGIFPath(folder string) (string, error) {
	for i := 0; i < 30; i++ {
		path := filepath.Join(folder, randomGIFName())
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path, nil
		}
	}

	return "", fmt.Errorf("could not find an unused random filename")
}

func randomGIFName() string {
	adjectives := []string{
		"turbo",
		"cosmic",
		"spicy",
		"velvet",
		"neon",
		"crystal",
		"wobbly",
		"golden",
		"midnight",
		"electric",
		"lucky",
		"rapid",
		"glitter",
		"moonlit",
		"fuzzy",
		"atomic",
	}
	nouns := []string{
		"pancakes",
		"sloths",
		"comets",
		"waffles",
		"rockets",
		"sunbeams",
		"spinners",
		"fireflies",
		"popcorn",
		"highlights",
		"skylines",
		"orbiters",
		"wiggles",
		"sprinkles",
		"drifters",
		"snapshots",
	}

	return fmt.Sprintf("%s-%s.gif", randomWord(adjectives), randomWord(nouns))
}

func randomWord(words []string) string {
	if len(words) == 0 {
		return "wiggle"
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(words))))
	if err != nil {
		return words[time.Now().UnixNano()%int64(len(words))]
	}

	return words[n.Int64()]
}
