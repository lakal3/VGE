package vimgui

import (
	"fmt"
	"github.com/lakal3/vge/vge/vk"
	"math"
	"strconv"
)

func Increment(uf *UIFrame, id vk.Key, min, max int, describe func(val int) string, val *int) bool {
	if uf.IsHidden() {
		return false
	}
	var styles = []string{"*increment"}
	if uf.MouseHover() {
		styles = []string{":hover", "*increment"}
	}
	s := uf.GetStyles(styles...)
	sd := uf.GetStyles(append(styles, ":disabled")...)
	pi := s.Get(PrefixIcons{}).(PrefixIcons)
	fc := s.Get(ForeColor{}).(ForeColor)
	fcd := sd.Get(ForeColor{}).(ForeColor)
	inside := DrawBorder(uf, s)
	uf.PushControlArea()
	uf.ControlArea = inside
	changed := false
	if pi.Font != nil && len(pi.Icons) >= 2 {
		uf.ControlArea.To[0] = uf.ControlArea.From[0]
		uf.NewColumn(pi.Size, 0)
		if *val > min {
			if uf.MouseClick(1) {
				*val--
				changed = true
			}
			uf.Canvas().DrawText(pi.Font, pi.Size, uf.ControlArea.From, &fc.Brush, string(pi.Icons[0]))
		} else {
			uf.Canvas().DrawText(pi.Font, pi.Size, uf.ControlArea.From, &fcd.Brush, string(pi.Icons[0]))
		}
		uf.NewColumn(pi.Size, 0)
		if *val < max {
			if uf.MouseClick(1) {
				*val++
				changed = true
			}
			uf.Canvas().DrawText(pi.Font, pi.Size, uf.ControlArea.From, &fc.Brush, string(pi.Icons[1]))
		} else {
			uf.Canvas().DrawText(pi.Font, pi.Size, uf.ControlArea.From, &fcd.Brush, string(pi.Icons[1]))
		}
		uf.ControlArea.From[0] = uf.ControlArea.To[0] + pi.Padding
		uf.ControlArea.To[0] = inside.To[0]
		if *val > min && uf.MouseClick(2) {
			*val--
			changed = true
		}
		if *val < max && uf.MouseClick(1) {
			*val++
			changed = true
		}
	}
	title := describe(*val)
	DrawLabel(uf, title, s)
	uf.Pop()
	return changed
}

func Number(uf *UIFrame, id vk.Key, precision int, val *float64) (changed bool) {
	changed = false
	ParsedTextBox(uf, id, func() string {
		return strconv.FormatFloat(*val, 'f', precision, 64)
	}, func(sv string) (err error) {
		v, err := strconv.ParseFloat(sv, 64)
		if err != nil {
			return err
		}
		p10 := math.Pow10(precision)
		*val = math.Round(v*p10) / p10
		changed = true
		return nil
	})
	return
}

func LimitNumber(uf *UIFrame, id vk.Key, precision int, min, max float64, val *float64) (changed bool) {
	changed = false
	ParsedTextBox(uf, id, func() string {
		return strconv.FormatFloat(*val, 'f', precision, 64)
	}, func(sv string) (err error) {
		v, err := strconv.ParseFloat(sv, 64)
		if err != nil {
			return err
		}
		p10 := math.Pow10(precision)
		v = math.Round(v*p10) / p10
		if v < min {
			return fmt.Errorf("Value too small")
		}
		if v > max {
			return fmt.Errorf("Value too large")
		}
		changed = true
		return nil
	})
	return
}
