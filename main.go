package main

import (
	"log"
	"os"

	"github.com/alexanderi96/libcamera-tray/config"
	"github.com/alexanderi96/libcamera-tray/ui"
	"github.com/alexanderi96/libcamera-tray/camera"
	"github.com/alexanderi96/libcamera-tray/utils"

	"gioui.org/app"
	"gioui.org/unit"
)

func main() {
	go func() {
		// create new window
		w := app.NewWindow(
			app.Title(config.Properties.App.Title),
			app.Size(
				unit.Dp(config.Properties.App.Width),
				unit.Dp(config.Properties.App.Height),
			),
			app.MinSize(
				unit.Dp(800),  // Minimum width to accommodate preview
				unit.Dp(480),  // Minimum height
			),
		)
		if err := ui.Draw(w); err != nil {
			log.Fatal(err)
		}

		if utils.IsItRunning("libcamera-vid") {
			camera.StopPreview()
		}
		log.Println("Exiting.")
		os.Exit(0)
	}()
	app.Main()
}
