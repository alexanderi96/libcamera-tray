package ui

import (
	"bytes"
	"fmt"
	"log"
	"image"
	"time"
	"sync"

	"github.com/alexanderi96/libcamera-tray/camera"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"
	"gioui.org/text"
	"image/color"
)

type C = layout.Context
type D = layout.Dimensions

// UI state
var (
	showSettings  bool
	showGallery   bool
	isFullscreen  bool
	previewFrames chan image.Image
	currentFrame  image.Image
	statsText     string
	gallery       = NewGallery()
	ops           op.Ops
	previewMutex  sync.Mutex
)

// UI controls
var (
	shotButton       widget.Clickable
	settingsButton   widget.Clickable
	galleryButton    widget.Clickable
	loadConfigButton widget.Clickable
	backButton       widget.Clickable
)

// layoutPreview handles both fullscreen and normal preview layouts
func layoutPreview(gtx C, th *material.Theme, isFullscreen bool) D {
	if isFullscreen {
		gtx.Constraints.Min = gtx.Constraints.Max
	}

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			// Black background
			paint.Fill(gtx.Ops, color.NRGBA{A: 255})
			
			previewMutex.Lock()
			frame := currentFrame
			previewMutex.Unlock()
			
			if frame != nil {
				return layoutFrame(gtx, th, frame)
			}
			return layoutPlaceholder(gtx, th)
		}),
		layout.Stacked(layoutStats(th)),
	)
}

// layoutFrame handles the camera frame display
func layoutFrame(gtx C, th *material.Theme, frame image.Image) D {
	imageOp := paint.NewImageOp(frame)
	
	pointer := &widget.Clickable{}
	if pointer.Clicked() {
		isFullscreen = !isFullscreen
	}
	
	return pointer.Layout(gtx, func(gtx C) D {
		return layout.Center.Layout(gtx, func(gtx C) D {
			return widget.Image{
				Src:   imageOp,
				Fit:   widget.Fill,
			}.Layout(gtx)
		})
	})
}

// layoutPlaceholder shows waiting message when no stream is available
func layoutPlaceholder(gtx C, th *material.Theme) D {
	return layout.Center.Layout(gtx, func(gtx C) D {
		label := material.H6(th, "Waiting for camera stream...")
		label.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		label.Alignment = text.Middle
		return label.Layout(gtx)
	})
}

// layoutStats displays frame statistics
func layoutStats(th *material.Theme) layout.Widget {
	return func(gtx C) D {
		return layout.Inset{
			Top:  unit.Dp(4),
			Left: unit.Dp(4),
		}.Layout(gtx, func(gtx C) D {
			label := material.Label(th, unit.Sp(12), statsText)
			label.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			return label.Layout(gtx)
		})
	}
}

// layoutControls handles the right-side control panel
func layoutControls(gtx C, th *material.Theme) D {
	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	return layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   layout.SpaceBetween,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Right: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
						btn := material.Button(th, &galleryButton, "🖼")
						btn.Background = th.Palette.Bg
						btn.Color = th.Palette.Fg
						btn.TextSize = unit.Sp(18)
						btn.Inset = layout.UniformInset(unit.Dp(8))
						return btn.Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Right: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
						btn := material.Button(th, &settingsButton, "⚙")
						btn.Background = th.Palette.Bg
						btn.Color = th.Palette.Fg
						btn.TextSize = unit.Sp(18)
						btn.Inset = layout.UniformInset(unit.Dp(8))
						return btn.Layout(gtx)
					})
				}),
			)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Center.Layout(gtx, func(gtx C) D {
				return layout.Stack{}.Layout(gtx,
					layout.Expanded(func(gtx C) D {
						btn := material.Button(th, &shotButton, "Take a Shot")
						btn.TextSize = unit.Sp(24)
						btn.Background = th.Palette.ContrastBg
						btn.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
						btn.Inset = layout.UniformInset(unit.Dp(48))
						return btn.Layout(gtx)
					}),
					layout.Expanded(func(gtx C) D {
						defer clip.Stroke{
							Path:  clip.Rect{Max: gtx.Constraints.Min}.Path(),
							Width: float32(gtx.Dp(1)),
						}.Op().Push(gtx.Ops).Pop()
						paint.ColorOp{Color: color.NRGBA{R: 255, G: 255, B: 255, A: 255}}.Add(gtx.Ops)
						paint.PaintOp{}.Add(gtx.Ops)
						return D{Size: gtx.Constraints.Min}
					}),
				)
			})
		}),
		layout.Rigid(func(gtx C) D {
			size := gtx.Dp(unit.Dp(72))
			return D{Size: image.Point{X: size, Y: size}}
		}),
	)
}

