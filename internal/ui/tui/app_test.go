package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/divijg19/Grimoire/internal/engine"
)

func TestSplitColumnOuterWidths_PreservesWidth(t *testing.T) {
	for total := 60; total <= 120; total++ {
		left, right := splitColumnOuterWidths(total)
		if left <= 0 || right <= 0 {
			t.Fatalf("invalid split for width %d: left=%d right=%d", total, left, right)
		}
		if left+right+columnGapCols > total {
			t.Fatalf("split overflow for width %d: left=%d right=%d", total, left, right)
		}
	}
}

func TestRenderInputPanel_HasConstantHeight(t *testing.T) {
	short := renderInputPanel(40, "x")
	long := renderInputPanel(40, strings.Repeat("a", 200))

	shortHeight := lipgloss.Height(short)
	longHeight := lipgloss.Height(long)

	minimum := inputPanelStyle.GetVerticalFrameSize() + promptContentHeight
	if shortHeight < minimum {
		t.Fatalf("input panel height=%d below minimum=%d", shortHeight, minimum)
	}
	if longHeight < shortHeight {
		t.Fatalf("input panel height regressed: short=%d long=%d", shortHeight, longHeight)
	}
	if longHeight-shortHeight > 2 {
		t.Fatalf("input panel height grew unexpectedly: short=%d long=%d", shortHeight, longHeight)
	}
}

func TestView_DoesNotOverflowHeightAtNarrowWidths(t *testing.T) {
	state := engine.DefaultState()
	m := newModel(&state, nil, nil)
	m.height = 20

	for width := 60; width <= 83; width++ {
		m.width = width
		m.layout()
		view := m.View()
		if strings.Contains(view, "Terminal too small") {
			continue
		}
		if h := lipgloss.Height(view); h > m.height {
			t.Fatalf("render overflow at width %d: viewHeight=%d termHeight=%d", width, h, m.height)
		}
		if w := lipgloss.Width(view); w > m.width {
			t.Fatalf("render overflow at width %d: viewWidth=%d termWidth=%d", width, w, m.width)
		}
	}
}

func TestView_SizeWarningStable(t *testing.T) {
	state := engine.DefaultState()
	m := newModel(&state, nil, nil)
	m.width = minTerminalWidth - 1
	m.height = 20

	first := m.View()
	second := m.View()

	if first != second {
		t.Fatalf("size warning changed between calls")
	}
	if !strings.Contains(first, "Terminal too small") {
		t.Fatalf("expected terminal too small warning, got: %q", first)
	}
}

func TestLayout_ProducesPositiveViewportDimensions(t *testing.T) {
	state := engine.DefaultState()
	m := newModel(&state, nil, nil)
	m.width = minTerminalWidth
	m.height = 20
	m.layout()

	if m.viewport.Width < 1 || m.viewport.Height < 1 {
		t.Fatalf("invalid viewport size: %dx%d", m.viewport.Width, m.viewport.Height)
	}
}

func TestWrapLogLines_CarriesTextToNextLines(t *testing.T) {
	wrapped := wrapLogLines([]string{strings.Repeat("word ", 20)}, 16)
	parts := strings.Split(wrapped, "\n")
	if len(parts) < 2 {
		t.Fatalf("expected wrapped output to span multiple lines")
	}
	for i, line := range parts {
		if w := lipgloss.Width(line); w > 16 {
			t.Fatalf("wrapped line %d exceeds width: got=%d max=%d", i, w, 16)
		}
	}
}

func TestLayout_WrapsLogsWithinViewportWidth(t *testing.T) {
	state := engine.DefaultState()
	m := newModel(&state, nil, nil)
	m.height = 20
	m.width = 70
	m.logs = []string{strings.Repeat("This is a long event log line ", 10)}
	m.layout()

	content := m.viewport.View()
	for i, line := range strings.Split(content, "\n") {
		if w := lipgloss.Width(line); w > m.viewport.Width {
			t.Fatalf("viewport line %d exceeds width: got=%d max=%d", i, w, m.viewport.Width)
		}
	}
}

func TestLeftColumnPanels_UseConsistentOuterWidth(t *testing.T) {
	state := engine.DefaultState()
	outerWidth := 42

	hud := renderHUDPanel(&state, outerWidth)
	logPane := logPanelStyle.Width(max(1, outerWidth-logPanelStyle.GetHorizontalFrameSize())).Render("Event Log\nentry")
	input := renderInputPanel(outerWidth, "explore")

	hudW := lipgloss.Width(hud)
	logW := lipgloss.Width(logPane)
	inputW := lipgloss.Width(input)

	if hudW <= 0 || logW <= 0 || inputW <= 0 {
		t.Fatalf("invalid panel widths: hud=%d log=%d input=%d", hudW, logW, inputW)
	}
	if hudW != logW || hudW != inputW {
		t.Fatalf("left column panel widths inconsistent: hud=%d log=%d input=%d", hudW, logW, inputW)
	}
	if hudW > outerWidth {
		t.Fatalf("panel width exceeds target outer width: panel=%d outer=%d", hudW, outerWidth)
	}
}

func TestRenderResizeWarning_StaysWithinTerminalWidth(t *testing.T) {
	termWidth := 60
	warning := renderResizeWarning(termWidth, "Terminal too small. Current: 58x18. Required: at least 60x20.")
	if w := lipgloss.Width(warning); w > termWidth {
		t.Fatalf("resize warning overflowed terminal width: got=%d max=%d", w, termWidth)
	}
}
