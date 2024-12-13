package components

import (
	"database/sql"
	"fmt"
	"main/controllers"
	"main/models"
	"strings"
)

func CurlBreaker(curl string, db *sql.DB) (models.Requests, error) {

	splited := strings.Split(curl, "-H")

	var item models.Requests
	var headers string
	item.Name = "curl added"
	item.Method = "GET"

	for i, v := range splited {
		value := strings.ReplaceAll(v, "'", "")

		if i == 0 {
			route := strings.ReplaceAll(value, "curl", "")
			routeSplit := strings.Split(route, "?")
			if routeSplit[0] != "" {

				urlSlipt := strings.Split(routeSplit[0], "-X")
				if urlSlipt[0] != "" {
					item.Route = strings.ReplaceAll(urlSlipt[0], " ", "")
				}

				if len(urlSlipt) > 1 && urlSlipt[1] != "" {
					item.Method = strings.ReplaceAll(urlSlipt[1], " ", "")
				}
			}
			if len(routeSplit) > 1 && routeSplit[1] != "" {
				item.Params = routeSplit[1]
			}
		} else {
			header := ""
			if strings.Contains(value, "--data-raw") {
				valueSplit := strings.Split(value, "--data-raw")
				header = valueSplit[0]
				item.Params = strings.TrimSpace(valueSplit[1])
			} else {
				header = value
			}

			header = strings.Replace(header, ":", "=", 1)
			headers += header + "|"

		}
	}

	item.Headers = headers

	err := controllers.AddItemsToTable(db, item)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return models.Requests{}, err
	}

	saved_item, err := controllers.GetItemFromTable(db, item.Method, item.Route)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return models.Requests{}, err
	}

	return saved_item, nil
}
