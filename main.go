package main

import (
	"database/sql"
	"flag"
	"fmt"
	"main/components"
	"main/controllers"
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

func initialModel(db *sql.DB, item models.Requests) model {
	m := model{
		padding:  2,
		db:       db,
		lg:       lipgloss.DefaultRenderer(),
		sent:     false,
		requests: item,
		selected: "form",
	}

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

	case models.ReturnRequest:
		if msg.Error != nil {
			m.preview = msg.Response
			m.viewport.SetContent(msg.Error.Error())
		} else {
			m.preview = msg.Response
			m.viewport.SetContent(msg.Response)
		}
		m.selected = "preview"
		m.loading = false

	case models.ReturnTable:
		m.table = msg.Table
		m.selected = "table"
		m.loading = false

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {

			m.viewport = viewport.New(msg.Width, msg.Height-15)
			m.viewport.YPosition = m.height
			m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
			m.viewport.SetContent(m.preview)
      m.viewport.GotoTop()
      m.ready = true

		}

		if m.selected == "table" {
			cmd = components.Table(m.db, m.widthCalc(1), m.height)
			cmds = append(cmds, cmd)
		}

		if useHighPerformanceRenderer {
			cmds = append(cmds, viewport.Sync(m.viewport))
		}

	case tea.KeyMsg:
		switch msg.String() {

		case "D":
			switch m.selected {
			case "table":
				err := controllers.DeleteItemFromTable(m.db, m.table.SelectedRow()[0])
				if err != nil {
					m.preview = err.Error()
				} else {
					m.viewport.SetContent("Item Deleted Successifully")
					m.selected = "preview"
				}
			}
		case "ctrl+w":
			switch m.selected {
			case "form":
				m.selected = "preview"
			case "preview":
				m.loading = true
				cmd = components.Table(m.db, m.widthCalc(1), m.height)
				cmds = append(cmds, cmd)
			case "table":
				m.selected = "form"
			}

		case "ctrl+n":
			switch m.selected {
			case "form", "preview":
				empty_form := models.Requests{}
				m.requests = empty_form
				m.form = components.CreateForm(&empty_form)
				m.selected = "form"
				m.sent = false

			}

		case "ctrl+c":
			return m, tea.Quit

		case "enter":

			switch m.selected {
			case "preview", "form":
				if m.sent {
					m.loading = true
					cmd = requests.MakeRequest(m.requests, m.db)
					cmds = append(cmds, cmd)
				}

			case "table":
				request_form := models.Requests{}
				request_form.Id, _ = strconv.Atoi(m.table.SelectedRow()[0])
				request_form.Name = m.table.SelectedRow()[1]
				request_form.Method = m.table.SelectedRow()[2]
				request_form.Route = m.table.SelectedRow()[3]
				request_form.Params = m.table.SelectedRow()[4]
				request_form.Headers = m.table.SelectedRow()[5]
				m.requests = request_form
				m.form = components.CreateForm(&request_form)
				m.sent = false
				m.selected = "form"
			}

		case "esc":
			m.selected = "form"
			m.sent = false
			m.form = components.CreateForm(&m.requests)
			return m, m.form.Init()

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
			request_form.Headers = m.form.GetString("headers")
			send := m.form.GetBool("send")

			if send == true {
				m.loading = true
				cmd = requests.MakeRequest(m.checkForm(request_form), m.db)
				cmds = append(cmds, cmd)
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

func (m model) checkForm(request_form models.Requests) models.Requests {

	if m.requests.Id == 0 {
		return request_form
	}

	checked_form := m.requests
	checked_form.Name = request_form.Name
	checked_form.Route = request_form.Route

	if len(request_form.Params) < 400 {
		checked_form.Params = request_form.Params
	}

	/*
		if len(request_form.Headers) < 400 {
			checked_form.Headers = request_form.Headers
		}
	*/

	return checked_form

}

func (m model) widthCalc(v_width float64) int {
	width := (float64(m.width) * v_width) - float64(m.padding)
	return int(width)
}

func (m model) formView(v_width float64) string {

	width := m.widthCalc(v_width)
	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(v + " ⸱ esc return to form ⸱ crtl+n to new form")

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

	content := ""

	if m.loading {
		content = "Loading ... "
	} else {
		content = fmt.Sprintf("%s \n %s D to delete row ", m.table.View(), m.table.HelpView())
	}
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
	width := m.widthCalc(v_width)
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

	view := "\n"
	//view = fmt.Sprintf("selected %s \n Debug %s \n width %d \n\n", m.selected, m.debug, m.width)

	view += m.tabsView() + "\n\n"
	if m.selected == "table" {
		view += m.tableView()
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

	item_request := models.Requests{Method: "GET"}
	db := database.SqliteDB()

	database.Migrations(db)
	defer db.Close()

	curl := flag.String("curl", "", "")
	flag.Parse()

	// Check if flag was provided
	curlProvided := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "curl" {
			curlProvided = true
		}
	})

	if curlProvided {
		if *curl != "" {
			item, err := components.CurlBreaker(*curl, db)
			if err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			}
			item_request = item
			fmt.Println("item_request ")
		}
	}

	p := tea.NewProgram(
		initialModel(db, item_request),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
