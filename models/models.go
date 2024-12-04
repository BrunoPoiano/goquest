package models

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


