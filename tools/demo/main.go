// Command demo renders a scripted tour of the real ReleaseForge TUI into a
// self-contained HTML player plus raw ANSI frame captures.
//
//	go run ./tools/demo -out demo
//
// Every frame is the actual app.Model.View() output, so the recording cannot
// drift from the implementation. See demo/README.md.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ojilon/releaseforge/internal/app"
)

type frame struct {
	Title   string `json:"title"`
	Caption string `json:"caption"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	HTML    string `json:"html"`
}

func main() {
	out := flag.String("out", "demo", "output directory for index.html and frames/")
	flag.Parse()
	if err := generate(*out); err != nil {
		fmt.Fprintln(os.Stderr, "demo:", err)
		os.Exit(1)
	}
}

func generate(out string) error {
	frames := app.DemoFrames()
	if len(frames) == 0 {
		return fmt.Errorf("no frames produced")
	}
	framesDir := filepath.Join(out, "frames")
	if err := os.MkdirAll(framesDir, 0o755); err != nil {
		return err
	}
	// Remove stale captures so renumbering/retiling never leaves ghosts.
	if old, err := os.ReadDir(framesDir); err == nil {
		for _, e := range old {
			if strings.HasSuffix(e.Name(), ".ans") {
				_ = os.Remove(filepath.Join(framesDir, e.Name()))
			}
		}
	}
	payload := make([]frame, 0, len(frames))
	for i, f := range frames {
		name := fmt.Sprintf("%02d-%s.ans", i+1, slug(f.Title))
		if err := os.WriteFile(filepath.Join(framesDir, name), []byte(f.ANSI), 0o644); err != nil {
			return err
		}
		payload = append(payload, frame{
			Title: f.Title, Caption: f.Caption, Width: f.Width, Height: f.Height,
			HTML: ansiToHTML(f.ANSI),
		})
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	// Avoid a literal </script> inside the embedded JSON.
	framesJSON := strings.ReplaceAll(string(data), "<", `\u003c`)
	page := strings.Replace(demoHTML, "/*FRAMES*/", "const frames = "+framesJSON+";", 1)
	page = strings.Replace(page, "/*COUNT*/", strconv.Itoa(len(payload)), 1)
	if err := os.WriteFile(filepath.Join(out, "index.html"), []byte(page), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d frames, %d ANSI captures)\n",
		filepath.Join(out, "index.html"), len(payload), len(payload))
	return nil
}

func slug(s string) string {
	s = strings.ToLower(s)
	if i := strings.Index(s, " · "); i >= 0 {
		s = s[i+len(" · "):]
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '·' || r == '-':
			if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
				b.WriteRune('-')
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// --- ANSI SGR -> HTML spans -------------------------------------------------

type styleState struct {
	fg, bg          string
	bold, underline bool
	italic          bool
}

func (s styleState) css() string {
	var parts []string
	if s.fg != "" {
		parts = append(parts, "color:"+s.fg)
	}
	if s.bg != "" {
		parts = append(parts, "background-color:"+s.bg)
	}
	if s.bold {
		parts = append(parts, "font-weight:700")
	}
	if s.underline {
		parts = append(parts, "text-decoration:underline")
	}
	if s.italic {
		parts = append(parts, "font-style:italic")
	}
	return strings.Join(parts, ";")
}

var basicColors = []string{
	"#1d2029", "#f7768e", "#9ece6a", "#e0af68", "#7aa2f7", "#bb9af7", "#7dcfff", "#c0caf5",
}
var brightColors = []string{
	"#565f89", "#ff9eaf", "#b8e986", "#ffcf8b", "#9ec5ff", "#d5baff", "#a5e8ff", "#e6e9f8",
}

func ansiToHTML(s string) string {
	var out strings.Builder
	var cur strings.Builder
	st := styleState{}
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		css := st.css()
		if css == "" {
			out.WriteString(html.EscapeString(cur.String()))
		} else {
			out.WriteString(`<span style="` + css + `">` + html.EscapeString(cur.String()) + `</span>`)
		}
		cur.Reset()
	}
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && (s[j] == ';' || (s[j] >= '0' && s[j] <= '9')) {
				j++
			}
			if j < len(s) && s[j] == 'm' {
				flush()
				applySGR(&st, s[i+2:j])
				i = j + 1
				continue
			}
		}
		if s[i] == '\n' {
			flush()
			out.WriteByte('\n')
			i++
			continue
		}
		cur.WriteByte(s[i])
		i++
	}
	flush()
	return out.String()
}

func applySGR(st *styleState, params string) {
	if params == "" {
		params = "0"
	}
	parts := strings.Split(params, ";")
	num := func(i int) int {
		if i >= len(parts) {
			return 0
		}
		n, _ := strconv.Atoi(parts[i])
		return n
	}
	for i := 0; i < len(parts); i++ {
		n := num(i)
		switch {
		case n == 0:
			*st = styleState{}
		case n == 1:
			st.bold = true
		case n == 22:
			st.bold = false
		case n == 3:
			st.italic = true
		case n == 4:
			st.underline = true
		case n == 24:
			st.underline = false
		case n == 39:
			st.fg = ""
		case n == 49:
			st.bg = ""
		case n >= 30 && n <= 37:
			st.fg = basicColors[n-30]
		case n >= 90 && n <= 97:
			st.fg = brightColors[n-90]
		case n >= 40 && n <= 47:
			st.bg = basicColors[n-40]
		case n >= 100 && n <= 107:
			st.bg = brightColors[n-100]
		case n == 38 && num(i+1) == 2:
			st.fg = fmt.Sprintf("rgb(%d,%d,%d)", num(i+2), num(i+3), num(i+4))
			i += 4
		case n == 48 && num(i+1) == 2:
			st.bg = fmt.Sprintf("rgb(%d,%d,%d)", num(i+2), num(i+3), num(i+4))
			i += 4
		case n == 38 && num(i+1) == 5:
			st.fg = color256(num(i + 2))
			i += 2
		case n == 48 && num(i+1) == 5:
			st.bg = color256(num(i + 2))
			i += 2
		}
	}
}

func color256(n int) string {
	if n < 8 {
		return basicColors[n]
	}
	if n < 16 {
		return brightColors[n-8]
	}
	if n < 232 {
		n -= 16
		r, g, b := n/36, (n%36)/6, n%6
		vals := []int{0, 95, 135, 175, 215, 255}
		return fmt.Sprintf("rgb(%d,%d,%d)", vals[r], vals[g], vals[b])
	}
	if n > 255 {
		n = 255
	}
	v := 8 + (n-232)*10
	return fmt.Sprintf("rgb(%d,%d,%d)", v, v, v)
}

const demoHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>ReleaseForge — TUI demo</title>
<style>
  :root { --bg:#0d0f16; --panel:#151822; --line:#2a3044; --muted:#8b93b5; --accent:#8aadf4; --text:#d5daf0; }
  * { box-sizing:border-box; }
  body { margin:0; background:var(--bg); color:var(--text);
         font-family:ui-sans-serif,system-ui,-apple-system,"Segoe UI",sans-serif; }
  header { padding:28px 32px 12px; max-width:1400px; margin:auto; }
  h1 { margin:0 0 6px; font-size:22px; letter-spacing:.02em; }
  h1 span { color:var(--accent); }
  .sub { color:var(--muted); font-size:14px; line-height:1.5; max-width:900px; }
  .wrap { max-width:1400px; margin:auto; padding:12px 32px 48px; }
  .bar { display:flex; flex-wrap:wrap; align-items:center; gap:10px; margin:14px 0; }
  button { background:var(--panel); color:var(--text); border:1px solid var(--line);
           border-radius:8px; padding:8px 14px; font-size:13px; cursor:pointer; }
  button:hover { border-color:var(--accent); }
  button.primary { background:#222a45; border-color:var(--accent); color:#dfe7ff; }
  .counter { color:var(--muted); font-size:13px; margin-left:auto; font-variant-numeric:tabular-nums; }
  .window { border:1px solid var(--line); border-radius:12px; overflow:hidden;
            background:#11131c; box-shadow:0 24px 60px rgba(0,0,0,.45); }
  .titlebar { display:flex; align-items:center; gap:7px; padding:11px 14px;
              background:#1a1e2b; border-bottom:1px solid var(--line); }
  .dot { width:11px; height:11px; border-radius:50%; }
  .d1 { background:#f7768e; } .d2 { background:#e0af68; } .d3 { background:#9ece6a; }
  .titlebar .name { margin-left:10px; color:var(--muted); font-size:12px;
                    font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace; }
  .titlebar .dims { margin-left:auto; color:#6b7395; font-size:12px;
                    font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace; }
  pre { margin:0; padding:18px 20px; overflow:auto; font-size:13px; line-height:1.22;
        font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,"Liberation Mono",monospace;
        white-space:pre; color:var(--text); min-height:420px; }
  .caption { display:flex; gap:14px; align-items:baseline; padding:16px 4px 0; }
  .caption h2 { margin:0; font-size:15px; color:var(--accent); white-space:nowrap; }
  .caption p { margin:0; color:var(--muted); font-size:13px; line-height:1.5; }
  .dots { display:flex; gap:6px; margin-top:16px; flex-wrap:wrap; }
  .dots button { width:auto; padding:5px 9px; font-size:11px; }
  .dots button.on { border-color:var(--accent); color:#dfe7ff; background:#222a45; }
  .keys { margin-top:18px; color:#6b7395; font-size:12px; line-height:1.7; }
  code { background:#1a1e2b; padding:2px 5px; border-radius:4px; color:#b9c6ef; }
</style>
</head>
<body>
<header>
  <h1><span>ReleaseForge</span> — TUI demo</h1>
  <p class="sub">A scripted tour of the actual Bubble Tea interface. Each frame is the real
  <code>Model.View()</code> output captured with colors, then translated from ANSI to HTML —
  not a mock-up. The app's intent is unchanged: a local command bar plus live windows for
  scan, build, test, git, logs and release work.</p>
</header>
<div class="wrap">
  <div class="bar">
    <button id="prev">← Prev</button>
    <button id="next" class="primary">Next →</button>
    <button id="play">▶ Autoplay</button>
    <span class="counter" id="counter"></span>
  </div>
  <div class="window">
    <div class="titlebar">
      <span class="dot d1"></span><span class="dot d2"></span><span class="dot d3"></span>
      <span class="name" id="winname">releaseforge tui</span>
      <span class="dims" id="dims"></span>
    </div>
    <pre id="screen"></pre>
  </div>
  <div class="caption"><h2 id="title"></h2><p id="caption"></p></div>
  <div class="dots" id="dots"></div>
  <div class="keys">
    <strong>In the real app:</strong> <code>1..6</code> / <code>alt+1..6</code> switch windows ·
    <code>ctrl+p</code> palette · <code>ctrl+r</code> history · <code>ctrl+o</code> folder picker ·
    <code>ctrl+w</code> focus a window · <code>tab</code> complete · <code>esc</code> back ·
    <code>ctrl+x</code> cancel · <code>ctrl+l</code> clear · <code>ctrl+c</code> quit.<br>
    Demo player: <code>←</code> / <code>→</code> to step, <code>space</code> to toggle autoplay.
    Raw ANSI captures live in <code>frames/</code> for terminal playback.
  </div>
</div>
<script>
/*FRAMES*/
let i = 0, timer = null;
const $ = id => document.getElementById(id);
function render() {
  const f = frames[i];
  $('screen').innerHTML = f.html;
  $('title').textContent = f.title;
  $('caption').textContent = f.caption;
  $('dims').textContent = f.width + ' × ' + f.height;
  $('counter').textContent = (i+1) + ' / /*COUNT*/';
  $('winname').textContent = 'releaseforge tui — ' + f.title;
  [...$('dots').children].forEach((b,n) => b.classList.toggle('on', n === i));
}
function step(d) { i = (i + d + frames.length) % frames.length; render(); }
function toggle() {
  if (timer) { clearInterval(timer); timer = null; $('play').textContent = '▶ Autoplay'; }
  else { timer = setInterval(() => step(1), 4500); $('play').textContent = '❚❚ Pause'; }
}
frames.forEach((f,n) => {
  const b = document.createElement('button');
  b.textContent = String(n+1).padStart(2,'0');
  b.onclick = () => { i = n; render(); };
  $('dots').appendChild(b);
});
$('prev').onclick = () => step(-1);
$('next').onclick = () => step(1);
$('play').onclick = toggle;
document.addEventListener('keydown', e => {
  if (e.key === 'ArrowRight') step(1);
  else if (e.key === 'ArrowLeft') step(-1);
  else if (e.key === ' ') { e.preventDefault(); toggle(); }
});
render();
</script>
</body>
</html>
`
