package main

import (
	"flag"
	"bytes"
	"io/ioutil"
	"fmt"
	"log"
	"os/exec"
	"net/http"
	"encoding/json"
// 	"strings"
)

var port_flag = flag.Int("port", 8988, "Bind port")
var port int

func init() {
	flag.Parse()
	port = *port_flag

	// out, err := exec.Command("ls", "-Shl", "/tmp/").Output()
	// log.Println(fmt.Sprintf("err: %q, out: %s", err, out))
}

func exec_cmd(name string, args ...string) []byte {
	out, err := exec.Command(name, args...).Output()
	log.Println(fmt.Sprintf("err: %q, out: %s", err, out))
	return out
}


func is_purecolor(jreq map[string]interface{}, jresp *map[string]interface{}) {
}

func api(w http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	
	var jreq, jresp map[string]interface{}
	json.Unmarshal(body, &jreq)

	action := jreq["action"].(string)
	jresp = make(map[string]interface{})
	switch action {
	case "is_purecolor":
		is_purecolor(jreq, &jresp)
	}

	bResp, err := json.Marshal(jresp)
	if err != nil {
		log.Println("Marshal resp error:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bResp)
}

func main() {
	http.HandleFunc("/api", api)
	addr := fmt.Sprintf(":%d", port)
	log.Println(fmt.Sprintf("Listen on: <%q>", addr))
	log.Fatal(http.ListenAndServe(addr, nil))
}


/* ============================================================================
 *  Test
 * ==========================================================================*/


func test_cmd() {
	cmd := exec.Command("ls", "-Shl", "/tmp/")
	// cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	println(out.String())
	// fmt.Printf("in all caps: %q\n", out.String())
}
