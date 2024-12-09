package models

import (
	"github.com/charmbracelet/bubbles/table"
)

type Requests struct {
	Id     int
	Name   string
	Method string
	Route  string
	Params string
	//Params url.Values
}

type RequestForm struct {
	Name   string
	Body   string
	Method string
	Send   bool
}

type ReturnRequest struct {
	Response string
	Error    error
}

type ReturnTable struct {
	Table table.Model
	Error error
}
