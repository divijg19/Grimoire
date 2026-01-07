package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/divijg19/Grimoire/internal/engine"
)

// ================================
// Color handling
// ================================

var noColor = os.Getenv("NO_COLOR") != ""

type color string

const (
	red     color = "\x1b[31m"
	green   color = "\x1b[32m"
	yellow  color = "\x1b[33m"
	blue    color = "\x1b[34m"
	magenta color = "\x1b[35m"
	cyan    color = "\x1b[36m"
	bold    color = "\x1b[1m"
	dim     color = "\x1b[2m"
	reset   color = "\x1b[0m"
)

func c(s string, clr color) string {
	if noColor {
		return s
	}
	return string(clr) + s + string(reset)
}

// cs applies multiple color/format codes then a single reset.
func cs(s string, clrs ...color) string {
	if noColor {
		return s
	}
	pref := ""
	for _, cl := range clrs {
		pref += string(cl)
	}
	return pref + s + string(reset)
}

// ================================
// Symbols
// ================================

const (
	heart = "♥"
	spark = "⚡"
)

// ================================
// HUD
// ================================

func RenderHUD(state *engine.State) {
	p := state.Player
	width := 64

	hr := "+" + repeat("-", width-2) + "+"
	title := fmt.Sprintf(" %s (%s) - Lv %d ", p.Name, p.Class, p.Level)
	loc := state.Meta.Location

	//header := "|" + padRight(title, width-2-len(loc)) + loc + "|"

	fmt.Println(c(hr, cyan))
	// keep padding calculations on raw strings, then apply bold+cyan to the
	// title portion when printing so alignment stays correct.
	leftRaw := padRight(title, width-2-len(loc))
	fmt.Println(cs("|"+leftRaw, bold, cyan) + cs(loc+"|", cyan))

	// HP bar
	hpBar := bar(p.HP, p.MaxHP, 30)
	hpLine := fmt.Sprintf(
		"| HP %s %s %d/%d",
		hpBar, heart, p.HP, p.MaxHP,
	)
	fmt.Println(colorByRatio(hpLine, p.HP, p.MaxHP))

	// SP bar
	spBar := "[" + repeat("●", min(p.SP, 12)) + repeat(" ", max(0, 12-p.SP)) + "]"
	spLine := fmt.Sprintf("| SP %s %s %d", spBar, spark, p.SP)
	fmt.Println(cs(padRight(spLine, width-1)+"|", magenta, bold))

	// XP
	need := engine.XPToNext(p.Level)
	xpBar := bar(p.XP, need, 30)
	xpLine := fmt.Sprintf("| XP %s %d/%d", xpBar, p.XP, need)
	fmt.Println(cs(padRight(xpLine, width-1)+"|", blue, bold))

	// Resources
	res := fmt.Sprintf(
		"| Gold: %d | Commands: %d",
		p.Gold, state.Meta.CommandCount,
	)
	fmt.Println(cs(padRight(res, width-1)+"|", cyan, bold))

	// Inventory
	fmt.Println(cs("| Inventory:", bold, cyan))
	if len(p.Inventory) == 0 {
		fmt.Println(c("|  (empty)", dim))
	} else {
		renderInventory(p, width)
	}

	fmt.Println(c(hr, cyan))
}

// RenderHP prints a compact view showing only the HP line (used after commands)
func RenderHP(state *engine.State) {
	p := state.Player
	width := 64

	hr := "+" + repeat("-", width-2) + "+"
	title := fmt.Sprintf(" %s (%s) - Lv %d ", p.Name, p.Class, p.Level)
	loc := state.Meta.Location

	leftRaw := padRight(title, width-2-len(loc))
	// compact header with bold title
	fmt.Println(c(hr, cyan))
	fmt.Println(cs("|"+leftRaw, bold, cyan) + cs(loc+"|", cyan))

	// HP line only
	hpBar := bar(p.HP, p.MaxHP, 30)
	hpLine := fmt.Sprintf("| HP %s %s %d/%d", hpBar, heart, p.HP, p.MaxHP)
	fmt.Println(colorByRatio(hpLine, p.HP, p.MaxHP))

	fmt.Println(c(hr, cyan))
}

// ================================
// Inventory
// ================================

func renderInventory(p engine.Player, width int) {
	type entry struct {
		id    string
		count int
	}

	var items []entry
	for id, cnt := range p.Inventory {
		items = append(items, entry{id, cnt})
	}

	// keep inventory display deterministic
	sort.Slice(items, func(i, j int) bool { return items[i].id < items[j].id })

	colW := (width - 6) / 2
	for i := 0; i < len(items); i += 2 {
		left := fmt.Sprintf("  - %s x%d", items[i].id, items[i].count)
		right := ""
		if i+1 < len(items) {
			right = fmt.Sprintf("  - %s x%d", items[i+1].id, items[i+1].count)
		}
		line := padRight(left, colW) + "  " + padRight(right, colW)
		fmt.Println(c("|"+padRight(line, width-2)+"|", dim))
	}
}

// ================================
// Helpers
// ================================

func bar(cur, max, w int) string {
	if max <= 0 {
		return "[" + repeat(" ", w) + "]"
	}
	filled := (cur * w) / max
	return "[" + repeat("█", filled) + repeat(" ", w-filled) + "]"
}

func colorByRatio(line string, cur, max int) string {
	ratio := float64(cur) / float64(max)
	switch {
	case ratio < 0.4:
		return c(padRight(line, 63)+"|", red)
	case ratio < 0.75:
		return c(padRight(line, 63)+"|", yellow)
	default:
		return c(padRight(line, 63)+"|", green)
	}
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func padRight(s string, n int) string {
	for len(s) < n {
		s += " "
	}
	return s
}
