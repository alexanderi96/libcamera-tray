package ui

import (
	"image"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"golang.org/x/image/draw"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type Gallery struct {
	images        []string
	thumbnails    map[string]image.Image
	selected      int
	gridMode      bool
	list          widget.List
	backBtn       widget.Clickable
	gridBtn       widget.Clickable
	maxThumbnails int
	loadingMutex  sync.RWMutex
	loadQueue     chan string
	workerCount   int
}

const (
	defaultMaxThumbnails = 50
	thumbnailSize        = 300 // Max dimension for thumbnails
	workerCount         = 4   // Number of concurrent image loading workers
)

func NewGallery() *Gallery {
	g := &Gallery{
		thumbnails:    make(map[string]image.Image),
		gridMode:      true,
		maxThumbnails: defaultMaxThumbnails,
		list: widget.List{
			List: layout.List{
				Axis:        layout.Vertical,
				ScrollToEnd: false,
			},
		},
		loadQueue:   make(chan string, 100),
		workerCount: workerCount,
	}
	
	// Start worker pool
	for i := 0; i < g.workerCount; i++ {
		go g.thumbnailWorker()
	}
	
	return g
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

func (g *Gallery) thumbnailWorker() {
	for path := range g.loadQueue {
		g.loadingMutex.RLock()
		_, exists := g.thumbnails[path]
		g.loadingMutex.RUnlock()
		if exists {
			continue
		}

		file, err := os.Open(path)
		if err != nil {
			continue
		}

		img, _, err := image.Decode(file)
		file.Close()
		if err != nil {
			continue
		}

		// Resize image to thumbnail size
		bounds := img.Bounds()
		width, height := bounds.Dx(), bounds.Dy()
		scale := float64(thumbnailSize) / float64(max(width, height))
		if scale < 1 {
			width = int(float64(width) * scale)
			height = int(float64(height) * scale)
			thumb := image.NewRGBA(image.Rect(0, 0, width, height))
			draw.BiLinear.Scale(thumb, thumb.Bounds(), img, bounds, draw.Over, nil)
			img = thumb
		}

		g.loadingMutex.Lock()
		if len(g.thumbnails) >= g.maxThumbnails {
			g.Cleanup()
		}
		g.thumbnails[path] = img
		g.loadingMutex.Unlock()
	}
}

func (g *Gallery) loadThumbnail(path string) (image.Image, error) {
	g.loadingMutex.RLock()
	thumb, ok := g.thumbnails[path]
	g.loadingMutex.RUnlock()
	
	if ok {
		return thumb, nil
	}

	// Queue image for loading if not already loaded
	select {
	case g.loadQueue <- path:
	default:
		// Queue is full, skip for now
	}

	return nil, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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
	img, _ := g.loadThumbnail(path)

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
				if img == nil {
					// Show loading placeholder while image is being loaded
					return material.Body1(th, "Loading...").Layout(gtx)
				}
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

	img, _ := g.loadThumbnail(g.images[g.selected])
	if img == nil {
		return layout.Center.Layout(gtx, func(gtx C) D {
			return material.Body1(th, "Loading...").Layout(gtx)
		})
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
