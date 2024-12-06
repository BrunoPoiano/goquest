package main

import (
	"database/sql"
	"fmt"
	"main/components"
	"main/database"
	"main/models"
	"main/requests"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	debug string

	form     *huh.Form
	sent     bool // Track if request was sent
	requests models.Requests
	lg       *lipgloss.Renderer

	selected string
	preview  string
	loading  bool

	padding int
	width   int
	height  int

	db *sql.DB

	ready    bool
	viewport viewport.Model

	table table.Model
}

const useHighPerformanceRenderer = false

func initialModel(db *sql.DB) model {
	m := model{
		padding:  2,
		db:       db,
		lg:       lipgloss.DefaultRenderer(),
		sent:     false,
		requests: models.Requests{Method: "GET"},
		selected: "form",
	}

	m.table = components.Table(db)
	m.form = components.CreateForm(&m.requests)

	return m
}

func (m model) Init() tea.Cmd {

	return m.form.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {

			m.viewport = viewport.New(msg.Width, msg.Height-30)
			m.viewport.YPosition = m.height
			m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
			m.viewport.SetContent(m.preview)
			m.ready = true

		}

		if useHighPerformanceRenderer {
			cmds = append(cmds, viewport.Sync(m.viewport))
		}

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+w":

			switch m.selected {
			case "form":
				m.selected = "preview"
			case "preview":
				m.selected = "table"
			case "table":
				m.selected = "form"
			}
		case "ctrl+c":
			return m, tea.Quit

		case "enter":

			switch m.selected {
			case "preview":
			case "form":
				if m.sent {
					requestReturn, err := requests.MakeRequest(m.requests, m.db)
					if err != nil {
						m.preview = err.Error()
					} else {
						m.preview = requestReturn
						m.viewport.SetContent(requestReturn)
					}
				}

			case "table":
				request_form := models.Requests{}
				request_form.Name = m.table.SelectedRow()[1]
				request_form.Method = m.table.SelectedRow()[2]
				request_form.Route = m.table.SelectedRow()[3]
				m.requests = request_form
				m.form = components.CreateForm(&request_form)
				m.sent = false
				m.selected = "form"
			}

		case "esc":
			m.selected = "form"
			if m.sent {
				// Reset the form
				m.sent = false
				m.form = components.CreateForm(&m.requests)
				return m, m.form.Init()
			}
		}
	}

	switch m.selected {
	case "form":
		form, cmd := m.form.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		if f, ok := form.(*huh.Form); ok {
			m.form = f
			request_form := models.Requests{}
			request_form.Method = m.form.GetString("method")
			request_form.Name = m.form.GetString("name")
			request_form.Route = m.form.GetString("route")
			request_form.Params = m.form.GetString("params")
			m.requests = request_form
			send := m.form.GetBool("send")

			if send == true {
				m.loading = true
				requestReturn, err := requests.MakeRequest(request_form, m.db)
				if err != nil {
					m.preview = err.Error()
				} else {
					m.preview += "requestReturn"
					m.preview += requestReturn
					m.viewport.SetContent(requestReturn)
				}
				m.selected = "preview"
				m.sent = true
				m.loading = false
			}
		}

	case "table":
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)

	case "preview":
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

	}

	return m, tea.Batch(cmds...)
}

func widthCalc(m_width int, padding int, v_width float64) int {
	width := (float64(m_width) * v_width) - float64(padding)
	return int(width)
}

func (m model) formView(v_width float64) string {
	if m.sent {

		content := fmt.Sprintf("%s: %s\nBody: %s\n\n", m.requests.Method, m.requests.Route, m.requests.Params)
		content += "Request sent!\n\nPress ESC to create a new request\n\nPress Enter to Resend "

		width := widthCalc(m.width, m.padding, v_width)
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("1")).
			Width(width).
			Align(lipgloss.Center).
			Render(content)
	}

	width := widthCalc(m.width, m.padding, v_width)

	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(v)

	if m.selected == "form" {

		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("1")).
			Width(width).
			Render(form)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFF")).
		Width(width).
		Render(form)
}

func (m model) tableView() string {

	focused := m.table.Focused()

	content := m.table.View() + "\n  " + m.table.HelpView() + "\n " + strconv.FormatBool(focused)
	if m.selected == "table" {

		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("1")).
			Render(content)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFF")).
		Render(content)

}

func (m model) previewView(v_width float64) string {
	width := widthCalc(m.width, m.padding, v_width)
	m.viewport.Width = width - m.padding - 5

	content := ""
	if m.loading {
		content = "Loading ..."
	} else {
		content = fmt.Sprintf("%s\n%s\n%s", components.HeaderView(m.viewport), m.viewport.View(), components.FooterView(m.viewport))
	}

	if m.selected == "preview" {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("1")).
			Padding(m.padding).
			Width(width).
			Render(content)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFF")).
		Padding(m.padding).
		Width(width).
		Render(content)
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func (m model) tabsView() string {

	type MenuItem struct {
		Name     string
		Selected []string
	}

	menu := []MenuItem{
		{Name: "Main", Selected: []string{"preview", "form"}},
		{Name: "Table", Selected: []string{"table"}},
	}

	actions := "|"
	for _, item := range menu {

		menuItem := fmt.Sprintf(" %s ", item.Name)
		if contains(item.Selected, m.selected) {
			actions += lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("1")).Foreground(lipgloss.Color("#FFFFFF")).Render(menuItem)
		} else {
			actions += lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(menuItem)
		}

		actions += "|"
	}

	return actions
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	view := m.tabsView() + " \n" + m.debug + " \n"
	if m.selected == "table" {
		view += "\n\n" + m.tableView()
	} else {
		view += lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.formView(0.4),
			m.previewView(0.6),
		)
	}

	return view
}

func main() {

	db := database.SqliteDB()

	database.Migrations(db)
	defer db.Close()

	p := tea.NewProgram(
		initialModel(db),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
