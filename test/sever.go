package server

import "net/http"

//Server is the exported http server instance
var Server = &http.Server{Addr: ":8080"}

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
}

func main() {
	Start()
}

//Start starts the http server
func Start() {
	go Server.ListenAndServe()
}

//Stop stops the http server
func Stop() {
	//Server.Close()
	panic("meh")
}
