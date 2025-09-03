package view

import (
	"testing"
)

func TestEq(t *testing.T) {
	var out int32 = 1
	got, err := Eq(out, 1)
	if err != nil {
		t.Errorf("Eq(%v, %v) = %v, want %v", out, 1, got, true)
	}
	// Test cases for equality
	tests := []struct {
		a    any
		b    any
		want bool
	}{
		{"hello", "hello", true},
		{123, 123, true},
		{123, 456, false},
		{nil, nil, true},
		{nil, "not nil", false},
		{struct{}{}, struct{}{}, true},
		{[]int{1, 2, 3}, []int{1, 2, 3}, false}, // Slices are not comparable with ==
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got, err := Eq(tt.a, tt.b)
			if err != nil {
				t.Errorf("Eq(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
