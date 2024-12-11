package components

import (
	"fmt"
	"main/models"
	"net/url"

	"github.com/charmbracelet/huh"
)

func CreateForm(rf *models.Requests) *huh.Form {
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
				Key("headers").
				Value(&rf.Headers).
				Title("Headers"),

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
		WithShowHelp(true).
		WithShowErrors(true)
}
