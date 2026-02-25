package engine

import "testing"

func TestXPToNext_MinimumLevelFloor(t *testing.T) {
	if got := XPToNext(0); got != 100 {
		t.Fatalf("expected XPToNext(0)=100, got %d", got)
	}
	if got := XPToNext(5); got != 500 {
		t.Fatalf("expected XPToNext(5)=500, got %d", got)
	}
}
