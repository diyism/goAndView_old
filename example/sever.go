package exampleserver

import "net/http"

func init() {
	//just add some endpoints to the server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
	//or you can also use your favorite router and then paste it to http.DefaultServeMux
}
