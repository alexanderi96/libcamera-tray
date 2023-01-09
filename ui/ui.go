package ui

import (
	//"fmt"
	"bytes"
	"log"
	"os/exec"
	"strconv"
	//"strings"
	//"image/color"

	"github.com/alexanderi96/libcamera-tray/camera"
	"github.com/alexanderi96/libcamera-tray/config"
	"github.com/alexanderi96/libcamera-tray/utils"
	//"github.com/alexanderi96/libcamera-tray/types"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"
	//"gioui.org/text"
)

type C = layout.Context
type D = layout.Dimensions

type Point struct {
	X, Y float32
}

var (
	windowPositioned   bool = false
	previewing         bool = false
	customConfigLoaded bool = false

	// ops are the operations from the UI
	ops op.Ops

	// shotButton is a clickable widget
	shotButton       widget.Clickable
	previewButton    widget.Clickable
	loadConfigButton widget.Clickable

	previewCheckbox = &widget.Bool{
		Value: config.Properties.Preview.Enabled,
	}

	infoTextField widget.Editor

	optionsList = []layout.Widget{}

	list = &widget.List{
		List: layout.List{
			Axis: layout.Vertical,
		},
	}
)

func Draw(w *app.Window) error {
	// th defines the material design style
	th := material.NewTheme(gofont.Collection())

	expl := explorer.NewExplorer(w)

	managePreview := func() {
		if previewCheckbox.Value {
			camera.StartPreview()
		} else {
			camera.StopPreview()
		}
	}

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

				if previewCheckbox.Changed() {
					managePreview()
				}

				if loadConfigButton.Clicked() {
					go func() {
						log.Println("Loading config file.")
						file, err := expl.ChooseFile("lctp")
						if err != nil {
							log.Println(err)
							return
						}

						buf := new(bytes.Buffer)
						buf.ReadFrom(file)

						camera.StopPreviewAndReload(func() {
							log.Println("Settings loaded configs.")
							camera.CustomParams.LoadParamsMap(buf.Bytes())
						})

					}()
				}

				layout.Flex{
					// Vertical alignment, from top to bottom
					Axis:    layout.Vertical,
					Spacing: layout.SpaceStart,
				}.Layout(gtx,
					layout.Rigid(
						func(gtx C) D {
							return material.List(th, list).Layout(gtx, len(optionsList), func(gtx C, i int) D {
								return layout.UniformInset(unit.Dp(16)).Layout(gtx, optionsList[i])
							})
						},
					),
					layout.Rigid(
						material.CheckBox(th, previewCheckbox, "Preview").Layout,
					),
					layout.Rigid(
						func(gtx C) D {
							in := layout.UniformInset(unit.Dp(8))
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									return in.Layout(gtx, material.Button(th, &shotButton, "Take a Shot").Layout)
								}),
								layout.Rigid(func(gtx C) D {
									return in.Layout(gtx, material.Button(th, &loadConfigButton, "Load Conf").Layout)
								}),
							)
						},
					),
				)
				e.Frame(gtx.Ops)

				//ugly workaroung in order to position the app at startup
				if !windowPositioned {
					moveWindow()
					windowPositioned = true

					managePreview()
				}
			}

		}
	}
	return nil
}

func moveWindow() {
	cmd := exec.Command("xdotool",
		"search",
		"--class",
		config.Properties.App.Title,
		"windowmove",
		strconv.Itoa(config.Properties.App.X),
		strconv.Itoa(config.Properties.App.Y),
	)
	utils.Exec(cmd, true)
}

func generateOptionsList(gtx layout.Context, th *material.Theme) {
	optionsList = []layout.Widget{}

	for _, opt := range camera.CustomParams.GetKeysList() {
		optionsList = append(optionsList,
			func(gtx C) D {
				return material.Editor(th, new(widget.Editor), opt).Layout(gtx)
			},
		)
	}
}
