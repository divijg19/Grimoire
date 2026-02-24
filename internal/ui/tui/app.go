package tui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/divijg19/Grimoire/internal/engine"
	"github.com/divijg19/Grimoire/internal/ports"
)

type App struct {
	state *engine.State
	store ports.Store
	rng   ports.RNG
}

func NewApp(state *engine.State, store ports.Store, rng ports.RNG) *App {
	return &App{state: state, store: store, rng: rng}
}

func (a *App) Run() error {
	m := newModel(a.state, a.store, a.rng)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

type model struct {
	state *engine.State
	store ports.Store
	rng   ports.RNG

	input      textinput.Model
	viewport   viewport.Model
	logs       []string
	history    []string
	historyPos int

	width  int
	height int

	quitting bool
}

func newModel(state *engine.State, store ports.Store, rng ports.RNG) model {
	input := textinput.New()
	input.Placeholder = "Type command (help, explore, hunt 2, rest 1, use healing_potion, save, exit)"
	input.Focus()
	input.Prompt = "❯ "
	input.CharLimit = 256

	vp := viewport.New(0, 0)
	vp.SetContent("")

	m := model{
		state:      state,
		store:      store,
		rng:        rng,
		input:      input,
		viewport:   vp,
		historyPos: -1,
	}
	m.addLines(
		"Welcome to Grimoire.",
		"This is full-screen mode. Type 'help' for commands.",
	)
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			_ = m.store.Save(m.state)
			m.quitting = true
			return m, tea.Quit

		case "up":
			if len(m.history) == 0 {
				break
			}
			if m.historyPos < 0 {
				m.historyPos = len(m.history) - 1
			} else if m.historyPos > 0 {
				m.historyPos--
			}
			m.input.SetValue(m.history[m.historyPos])
			m.input.CursorEnd()
			return m, nil

		case "down":
			if len(m.history) == 0 || m.historyPos < 0 {
				break
			}
			if m.historyPos < len(m.history)-1 {
				m.historyPos++
				m.input.SetValue(m.history[m.historyPos])
			} else {
				m.historyPos = -1
				m.input.SetValue("")
			}
			m.input.CursorEnd()
			return m, nil

		case "enter":
			line := strings.TrimSpace(m.input.Value())
			if line == "" {
				return m, nil
			}

			m.history = append(m.history, line)
			m.historyPos = -1
			m.addLines(promptStyle.Render("❯ ") + line)
			m.input.SetValue("")

			if m.execute(line) {
				m.quitting = true
				return m, tea.Quit
			}
			m.layout()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *model) execute(line string) bool {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return false
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "help", "?":
		m.addLines(helpLines()...)
		return false

	case "status":
		m.addLines("Status refreshed.")
		return false

	case "explore":
		events, err := engine.Explore(m.state, m.rng)
		m.handle(events, err)
		return false

	case "hunt":
		extra := 0
		if len(args) > 0 {
			v, err := strconv.Atoi(args[0])
			if err != nil {
				m.addError("hunt expects an integer extra_sp")
				return false
			}
			extra = v
		}
		events, err := engine.Hunt(m.state, extra, m.rng)
		m.handle(events, err)
		return false

	case "rest":
		sp := 1
		if len(args) > 0 {
			v, err := strconv.Atoi(args[0])
			if err != nil {
				m.addError("rest expects an integer sp amount")
				return false
			}
			sp = v
		}
		events, err := engine.Rest(m.state, sp)
		m.handle(events, err)
		return false

	case "use":
		if len(args) == 0 {
			m.addError("usage: use <item_id>")
			return false
		}
		events, err := engine.UseItem(m.state, args[0], m.rng)
		m.handle(events, err)
		return false

	case "save":
		if err := m.store.Save(m.state); err != nil {
			m.addError("save failed: " + err.Error())
			return false
		}
		m.addLines(successStyle.Render("Game saved."))
		return false

	case "exit", "quit":
		if err := m.store.Save(m.state); err != nil {
			m.addError("save failed on exit: " + err.Error())
		} else {
			m.addLines(successStyle.Render("Game saved. Goodbye."))
		}
		return true

	default:
		m.addError("unknown command. Type 'help'.")
		return false
	}
}

