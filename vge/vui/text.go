package vui

import (
	"image"

	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vglyph"
	"github.com/lakal3/vge/vge/vmodel"
)

type Label struct {
	Style Style
	Class string
	Text  string
}

func NewLabel(text string) *Label {
	return &Label{Text: text}
}

func (l *Label) SetClass(class string) *Label {
	l.Class = class
	return l
}

func (l *Label) Measure(owner Owner, freeWidth int) (optimalSize image.Point) {
	if l.Style == nil {
		l.Style = owner.Theme().GetStyle(l, l.Class)
	}
	if l.Style != nil {
		font, fh := l.Style.GetFont(owner, l, 0)
		w := font.MeasureString(l.Text, fh)
		return image.Pt(w, fh)
	}
	return image.Pt(0, 0)
}

func (l *Label) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	if l.Style == nil {
		l.Style = owner.Theme().GetStyle(l, l.Class)
	}
	if l.Style != nil {
		l.Style.DrawString(owner, l, dc, pos, 0, l.Text)

	}
}

func (l *Label) Event(owner Owner, ev vapp.Event) {
}

func (l *Label) AssignTo(to **Label) *Label {
	*to = l
	return l
}

type Caret struct {
	Style    Style
	CaretPos int
}

func (c *Caret) Measure(owner Owner, freeWidth int) (optimalSize image.Point) {
	return
}

func (c *Caret) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	if c.Style != nil {
		c.Style.Draw(owner, c, dc, pos, 0)
	}
}

func (c *Caret) CalcPos(owner Owner, text string, pos vglyph.Position) vglyph.Position {
	if c.CaretPos > len(text) {
		c.CaretPos = len(text)
	}
	font, fh := c.Style.GetFont(owner, c, 0)
	w := font.MeasureString(text[:c.CaretPos], fh)
	pos.GlyphArea.Min.X += w
	pos.GlyphArea.Max.X = pos.GlyphArea.Min.X + 2
	// pos.GlyphArea.Max.Y -= pos.GlyphArea.Size().Y / 4
	return pos
}

func (c *Caret) Event(owner Owner, ev vapp.Event) {

}

type TextBox struct {
	Field
	caret      Caret
	CharLength int
	Text       string
	ms         MouseState
	focusState State
}

func NewTextBox(id string, charLen int, text string) *TextBox {
	return &TextBox{Field: Field{ID: id}, CharLength: charLen, Text: text}
}

func (t *TextBox) Measure(owner Owner, freeWidth int) (optimalSize image.Point) {

	if t.Style == nil {
		t.Style = owner.Theme().GetStyle(t, t.Class)
	}
	if t.Style != nil {
		_, fh := t.Style.GetFont(owner, t, t.GetState(owner))
		cp := t.Style.ContentPadding()
		return image.Pt(fh*(1+t.CharLength)/2, fh*3/2).Add(image.Pt(cp.Min.X+cp.Max.X, cp.Min.Y+cp.Max.Y))
	}
	return image.Pt(0, 0)
}

func (t *TextBox) GetState(owner Owner) State {
	s := t.ms.State
	if owner.GetFocus() == t {
		s |= STATEFocus
	}
	return t.Field.GetState(s)
}

func (t *TextBox) Render(owner Owner, dc *vmodel.DrawContext, pos vglyph.Position) {
	t.ms.SetArea(pos)
	s := t.GetState(owner)
	if t.Style == nil {
		t.Style = owner.Theme().GetStyle(t, t.Class)
	}
	if t.Style != nil {
		t.Style.Draw(owner, t, dc, pos, s)
		cp := t.Style.ContentPadding()
		pText := pos.Inset(cp.Min, cp.Max)
		t.Style.DrawString(owner, t, dc, pText, s, t.Text)
		if s.HasState(STATEFocus) {
			if t.caret.Style == nil {
				t.caret.Style = owner.Theme().GetStyle(&t.caret, t.Class)
			}
			pCaret := t.caret.CalcPos(owner, t.Text, pText)
			t.caret.Render(owner, dc, pCaret)
		}
	}

}

func (t *TextBox) Event(owner Owner, ev vapp.Event) {
	if t.Disabled {
		return
	}
	if CheckSetFocus(t, owner, ev) {
		t.caret.CaretPos = len(t.Text)
		return
	}
	if t.ms.Event(owner, ev) {
		// TODO: Set focus
		owner.SetFocus(t)
	}
	if t.checkKeyEvent(owner, ev) {
		return
	}
}

func (t *TextBox) checkKeyEvent(owner Owner, ev vapp.Event) bool {
	che, ok := ev.(*vapp.CharEvent)
	if ok {
		t.Text = t.Text[:t.caret.CaretPos] + string(che.Char) + t.Text[t.caret.CaretPos:]
		t.caret.CaretPos++
		che.SetHandled()
		return true
	}
	ke, ok := ev.(*vapp.KeyUpEvent)
	if ok {
		if ke.KeyCode == vapp.GLFWKeyLeft && ke.CurrentMods == 0 {
			if t.caret.CaretPos > 0 {
				t.caret.CaretPos--
			}
			ke.SetHandled()
			return true
		}
		if ke.KeyCode == vapp.GLFWKeyRight && ke.CurrentMods == 0 {
			if t.caret.CaretPos < len(t.Text) {
				t.caret.CaretPos++
			}
			ke.SetHandled()
			return true
		}
		if ke.KeyCode == vapp.GLFWKeyBackspace && ke.CurrentMods == 0 {
			if t.caret.CaretPos > 0 {
				t.Text = t.Text[:t.caret.CaretPos-1] + t.Text[t.caret.CaretPos:]
				t.caret.CaretPos--
			}
			ke.SetHandled()
			return true
		}
		if ke.KeyCode == vapp.GLFWKeyDelete && ke.CurrentMods == 0 {
			if t.caret.CaretPos < len(t.Text) {
				t.Text = t.Text[:t.caret.CaretPos] + t.Text[t.caret.CaretPos+1:]
			}
			ke.SetHandled()
			return true
		}
	}
	return false
}
