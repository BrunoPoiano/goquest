package models

import (
	"github.com/charmbracelet/bubbles/table"
)

type Requests struct {
	Id      int
	Name    string
	Method  string
	Route   string
	Params  string
	Headers string
	//Params url.Values
}

type HeadersForm struct {
	Header string
	value  string
}

type ReturnRequest struct {
	Response string
	Error    error
}

type ReturnTable struct {
	Table table.Model
	Error error
}
