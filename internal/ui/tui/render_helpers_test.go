package tui

import (
	"strings"
	"testing"

	"github.com/divijg19/Grimoire/internal/engine"
)

func TestRenderArea_AccountsForOuterMargins(t *testing.T) {
	w, h := renderArea(100, 30)
	if w != 100-outerMarginLeft-outerMarginRight {
		t.Fatalf("unexpected width: %d", w)
	}
	if h != 30-outerMarginTop-outerMarginBottom {
		t.Fatalf("unexpected height: %d", h)
	}
}

func TestTruncateText_AppendsEllipsis(t *testing.T) {
	got := truncateText("abcdefgh", 5)
	if got != "abcd…" {
		t.Fatalf("truncateText returned %q", got)
	}
}

func TestItemDisplayName_NormalizesLookup(t *testing.T) {
	got := itemDisplayName("Healing Potion")
	if got != engine.Items["healing_potion"].Name {
		t.Fatalf("unexpected display name %q", got)
	}
}

func TestResizeMessage_ContainsBounds(t *testing.T) {
	msg := resizeMessage(58, 18, 60, 20)
	if !strings.Contains(msg, "58x18") || !strings.Contains(msg, "60x20") {
		t.Fatalf("resize message missing current/required bounds: %q", msg)
	}
}

func TestRenderFooter_ZeroWidthIsEmpty(t *testing.T) {
	if got := renderFooter(0); got != "" {
		t.Fatalf("expected empty footer for zero width, got %q", got)
	}
}

func TestFormatEvent_ExplorationTreasureText(t *testing.T) {
	msg := formatEvent(engine.ExplorationResult{Kind: "treasure"})
	if !strings.Contains(msg, "treasure") {
		t.Fatalf("unexpected formatted event text: %q", msg)
	}
}
