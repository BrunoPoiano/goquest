package components

import (
	"database/sql"
	"main/controllers"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

func Table(db *sql.DB) table.Model {

	items := controllers.GetItemsFromTable(db)

	var rows []table.Row

	for _, v := range items {
		rows = append(rows, table.Row{strconv.Itoa(v.Id), v.Name, v.Method, v.Route})
	}

	columns := []table.Column{
		{Title: "Id", Width: 4},
		{Title: "Name", Width: 10},
		{Title: "Method", Width: 10},
		{Title: "URL", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
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

	return t
}
