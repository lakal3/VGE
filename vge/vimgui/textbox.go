package vimgui

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vdraw"
)

type TextSelection struct {
	From uint32
	To   uint32
}

type textBox struct {
	id       string
	sel      TextSelection
	t        string
	hasFocus bool
	ss       StyleSet
	caret    vdraw.Area
	fs       FontStyle
}

func (tb *textBox) handleKey(uf *UIFrame) (changed bool) {
	ml := uint32(len(tb.t))
	tb.sel = uf.GetState(tb.id, TextSelection{From: ml, To: ml}).(TextSelection)
	if tb.sel.From > ml {
		tb.sel.From = ml
	}
	if tb.sel.To > ml {
		tb.sel.To = ml
	}
	if tb.sel.From > tb.sel.To {
		tb.sel.To = tb.sel.From
	}
	if uf.Ev.Char >= 32 {
		tb.t = tb.t[:tb.sel.From] + string(uf.Ev.Char) + tb.t[tb.sel.To:]
		tb.sel.From++
		tb.sel.To++
		uf.SetState(tb.id, tb.sel)
		return true
	}
	if uf.Ev.KeyDown == vapp.GLFWKeyLeft || uf.Ev.KeyDown == vapp.GLFWKeyNumpad0+4 {
		if uf.HasMods(0) && tb.sel.From > 0 {
			tb.sel.From--
			tb.sel.To = tb.sel.From
			uf.SetState(tb.id, tb.sel)
		}
		if uf.HasMods(vapp.MODShift) && tb.sel.From > 0 {
			tb.sel.From--
			uf.SetState(tb.id, tb.sel)
		}
		if uf.HasMods(vapp.MODCtrl) && tb.sel.To > 0 && tb.sel.To > tb.sel.From {
			tb.sel.To--
			uf.SetState(tb.id, tb.sel)
		}
	}
	if uf.Ev.KeyDown == vapp.GLFWKeyRight || uf.Ev.KeyDown == vapp.GLFWKeyNumpad0+6 {
		if uf.HasMods(0) && tb.sel.From < ml {
			tb.sel.From++
			tb.sel.To = tb.sel.From
			uf.SetState(tb.id, tb.sel)
		}
		if uf.HasMods(vapp.MODShift) && tb.sel.To < ml {
			tb.sel.To++
			uf.SetState(tb.id, tb.sel)
		}
		if uf.HasMods(vapp.MODCtrl) && tb.sel.From < ml && tb.sel.To > tb.sel.From {
			tb.sel.From++
			uf.SetState(tb.id, tb.sel)
		}
	}
	del := false
	if (uf.Ev.KeyDown == vapp.GLFWKeyCode('C') || uf.Ev.KeyDown == vapp.GLFWKeyCode('X')) && uf.HasMods(vapp.MODCtrl) {
		if tb.sel.From < tb.sel.To {
			ct := tb.t[tb.sel.From:tb.sel.To]
			vapp.SetClipboard(ct)
			if uf.Ev.KeyDown == vapp.GLFWKeyCode('X') {
				del = true
			}
		}
	}
	if uf.Ev.KeyDown == vapp.GLFWKeyBackspace {
		if uf.HasMods(0) {
			if tb.sel.From != tb.sel.To {
				del = true
			} else if tb.sel.From > 0 {
				tb.sel.From--
				tb.sel.To--
				del = true
			}
		}
	}

	if uf.Ev.KeyDown == vapp.GLFWKeyDelete {
		if uf.HasMods(0) {
			del = true
		}
	}

	if del {
		if tb.sel.From < tb.sel.To {
			tb.t = tb.t[:tb.sel.From] + tb.t[tb.sel.To:]
			tb.sel.To = tb.sel.From
			uf.SetState(tb.id, tb.sel)
			return true
		}
		if tb.sel.From < ml {
			tb.t = tb.t[:tb.sel.From] + tb.t[tb.sel.From+1:]
			return true
		}
	}

	if uf.Ev.KeyDown == vapp.GLFWKeyCode('A') && uf.HasMods(vapp.MODCtrl) {
		tb.sel.From = 0
		tb.sel.To = ml
		uf.SetState(tb.id, tb.sel)
		return false
	}
	if uf.Ev.KeyDown == vapp.GLFWKeyCode('V') && uf.HasMods(vapp.MODCtrl) {
		cpTxt := vapp.GetClipboard()
		tb.t = tb.t[:tb.sel.From] + cpTxt + tb.t[tb.sel.To:]
		tb.sel.To = tb.sel.From + uint32(len(cpTxt))
		uf.SetState(tb.id, tb.sel)
		return true
	}
	return false
}

