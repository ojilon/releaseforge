package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/build"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/storage"
)

// linesMsg carries a batch of streamed output lines into the model.
type linesMsg struct {
	lines []rflog.Line
}

// doneMsg marks completion of an async command.
type doneMsg struct {
	label   string
	err     error
	report  []string
	summary string
}

// Run starts the interactive TUI. The data root must be initialised
// (run `releaseforge init`); otherwise it returns a clear error.
func Run(dataRoot, projectDir string) error {
	if strings.TrimSpace(dataRoot) == "" {
		var err error
		dataRoot, err = storage.ResolveRoot("")
		if err != nil {
			return err
		}
	}
	if !storage.Exists(storage.GlobalConfigPath(dataRoot)) {
		return fmt.Errorf("data root not initialised (%s missing) — run `releaseforge init` first", storage.GlobalConfigPath(dataRoot))
	}
	p := tea.NewProgram(New(dataRoot, projectDir), tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

// waitLine returns a Cmd delivering the next output batch, or the final
// doneMsg once the process ends and its output is drained. Batching (first
// line blocking, rest opportunistic up to 100) keeps fast output cheap.
func (m *Model) waitLine() tea.Cmd {
	h, okL, failL := m.handle, m.pendingOK, m.pendingErr
	return func() tea.Msg {
		first, ok := <-h.Lines
		if !ok {
			return finishRun(h, okL, failL)
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
		return linesMsg{lines: batch}
	}
}

// finishRun drains the result after output is exhausted.
func finishRun(h *build.Handle, okL, failL string) tea.Msg {
	res := <-h.Done
	if res.Success {
		return doneMsg{label: okL}
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
