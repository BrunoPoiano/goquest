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
	name string
	body string
	send bool
}

type model struct {
	choices      []string
	cursor       int
	selected     map[int]struct{}
	focused      string
	request_form RequestForm
	requests     Requests
	form         *huh.Form
	lg          *lipgloss.Renderer
	preview      string
	padding      int
	width        int
	height       int
}

func (m model) Init() tea.Cmd {
	return m.form.Init()
}

func initialModel() model {
	m := model{
		choices:      []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},
		padding:      2,
		selected:     make(map[int]struct{}),
		focused:      "name",
		lg:          lipgloss.DefaultRenderer(),
		request_form: RequestForm{},
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("name").
				Title("URL").
				Value(&m.request_form.name).
				Prompt("?"),
			huh.NewText().
				Key("body").
				Value(&m.request_form.body).
				Title("Body"),
			huh.NewConfirm().
				Key("done").
				Title("All done?").
				Affirmative("Send").
				Value(&m.request_form.send),
		),
	).
		WithWidth(45).
		WithShowHelp(false).
		WithShowErrors(false)

	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Handle form updates
	form, formCmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		
		// Update preview whenever the form updates
		m.preview = fmt.Sprintf("URL: %s\nBody: %s\nSend: %v",
			m.request_form.name,
			m.request_form.body,
			m.request_form.send,
		)

		// If send is true, handle the submission
		if m.request_form.send {
			m.preview = fmt.Sprintf("Sending request...\nURL: %s\nBody: %s",
				m.request_form.name,
				m.request_form.body,
			)
			// Reset the send flag
			m.request_form.send = false
			
			// Create a new form with the same fields but reset state
			m.form = huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Key("name").
						Title("URL").
						Value(&m.request_form.name).
						Prompt("?"),
					huh.NewText().
						Key("body").
						Value(&m.request_form.body).
						Title("Body"),
					huh.NewConfirm().
						Key("done").
						Title("All done?").
						Affirmative("Send").
						Value(&m.request_form.send),
				),
			).
				WithWidth(45).
				WithShowHelp(false).
				WithShowErrors(false)
		}
	}

	return m, tea.Batch(cmd, formCmd)
}

func widthCalc(m_width int, padding int, v_width float64) int {
	width := (float64(m_width) * v_width) - float64(padding)
	return int(width)
}

func (m model) formView(v_width float64) string {
	width := widthCalc(m.width, m.padding, v_width)
	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(v)
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("1")).
		Width(width)
	return style.Render(form)
}

func (m model) previewView(v_width float64) string {
	width := widthCalc(m.width, m.padding, v_width)
	render := fmt.Sprintf("Preview\n\n%s", m.preview)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("1")).
		Padding(m.padding).
		Width(width).
		Render(render)
}

func (m model) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Left, m.formView(0.3), m.previewView(0.7))
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
