// Package app is the Bubble Tea root model: tabs, command bar, live panes.
package app

// Root model will own:
//   - current project context
//   - active tab (Overview | Build | Logs | Git | Metrics | Config)
//   - persistent command input (Bubbles textinput + history)
//   - live log viewport
//   - message routing from build/test runners
//
// See docs/04-tui-design.md.
