package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/ogpourya/dnsbro/internal/daemon"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item daemon.QueryEvent

func (i item) Title() string { return strings.TrimSuffix(i.Domain, ".") }
func (i item) Description() string {
	if i.Blocked {
		return fmt.Sprintf("client %s blocked (%s)", i.Client, i.Duration)
	}
	if i.Err != nil {
		return fmt.Sprintf("client %s error: %v", i.Client, i.Err)
	}
	return fmt.Sprintf("client %s via %s [%v]", i.Client, i.Upstream, i.ResponseIPs)
}
func (i item) FilterValue() string { return i.Domain }

// Model renders a small dashboard of recent queries.
type Model struct {
	list   list.Model
	events <-chan daemon.QueryEvent
	stats  Stats
}

// Stats mirrors daemon counters without embedding locks for the TUI.
type Stats struct {
	Queries  int
	Blocked  int
	Failures int
	Last     daemon.QueryEvent
}

// NewModel constructs the UI model.
func NewModel(events <-chan daemon.QueryEvent) Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "dnsbro"
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	return Model{
		list:   l,
		events: events,
	}
}

type tickMsg time.Time
type queryMsg daemon.QueryEvent

// Init subscribes to ticks and event stream.
func (m Model) Init() tea.Cmd {
	return tea.Batch(waitTick(), waitEvent(m.events))
}

func waitTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func waitEvent(ch <-chan daemon.QueryEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return nil
		}
		return queryMsg(ev)
	}
}

// Update ingests query events and ticks to refresh the UI.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg == nil {
		return m, nil
	}
	switch msg := msg.(type) {
	case tickMsg:
		return m, waitTick()
	case queryMsg:
		ev := daemon.QueryEvent(msg)
		m.list.InsertItem(0, item(ev))
		if len(m.list.Items()) > 20 {
			m.list.RemoveItem(len(m.list.Items()) - 1)
		}
		m.stats.Last = ev
		m.stats.Queries++
		if ev.Blocked {
			m.stats.Blocked++
		}
		if ev.Err != nil {
			m.stats.Failures++
		}
		return m, waitEvent(m.events)
	default:
		return m, nil
	}
}

// View renders a compact dashboard.
func (m Model) View() string {
	header := "dnsbro – DoH resolver\n"
	stats := fmt.Sprintf("queries: %d   blocked: %d   failures: %d\n", m.stats.Queries, m.stats.Blocked, m.stats.Failures)
	footer := ""
	if m.stats.Queries > 0 {
		footer = fmt.Sprintf("\nlast: %s -> %v (rcode %d)", strings.TrimSuffix(m.stats.Last.Domain, "."), m.stats.Last.ResponseIPs, m.stats.Last.RCode)
	}
	return lipgloss.NewStyle().Padding(0, 1).Render(header + stats + m.list.View() + footer)
}
