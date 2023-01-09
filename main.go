package main

import (
	"os"
	"log"

  "github.com/alexanderi96/libcamera-tray/config"
  "github.com/alexanderi96/libcamera-tray/ui"
	
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
    )
    if err := ui.Draw(w); err != nil {
      log.Fatal(err)
    }

    log.Println("Exiting.")
    os.Exit(0)
  }()
  app.Main()
}