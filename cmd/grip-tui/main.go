// Package main provides a TUI client for GRIP.
// It replicates the Rosé Pine Moon aesthetic of the web interface
// and consumes the core logic engine directly.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/Numpkens/grip/internal/logic"
	"github.com/Numpkens/grip/internal/logic/sources"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Rosé Pine Moon palette - keep these in sync with index.html
const (
	colorSurface = "#2a273f"
	colorRose    = "#ea9a97"
	colorGold    = "#f6c177"
	colorPine    = "#3e8fb0"
	colorText    = "#e0def4"
	colorMuted   = "#6e6a86"
)

var (
	// cardStyle defines the look of an article entry
	cardStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colorSurface)).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorPine)).
			Padding(1, 2).
			MarginLeft(2).
			MarginBottom(1).
			Width(75)

	// activeStyle is our TUI version of a CSS :hover
	activeStyle = cardStyle.Copy().
			BorderForeground(lipgloss.Color(colorRose)).
			Border(lipgloss.ThickBorder())
)

// resultsMsg is passed when the engine finishes its fan-out
type resultsMsg []logic.Post

type model struct {
	engine  *logic.Engine
	posts   []logic.Post
	cursor  int
	loading bool
	spinner spinner.Model
	query   string
}

// fetchCmd wraps the engine call to avoid blocking the UI thread
func fetchCmd(m model) tea.Cmd {
	return func() tea.Msg {
		// engine.Collect handles the 2s timeout and min-heap logic
		res := m.engine.Collect(context.Background(), m.query)
		return resultsMsg(res)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchCmd(m))
}

// Update handles our "Handwritten" logic for HJKL and Mouse support
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		// Vertical movement: JK + Arrows
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.posts)-1 {
				m.cursor++
			}

		case "enter":
			if len(m.posts) > 0 {
				launchBrowser(m.posts[m.cursor].URL)
			}
		}

	case tea.MouseMsg:
		// Mouse wheel integration
		if msg.Type == tea.MouseWheelUp && m.cursor > 0 {
			m.cursor--
		} else if msg.Type == tea.MouseWheelDown && m.cursor < len(m.posts)-1 {
			m.cursor++
		}

	case resultsMsg:
		m.posts = msg
		m.loading = false

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		// Explicitly capture both arrows and hjkl
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.posts)-1 {
				m.cursor++
			}

		case "enter":
			if len(m.posts) > 0 {
				launchBrowser(m.posts[m.cursor].URL)
			}
		}

	case tea.WindowSizeMsg:
		// Dynamic resizing: update card width to fit terminal
		cardStyle = cardStyle.Width(msg.Width - 10)
		activeStyle = activeStyle.Width(msg.Width - 10)

	case resultsMsg:
		m.posts = msg
		m.loading = false

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	if m.loading {
		return fmt.Sprintf("\n  %s Scoping sources for '%s'...\n", m.spinner.View(), m.query)
	}

	var b strings.Builder
	b.WriteString(headerStyle.Render(" GRIP // BLOG AGGREGATOR") + "\n\n")

	for i, p := range m.posts {
		style := cardStyle
		prefix := "  "

		if i == m.cursor {
			style = activeStyle
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color(colorRose)).Render("→ ")
		}

		// Use lipgloss.JoinVertical for better alignment
		meta := lipgloss.NewStyle().Foreground(lipgloss.Color(colorGold)).
			Render(fmt.Sprintf("%s | %s", p.PublishedAt.Format("02 Jan"), p.Source))
		
		title := lipgloss.NewStyle().Foreground(lipgloss.Color(colorText)).
			Bold(true).Render(p.Title)

		// Join content before putting it in the styled box
		content := lipgloss.JoinVertical(lipgloss.Left, meta, "", title)
		
		b.WriteString(fmt.Sprintf("%s%s\n", prefix, style.Render(content)))
	}

	return b.String()
}
func main() {
	client := &http.Client{Timeout: 10 * time.Second}
	
	// Direct engine assembly
	engine := logic.NewEngine([]logic.Source{
		&sources.DevTo{Client: client},
		&sources.HackerNews{Client: client},
		&sources.Hashnode{Client: client},
		&sources.BootDev{Client: client},
		&sources.Lobsters{Client: client},
		&sources.FreeCodeCamp{Client: client},
	})

	spin := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(colorRose))

	m := model{
		engine:  engine,
		query:   "golang", 
		loading: true,
		spinner: spin,
	}

	// We use WithAltScreen to keep the user's terminal scrollback clean
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Program error: %v\n", err)
		os.Exit(1)
	}
}