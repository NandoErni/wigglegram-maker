package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed assets/icon.png
var iconBytes []byte

var appIcon = fyne.NewStaticResource("icon.png", iconBytes)
