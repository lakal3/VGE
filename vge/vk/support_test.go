package vk

import "testing"

func TestNewHashKey(t *testing.T) {
	h1 := NewHashKey("Hello", "World")
	h2 := NewHashKey("HelloWorld")
	h3 := NewHashKey("HelLo", "World")
	h4 := NewHashKey("Hello", "World")
	t.Log("H1: ", h1, " H2: ", h2, " H3: ", h3, " H4:", h4)
	if h1 == h2 {
		t.Error("H1 == H2")
	}
	if h1 < hashKeyOffset {
		t.Error("H1 too small ", h1)
	}
	if h2 < hashKeyOffset {
		t.Error("H2 too small ", h2)
	}
	if h4 < hashKeyOffset {
		t.Error("H4 too small ", h4)
	}
	if h1 == h3 {
		t.Error("H1 == H3")
	}
	if h1 != h4 {
		t.Error("H1 != H4")
	}
	if int64(h1) > 0 {
		t.Error("H1 to int64 should be negative")
	}
}
