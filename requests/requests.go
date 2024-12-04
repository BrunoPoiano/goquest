package requests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"main/controllers"
	"main/models"
	"net/http"
	"net/url"
	"strings"
)

func MakeRequest(request models.Requests, db *sql.DB) (string, error) {

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

  request_params, err := url.ParseQuery(request.Params)

	if requestMethod == http.MethodGet {
		fullURL := fmt.Sprintf("%s?%s", request.Route, request_params.Encode())
		response, err = http.NewRequest(requestMethod, fullURL, nil)
	} else {
		response, err = http.NewRequest(requestMethod, request.Route, strings.NewReader(request_params.Encode()))
		response.Header.Set("Content-Type", "application/x-www-form-urlencoded") // For form data
	}

	if err != nil {
		return "", err
	}

	controllers.AddItemsToTable(db, request)
	prettyRes, err := responseParser(response)

	if err != nil {
		return "", err
	}
	return prettyRes, nil

}

func responseParser(response *http.Request) (string, error) {
	res, err := http.DefaultClient.Do(response)
	if err != nil {
		return "error making http request:", err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "error reading response body", err
	}

	prettyRes, err := prettyString(string(resBody))
	if err != nil {
		return "error formating Json", err
	}
	return prettyRes, nil
}

func prettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "error prettyString func", err
	}
	return prettyJSON.String(), nil
}
