package main

import "net/http"

func main() {
	r := newRoom()
	http.Handle("/room", r)
	go r.run()

	http.Handle("/", http.FileServer(http.Dir("./")))
	http.ListenAndServe(":8080", nil)
}
