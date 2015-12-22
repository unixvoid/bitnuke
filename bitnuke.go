package main

import (
	//"fmt"
	//"github.com/gorilla/mux"
	"net/http"
)

// working
//func main() {
//	router := mux.NewRouter()
//	stat := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
//	router.PathPrefix("/static/").Handler(stat)
//
//	http.Handle("/", router)
//	http.ListenAndServe(":80", router)
//}

func main() {
	//repoFrontend := "./tmpnuke/"
	//http.Handle("/", http.FileServer(http.Dir(repoFrontend)))
	http.Handle("/", handlerstatic)
	err := http.ListenAndServe(":80", nil)
	if nil != err {
		panic(err)
	}
}

func handlerstatic(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.Dir("./tmpnuke/"))
}
