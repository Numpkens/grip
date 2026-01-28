// Package main provides a Terminal User Interface (TUI) client for GRIP.
//
// It replicates the Rosé Pine Moon aesthetic of the web interface,
// featuring a dynamic grid layout, responsive search, and real-time 
// performance metrics for engine latency.
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

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	colorBase    = "#232136"
	colorSurface = "#2a273f"
	colorRose    = "#ea9a97"
	colorGold    = "#f6c177"
	colorPine    = "#3e8fb0"
	colorText    = "#e0def4"
	colorMuted   = "#6e6a86"
)

// keyMap defines the keyboard shortcuts for the application navigation.
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Search key.Binding
	Enter  key.Binding
	Quit   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Search, k.Enter, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down}, {k.Search, k.Enter, k.Quit}}
}

var keys = keyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Search: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

var (
	cardWidth  = 34
	cardHeight = 10

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorRose)).
			Bold(true).
			MarginTop(1)

	searchLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorGold)).
				Bold(true).
				MarginRight(1)

	latencyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorPine)).
			Italic(true)

	cardStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colorSurface)).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorMuted)).
			Padding(1, 2).
			Width(cardWidth).
			Height(cardHeight)

	activeStyle = cardStyle.Copy().
			BorderForeground(lipgloss.Color(colorRose)).
			Border(lipgloss.ThickBorder())

	helpStyle = lipgloss.NewStyle().MarginTop(1).MarginLeft(2).PaddingBottom(1)
)

type resultsMsg struct {
	posts   []logic.Post
	latency time.Duration
}

type model struct {
	engine      *logic.Engine
	posts       []logic.Post
	latency     time.Duration
	cursor      int
	loading     bool
	spinner     spinner.Model
	viewport    viewport.Model
	searchInput textinput.Model
	help        help.Model
	keys        keyMap
	searching   bool
	ready       bool
	width       int
	height      int
}

func fetchCmd(m model) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		res := m.engine.Collect(context.Background(), m.searchInput.Value())
		return resultsMsg{
			posts:   res,
			latency: time.Since(start),
		}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchCmd(m))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "enter":
				m.searching = false
				m.loading = true
				m.cursor = 0
				m.searchInput.Blur()
				return m, fetchCmd(m)
			case "esc":
				m.searching = false
				m.searchInput.Blur()
				return m, nil
			}
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Search):
			m.searching = true
			m.searchInput.Focus()
			return m, nil
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
				m.viewport.LineUp(1)
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.posts)-1 {
				m.cursor++
				m.viewport.LineDown(1)
			}
		case key.Matches(msg, m.keys.Enter):
			if len(m.posts) > 0 {
				launchBrowser(m.posts[m.cursor].URL)
			}
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		headerHeight := 16
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-headerHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerHeight
		}
		m.help.Width = msg.Width

	case resultsMsg:
		m.posts = msg.posts
		m.latency = msg.latency
		m.loading = false
		m.cursor = 0
		m.viewport.YOffset = 0
		m.viewport.SetContent(m.renderGrid())

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Only update content if we aren't loading, to reflect cursor changes
	if m.ready && !m.loading && len(m.posts) > 0 {
		m.viewport.SetContent(m.renderGrid())
	}

	// Process viewport updates (handles mouse wheel automatically)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) renderGrid() string {
	if len(m.posts) == 0 {
		return lipgloss.Place(m.width, 10, lipgloss.Center, lipgloss.Center, "NO_DATA_RETURNED")
	}

	cols := m.width / (cardWidth + 4)
	if cols < 1 {
		cols = 1
	}

	var rows []string
	var currentRow []string

	for i, p := range m.posts {
		style := cardStyle
		if i == m.cursor {
			style = activeStyle
		}

		meta := lipgloss.NewStyle().Foreground(lipgloss.Color(colorGold)).
			Render(fmt.Sprintf("%s\n[%s]", p.PublishedAt.Format("02 Jan 2006"), strings.ToUpper(p.Source)))
		title := lipgloss.NewStyle().Foreground(lipgloss.Color(colorText)).Bold(true).Render(p.Title)

		cardContent := lipgloss.JoinVertical(lipgloss.Left, meta, "\n", title)
		currentRow = append(currentRow, style.Render(cardContent))

		if len(currentRow) == cols || i == len(m.posts)-1 {
			rowStr := lipgloss.JoinHorizontal(lipgloss.Top, currentRow...)
			rows = append(rows, lipgloss.PlaceHorizontal(m.width, lipgloss.Center, rowStr))
			currentRow = []string{}
		}
	}
	return lipgloss.JoinVertical(lipgloss.Center, rows...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	latStr := ""
	if m.latency > 0 {
		latStr = latencyStyle.Render(fmt.Sprintf("LATENCY: %v", m.latency))
	}
	topBar := lipgloss.PlaceHorizontal(m.width, lipgloss.Right, latStr)

	largeTitle := titleStyle.Render(" ██████╗ ██████╗ ██╗██████╗ \n ██╔════╝ ██╔══██╗██║██╔══██╗\n ██║  ███╗██████╔╝██║██████╔╝\n ██║   ██║██╔══██╗██║██╔═══╝ \n ╚██████╔╝██║  ██║██║██║     \n  ╚═════╝ ╚═╝  ╚═╝╚═╝╚═╝     ")
	headerTitle := lipgloss.Place(m.width, 6, lipgloss.Center, lipgloss.Center, largeTitle)

	searchBar := lipgloss.JoinHorizontal(lipgloss.Center,
		searchLabelStyle.Render("SEARCH:"),
		m.searchInput.View(),
	)
	centeredSearch := lipgloss.Place(m.width, 3, lipgloss.Center, lipgloss.Center, searchBar)

	var mainContent string
	if m.loading {
		mainContent = lipgloss.Place(m.width, m.viewport.Height, lipgloss.Center, lipgloss.Center,
			fmt.Sprintf("%s Scoping sources...", m.spinner.View()))
	} else {
		mainContent = m.viewport.View()
	}

	helpView := helpStyle.Render(m.help.View(m.keys))

	return lipgloss.JoinVertical(lipgloss.Left, topBar, headerTitle, centeredSearch, mainContent, helpView)
}

func launchBrowser(url string) {
	cmd := "xdg-open"
	if runtime.GOOS == "windows" {
		cmd = "rundll32"
	} else if runtime.GOOS == "darwin" {
		cmd = "open"
	}
	args := []string{url}
	if runtime.GOOS == "windows" {
		args = append([]string{"url.dll,FileProtocolHandler"}, args...)
	}
	_ = exec.Command(cmd, args...).Start()
}

func main() {
	client := &http.Client{Timeout: 10 * time.Second}
	engine := logic.NewEngine([]logic.Source{
		&sources.DevTo{Client: client}, &sources.HackerNews{Client: client},
		&sources.Hashnode{Client: client}, &sources.BootDev{Client: client},
		&sources.Lobsters{Client: client}, &sources.FreeCodeCamp{Client: client},
	})

	ti := textinput.New()
	ti.Placeholder = "type and press enter..."
	ti.SetValue("golang")
	ti.CharLimit = 50

	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(lipgloss.Color(colorGold))
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(lipgloss.Color(colorText))
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))

	spin := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(colorRose))

	m := model{
		engine:      engine,
		loading:     true,
		spinner:     spin,
		searchInput: ti,
		keys:        keys,
		help:        h,
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Program error: %v\n", err)
		os.Exit(1)
	}
}