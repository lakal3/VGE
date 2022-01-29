package vimgui

import (
	"github.com/lakal3/vge/vge/vk"
	"sort"
	"sync"
)

type Style struct {
	Tags     []string
	Priority float64
	vk.State
}

func (s Style) Clone() Style {
	sNew := Style{Priority: s.Priority}
	if s.Tags != nil {
		sNew.Tags = make([]string, len(s.Tags))
		copy(sNew.Tags, s.Tags)
	}
	sNew.State = s.State.Clone()
	return sNew
}

type StyleSet struct {
	styles []Style
}

// Has check if any of styles in style set has given value
func (ss StyleSet) Has(valueType interface{}) bool {
	for _, s := range ss.styles {
		if s.Has(valueType) {
			return true
		}
	}
	return false
}

// Get retrieves first value from style set (first is one with the highest priority).
// If multiple styles have same priority order is undefined
func (ss StyleSet) Get(defaultValue interface{}) interface{} {
	for _, s := range ss.styles {
		v, ok := s.GetExists(defaultValue)
		if ok {
			return v
		}
	}
	return defaultValue
}

type Theme struct {
	styles []Style
	cache  map[vk.Key]StyleSet
	mx     *sync.Mutex
}

// Tags is helper to build string arrays
func Tags(tags ...string) []string {
	return tags
}

func NewTheme() *Theme {
	return &Theme{
		mx: &sync.Mutex{},
	}
}

// AddStyle adds new style to theme.
// You should add all styles to Theme before using it. Adding new styles while Theme is used will have performance penalty
func (t *Theme) AddStyle(st Style) *Theme {
	t.mx.Lock()
	t.cache = nil
	t.styles = append(t.styles, st.Clone())
	t.mx.Unlock()
	return t
}

// Add new style with given priority, tags and properties
// You should add all styles to Theme before using it. Adding new styles while Theme is used will have performance penalty
func (t *Theme) Add(priority float64, tags []string, properties ...interface{}) *Theme {
	s := Style{Priority: priority, Tags: tags}
	s.Set(properties...)
	t.AddStyle(s)
	return t
}

// GetStyles retrieve all styles matching given style tags
func (t *Theme) GetStyles(tags ...string) StyleSet {
	h := vk.NewHashKey(tags...)
	t.mx.Lock()
	defer t.mx.Unlock()
	if t.cache == nil {
		t.cache = make(map[vk.Key]StyleSet)
	}
	ss, ok := t.cache[h]
	if ok {
		return ss
	}
	ss = t.buildSet(tags)
	t.cache[h] = ss
	return ss
}

func (t *Theme) buildSet(tags []string) StyleSet {
	var ss StyleSet
	for _, s := range t.styles {
		if t.isValid(tags, s) {
			ss.styles = append(ss.styles, s)
		}
	}
	sort.Slice(ss.styles, func(i, j int) bool {
		return ss.styles[i].Priority > ss.styles[j].Priority
	})
	return ss
}

func (t *Theme) isValid(tags []string, s Style) bool {
	for _, ts := range s.Tags {
		found := false
		for _, t := range tags {
			if ts == t {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}