func (m *model) handle(events engine.Events, err error) {
	if err != nil {
		m.addError(err.Error())
		return
	}

	if len(events) == 0 {
		m.addLines(dimStyle.Render("No events."))
	}

	for _, ev := range events {
		m.addLines(formatEvent(ev))
	}

	if saveErr := m.store.Save(m.state); saveErr != nil {
		m.addError("auto-save failed: " + saveErr.Error())
	}
}

func (m *model) addError(message string) {
	m.addLines(errorStyle.Render("Error: " + message))
}

func (m *model) addLines(lines ...string) {
	m.logs = append(m.logs, lines...)
	if len(m.logs) > 300 {
		m.logs = m.logs[len(m.logs)-300:]
	}
	m.viewport.SetContent(strings.Join(m.logs, "\n"))
	m.viewport.GotoBottom()
}

func (m *model) layout() {
	if m.width == 0 || m.height == 0 {
		return
	}
	leftWidth, _ := splitMainWidths(m.width)
	hud := renderHUDPanel(m.state, leftWidth)
	hudHeight := lipgloss.Height(hud)

	inputHeight := lipgloss.Height(inputPanelStyle.Width(max(18, leftWidth)).Render(m.input.View()))
	footerHeight := lipgloss.Height(footerStyle.Render("x"))

	logHeight := m.height - hudHeight - inputHeight - footerHeight
	if logHeight < 4 {
		logHeight = 4
	}

	innerLogWidth := max(10, leftWidth-logPanelStyle.GetHorizontalFrameSize())
	m.viewport.Width = innerLogWidth
	m.viewport.Height = logHeight
	m.viewport.SetContent(strings.Join(m.logs, "\n"))
	m.viewport.GotoBottom()
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}
	if m.width < 60 || m.height < 16 {
		return warnStyle.Render("Terminal too small. Resize to at least 60x16.")
	}

	leftWidth, rightWidth := splitMainWidths(m.width)

	hud := renderHUDPanel(m.state, leftWidth)

	logTitle := titleStyle.Render("Event Log")
	logPane := logPanelStyle.Width(max(18, leftWidth)).Render(logTitle + "\n" + m.viewport.View())

	footer := footerStyle.Render("Enter: run  •  ↑/↓: history  •  Ctrl+C: save & quit")
	input := inputPanelStyle.Width(max(18, leftWidth)).Render(m.input.View())

	leftColumn := lipgloss.JoinVertical(lipgloss.Left, hud, logPane, input)
	rightPanel := renderInventoryPanel(m.state, rightWidth, lipgloss.Height(leftColumn))
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, " ", rightPanel)

	return lipgloss.JoinVertical(lipgloss.Left, mainRow, footer)
}

func splitMainWidths(totalWidth int) (int, int) {
	gap := 1
	right := totalWidth / 3
	if right < 24 {
		right = 24
	}
	if right > 36 {
		right = 36
	}

	left := totalWidth - right - gap
	if left < 30 {
		left = 30
		right = max(24, totalWidth-left-gap)
	}

	return left, right
}

func renderHUDPanel(state *engine.State, width int) string {
	p := state.Player
	need := engine.XPToNext(p.Level)

	lines := []string{
		titleStyle.Render(fmt.Sprintf("%s (%s)", p.Name, p.Class)),
		dimStyle.Render(state.Meta.Location),
		"",
		fmt.Sprintf("Level %d", p.Level),
		fmt.Sprintf("HP %d/%d %s", p.HP, p.MaxHP, ratioBar(p.HP, p.MaxHP, 18)),
		fmt.Sprintf("SP %d %s", p.SP, simpleBar(min(p.SP, 12), 12, 12)),
		fmt.Sprintf("XP %d/%d %s", p.XP, need, ratioBar(p.XP, need, 18)),
		fmt.Sprintf("Gold %d", p.Gold),
		fmt.Sprintf("Commands %d", state.Meta.CommandCount),
	}
	return sidePanelStyle.Width(width).Render(strings.Join(lines, "\n"))
}

