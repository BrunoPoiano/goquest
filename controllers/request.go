package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"main/models"
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

func AddItemsToTable(db *sql.DB, item models.Requests) {

	var quantity string
	quantity_query := db.QueryRow("SELECT count(id) as quantity FROM requests WHERE name = ?", item.Name).Scan(&quantity)

	if quantity_query != nil && quantity_query != sql.ErrNoRows {
		return
	}

	params_encoded := item.Params

	if quantity > "0" {

		update_query := `UPDATE requests SET name=?, method=?, route=?, params=?, headers=? WHERE route=?`
		_, err := db.Exec(update_query, item.Name, item.Method, item.Route, params_encoded, item.Headers, item.Route)
		if err != nil {
			return
		}

		return
	}

	insert_query := `INSERT INTO requests (name, method, route, params, headers) VALUES (?, ?, ?, ?,?)`
	_, err := db.Exec(insert_query, item.Name, item.Method, item.Route, params_encoded, item.Headers)
	if err != nil {
		return
	}
}
