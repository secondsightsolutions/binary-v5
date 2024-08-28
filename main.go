package main

import "net/http"

func main() {
	http.HandleFunc("/", httpRun)
	http.ListenAndServe(":80", nil)
	c := &cache{}
	c.Find("", "")
}