// layoutSettings displays the settings view
func layoutSettings(gtx C, th *material.Theme) D {
	gtx.Constraints.Min = gtx.Constraints.Max
	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceStart,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Top: unit.Dp(8), Left: unit.Dp(8)}.Layout(gtx,
				material.Button(th, &backButton, "Back").Layout,
			)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Top: unit.Dp(8), Left: unit.Dp(8)}.Layout(gtx,
				material.Button(th, &loadConfigButton, "Load Config").Layout,
			)
		}),
		layout.Rigid(func(gtx C) D {
			gtx.Constraints.Min.X = gtx.Dp(unit.Dp(300))
			gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(600))
			return material.List(th, list).Layout(gtx, len(OptionsList), func(gtx C, i int) D {
				return layout.UniformInset(unit.Dp(0)).Layout(gtx, OptionsList[i])
			})
		}),
	)
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

// handleEvents processes all UI events and button clicks
func handleEvents(gtx C, th *material.Theme, expl *explorer.Explorer, w *app.Window) {
	if shotButton.Clicked() {
		log.Println("Shot button clicked")
		
		previewMutex.Lock()
		// Reset current frame and take shot
		currentFrame = nil
		previewFrames = nil // Clear the channel reference
		previewMutex.Unlock()
		
		w.Invalidate() // Force UI update to show placeholder
		
		// Take shot and wait for completion
		done := camera.Shot()
		
		// Wait for shot to complete before restarting preview
		go func() {
			log.Println("Waiting for shot completion")
			<-done // Wait for shot completion signal
			log.Println("Shot completed, restarting preview")
			
			previewMutex.Lock()
			// Get new preview channel and update UI
			previewFrames = camera.StartPreview()
			previewMutex.Unlock()
			
			// Force UI update
			w.Invalidate()
			log.Println("Preview restarted")
		}()
	}
	if settingsButton.Clicked() {
		showSettings = true
	}
	if backButton.Clicked() {
		showSettings = false
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
		gallery.Cleanup()
	}
	if gallery.gridBtn.Clicked() {
		gallery.gridMode = !gallery.gridMode
	}
}

// layoutMainView handles the main camera preview with controls
func layoutMainView(gtx C, th *material.Theme) D {
	if isFullscreen {
		return layoutPreview(gtx, th, true)
	}
	
	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceBetween,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			return layout.Inset{
				Top: unit.Dp(8),
				Bottom: unit.Dp(8),
				Left: unit.Dp(8),
			}.Layout(gtx, func(gtx C) D {
				return layoutPreview(gtx, th, false)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return layoutControls(gtx, th)
		}),
	)
}

func Draw(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	th.Palette.Bg = color.NRGBA{R: 48, G: 48, B: 48, A: 255}
	th.Palette.ContrastBg = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	expl := explorer.NewExplorer(w)
	
	settings(layout.NewContext(&ops, system.FrameEvent{}), th)
	previewFrames = camera.StartPreview()
	
	// Start frame handling goroutine
	go func() {
		ticker := time.NewTicker(time.Second / 30) // Match camera framerate
		defer ticker.Stop()
		
		statsTicker := time.NewTicker(time.Second)
		defer statsTicker.Stop()
		
		for {
			select {
			case <-ticker.C:
				previewMutex.Lock()
				if previewFrames == nil {
					currentFrame = nil // Clear current frame when preview is stopped
					previewMutex.Unlock()
					w.Invalidate() // Force UI update to show placeholder
					continue
				}
				
				select {
				case frame, ok := <-previewFrames:
					if !ok {
						log.Println("Preview channel closed")
						previewFrames = nil
						currentFrame = nil
						previewMutex.Unlock()
						w.Invalidate() // Force UI update to show placeholder
						continue
					}
					currentFrame = frame
					previewMutex.Unlock()
					w.Invalidate()
				default:
					previewMutex.Unlock()
					// No new frame available, continue
				}
				
			case <-statsTicker.C:
				_, dropped, errors, fps := camera.GetFrameStats()
				statsText = fmt.Sprintf("FPS: %.1f | Dropped: %d | Errors: %d", fps, dropped, errors)
				w.Invalidate()
			}
		}
	}()

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err

			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				handleEvents(gtx, th, expl, w)

				// Layout current view
				if showGallery {
					gtx.Constraints.Min = gtx.Constraints.Max
					if len(gallery.images) == 0 {
						if err := gallery.LoadImages(); err != nil {
							log.Printf("Error loading gallery images: %v", err)
						}
					}
					gallery.Layout(gtx, th)
				} else if showSettings {
					layoutSettings(gtx, th)
				} else {
					layoutMainView(gtx, th)
				}

				e.Frame(gtx.Ops)
			}
		}
	}
}
