package app

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/build"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/storage"
)

// linesMsg carries a batch of streamed output lines into the console.
type linesMsg struct{ lines []rflog.Line }

// doneMsg marks completion of an async command. cancelled is set when the
// user asked for it, so a late doneMsg never masquerades as a failure.
type doneMsg struct {
	label     string
	err       error
	report    []string
	summary   string
	cancelled bool
}

// tickMsg drives the 1s elapsed-time clock while a run is active.
type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Run starts the interactive TUI. The data root must be initialised.
func Run(dataRoot, projectDir string) error {
	if strings.TrimSpace(dataRoot) == "" {
		var err error
		dataRoot, err = storage.ResolveRoot("")
		if err != nil {
			return err
		}
	}
	if !storage.Exists(storage.GlobalConfigPath(dataRoot)) {
		return fmt.Errorf("data root not initialised (%s missing) — run `releaseforge init` first",
			storage.GlobalConfigPath(dataRoot))
	}
	p := tea.NewProgram(New(dataRoot, projectDir),
		tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

// waitLine returns a Cmd delivering the next output batch, or the final
// doneMsg once the process ends and its output is drained. Batching (first
// line blocking, rest opportunistic up to 100) keeps fast output cheap.
func (m *Model) waitLine() tea.Cmd {
	if m.run == nil || m.run.Handle == nil {
		return nil
	}
	r := m.run // cancellation may land after this Cmd was created
	h := r.Handle
	return func() tea.Msg {
		first, ok := <-h.Lines
		if !ok {
			return finishRun(h, r.OKLabel, r.FailLabel, r.Cancelled)
		}
		batch := []rflog.Line{first}
		for len(batch) < 100 {
			select {
			case ln, ok := <-h.Lines:
				if !ok {
					return linesMsg{lines: batch}
				}
				batch = append(batch, ln)
			default:
				return linesMsg{lines: batch}
			}
		}
		return tea.Msg(linesMsg{lines: batch})
	}
}

// finishRun drains the result after output is exhausted.
func finishRun(h *build.Handle, okL, failL string, cancelled bool) tea.Msg {
	res := <-h.Done
	if res.Success && !cancelled {
		return doneMsg{label: okL}
	}
	if cancelled {
		return doneMsg{label: "cancelled " + failL, cancelled: true,
			err: fmt.Errorf("stopped by user")}
	}
	var rep []string
	summary := ""
	if res.Report != nil {
		rep = res.Report.Format()
		summary = res.Report.Summary
	}
	return doneMsg{label: failL, err: firstError(res), report: rep, summary: summary}
}

// firstError summarizes a failed Result for doneMsg.
func firstError(res build.Result) error {
	if len(res.Errors) > 0 {
		return fmt.Errorf("%s", res.Errors[0])
	}
	return fmt.Errorf("exit %d", res.ExitCode)
}
