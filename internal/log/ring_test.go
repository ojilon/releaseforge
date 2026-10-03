package log

import (
	"fmt"
	"testing"
)

func TestRingOverwriteOrder(t *testing.T) {
	r := NewRing(3)
	for i := 1; i <= 5; i++ {
		r.Append(fmt.Sprintf("l%d", i))
	}
	if r.Len() != 3 {
		t.Fatalf("len %d", r.Len())
	}
	got := r.Snapshot()
	want := []string{"l3", "l4", "l5"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestRingUnfull(t *testing.T) {
	r := NewRing(5)
	r.Append("a")
	r.Append("b")
	got := r.Snapshot()
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("got %v", got)
	}
}

func TestRingCapFloor(t *testing.T) {
	if NewRing(0).Len() != 0 {
		t.Fatal("empty ring must have len 0")
	}
}
