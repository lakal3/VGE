package mintheme

import (
	_ "embed"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vdraw"
	"github.com/lakal3/vge/vge/vimgui"
	"github.com/lakal3/vge/vge/vimgui/materialicons"
)

var PrimaryFont *vdraw.Font
var Theme *vimgui.Theme

func BuildMinTheme() (err error) {
	if Theme != nil {
		return nil
	}
	PrimaryFont, err = vdraw.LoadFont(opensans_regular)
	if err != nil {
		return err
	}
	err = materialicons.LoadIcons()
	if err != nil {
		return err
	}
	Theme = vimgui.NewTheme()
	root := vimgui.Style{}
	root.Set(vimgui.FontStyle{Font: PrimaryFont, Size: 16},
		vimgui.ForeColor{Brush: vdraw.SolidColor(mgl32.Vec4{1, 1, 1, 1})},
		vimgui.BorderColor{Brush: vdraw.SolidColor(mgl32.Vec4{1, 1, 1, 1})},
		vimgui.ThumbSize{ThumbSize: 3},
		vimgui.IconStyle{Font: materialicons.Icons, Size: 20, Padding: 5},
	)
	Theme.AddStyle(root)
	label := vimgui.Style{Tags: []string{"*label"}, Priority: 1}
	label.Set(vimgui.ForeColor{Brush: vdraw.SolidColor(mgl32.Vec4{0.75, 0.75, 0.75, 1})})
	Theme.AddStyle(label)

	btn := vimgui.Style{Tags: []string{"*button"}, Priority: 1}
	btn.Set(vimgui.BorderRadius{Corners: vdraw.UniformCorners(8)}, vimgui.BorderThickness{Edges: vdraw.UniformEdge(2)})
	Theme.AddStyle(btn)
	Theme.Add(1, vimgui.Tags(":hover"),
		vimgui.BackgroudColor{Brush: vdraw.SolidColor(mgl32.Vec4{0.5, 0.5, 0.5, 0.2})})
	Theme.Add(1, vimgui.Tags(":disabled"),
		vimgui.ForeColor{Brush: vdraw.SolidColor(mgl32.Vec4{0.5, 0.5, 0.5, 1})})
	Theme.Add(0, vimgui.Tags(),
		vimgui.SelectedTextColor{Text: vdraw.SolidColor(mgl32.Vec4{1, 1, 1, 1}),
			Caret: vdraw.SolidColor(mgl32.Vec4{0.0, 0.8, 0.0, 1})})
	Theme.Add(10, vimgui.Tags("radiobutton", "*togglebutton"),
		vimgui.PrefixIcons{Size: 20, Padding: 5, Font: materialicons.Icons, Icons: materialicons.GetRunes(
			"radio_button_unchecked", "radio_button_checked")})
	Theme.Add(10, vimgui.Tags("checkbox", "*togglebutton"),
		vimgui.PrefixIcons{Size: 20, Padding: 5, Font: materialicons.Icons, Icons: materialicons.GetRunes(
			"check_box_outline_blank", "check_box")})
	Theme.Add(10, vimgui.Tags("*togglebutton", "tab", ":checked"),
		vimgui.UnderlineThickness{Thickness: 2})
	Theme.Add(20, vimgui.Tags("primary"),
		vimgui.ForeColor{Brush: vdraw.SolidColor(mgl32.Vec4{0.3, 1, 0.3, 1})},
		vimgui.BorderColor{Brush: vdraw.SolidColor(mgl32.Vec4{0.3, 1, 0.3, 1})},
	)
	Theme.Add(20, vimgui.Tags("error"),
		vimgui.ForeColor{Brush: vdraw.SolidColor(mgl32.Vec4{1, 0.2, 0.2, 1})},
		vimgui.BorderColor{Brush: vdraw.SolidColor(mgl32.Vec4{1, 0.2, 0.2, 1})},
	)
	Theme.Add(1, vimgui.Tags("*slider"),
		vimgui.BorderColor{Brush: vdraw.SolidColor(mgl32.Vec4{0.8, 0.8, 0.8, 0.8})})
	Theme.Add(1, vimgui.Tags("*divider"),
		vimgui.BackgroudColor{Brush: vdraw.SolidColor(mgl32.Vec4{0.4, 0.4, 0.4, 1})})
	Theme.Add(1, vimgui.Tags("*border"),
		vimgui.BackgroudColor{Brush: vdraw.SolidColor(mgl32.Vec4{0.4, 0.4, 0.4, 1})})
	Theme.Add(1, vimgui.Tags("*textbox"), vimgui.UnderlineThickness{Thickness: 2},
		vimgui.BorderColor{Brush: vdraw.SolidColor(mgl32.Vec4{0.4, 0.4, 0.4, 1})})
	Theme.Add(2, vimgui.Tags("*textbox", ":focus"),
		vimgui.BorderColor{Brush: vdraw.SolidColor(mgl32.Vec4{0.8, 0.8, 0.8, 1})})

	Theme.Add(1, vimgui.Tags("*panel"),
		vimgui.PanelStyle{Edges: vdraw.Edges{Left: 10, Top: 5, Right: 10, Bottom: 5}, TitleHeight: 30, TitleBg: vdraw.SolidColor(mgl32.Vec4{0.8, 0.8, 0.8, 0.5})},
		vimgui.BorderColor{vdraw.SolidColor(mgl32.Vec4{0.8, 0.8, 0.8, 0.5})},
		vimgui.BorderRadius{Corners: vdraw.UniformCorners(10)},
		vimgui.BorderThickness{Edges: vdraw.UniformEdge(3)},
	)
	Theme.Add(10, vimgui.Tags("notitle", "*panel"),
		vimgui.PanelStyle{Edges: vdraw.Edges{Left: 10, Top: 5, Right: 10, Bottom: 5}, TitleHeight: 0})
	Theme.Add(10, vimgui.Tags("dialog", "*panel"),
		vimgui.BackgroudColor{Brush: vdraw.SolidColor(mgl32.Vec4{0.2, 0.2, 0.2, 1})})
	Theme.Add(10, vimgui.Tags("alert", "dialog", "*panel"),
		vimgui.BorderColor{Brush: vdraw.SolidColor(mgl32.Vec4{1, 0.2, 0.2, 1})})
	Theme.Add(10, vimgui.Tags("h2"),
		vimgui.FontStyle{Font: PrimaryFont, Size: 24})

	lr := materialicons.GetRunes("arrow_left", "arrow_right")
	Theme.Add(1, vimgui.Tags("*increment"),
		vimgui.IncrementIcons{Font: materialicons.Icons, Size: 20, Padding: 5, Decrement: lr[0], Increment: lr[1]})
	return nil
}

//go:embed SourceSansPro-Regular.ttf
var opensans_regular []byte