func renderInventoryPanel(state *engine.State, width, height int) string {
	lines := []string{titleStyle.Render("Inventory"), ""}
	if len(state.Player.Inventory) == 0 {
		lines = append(lines, dimStyle.Render("(empty)"))
		return sidePanelStyle.Width(width).Height(max(4, height)).Render(strings.Join(lines, "\n"))
	}

	keys := make([]string, 0, len(state.Player.Inventory))
	for k := range state.Player.Inventory {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		count := state.Player.Inventory[k]
		itemName := itemDisplayName(k)
		lines = append(lines, fmt.Sprintf("• %s x%d", itemName, count))
	}

	return sidePanelStyle.Width(width).Height(max(4, height)).Render(strings.Join(lines, "\n"))
}

func helpLines() []string {
	return []string{
		titleStyle.Render("Commands"),
		"  help                Show this help",
		"  status              Show current HUD",
		"  explore             Explore once",
		"  hunt [extra_sp]     Hunt with optional SP stake",
		"  rest [sp]           Convert SP to HP (default 1)",
		"  use <item_id>       Use item, e.g. healing_potion",
		"  save                Save game",
		"  exit | quit         Save and exit",
	}
}

func formatEvent(e engine.Event) string {
	switch ev := e.(type) {
	case engine.ExplorationResult:
		switch ev.Kind {
		case "treasure":
			return successStyle.Render("You uncover a hidden treasure cache.")
		case "item":
			return infoStyle.Render("You find a useful item.")
		case "gold":
			return infoStyle.Render("You discover scattered gold.")
		default:
			return dimStyle.Render("The path yields nothing this time.")
		}
	case engine.EncounterStarted:
		return warnStyle.Render("Encounter: " + prettyID(ev.EnemyID))
	case engine.DamageDealt:
		if ev.Target == "player" {
			return errorStyle.Render(fmt.Sprintf("You take %d damage (%d HP left)", ev.Amount, ev.HPLeft))
		}
		return successStyle.Render(fmt.Sprintf("You deal %d damage (%d enemy HP left)", ev.Amount, ev.HPLeft))
	case engine.EnemyDefeated:
		return successStyle.Render(fmt.Sprintf("Defeated %s • +%d XP • +%d gold", prettyID(ev.EnemyID), ev.XP, ev.Gold))
	case engine.PlayerDefeated:
		return errorStyle.Render("You were defeated.")
	case engine.XPGained:
		return infoStyle.Render(fmt.Sprintf("+%d XP", ev.Amount))
	case engine.LevelUp:
		return successStyle.Bold(true).Render(fmt.Sprintf("Level up! Now level %d (Max HP %d)", ev.NewLevel, ev.NewMaxHP))
	case engine.ItemAdded:
		return infoStyle.Render(fmt.Sprintf("Obtained %s x%d", itemDisplayName(ev.ItemID), ev.Count))
	case engine.ItemRemoved:
		return dimStyle.Render(fmt.Sprintf("Used %s x%d", itemDisplayName(ev.ItemID), ev.Count))
	case engine.GoldGained:
		return successStyle.Render(fmt.Sprintf("+%d gold", ev.Amount))
	case engine.SPSpent:
		return dimStyle.Render(fmt.Sprintf("Spent %d SP", ev.Amount))
	case engine.HPRestored:
		return successStyle.Render(fmt.Sprintf("Restored %d HP", ev.Amount))
	default:
		return dimStyle.Render("Event: " + e.EventType())
	}
}

func itemDisplayName(itemID string) string {
	if it, ok := engine.Items[itemID]; ok {
		return it.Name
	}
	return prettyID(itemID)
}

func prettyID(id string) string {
	return strings.Title(strings.ReplaceAll(id, "_", " "))
}

func simpleBar(cur, maxV, width int) string {
	if maxV <= 0 {
		return "[" + strings.Repeat(" ", width) + "]"
	}
	filled := cur * width / maxV
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("●", filled) + strings.Repeat(" ", width-filled) + "]"
}

func ratioBar(cur, maxV, width int) string {
	if maxV <= 0 {
		return "[" + strings.Repeat(" ", width) + "]"
	}
	filled := cur * width / maxV
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	promptStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)

	sidePanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1)

	logPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1)

	inputPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)
