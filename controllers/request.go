package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"goquest/models"
)

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
