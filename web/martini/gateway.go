

package main

// Created at: [2014-06-28 13:04]

import (
	"log"
	"net/http"
	"github.com/go-martini/martini"
// 	"github.com/martini-contrib/auth"
)


var m *martini.Martini

func init() {
	m = martini.New()
	m.Use(martini.Logger())
	
	r := martini.NewRouter()
	r.Get(`/`, index)
	
	m.Action(r.Handle)
}


func index() (int, string) {
	return 500, "This is index."
}


func main() {
	log.Println("Listening on :9001")
	err := http.ListenAndServeTLS(":9001", "cert.pem", "key.pem", m);
	if err != nil {
		log.Fatal(err)
	}
}
