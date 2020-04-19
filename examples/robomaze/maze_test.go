package main

import "testing"

func Test_Maze(t *testing.T) {
	m := &maze{size: 8}
	m.buildSquares()
	for idx, s := range m.squares {
		if s.open == 0 {
			t.Errorf("Closed square at %d", idx)
		}
	}
}
