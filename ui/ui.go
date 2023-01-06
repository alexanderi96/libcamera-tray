package ui

import (
  "github.com/alexanderi96/libcamera-tray/camera"
  "github.com/alexanderi96/libcamera-tray/utils"


  "gioui.org/app"
  "gioui.org/font/gofont"
  "gioui.org/io/system"
  "gioui.org/layout"
  "gioui.org/op"
    "gioui.org/unit"
  "gioui.org/widget"
  "gioui.org/widget/material"
)

type C = layout.Context
type D = layout.Dimensions

type Point struct {
    X, Y float32
}

var (
  previewing bool = false
  customConfigLoaded bool = false
)

func Draw(w *app.Window) error {
   // ops are the operations from the UI
   var ops op.Ops

   // shotButton is a clickable widget
   var shotButton widget.Clickable
   var previewButton widget.Clickable
   var loadConfigButton widget.Clickable

   // th defines the material design style
   th := material.NewTheme(gofont.Collection())

   // listen for events in the window.
  for {

    // detect what type of event
    select {

    // this is sent when the application should re-render.
    case e := <-w.Events():
      switch e := e.(type) {
      case system.DestroyEvent:
        return e.Err

      case system.FrameEvent:
        gtx := layout.NewContext(&ops, e)
        // Let's try out the flexbox layout concept:
        if shotButton.Clicked() {
            camera.Shot()
        }

        if previewButton.Clicked() {
            previewing = camera.TogglePreview()
        }

        if loadConfigButton.Clicked() {
            camera.Params.LoadParamsMap(utils.OpenFile("/home/stego/Documents/libcamera-tray/custom.lctp"))
        }

        layout.Flex{
            // Vertical alignment, from top to bottom
            Axis: layout.Vertical,
        }.Layout(gtx,
          layout.Rigid(
                func(gtx C) D {
                    // ONE: First define margins around the button using layout.Inset ...
                    margins := layout.Inset{
                      Top:    unit.Dp(10),
                      Bottom: unit.Dp(10),
                      Right:  unit.Dp(15),
                      Left:   unit.Dp(15),
                    }

                    // TWO: ... then we lay out those margins ...
                    return margins.Layout(gtx,

                        // THREE: ... and finally within the margins, we define and lay out the button
                        func(gtx C) D {
                            btn := material.Button(th, &loadConfigButton, "Load Custom Config")
                            return btn.Layout(gtx)
                        },

                    )

                },
            ),
          layout.Rigid(
                func(gtx C) D {
                    // ONE: First define margins around the button using layout.Inset ...
                    margins := layout.Inset{
                      Top:    unit.Dp(10),
                      Bottom: unit.Dp(10),
                      Right:  unit.Dp(15),
                      Left:   unit.Dp(15),
                    }

                    // TWO: ... then we lay out those margins ...
                    return margins.Layout(gtx,

                        // THREE: ... and finally within the margins, we define and lay out the button
                        func(gtx C) D {
                            var text string
                            if !previewing {
                              text = "Start preview"
                            } else {
                              text = "Stop preview"
                            }
                            btn := material.Button(th, &previewButton, text)
                            return btn.Layout(gtx)
                        },

                    )

                },
            ),
            layout.Rigid(
                func(gtx C) D {
                    // ONE: First define margins around the button using layout.Inset ...
                    margins := layout.Inset{
                    	Top:    unit.Dp(10),
                    	Bottom: unit.Dp(10),
                    	Right:  unit.Dp(15),
                    	Left:   unit.Dp(15),
                    }

                    // TWO: ... then we lay out those margins ...
                    return margins.Layout(gtx,

                        // THREE: ... and finally within the margins, we define and lay out the button
                        func(gtx C) D {
                            btn := material.Button(th, &shotButton, "Shot")
                            return btn.Layout(gtx)
                        },

                    )

                },
            ),
        )
        e.Frame(gtx.Ops)
      }

    }
  }
  return nil    
}


