package vui

import (
	"github.com/lakal3/vge/vge/vapp"
	"github.com/lakal3/vge/vge/vui/materialicons"
)

func NewCheckbox(title string, labelClass string) *ToggleButton {
	tb := &ToggleButton{Field: Field{ID: MakeID()}}
	on := NewHStack(4, NewLabel(string(materialicons.Check_box)).SetClass("icon "+labelClass),
		NewLabel(title).SetClass(labelClass))
	off := NewHStack(4, NewLabel(string(materialicons.Check_box_outline_blank)).SetClass("icon"+labelClass),
		NewLabel(title).SetClass(labelClass))
	tb.CheckedContent, tb.Content = on, off
	return tb
}

func NewRadioButton(title string) *ToggleButton {
	tb := &ToggleButton{Field: Field{ID: MakeID()}}
	on := NewHStack(4, NewLabel(string(materialicons.Radio_button_checked)).SetClass("icon"),
		NewLabel(title))
	off := NewHStack(4, NewLabel(string(materialicons.Radio_button_unchecked)).SetClass("icon"),
		NewLabel(title))
	tb.CheckedContent, tb.Content = on, off
	return tb
}

type RadioGroup struct {
	ID        string
	Value     int
	OnChanged func(value int)
	buttons   []*ToggleButton
}

func (rg *RadioGroup) SetValue(value int) *RadioGroup {
	rg.Value = value
	for idx, b := range rg.buttons {
		b.Checked = value == idx
	}
	return rg
}

func NewRadioGroup(buttons ...*ToggleButton) *RadioGroup {
	rg := &RadioGroup{ID: MakeID(), buttons: buttons}
	for idx, btn := range buttons {
		val := idx
		btn.OnChanged = func(checked bool) {
			rg.SetValue(val)
			if rg.OnChanged != nil {
				rg.OnChanged(rg.Value)
			} else {
				vapp.Post(&ValueChangedEvent{ID: rg.ID, NewValue: rg.Value})
			}
		}
	}
	rg.SetValue(0)
	return rg
}
