package main

// Created at: [2014-06-28 13:04]

import (
	"log"
	"net/http"
	"github.com/go-martini/martini"
 	"github.com/martini-contrib/auth"
)


var m *martini.Martini
const AuthUser = "user"
const AuthPasswd = "passwd"

func init() {
	m = martini.New()
	m.Use(martini.Recovery())
	m.Use(martini.Logger())
	m.Use(auth.Basic(AuthUser, AuthPasswd))
	
	r := martini.NewRouter()
	r.Get(`/`, index)
	
	m.Action(r.Handle)
}


func main() {
	log.Println("Listening on :9001")
	err := http.ListenAndServeTLS(":9001", "cert.pem", "key.pem", m);
	if err != nil {
		log.Fatal(err)
	}
}
