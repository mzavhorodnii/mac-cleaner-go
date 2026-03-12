package ui

import (
	"fmt"
	"sort"
	"time"

	"github.com/mzavhorodnii/mac-cleaner-go/internal/cleaner"
	"github.com/mzavhorodnii/mac-cleaner-go/internal/model"
	"github.com/mzavhorodnii/mac-cleaner-go/internal/scanner"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// STYLES

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1).
			Bold(true)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#04B575")).
				Italic(true).
				MarginTop(1)

	errorMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF4672")).
				Bold(true).
				MarginTop(1)

	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#696969"))
)

// LIST ITEMS

type item struct {
	dir model.Dir
}

func (i item) Title() string       { return i.dir.Path }
func (i item) Description() string { return fmt.Sprintf("%.2f GB", i.dir.SizeGB()) }
func (i item) FilterValue() string { return i.dir.Path }

type clearAllItem struct {
	paths     []string
	totalSize int64
}

func (c clearAllItem) Title() string { return "🧹 Clear All Items" }
func (c clearAllItem) Description() string {
	totalGB := float64(c.totalSize) / 1e9
	return fmt.Sprintf("Delete all %d discovered items (%.2f GB total). Cannot be undone.", len(c.paths), totalGB)
}
func (c clearAllItem) FilterValue() string { return "clear all items" }

// MODEL & STATE

type state int

const (
	stateScanning state = iota
	stateList
)

type tuiModel struct {
	root        string
	state       state
	spinner     spinner.Model
	list        list.Model
	status      string
	isError     bool
	width       int
	height      int
	startTime   time.Time
	isCleaning  bool
	cleanStatus string
}

// MESSAGES

type scanFinishedMsg struct {
	dirs []model.Dir
	err  error
}

type cleanProgressMsg struct {
	path string
	err  error
	done bool
}

// INIT

func StartUI(root string) {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = spinnerStyle

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#25A065")).
		BorderLeftForeground(lipgloss.Color("#25A065"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#89F0C1")).
		BorderLeftForeground(lipgloss.Color("#25A065"))

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Mac Cleaner"
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(false)

	m := tuiModel{
		root:      root,
		state:     stateScanning,
		spinner:   s,
		list:      l,
		startTime: time.Now(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error starting app:", err)
	}
}

func (m tuiModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			dirs, err := scanner.Scan(m.root)
			if err == nil {
				sort.Slice(dirs, func(i, j int) bool {
					return dirs[i].Size > dirs[j].Size
				})
			}
			return scanFinishedMsg{dirs: dirs, err: err}
		},
	)
}

// UPDATE

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-3)

	case tea.KeyMsg:
		if m.isCleaning && msg.String() != "ctrl+c" {
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	switch m.state {
	case stateScanning:
		switch msg := msg.(type) {
		case scanFinishedMsg:
			if msg.err != nil {
				m.isError = true
				m.status = fmt.Sprintf("Error scanning: %v", msg.err)
				m.state = stateList
				return m, nil
			}

			var items []list.Item
			var allPaths []string
			var totalSize int64

			for _, d := range msg.dirs {
				if d.Size == 0 {
					continue
				}
				allPaths = append(allPaths, d.Path)
				totalSize += d.Size
				items = append(items, item{dir: d})
			}

			if len(allPaths) > 0 {
				items = append([]list.Item{clearAllItem{paths: allPaths, totalSize: totalSize}}, items...)
			}

			m.list.SetItems(items)
			m.state = stateList
			return m, nil

		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case stateList:
		switch msg := msg.(type) {
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)

		case tea.KeyMsg:
			if m.isCleaning {
				break
			}
			switch msg.String() {
			case "enter":
				selected := m.list.SelectedItem()
				if selectedItem, ok := selected.(item); ok {
					m.isCleaning = true
					m.isError = false
					m.cleanStatus = fmt.Sprintf("Cleaning %s...", selectedItem.dir.Path)

					cmds = append(cmds, m.spinner.Tick, func() tea.Msg {
						err := cleaner.Clean(selectedItem.dir.Path)
						return cleanProgressMsg{path: selectedItem.dir.Path, err: err, done: true}
					})

				} else if clearAll, ok := selected.(clearAllItem); ok {
					m.isCleaning = true
					m.isError = false
					m.cleanStatus = fmt.Sprintf("Cleaning %d items...", len(clearAll.paths))

					cmds = append(cmds, m.spinner.Tick, func() tea.Msg {
						for i, p := range clearAll.paths {
							err := cleaner.Clean(p)
							if err != nil {
							}
							_ = i
						}
						return cleanProgressMsg{done: true, path: "ALL"}
					})
				}
			}

		case cleanProgressMsg:
			m.isCleaning = false
			m.cleanStatus = ""

			if msg.err != nil {
				m.isError = true
				m.status = fmt.Sprintf("Error cleaning: %v", msg.err)
			} else if msg.path == "ALL" {
				m.status = "Successfully cleaned all folders! 🎉"
				m.list.SetItems(nil)
			} else {
				m.status = "Successfully cleaned: " + msg.path
				for idx, it := range m.list.Items() {
					if i, ok := it.(item); ok && i.dir.Path == msg.path {
						m.list.RemoveItem(idx)
						break
					}
				}
				m.recalculateClearAll()
			}
		}

		newList, cmd := m.list.Update(msg)
		m.list = newList
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *tuiModel) recalculateClearAll() {
	var normalItems []item
	var totalSize int64
	var paths []string

	for _, it := range m.list.Items() {
		if i, ok := it.(item); ok {
			normalItems = append(normalItems, i)
			totalSize += i.dir.Size
			paths = append(paths, i.dir.Path)
		}
	}

	if len(normalItems) == 0 {
		m.list.SetItems(nil)
		return
	}

	var newItems []list.Item
	newItems = append(newItems, clearAllItem{paths: paths, totalSize: totalSize})
	for _, it := range normalItems {
		newItems = append(newItems, it)
	}
	m.list.SetItems(newItems)
}

// VIEW

func (m tuiModel) View() string {
	switch m.state {
	case stateScanning:
		duration := time.Since(m.startTime).Round(time.Second)
		s := fmt.Sprintf("\n  %s %s Scanning %s...\n\n", m.spinner.View(), duration, m.root)
		return appStyle.Render(s)

	case stateList:
		view := m.list.View()

		var status string
		if m.isCleaning {
			status = statusMessageStyle.Render(fmt.Sprintf("%s %s", m.spinner.View(), m.cleanStatus))
		} else if m.status != "" {
			if m.isError {
				status = errorMessageStyle.Render(m.status)
			} else {
				status = statusMessageStyle.Render(m.status)
			}
		} else {
			status = statusMessageStyle.Render("Press 'enter' to clean an item, 'q' to quit")
		}

		return appStyle.Render(view + "\n" + status)
	}

	return ""
}
