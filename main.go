package main

import (
	"fmt"
	//"github.com/gomodule/redigo/redis"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type urlData struct {
	Url string `json:"url"`
}

type response struct {
	BaseUrl  string `json:"BaseUrl"`
	ShortUrl string `json:"ShortUrl"`
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	u := &urlData{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&u)
	fmt.Printf("%+v\n", u)

	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	resp := response{
		BaseUrl:  u.Url,
		ShortUrl: "kek",
	}
	jsonResp, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "Failed to marshal", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func main() {
	// conn, err := redis.Dial("tcp", ":6379")
	// if err != nil {
	// 	panic(err)
	// }
	// defer conn.Close()

	r := mux.NewRouter()

	r.HandleFunc("/shorten", shortenHandler).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public"))).Methods("GET")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":8080", r))
}
