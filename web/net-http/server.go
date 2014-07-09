package main


import (
	"fmt"
	"log"
	"flag"
	"runtime"
	"io/ioutil"
	"net/http"
)

var (
	port = flag.Int("port", 9001, "Bind port")
	processes = flag.Int("processes", 4, "GOMAXPROCS size")
)

func init() {
	runtime.GOMAXPROCS(*processes)
	
	http.HandleFunc("/query", query)
}


func query(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "Proto : %s\n", req.Proto)
	fmt.Fprintf(rw, "Method : %s\n", req.Method)
	fmt.Fprintf(rw, "Host : %s\n", req.Host)
	fmt.Fprintf(rw, "ContentLength : %d\n", req.ContentLength)
	fmt.Fprintf(rw, "TransferEncoding : %s\n", req.TransferEncoding)
	fmt.Fprintf(rw, "RemoteAddr : %s\n", req.RemoteAddr)
	fmt.Fprintf(rw, "RequestURI : %s\n", req.RequestURI)
	body, _ := ioutil.ReadAll(req.Body)
	fmt.Fprintf(rw, "Body : %s\n", body)
	
	fmt.Fprintf(rw, "\nHeaders\n=======\n")
	for k, v := range req.Header {
		fmt.Fprintf(rw, "%s : %s\n", k, v)
	}


	fmt.Fprintf(rw, "\nURL\n====\n")
	url := *req.URL
	fmt.Fprintf(rw, "Scheme : %s\n", url.Scheme)
	fmt.Fprintf(rw, "Opaque : %s\n", url.Opaque)
	fmt.Fprintf(rw, "User : %s\n", url.User) // *Userinfo
	fmt.Fprintf(rw, "Host : %s\n", url.Host)
	fmt.Fprintf(rw, "Path : %s\n", url.Path)
	fmt.Fprintf(rw, "RawQuery : %s\n", url.RawQuery)
	fmt.Fprintf(rw, "Fragment : %s\n", url.Fragment)

	req.ParseForm()
	fmt.Fprintf(rw, "\nForm (%s)\n====\n", req.Form)
	for k, v := range req.Form {
		fmt.Fprintf(rw, "%s : %s\n", k, v)
	}
	
	fmt.Fprintf(rw, "\nPostForm (%s)\n========\n", req.PostForm)
	for k, v := range req.PostForm {
		fmt.Fprintf(rw, "%s : %s\n", k, v)
	}

	req.ParseMultipartForm(1024 * 1024 * 4)
	fmt.Fprintf(rw, "\nMultipartForm (%s)\n===========\n", req.MultipartForm)
	if req.MultipartForm != nil {
		fmt.Fprintf(rw, ".Value:\n")
		for k, v := range req.MultipartForm.Value {
			fmt.Fprintf(rw, "%s : %s\n", k, v)
		}
		fmt.Fprintf(rw, ".File:\n")
		for k, v := range req.MultipartForm.File {
			fmt.Fprintf(rw, "%s : %s\n", k, v)
		}
	} 
}


func main() {
	
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Listen on >> %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
