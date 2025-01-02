package requests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"goquest/controllers"
	"goquest/models"
	"io"
	"net/http"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func MakeRequest(request models.Requests, db *sql.DB) tea.Cmd {

	/*
	   	if strconv.Itoa(request.Id) != "0"{
	   		return func() tea.Msg {
	   			return models.ReturnRequest{
	   				Response: "TEM ID " + request.Headers,
	   				Error:    nil,
	   			}
	   		}
	     }else{
	   		return func() tea.Msg {
	   			return models.ReturnRequest{
	   				Response: " sem id" +strconv.Itoa(request.Id),

	   				Error:    nil,
	   			}
	   		}

	     }
	*/

	requestMethod := http.MethodGet
	var response *http.Request
	var err error

	switch request.Method {
	case "POST":
		requestMethod = http.MethodPost
	case "GET":
		requestMethod = http.MethodGet
	case "DELETE":
		requestMethod = http.MethodDelete
	case "PUT":
		requestMethod = http.MethodPut
	}

	switch requestMethod {

	case http.MethodGet:
		fullURL := fmt.Sprintf("%s?%s", request.Route, request.Params)
		response, err = http.NewRequest(requestMethod, fullURL, nil)

	default:
		response, err = http.NewRequest(requestMethod, request.Route, bytes.NewBuffer([]byte(request.Params)))
	}

	response.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if len(request.Headers) > 0 {
		headersSplit := strings.Split(request.Headers, "|")

		for _, values := range headersSplit {
			item := strings.SplitN(values, "=", 2)
			if len(item) > 1 && item[0] != "" {
				name := strings.TrimSpace(item[0])
				value := strings.TrimSpace(item[1])
				response.Header.Set(name, value)
			}
		}
	}

	/*
		    //Check Headers

				return func() tea.Msg {
					return models.ReturnRequest{
						Response: formatHeaders(response.Header),
						Error:    err,
					}
				}
	*/

	if err != nil {
		return func() tea.Msg {
			return models.ReturnRequest{
				Response: err.Error(),
			}
		}
	}

	controllers.AddItemsToTable(db, request)
	prettyRes := responseParser(response)

	return func() tea.Msg {
		return models.ReturnRequest{
			Response: prettyRes,
		}
	}

}

func responseParser(response *http.Request) string {

	res, err := http.DefaultClient.Do(response)
	if err != nil {
		return "Error making http request \n" + err.Error()
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "Error reading response body \n" + err.Error()
	}

	prettyRes := prettyString(string(resBody))
	if err != nil {
		return string(resBody)
	}

	return prettyRes

}

func prettyString(str string) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return str
	}
	return prettyJSON.String()
}

func formatHeaders(headers http.Header) string {
	var headerStr string
	for key, values := range headers {
		for _, value := range values {
			headerStr += fmt.Sprintf("%s: %s\n", key, value)
		}
	}
	return headerStr
}
