package main

import (
	"database/sql"
	"fmt"
	"main/database"
	"main/models"
	"main/requests"
	"net/url"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	form     *huh.Form
	requests models.Requests
	lg       *lipgloss.Renderer
	preview  string
	padding  int
	width    int
	height   int
	sent     bool // Track if request was sent
	db       *sql.DB
	loading  bool

	ready    bool
	viewport viewport.Model
}

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)
const useHighPerformanceRenderer = false

func initialModel(db *sql.DB) model {
	m := model{
		padding:  2,
		db:       db,
		lg:       lipgloss.DefaultRenderer(),
		sent:     false,
		requests: models.Requests{Method: "GET"},
	}

	m.form = createForm(&m.requests)

	return m
}

func createForm(rf *models.Requests) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("method").
				Title("Method").
				Options(
					huh.NewOption("GET", "GET"),
					huh.NewOption("POST", "POST"),
					huh.NewOption("PUT", "PUT"),
					huh.NewOption("DELETE", "DELETE"),
				).
				Value(&rf.Method),
			huh.NewInput().
				Key("name").
				Value(&rf.Name).
				Title("name"),

			huh.NewInput().
				Key("route").
				Title("URL").
				Value(&rf.Route).
				Validate(func(s string) error {
					if s == "" {
						return nil
					}
					if _, err := url.Parse(s); err != nil {
						return fmt.Errorf("invalid URL")
					}
					return nil
				}),
			huh.NewText().
				Key("params").
				Value(&rf.Params).
				Title("Body"),

			huh.NewConfirm().
				Key("send").
				Title("Send Request?").
				Affirmative("Send"),
		),
	).
		WithWidth(45).
		WithShowHelp(true).
		WithShowErrors(true)
}

func (m model) Init() tea.Cmd {
	return m.form.Init()
}




func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

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
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			if m.sent {
				requestReturn, err := requests.MakeRequest(m.requests, m.db)
				if err != nil {
					m.preview = err.Error()
				} else {
					m.preview = requestReturn
					m.viewport.SetContent(requestReturn)
				}
			}
		case "esc":
			if m.sent {
				// Reset the form
				m.sent = false
				m.form = createForm(&m.requests)
				return m, m.form.Init()
			}
		}
	}

	// Handle form updates
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
			requestReturn, err := requests.MakeRequest(request_form, m.db)
			if err != nil {
				m.preview = err.Error()
			} else {
        m.preview += "requestReturn"
				m.preview += requestReturn
				m.viewport.SetContent(requestReturn)
			}
			m.sent = true
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func widthCalc(m_width int, padding int, v_width float64) int {
	width := (float64(m_width) * v_width) - float64(padding)
	return int(width)
}

func (m model) formView(v_width float64) string {
	if m.sent {
		width := widthCalc(m.width, m.padding, v_width)
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("1")).
			Width(width).
			Align(lipgloss.Center).
			Render("Request sent!\n\nPress ESC to create a new reques\n\nPress Enter to Resend")
	}

	width := widthCalc(m.width, m.padding, v_width)

	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(v)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("1")).
		Width(width).
		Render(form)
}

func (m model) headerView() string {
	title := titleStyle.Render("Preview")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (m model) previewView(v_width float64) string {
	width := widthCalc(m.width, m.padding, v_width)
	m.viewport.Width = width - m.padding - 5

	content := fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())


	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("1")).
		Padding(m.padding).
		Width(width).
		Render(content)
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.formView(0.4),
		m.previewView(0.6),
	)
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
