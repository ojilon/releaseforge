package history

import (
	"fmt"
	"testing"
)

func TestCommandsRoundTrip(t *testing.T) {
	root := t.TempDir()
	if got := LoadCommands(root); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
	AppendCommand(root, "status")
	AppendCommand(root, "  ")
	AppendCommand(root, "scan /x")
	got := LoadCommands(root)
	if len(got) != 2 || got[0] != "status" || got[1] != "scan /x" {
		t.Fatalf("got %v", got)
	}
}

func TestCommandsCap(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < MaxCommands+50; i++ {
		AppendCommand(root, fmt.Sprintf("cmd-%d", i))
	}
	got := LoadCommands(root)
	if len(got) != MaxCommands {
		t.Fatalf("got %d", len(got))
	}
	if got[0] != "cmd-50" {
		t.Fatalf("oldest not dropped: %q", got[0])
	}
}
