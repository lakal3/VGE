package vimgui

import "github.com/lakal3/vge/vge/vdraw"

type FontStyle struct {
	Font *vdraw.Font
	Size float32
}

type ForeColor struct {
	Brush vdraw.Brush
}

type BorderColor struct {
	Brush vdraw.Brush
}

type BackgroudColor struct {
	Brush vdraw.Brush
}

type BorderRadius struct {
	vdraw.Corners
}

type BorderThickness struct {
	vdraw.Edges
}

type UnderlineThickness struct {
	Thickness float32
}

type ThumbSize struct {
	// ThumbSize tells how much thumbnail is bigger / smaller (negative) that slider
	ThumbSize float32
}

type ScrollBarSize struct {
	// Scroll bar width or height in scroll area
	BarSize float32
	// Padding of scrollable area
	Padding float32
}

type PrefixIcons struct {
	Font    *vdraw.Font
	Size    float32
	Icons   []rune
	Padding float32
}

type SelectedTextColor struct {
	Caret vdraw.Brush
	Text  vdraw.Brush
}

type PanelStyle struct {
	Edges       vdraw.Edges
	TitleHeight float32
	TitleBg     vdraw.Brush
}
