package ui

import (
	"log"

	"gioui.org/layout"
	"gioui.org/widget"

	"gioui.org/unit"
		"gioui.org/widget/material"

		"golang.org/x/exp/shiny/materialdesign/icons"

			"github.com/alexanderi96/libcamera-tray/camera"


)

var (
	//EV: checkbox label text field +butt -butt
	evCheckbox = new(widget.Bool)
	evEditor = new(widget.Editor)

	evIncIconButton = new(widget.Clickable)
	evDecIconButton = new(widget.Clickable)


	OptionsList = []layout.Widget{}

	list = &widget.List{
		List: layout.List{
			Axis: layout.Vertical,
		},
	}
)

func settings(gtx *C, th *material.Theme) {
	
	
	OptionsList = []layout.Widget{
		// getSettingsRow(gtx, th, evCheckbox, "EV", evEditor, camera.Params["ev"].Value, evIncIconButton, evDecIconButton),	
	}

	for key, opt := range camera.Params {
		checkbox := new(widget.Bool)
		editor := new(widget.Editor)

		incIconButton := new(widget.Clickable)
		decIconButton := new(widget.Clickable)
		OptionsList = append(OptionsList, getSettingsRow(gtx, th, checkbox, key, editor, opt.Value, incIconButton, decIconButton))
	}

}

func getSettingsRow(gtx *C, th *material.Theme, checkbox *widget.Bool, checkboxLabel string, editor *widget.Editor, editorLabel string, incButton, decButton *widget.Clickable) layout.Widget {
	icAdd, err := widget.NewIcon(icons.ContentAdd)
	if err != nil {
		log.Fatal(err)
	}

	icRem, err := widget.NewIcon(icons.ContentRemove)
	if err != nil {
		log.Fatal(err)
	}

	return func(gtx C) D {
		in := layout.UniformInset(unit.Dp(8))
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				//checkbox
				return in.Layout(gtx, material.CheckBox(th, checkbox, checkboxLabel).Layout)
			}),
			layout.Rigid(func(gtx C) D {
				//editor
				return in.Layout(gtx, material.Editor(th, editor, editorLabel).Layout)
			}),
			layout.Rigid(func(gtx C) D {
				//increase
				return in.Layout(gtx, material.IconButton(th, incButton, icAdd, "Increment").Layout)
			}),
			layout.Rigid(func(gtx C) D {
				//decrease
				return in.Layout(gtx, material.IconButton(th, decButton, icRem, "Decrement").Layout)
			}),
		)
	}
}