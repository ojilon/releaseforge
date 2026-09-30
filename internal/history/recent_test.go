package history

import (
	"path/filepath"
	"testing"
)

func TestTouchRecentDedupes(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "proj")
	if err := TouchRecent(root, a, "proj", "go", true); err != nil {
		t.Fatal(err)
	}
	if err := TouchRecent(root, a, "proj", "go", false); err != nil {
		t.Fatal(err)
	}
	rf, err := LoadRecent(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(rf.Items) != 1 {
		t.Fatalf("got %d items", len(rf.Items))
	}
	if rf.Items[0].LastScan == "" {
		t.Fatal("expected LastScan preserved")
	}
}
