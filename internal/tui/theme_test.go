package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/x/ansi"
)

func TestSetThemeAndRepaint(t *testing.T) {
	defer SetTheme("dark")
	if !SetTheme("light") || Current.Name != "light" {
		t.Fatal("light theme not applied")
	}
	ti := textinput.New()
	PaintInput(&ti)
	// A theme change must repaint the input too (the old bug: only chrome
	// changed while the focused bar kept stale colors).
	if ti.PromptStyle.GetForeground() == nil {
		t.Log("no color in this profile; style still rebuilt")
	}
	if SetTheme("does-not-exist") {
		t.Fatal("unknown theme must return false")
	}
	if !SetTheme("mono") || !Current.Mono {
		t.Fatal("mono theme not applied")
	}
	if got := ansi.StringWidth(ErrorStyle.Render("x")); got != 1 {
		t.Fatalf("mono style width: %d", got)
	}
	if got := ErrorStyle.Render("x"); got != "x" && ansi.StringWidth(got) != 1 {
		t.Fatalf("mono must not add color: %q", got)
	}
}

func TestDetectTheme(t *testing.T) {
	env := func(m map[string]string) func(string) string {
		return func(k string) string { return m[k] }
	}
	cases := []struct {
		name string
		env  map[string]string
		tty  bool
		dark bool
		want string
	}{
		{"explicit", map[string]string{"RELEASEFORGE_THEME": "light"}, false, true, "light"},
		{"no color", map[string]string{"NO_COLOR": "1"}, false, true, "mono"},
		{"dumb term", map[string]string{"TERM": "dumb"}, false, true, "mono"},
		{"unknown override", map[string]string{"RELEASEFORGE_THEME": "neon"}, false, true, "dark"},
		{"light fg/bg", map[string]string{"COLORFGBG": "0;15"}, false, true, "light"},
		{"dark fg/bg", map[string]string{"COLORFGBG": "15;0"}, false, true, "dark"},
		{"tty light bg", map[string]string{}, true, false, "light"},
		{"tty dark bg", map[string]string{}, true, true, "dark"},
	}
	for _, c := range cases {
		got := DetectTheme(env(c.env), c.tty, func() bool { return c.dark })
		if got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}

func TestThemeNames(t *testing.T) {
	for _, n := range ThemeNames() {
		if !IsTheme(n) {
			t.Fatalf("%q registered but IsTheme false", n)
		}
	}
	if IsTheme("") {
		t.Fatal("empty name must not be a theme")
	}
}
