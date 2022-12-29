package handler

import (
	"io/ioutil"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("./static/view/signin.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func Home(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("./static/view/home.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
