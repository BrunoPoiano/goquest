package components

import (
	"database/sql"
	"goquest/controllers"
	"goquest/models"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func Table(db *sql.DB, width int, height int) tea.Cmd {

	items := controllers.GetItemsFromTable(db)

	var rows []table.Row

	for _, v := range items {
		rows = append(rows, table.Row{strconv.Itoa(v.Id), v.Name, v.Method, v.Route, v.Params, v.Headers})
	}
	column_w := (width - 24) / 4

	columns := []table.Column{
		{Title: "Id", Width: 4},
		{Title: "Name", Width: column_w},
		{Title: "Method", Width: 10},
		{Title: "URL", Width: column_w},
		{Title: "Params", Width: column_w},
		{Title: "Headers", Width: column_w},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height-15),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return func() tea.Msg {
		return models.ReturnTable{
			Table: t,
			Error: nil,
		}
	}
}
