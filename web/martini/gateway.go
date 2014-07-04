package main

// Created at: [2014-06-28 13:04]

/* ============================================================================
   Requirements
   ------------
   + https support (https://)
   + secure websocket (wss://)
   + Performance test >> http://ziutek.github.io/web_bench/
      - siege -c 200 -t 20s http://ADDRESS:PORT/Hello/100
 * ==========================================================================*/


import (
	"log"
	"net/http"
	"runtime"
	"github.com/go-martini/martini"
// 	"github.com/martini-contrib/auth"
)


var m *martini.Martini
const AuthUser = "user"
const AuthPasswd = "passwd"

func init() {
	runtime.GOMAXPROCS(4)
	
	m = martini.New()
	m.Use(martini.Recovery())
	m.Use(martini.Logger())
	// m.Use(auth.Basic(AuthUser, AuthPasswd))
	
	r := martini.NewRouter()
	r.Get(`/`, index)
	
	m.Action(r.Handle)
}


func main() {
	log.Println("Listening on :9001")
	log.Fatal(http.ListenAndServeTLS(":9001", "res/cert.pem", "res/key.pem", m))
}
