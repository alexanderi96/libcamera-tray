package ui

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/alexanderi96/libcamera-tray/camera"
	"github.com/alexanderi96/libcamera-tray/config"
	"github.com/alexanderi96/libcamera-tray/utils"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"
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
	previewWindowID    string = ""
	showSettings      bool = false
	showGallery       bool = false

	gallery = NewGallery()

	// ops are the operations from the UI
	ops op.Ops

	// Buttons
	shotButton       widget.Clickable
	settingsButton   widget.Clickable
	galleryButton    widget.Clickable
	loadConfigButton widget.Clickable
	backButton       widget.Clickable

	infoTextField widget.Editor
)

var PreviewCheckbox = &widget.Bool{
	Value: false, // Disable preview by default
}

func getPreviewWindowID() (string, error) {
	time.Sleep(1000 * time.Millisecond)
	cmd := exec.Command("xdotool", "search", "--name", "Camera")
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output[:len(output)-1]), nil
}

func positionPreviewWindow() error {
	previewID, err := getPreviewWindowID()
	if err != nil {
		return err
	}

	mainX := config.Properties.App.X
	mainY := config.Properties.App.Y
	previewX := mainX + config.Properties.App.Width
	previewY := mainY

	cmd := exec.Command("xdotool", "windowmove", previewID,
		strconv.Itoa(previewX),
		strconv.Itoa(previewY))
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
	return cmd.Run()
}

func Draw(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	expl := explorer.NewExplorer(w)

	managePreview := func() {
		if PreviewCheckbox.Value {
			camera.StartPreview()
			go func() {
				if err := positionPreviewWindow(); err != nil {
					log.Printf("Error positioning preview window: %v", err)
				}
			}()
		} else {
			camera.StopPreview()
			previewWindowID = ""
		}
	}

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err

			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				
				if shotButton.Clicked() {
					camera.Shot()
				}

				if settingsButton.Clicked() {
					showSettings = true
				}

				if backButton.Clicked() {
					showSettings = false
				}

				if PreviewCheckbox.Changed() {
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
							camera.Params.LoadParamsMap(buf.Bytes())
							settings(gtx, th)
						})
					}()
				}

				if galleryButton.Clicked() {
					showGallery = true
					if err := gallery.LoadImages(); err != nil {
						log.Printf("Error loading gallery images: %v", err)
					}
				}

				if gallery.backBtn.Clicked() {
					showGallery = false
				}

				if gallery.gridBtn.Clicked() {
					gallery.gridMode = !gallery.gridMode
				}

				if showGallery {
					// Gallery view in fullscreen
					gtx.Constraints.Min = gtx.Constraints.Max // Make fullscreen
					gallery.Layout(gtx, th)
				} else if showSettings {
					// Settings view in fullscreen
					gtx.Constraints.Min = gtx.Constraints.Max // Make fullscreen
					layout.Flex{
						Axis:    layout.Vertical,
						Spacing: layout.SpaceStart,
					}.Layout(gtx,
						// Back button at top
						layout.Rigid(
							func(gtx C) D {
								return layout.Inset{Top: unit.Dp(8), Left: unit.Dp(8)}.Layout(gtx,
									material.Button(th, &backButton, "Back").Layout,
								)
							},
						),
						// Load config button
						layout.Rigid(
							func(gtx C) D {
								return layout.Inset{Top: unit.Dp(8), Left: unit.Dp(8)}.Layout(gtx,
									material.Button(th, &loadConfigButton, "Load Config").Layout,
								)
							},
						),
						// Settings list
						layout.Rigid(
							func(gtx C) D {
								gtx.Constraints.Min.X = gtx.Dp(unit.Dp(300)) // Minimum width for touch
								gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(600)) // More vertical space
								return material.List(th, list).Layout(gtx, len(OptionsList), func(gtx C, i int) D {
									return layout.UniformInset(unit.Dp(0)).Layout(gtx, OptionsList[i])
								})
							},
						),
					)
				} else {
					// Main view
					layout.Flex{
						Axis:    layout.Vertical,
						Spacing: layout.SpaceStart,
					}.Layout(gtx,
						// Preview checkbox at top
						layout.Rigid(
							func(gtx C) D {
								return layout.Inset{Top: unit.Dp(8), Left: unit.Dp(8)}.Layout(gtx,
									material.CheckBox(th, PreviewCheckbox, "Preview").Layout,
								)
							},
						),
						// Center shot button
						layout.Flexed(1,
							func(gtx C) D {
								return layout.Center.Layout(gtx,
									func(gtx C) D {
										btn := material.Button(th, &shotButton, "Take a Shot")
										btn.TextSize = unit.Sp(24)
										btn.Background = th.Palette.ContrastBg
										btn.Inset = layout.UniformInset(unit.Dp(32))
										return btn.Layout(gtx)
									},
								)
							},
						),
						// Bottom buttons
						layout.Rigid(
							func(gtx C) D {
								return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
									layout.Flexed(1, func(gtx C) D {
										return layout.E.Layout(gtx, func(gtx C) D {
											return layout.Inset{Bottom: unit.Dp(16), Right: unit.Dp(16)}.Layout(gtx,
												func(gtx C) D {
													galleryBtn := material.Button(th, &galleryButton, "🖼")
													galleryBtn.Background = th.Palette.Bg
													galleryBtn.Color = th.Palette.Fg
													galleryBtn.TextSize = unit.Sp(18)
													galleryBtn.Inset = layout.UniformInset(unit.Dp(8))
													return galleryBtn.Layout(gtx)
												},
											)
										})
									}),
									layout.Rigid(func(gtx C) D {
										return layout.E.Layout(gtx, func(gtx C) D {
											return layout.Inset{Bottom: unit.Dp(16), Right: unit.Dp(16)}.Layout(gtx,
												func(gtx C) D {
													btn := material.Button(th, &settingsButton, "⚙")
													btn.Background = th.Palette.Bg
													btn.Color = th.Palette.Fg
													btn.TextSize = unit.Sp(18)
													btn.Inset = layout.UniformInset(unit.Dp(8))
													return btn.Layout(gtx)
												},
											)
										})
									}),
								)
							},
						),
					)
				}

				e.Frame(gtx.Ops)

				if !windowPositioned {
					settings(gtx, th)
					w.Invalidate()
					moveWindow()
					windowPositioned = true
					managePreview()
				}
			}
		}
	}
}

func moveWindow() {
	cmd := exec.Command("xdotool",
		"search",
		"--name",
		config.Properties.App.Title,
		"windowmove",
		strconv.Itoa(config.Properties.App.X),
		strconv.Itoa(config.Properties.App.Y),
	)
	utils.Exec(cmd, true)
}
