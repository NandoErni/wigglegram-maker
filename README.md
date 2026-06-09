# Wigglegram Maker

A small desktop app for making Nishika N8000-style wigglegrams from photo frame sequences.

## Features

- Select the exact `.jpg`, `.jpeg`, or `.png` frames you want to import
- Align frames by dragging the active image
- Adjust the crop box with corner handles
- Fit the crop to the largest shared image area after alignment
- Reorder frames with the scrollable thumbnail strip
- Preview the animation and export an animated GIF
- Choose export quality to trade image size for smaller GIF files
- Use the drag loupe for more precise frame alignment
- Save directly back to the source folder with a fun random filename
- Use Save As for a custom output path
- Uses the operating system's native file pickers

## Requirements

- Go 1.26 or newer
- Fyne desktop build dependencies for your platform
- Linux: install `zenity`, `matedialog`, or `qarma` for native file dialogs
- `make` if you want to use the Makefile targets

## Build

```sh
make build
```

The binary is written to `bin/wigglegram-maker` or `bin/wigglegram-maker.exe` on Windows.
On Windows, this target builds the app as a GUI program so it does not open a terminal window.

For a packaged Windows executable with the file icon embedded, install the Fyne CLI and run:

```sh
make package-windows
```

That writes the packaged executable to the project root.

## Run

```sh
make run
```

You can also run it directly with:

```sh
go run .
```

## Usage

Select the frames you want to import. The app loads the selected files in sorted filename order, so naming them in sequence is recommended.

Use the thumbnail buttons to choose the active frame, then drag in the canvas to align it. A magnified loupe appears while dragging to help with fine alignment. Drag a crop corner to resize the crop box, use Max Safe Crop to fit the largest area covered by every shifted frame, right-click to move the reference point, choose an export quality, and save the GIF.

`Save` writes to the source image folder with a random playful filename. `Save As` lets you choose a custom path and starts in the source folder when available.

## Platform Notes

The app is built with Fyne and can run on Windows, macOS, and Linux. You need to build it on each target platform or use the Fyne packaging tools for that platform. The native folder picker uses Windows dialogs on Windows, Cocoa dialogs on macOS, and GTK dialogs on Linux.
