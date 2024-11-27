package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	name textinput.Model
}

type model struct {
	choices  []string         // items on the to-do list
	cursor   int              // which to-do list item our cursor is pointing at
	selected map[int]struct{} // which to-do items are selected
	focused  string

	request_form RequestForm
  requests Requests

  preview string

	padding int
	width   int
	height  int
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func initialModel() model {
	return model{
		// Our to-do list is a grocery list
		choices: []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},
		padding: 2,
		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected: make(map[int]struct{}),
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}

		m.request_form.name, cmd = m.request_form.name.Update(msg)
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, cmd
}

func widthCalc(m_width int, padding int, v_width float64) int {

	width := (float64(m_width) * v_width) - float64(padding)
	return int(width)
}

func (m model) textAreaParamsView( width int) string {

  input_name := fmt.Sprintf(
		"Body \n\n%s",
		m.request_form.name.View(),
	) + "\n"

  return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color("1")).
		Padding(m.padding).
		Width(width).
		Render(input_name)
}


func (m model) inputNameView( width int) string {

  input_name := fmt.Sprintf(
		"URL\n\n%s",
		m.request_form.name.View(),
	) + "\n"

  return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color("1")).
		Width(width).
		Render(input_name)
}

func (m model) contentView(v_width float64) string {

	width := widthCalc(m.width, m.padding, v_width)

	//content := fmt.Sprintf("model width: %d | view Width: %d \n", m.width, int(width))

	return lipgloss.JoinVertical(lipgloss.Position(m.height), m.inputNameView(width), m.textAreaParamsView(width))
}

func (m model) previewView(v_width float64) string {
	
  width := widthCalc(m.width, m.padding, v_width)
 

  render := fmt.Sprintf("Preview \n\n %s", m.preview)
  return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color("1")).
		Padding(m.padding).
		Width(width).
		Render(render)
}

func (m model) View() string {
	// The header
	tea.ClearScreen()
	// Send the UI for rendering

	return lipgloss.JoinHorizontal(lipgloss.Left, m.contentView(0.5), m.previewView(0.5))

}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
