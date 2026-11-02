package tui

import "github.com/charmbracelet/lipgloss"

// Mode is a responsive breakpoint. The visual language is the same in all of
// them; only chrome density changes.
type Mode uint8

const (
	ModeTiny    Mode = iota // < 60 cols or < 14 rows: chrome collapses
	ModeCompact             // < 80 cols: one-line header, numeric tabs
	ModeNormal              // 80..119: normal shell
	ModeWide                // 120..159: room for side cards
	ModeUltra               // >= 160: room for side cards + wider tables
)

func (m Mode) String() string {
	switch m {
	case ModeTiny:
		return "tiny"
	case ModeCompact:
		return "compact"
	case ModeNormal:
		return "normal"
	case ModeWide:
		return "wide"
	default:
		return "ultra"
	}
}

// Rect is a content area (already inside any frame border).
type Rect struct {
	Width  int
	Height int
}

// Valid reports whether the rect can hold any text.
func (r Rect) Valid() bool { return r.Width > 0 && r.Height > 0 }

// Layout is the geometry of one frame. Shell chrome and windows both use it;
// no magic numbers live in internal/app anymore (docs/sweep-2mo/04).
type Layout struct {
	Width  int
	Height int
	Mode   Mode

	HeaderH int
	TabsH   int
	BarH    int
	FooterH int
	BodyH   int // total height of the framed body (borders included)

	Content   Rect // exact rect a window renders into
	SideWidth int  // width of an optional right-side card (0 when too narrow)
	Compact   bool
}

// Compute derives all rects from the terminal size. It never returns negative
// or zero dimensions: Bubble Tea sends WindowSizeMsg with 0 width/height on
// some terminals before the first real size.
func Compute(width, height int) Layout {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	mode := ModeNormal
	switch {
	case width < 60 || height < 14:
		mode = ModeTiny
	case width < 80:
		mode = ModeCompact
	case width < 120:
		mode = ModeNormal
	case width < 160:
		mode = ModeWide
	default:
		mode = ModeUltra
	}

	l := Layout{Width: width, Height: height, Mode: mode}
	l.Compact = mode == ModeCompact || mode == ModeTiny

	switch mode {
	case ModeTiny:
		l.HeaderH, l.TabsH, l.BarH, l.FooterH = 1, 0, 1, 1
	case ModeCompact:
		l.HeaderH, l.TabsH, l.BarH, l.FooterH = 3, 0, 3, 1
	default:
		l.HeaderH, l.TabsH, l.BarH, l.FooterH = 4, 1, 3, 1
	}

	fixed := l.HeaderH + l.TabsH + l.BarH + l.FooterH
	l.BodyH = height - fixed
	if l.BodyH < 3 {
		l.BodyH = 3
	}
	// A frame has a left+right border and a top+bottom border.
	innerW := width - 2
	if innerW < 1 {
		innerW = 1
	}
	innerH := l.BodyH - 2
	if innerH < 1 {
		innerH = 1
	}
	l.Content = Rect{Width: innerW, Height: innerH}

	if mode == ModeWide || mode == ModeUltra {
		side := innerW / 4
		if side > 38 {
			side = 38
		}
		if side >= 26 && innerW-side >= 44 {
			l.SideWidth = side
		}
	}
	return l
}

// BarRect is the content rect of the command bar frame.
func (l Layout) BarRect() Rect {
	w := l.Width - 2
	if w < 1 {
		w = 1
	}
	h := l.BarH - 2
	if h < 1 {
		h = 1
	}
	return Rect{Width: w, Height: h}
}

// HeaderRect is the content rect of the header frame.
func (l Layout) HeaderRect() Rect {
	w := l.Width - 2
	if w < 1 {
		w = 1
	}
	h := l.HeaderH - 2
	if h < 1 {
		h = 1
	}
	return Rect{Width: w, Height: h}
}

// Split returns the primary and side rects of the body content. Windows that
// have no side card simply use Primary and ignore Secondary.
func (l Layout) Split() (primary, secondary Rect) {
	if l.SideWidth <= 0 {
		return l.Content, Rect{}
	}
	gap := 2
	pw := l.Content.Width - l.SideWidth - gap
	if pw < 20 {
		return l.Content, Rect{}
	}
	return Rect{Width: pw, Height: l.Content.Height},
		Rect{Width: l.SideWidth, Height: l.Content.Height}
}

// PlaceCenter renders an overlay of size w x h centered on the screen,
// clipped to the terminal so a small window can never break the layout.
func PlaceCenter(screen Layout, content string, w, h int) string {
	if w > screen.Width {
		w = screen.Width
	}
	if h > screen.Height {
		h = screen.Height
	}
	if w < 1 || h < 1 {
		return ""
	}
	box := Clamp(content, w-2, h-2)
	return lipgloss.Place(screen.Width, screen.Height,
		lipgloss.Center, lipgloss.Center,
		Frame(box, w, h, OverlayBorder))
}

// VerticalGap joins shell blocks and asserts the result has exactly the
// expected number of lines. Used by tests to catch off-by-one chrome drift.
func VerticalGap(blocks ...string) string { return lipgloss.JoinVertical(lipgloss.Left, blocks...) }
