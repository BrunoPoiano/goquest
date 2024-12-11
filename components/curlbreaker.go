package components

import (
	"database/sql"
	"fmt"
	"main/controllers"
	"main/models"
	"strings"
)

func CurlBreaker(curl string, db *sql.DB) (error) {

	splited := strings.Split(curl, "-H")

	var item models.Requests
	var headers string

	for i, v := range splited {
		value := strings.ReplaceAll(v, `'`, "")

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
				item.Params = strings.ReplaceAll(routeSplit[1], "&", ";")
			}
		} else {
			header := strings.ReplaceAll(value, ":", "=")
			headers += header + "|"
		}
	}

	item.Headers = headers
	item.Name = "curl added"

  err := controllers.AddItemsToTable(db, item)
  if err != nil {
    fmt.Println("Error: ", err.Error())
    return err

  }
  
return nil
}
