package main

import (
	"os"
	"log"

    "github.com/alexanderi96/libcamera-tray/ui"
	  "gioui.org/app"

  "gioui.org/unit"
)

const (
    appName string = "libcamera-tray"
    //appDesc string = "libcamera-apps wrapper wreitten in go"
)

func main() {
  go func() {
    // create new window
    w := app.NewWindow(
      app.Title(appName),
      app.Size(unit.Dp(228), unit.Dp(423)),
    )
    if err := ui.Draw(w); err != nil {
      log.Fatal(err)
    }

    log.Println("Exiting.")
    os.Exit(0)
  }()
  app.Main()
}