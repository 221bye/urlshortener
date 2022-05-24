package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
)

var (
	conn     redis.Conn
	alphabet string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

type urlData struct {
	Url string `json:"url"`
}

type response struct {
	BaseUrl  string `json:"BaseUrl"`
	ShortUrl string `json:"ShortUrl"`
}

//converts n to baseM, M = len(alphabet)
func encode(n int64) string {
	alphabet := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	base := int64(len(alphabet))

	rem := n % base
	n /= base

	res := string(alphabet[rem])
	for n != 0 {
		rem = n % base
		res += string(alphabet[rem])
		n /= base
	}

	return res
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	u := &urlData{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&u)

	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	if !govalidator.IsURL(u.Url) {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		fmt.Println("Invalid URL")
		return
	}

	counter, err := redis.Int64(conn.Do("INCR", "counter"))
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	short := encode(counter)

	conn.Do("HSET", short, "baseUrl", u.Url)

	resp := response{
		BaseUrl:  u.Url,
		ShortUrl: short,
	}

	jsonResp, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "Failed to marshal", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	short := mux.Vars(r)["short"]

	baseUrl, _ := redis.String(conn.Do("HGET", short, "baseUrl"))
	fmt.Println("short: ", short)
	fmt.Println("baseUrl: ", baseUrl)

	if baseUrl == "" {
		http.Error(w, "Not found :(", http.StatusBadRequest)
		return
	} else {
		if !(strings.HasPrefix(baseUrl, "http://") ||
			strings.HasPrefix(baseUrl, "https://")) {
			baseUrl = "http://" + baseUrl
		}
		fmt.Println(baseUrl)
		http.Redirect(w, r, baseUrl, http.StatusMovedPermanently)
	}

}

func main() {
	var err error
	conn, err = redis.Dial("tcp", ":6379")

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	reply, err := conn.Do("GET", "counter")
	if err != nil {
		panic(err)
	}
	if reply == nil {
		conn.Do("SET", "counter", "0")
	}

	router := mux.NewRouter()

	router.HandleFunc("/shorten", shortenHandler).Methods("POST")
	router.HandleFunc("/r/{short}", redirectHandler).Methods("GET")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public"))).Methods("GET")
	http.Handle("/", router)

	log.Fatal(http.ListenAndServe(":8080", router))
}
