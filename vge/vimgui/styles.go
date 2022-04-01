package vimgui

import "github.com/lakal3/vge/vge/vdraw"

// FontStyle sets Font and Size for all controls that draw text
type FontStyle struct {
	Font *vdraw.Font
	Size float32
}

// IconStyle sets icon font for controls that use icons and don't have dedicated style
type IconStyle struct {
	Font    *vdraw.Font
	Size    float32
	Padding float32
}

// ForeColor describes Brush used to paint foreground elements
type ForeColor struct {
	Brush vdraw.Brush
}

// BorderColor describes Brush used to paint border elements
type BorderColor struct {
	Brush vdraw.Brush
}

// BackgroudColor describes Brush used to paint background elements
type BackgroudColor struct {
	Brush vdraw.Brush
}

type BorderRadius struct {
	vdraw.Corners
}

type BorderThickness struct {
	vdraw.Edges
}

// UnderlineThickness is set when you like control to draw just underline. BorderThickness and UnderlineThickness are exclusive
type UnderlineThickness struct {
	Thickness float32
}

// ThumbSize tells how much thumbnail is bigger / smaller (negative) that slider
type ThumbSize struct {
	ThumbSize float32
}

type ScrollBarSize struct {
	// Scroll bar width or height in scroll area
	BarSize float32
	// Padding of scrollable area
	Padding float32
}

// PrefixIcons describe icons for controls that need one or more standard icons to display state
// PrefixIcons area used for example by CheckBox, RadioButton and Increment control
type PrefixIcons struct {
	Font    *vdraw.Font
	Size    float32
	Icons   []rune
	Padding float32
}

// SelectedTextColor sets brushes used for selected text and caret
type SelectedTextColor struct {
	Caret vdraw.Brush
	Text  vdraw.Brush
}

// PanelStyle describes size of panel edges and title height + title background color
// If title height == 0, panel will only draw content
type PanelStyle struct {
	Edges       vdraw.Edges
	TitleHeight float32
	TitleBg     vdraw.Brush
}
