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
		t.Fatalf("long input panel height=%d less than short=%d", longHeight, shortHeight)
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
