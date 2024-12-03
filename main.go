package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type Requests struct {
	id     int
	name   string
	method string
	route  string
	params url.Values
}

type RequestForm struct {
	name   string
	body   string
	method string
	send   bool
}

type model struct {
	form         *huh.Form
	request_form RequestForm
	requests     Requests
	lg           *lipgloss.Renderer
	preview      string
	padding      int
	width        int
	height       int
	sent         bool // Track if request was sent
}

func initialModel() model {
	m := model{
		padding: 2,
		lg:      lipgloss.DefaultRenderer(),
		request_form: RequestForm{
			method: "GET",
		},
		sent: false,
	}

	m.form = createForm(&m.request_form)
	return m
}

func createForm(rf *RequestForm) *huh.Form {
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
				Value(&rf.method),
			huh.NewInput().
				Key("name").
				Title("URL").
				Value(&rf.name).
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
				Key("body").
				Value(&rf.body).
				Title("Body"),
			huh.NewConfirm().
				Key("send").
				Title("Send Request?").
				Affirmative("Send").
				Negative("Cancel").
				Value(&rf.send),
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

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			if m.sent {
				m.preview += "Resended \n"
				return m, tea.Batch(cmds...)

			}
		case "esc":
			if m.sent {
				// Reset the form
				m.sent = false
				m.form = createForm(&m.request_form)
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
		m.request_form.method = m.form.GetString("method")
		m.request_form.name = m.form.GetString("name")
		m.request_form.body = m.form.GetString("body")
		m.request_form.send = m.form.GetBool("send")

		if m.request_form.send == true {
			m.sent = true
		}
	}
	// Update preview for in-progress form
	m.updatePreview()
	return m, tea.Batch(cmds...)
}

func (m *model) updatePreview() {
	preview := "Current Request:\n\n"

	preview += fmt.Sprintf("Method: %s\n", m.request_form.method)
	preview += fmt.Sprintf("URL: %s\n", m.request_form.name)
	preview += fmt.Sprintf("Body: %s\n", m.request_form.body)
	preview += fmt.Sprintf("Send: %s\n", m.request_form.send)

	m.preview = preview
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

func (m model) previewView(v_width float64) string {
	width := widthCalc(m.width, m.padding, v_width)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("1")).
		Padding(m.padding).
		Width(width).
		Render(m.preview)
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
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
