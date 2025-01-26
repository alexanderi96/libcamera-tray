package ui

import (
	"image"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"sort"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type Gallery struct {
	images     []string
	thumbnails map[string]image.Image
	selected   int
	gridMode   bool
	list       widget.List
	backBtn    widget.Clickable
	gridBtn    widget.Clickable
	maxThumbnails int // Maximum number of thumbnails to keep in memory
}

const defaultMaxThumbnails = 50 // Adjust based on typical memory constraints

func NewGallery() *Gallery {
	return &Gallery{
		thumbnails: make(map[string]image.Image),
		gridMode:   true,
		maxThumbnails: defaultMaxThumbnails,
		list: widget.List{
			List: layout.List{
				Axis:        layout.Vertical,
				ScrollToEnd: false,
			},
		},
	}
}

// Cleanup releases memory used by thumbnails
func (g *Gallery) Cleanup() {
	g.thumbnails = make(map[string]image.Image)
}

func (g *Gallery) LoadImages() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	baseDir := filepath.Join(homeDir, "Pictures", "libcamera-tray")
	g.images = []string{}

	// Walk through all subdirectories
	err = filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".jpg" {
			g.images = append(g.images, path)
		}
		return nil
	})

	// Sort images by modification time (newest first)
	sort.Slice(g.images, func(i, j int) bool {
		iInfo, _ := os.Stat(g.images[i])
		jInfo, _ := os.Stat(g.images[j])
		return iInfo.ModTime().After(jInfo.ModTime())
	})

	return err
}

func (g *Gallery) loadThumbnail(path string) (image.Image, error) {
	if thumb, ok := g.thumbnails[path]; ok {
		return thumb, nil
	}

	// Clean up old thumbnails if we've exceeded the limit
	if len(g.thumbnails) >= g.maxThumbnails {
		g.Cleanup()
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	g.thumbnails[path] = img
	return img, nil
}

func (g *Gallery) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if len(g.images) == 0 {
		// Show message when no images
		return layout.Center.Layout(gtx, func(gtx C) D {
			return material.Body1(th, "No images found").Layout(gtx)
		})
	}

	// Top bar with navigation buttons
	flex := layout.Flex{Axis: layout.Vertical}
	return flex.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					btn := material.Button(th, &g.backBtn, "Back")
					return btn.Layout(gtx)
				}),
				layout.Flexed(1, func(gtx C) D {
					return layout.E.Layout(gtx, func(gtx C) D {
						icon := "⊞"
						if g.gridMode {
							icon = "⊟"
						}
						btn := material.Button(th, &g.gridBtn, icon)
						return btn.Layout(gtx)
					})
				}),
			)
		}),
		layout.Flexed(1, func(gtx C) D {
			if g.gridMode {
				return g.layoutGrid(gtx, th)
			}
			return g.layoutSingle(gtx, th)
		}),
	)
}

func (g *Gallery) layoutGrid(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Calculate thumbnail size based on window width
	width := float32(gtx.Constraints.Max.X)
	columns := int(width / 150) // Approximate thumbnail width
	if columns < 1 {
		columns = 1
	}
	thumbSize := int(width) / columns

	return g.list.List.Layout(gtx, (len(g.images)+columns-1)/columns, func(gtx C, i int) D {
		// Create a row of thumbnails
		return layout.Flex{}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				flex := layout.Flex{}
				var children []layout.FlexChild

				// Add thumbnails for this row
				for j := 0; j < columns && i*columns+j < len(g.images); j++ {
					idx := i*columns + j
					path := g.images[idx]
					children = append(children, layout.Rigid(func(gtx C) D {
						return g.layoutThumbnail(gtx, th, path, thumbSize, idx)
					}))
				}

				return flex.Layout(gtx, children...)
			}),
		)
	})
}

func (g *Gallery) layoutThumbnail(gtx layout.Context, th *material.Theme, path string, size int, idx int) layout.Dimensions {
	img, err := g.loadThumbnail(path)
	if err != nil {
		return material.Body1(th, "Error loading image").Layout(gtx)
	}

	gtx.Constraints.Min = image.Point{X: size, Y: size}
	gtx.Constraints.Max = image.Point{X: size, Y: size}

	btn := &widget.Clickable{}
	if btn.Clicked() {
		g.selected = idx
		g.gridMode = false
	}

	return btn.Layout(gtx, func(gtx C) D {
		return layout.Stack{}.Layout(gtx,
			layout.Stacked(func(gtx C) D {
				return widget.Image{
					Src:      paint.NewImageOp(img),
					Fit:      widget.Cover,
					Position: layout.Center,
				}.Layout(gtx)
			}),
		)
	})
}

func (g *Gallery) layoutSingle(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if g.selected < 0 || g.selected >= len(g.images) {
		return layout.Dimensions{}
	}

	img, err := g.loadThumbnail(g.images[g.selected])
	if err != nil {
		return material.Body1(th, "Error loading image").Layout(gtx)
	}

	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx C) D {
			return widget.Image{
				Src:      paint.NewImageOp(img),
				Fit:      widget.Contain,
				Position: layout.Center,
			}.Layout(gtx)
		}),
	)
}
