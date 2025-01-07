package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"goquest/models"
	"io"
	"net/http"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func MakeRequest(request models.Requests, db *sql.DB) tea.Cmd {

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

	if err != nil {
		return func() tea.Msg {
			return models.ReturnRequest{
				Response: err.Error(),
			}
		}
	}

	AddItemsToTable(db, request)
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

func DeleteItemFromTable(db *sql.DB, item_id string) error {

	_, err := db.Exec("DELETE FROM requests WHERE id=?", item_id)

	if err != nil {
		return errors.New("Error deleting item from requests")
	}

	return nil
}

func GetItemsFromTable(db *sql.DB) []models.Requests {

	var items []models.Requests
	rows, err := db.Query("Select * from requests")
	if err != nil {
		fmt.Println("Error getting table list", err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Requests
		if err := rows.Scan(&item.Id, &item.Name, &item.Method, &item.Route, &item.Params, &item.Headers); err != nil {
			return nil
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil
	}
	return items
}

func GetItemFromTable(db *sql.DB, method string, route string) (models.Requests, error) {

	var item models.Requests
	rows, err := db.Query("Select * from requests where method=? and route=?", method, route)
	if err != nil {
		return models.Requests{}, errors.New("Error getting table list")
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&item.Id, &item.Name, &item.Method, &item.Route, &item.Params, &item.Headers); err != nil {
			return models.Requests{}, errors.New("Error rows.Scan")
		}
	}

	if err := rows.Err(); err != nil {
		return models.Requests{}, errors.New("Error rows.Err")
	}

	return item, nil
}

func AddItemsToTable(db *sql.DB, item models.Requests) error {

	var quantity string
	quantity_query := db.QueryRow("SELECT count(id) as quantity FROM requests WHERE method=? and route = ?", item.Method, item.Route).Scan(&quantity)

	if quantity_query != nil && quantity_query != sql.ErrNoRows {
		return errors.New("Error runing quantity check")
	}

	params_encoded := item.Params

	if quantity > "0" {

		update_query := `UPDATE requests SET name=?, method=?, route=?, params=?, headers=? WHERE method=? and route=?`
		_, err := db.Exec(update_query, item.Name, item.Method, item.Route, params_encoded, item.Headers, item.Method, item.Route)
		if err != nil {
			return errors.New("Error updating Table")
		}

		return nil
	}

	insert_query := `INSERT INTO requests (name, method, route, params, headers) VALUES (?, ?, ?, ?,?)`
	_, err := db.Exec(insert_query, item.Name, item.Method, item.Route, params_encoded, item.Headers)
	if err != nil {
		return errors.New("Error inserting item in table")
	}

	return nil
}