func (tb *textBox) measureCaret(uf *UIFrame) {
	var from float32
	_, _ = tb.fs.Font.MeasureTextWith(tb.fs.Size, tb.t+"_", func(idx int, at mgl32.Vec2, advance float32, kern bool) (nextAt mgl32.Vec2) {
		if kern {
			return at.Add(mgl32.Vec2{advance, 0})
		}
		if uint32(idx) == tb.sel.From {
			from = at[0]
		}
		if uint32(idx) == tb.sel.To {
			if tb.sel.From == tb.sel.To {
				tb.caret = vdraw.Area{From: mgl32.Vec2{from, 0}, To: mgl32.Vec2{from + 1, tb.fs.Size}}
			} else {
				tb.caret = vdraw.Area{From: mgl32.Vec2{from, 0}, To: mgl32.Vec2{at[0], tb.fs.Size}}
			}
		}
		return at.Add(mgl32.Vec2{advance + 1, 0})
	})
}

func (tb *textBox) draw(uf *UIFrame) {
	ss := tb.ss
	fs := ss.Get(FontStyle{}).(FontStyle)
	fc := ss.Get(ForeColor{}).(ForeColor)
	stc := ss.Get(SelectedTextColor{}).(SelectedTextColor)
	uf.PushControlArea()
	defer uf.Pop()
	area := DrawBorder(uf, ss)
	c := uf.Canvas()
	sz := mgl32.Vec2{1, 1}
	if tb.hasFocus {
		c.Draw(vdraw.Rect{Area: tb.caret}, area.From, sz, &stc.Caret)
	}
	c.DrawTextWith(fs.Font, fs.Size, area.From, tb.t, func(idx int, at mgl32.Vec2, advance float32, kern bool) (nextAt mgl32.Vec2) {
		if kern {
			return at.Add(mgl32.Vec2{advance, 0})
		}
		return at.Add(mgl32.Vec2{advance + 1, 0})
	}, func(idx int, at mgl32.Vec2, ch vdraw.Drawable) {
		if tb.hasFocus && uint32(idx) >= tb.sel.From && uint32(idx) <= tb.sel.To && tb.sel.From != tb.sel.To {
			c.Draw(ch, at.Add(mgl32.Vec2{2, 0}), sz.Mul(fs.Size/64), &stc.Text)
		} else {
			c.Draw(ch, at.Add(mgl32.Vec2{2, 0}), sz.Mul(fs.Size/64), &fc.Brush)
		}
	})

}

func TextBox(uf *UIFrame, id string, text *string) (changed bool) {
	tb := textBox{id: id, t: *text}
	var styles = []string{"*textbox"}
	if uf.MousePress(1) {
		uf.SetFocus(id)
	}
	tb.hasFocus = uf.HasFocus(id)
	if tb.hasFocus {
		styles = append(styles, ":focus")
	}
	if uf.MouseHover() {
		styles = append(styles, ":hover")
	}
	tb.ss = uf.GetStyles(styles...)
	tb.fs = tb.ss.Get(FontStyle{}).(FontStyle)
	if tb.fs.Font == nil {
		return
	}
	if tb.hasFocus {
		changed = tb.handleKey(uf)
		if changed {
			*text = tb.t
		}
		tb.measureCaret(uf)
	}
	tb.draw(uf)
	return changed
}
