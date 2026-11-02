package app

// Phase is the run state machine (docs/sweep-2mo/40). Transitions are
// validated: a late doneMsg after a cancel must not silently turn a
// cancelled run into a failure.
type Phase uint8

const (
	PhaseIdle Phase = iota
	PhaseBuilding
	PhaseTesting
	PhaseOK
	PhaseFailed
	PhaseCancelled
)

func (p Phase) String() string {
	switch p {
	case PhaseBuilding:
		return "building"
	case PhaseTesting:
		return "testing"
	case PhaseOK:
		return "ok"
	case PhaseFailed:
		return "failed"
	case PhaseCancelled:
		return "cancelled"
	default:
		return "idle"
	}
}

// Busy reports whether a background job owns the console.
func (p Phase) Busy() bool { return p == PhaseBuilding || p == PhaseTesting }

var transitions = map[Phase][]Phase{
	PhaseIdle:      {PhaseBuilding, PhaseTesting, PhaseOK, PhaseFailed},
	PhaseBuilding:  {PhaseOK, PhaseFailed, PhaseCancelled},
	PhaseTesting:   {PhaseOK, PhaseFailed, PhaseCancelled},
	PhaseOK:        {PhaseBuilding, PhaseTesting, PhaseIdle},
	PhaseFailed:    {PhaseBuilding, PhaseTesting, PhaseIdle},
	PhaseCancelled: {PhaseBuilding, PhaseTesting, PhaseIdle},
}

// CanTransition is the pure transition table, exercised by phase_test.go.
func CanTransition(from, to Phase) bool {
	if from == to {
		return true
	}
	for _, allowed := range transitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// setPhase applies a transition. Illegal transitions are ignored (and
// reported) so state can never be corrupted by a racing message.
func (m *Model) setPhase(to Phase) bool {
	if !CanTransition(m.phase, to) {
		return false
	}
	m.phase = to
	return true
}

// Focus identifies the single consumer of keyboard input
// (docs/sweep-2mo/41): the bar, the active window, or an overlay.
type Focus uint8

const (
	FocusBar Focus = iota
	FocusWindow
	FocusPalette
	FocusPicker
)

func (f Focus) String() string {
	switch f {
	case FocusWindow:
		return "window"
	case FocusPalette:
		return "palette"
	case FocusPicker:
		return "picker"
	default:
		return "bar"
	}
}
