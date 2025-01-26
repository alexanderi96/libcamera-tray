package ui

import (
	"strconv"
	"strings"

	"github.com/alexanderi96/libcamera-tray/camera"
	"github.com/alexanderi96/libcamera-tray/types"
	
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

var (
	OptionsList = []layout.Widget{}
	list = &widget.List{
		List: layout.List{
			Axis: layout.Vertical,
		},
	}

	// Store sliders for numeric values
	sliders = make(map[string]*widget.Float)

	// Track expanded state of categories
	basicExpanded = &widget.Bool{Value: true}
	advancedExpanded = &widget.Bool{Value: false}
)

// Basic parameters that most users will need
var basicParams = map[string]bool{
	"brightness": true,
	"contrast": true,
	"saturation": true,
	"sharpness": true,
	"ev": true,
	"preview": true,
	"output": true,
}

func isNumeric(value string) bool {
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}

func settings(gtx C, th *material.Theme) {
	OptionsList = []layout.Widget{
		// Basic Settings category
		func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				// Category header with expand/collapse button
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: unit.Dp(16), Bottom: unit.Dp(8), Left: unit.Dp(16), Right: unit.Dp(16)}.Layout(gtx,
						func(gtx C) D {
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(material.CheckBox(th, basicExpanded, "").Layout),
								layout.Rigid(material.H6(th, "Basic Settings").Layout),
							)
						},
					)
				}),
				// Basic settings content
				layout.Rigid(func(gtx C) D {
					if !basicExpanded.Value {
						return D{}
					}
					return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							var widgets []layout.FlexChild
							for key, opt := range camera.Params {
								if basicParams[key] {
									widget := createSettingRow(gtx, th, key, opt)
									widgets = append(widgets, layout.Rigid(widget))
								}
							}
							return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceSides}.Layout(gtx, widgets...)
						}),
					)
				}),
			)
		},
		// Advanced Settings category
		func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				// Category header with expand/collapse button
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Top: unit.Dp(24), Bottom: unit.Dp(8), Left: unit.Dp(16), Right: unit.Dp(16)}.Layout(gtx,
						func(gtx C) D {
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(material.CheckBox(th, advancedExpanded, "").Layout),
								layout.Rigid(material.H6(th, "Advanced Settings").Layout),
							)
						},
					)
				}),
				// Advanced settings content
				layout.Rigid(func(gtx C) D {
					if !advancedExpanded.Value {
						return D{}
					}
					return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							var widgets []layout.FlexChild
							for key, opt := range camera.Params {
								if !basicParams[key] {
									widget := createSettingRow(gtx, th, key, opt)
									widgets = append(widgets, layout.Rigid(widget))
								}
							}
							return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceSides}.Layout(gtx, widgets...)
						}),
					)
				}),
			)
		},
	}
}

func createSettingRow(gtx C, th *material.Theme, key string, opt types.Parameter) layout.Widget {
	checkbox := &widget.Bool{
		Value: opt.Enabled,
	}

	if checkbox.Changed() {
		opt.Toggle()
		camera.Params[key] = opt
		// Restart preview to apply changes if enabled
		if PreviewCheckbox.Value {
			camera.StopPreviewAndReload(func() {
				camera.StartPreview()
			})
		}
	}

	// Initialize or get slider for numeric values
	if _, exists := sliders[key]; !exists && isNumeric(opt.Value) {
		sliders[key] = &widget.Float{
			Value: parseFloatValue(opt.Value),
		}
	}

	return getSettingsRow(gtx, th, checkbox, key, opt)
}

func parseFloatValue(value string) float32 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(value), 32)
	return float32(v)
}

func getSettingsRow(gtx C, th *material.Theme, checkbox *widget.Bool, key string, opt types.Parameter) layout.Widget {
	return func(gtx C) D {
		// Increase padding for better touch targets
		in := layout.UniformInset(unit.Dp(16))
		
		return in.Layout(gtx, func(gtx C) D {
			return layout.Flex{
				Axis: layout.Horizontal,
				Alignment: layout.Middle,
				Spacing: layout.SpaceBetween,
			}.Layout(gtx,
				// Checkbox with label in a fixed-width container
				layout.Rigid(func(gtx C) D {
					gtx.Constraints.Min.X = gtx.Dp(unit.Dp(120)) // Fixed width for labels
					return material.CheckBox(th, checkbox, key).Layout(gtx)
				}),
				
				// Value control (slider for numeric, editor for text)
				layout.Flexed(1, func(gtx C) D {
					if slider, ok := sliders[key]; ok && isNumeric(opt.Value) {
						// Slider for numeric values with increased touch area
						if slider.Changed() {
							camera.Params[key] = types.Parameter{
								Command: opt.Command,
								Value: strconv.FormatFloat(float64(slider.Value), 'f', 2, 32),
								Enabled: opt.Enabled,
								StillSpecific: opt.StillSpecific,
								Description: opt.Description,
							}
							// Restart preview to apply changes if enabled
							if PreviewCheckbox.Value {
								camera.StopPreviewAndReload(func() {
									camera.StartPreview()
								})
							}
						}
						
						sliderStyle := material.Slider(th, slider, getMinValue(key), getMaxValue(key))
						sliderStyle.Color = th.Palette.ContrastBg
						return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, sliderStyle.Layout)
					} else {
						// Text editor for non-numeric values
						editor := &widget.Editor{SingleLine: true}
						editor.SetText(opt.Value)
						return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, material.Editor(th, editor, "").Layout)
					}
				}),
			)
		})
	}
}

func getMinValue(key string) float32 {
	switch key {
	case "brightness", "contrast", "saturation", "sharpness":
		return -1.0
	case "ev":
		return -10.0
	default:
		return 0
	}
}

func getMaxValue(key string) float32 {
	switch key {
	case "brightness", "contrast", "saturation", "sharpness":
		return 1.0
	case "ev":
		return 10.0
	default:
		return 100
	}
}
