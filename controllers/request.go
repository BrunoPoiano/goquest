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
		var paramsStr string
		var item models.Requests
		if err := rows.Scan(&item.Id, &item.Name, &item.Method, &item.Route, &paramsStr); err != nil {
			//fmt.Println("Error scanning row:", err)
			return nil
		}

		// Decode the URL-encoded params string into a url.Values
    item.Params = paramsStr

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		//fmt.Println("Error iterating over rows:", err)
		return nil
	}
	return items
}

func AddItemsToTable(db *sql.DB, item models.Requests) {

	var quantity string
	quantity_query := db.QueryRow("SELECT count(id) as quantity FROM requests WHERE name = ?", item.Name).Scan(&quantity)

	if quantity_query != nil && quantity_query != sql.ErrNoRows {
		//fmt.Println("Error checking if item exists:", quantity_query)
		return
	}

	params_encoded := item.Params

	if quantity > "0" {

		update_query := `UPDATE requests SET name=?, method=?, route=?, params=? WHERE route=?`
		_, err := db.Exec(update_query, item.Name, item.Method, item.Route, params_encoded, item.Route)
		if err != nil {
			//fmt.Println("Error updating item into table:", err)
			return
		}

		//fmt.Println("Item successfully updated")
		return
	}

	insert_query := `INSERT INTO requests (name, method, route, params) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(insert_query, item.Name, item.Method, item.Route, params_encoded)
	if err != nil {
		//fmt.Println("Error inserting item into table:", err)
		return
	}

	//fmt.Println("Item successfully inserted into the 'requests' table!")
}
