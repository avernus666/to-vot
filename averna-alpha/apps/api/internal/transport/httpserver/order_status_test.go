package httpserver

import "testing"

func TestValidOrderTransition(t *testing.T) {
	tests := []struct {
		current, next string
		want          bool
	}{
		{"pending", "confirmed", true},
		{"pending", "cancelled", true},
		{"confirmed", "completed", true},
		{"confirmed", "cancelled", true},
		{"pending", "completed", false},
		{"completed", "cancelled", false},
		{"cancelled", "confirmed", false},
	}
	for _, tt := range tests {
		t.Run(tt.current+"_to_"+tt.next, func(t *testing.T) {
			if got := validOrderTransition(tt.current, tt.next); got != tt.want {
				t.Fatalf("validOrderTransition(%q, %q) = %v, want %v", tt.current, tt.next, got, tt.want)
			}
		})
	}
}
