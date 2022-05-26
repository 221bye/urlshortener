package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
)

type App struct {
	Conn redis.Conn
}

type request struct {
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

// 1. Gets url, checks if it is valid
// 2. gets counter (counter = total urls in db)
// 3. increments counter by one
// 4. converts this value to base62 (this will be shortened url)
// 5. and stores it in db
// 6. sends response to client
func (a *App) shortenHandler(w http.ResponseWriter, r *http.Request) {
	u := &request{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&u)

	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	if !govalidator.IsURL(u.Url) {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	counter, err := redis.Int64(a.Conn.Do("INCR", "counter"))
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	short := encode(counter)

	a.Conn.Do("HSET", short, "baseUrl", u.Url)

	resp := response{
		BaseUrl:  u.Url,
		ShortUrl: r.Host + "/r/" + short,
	}

	jsonResp, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "Failed to marshal", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func (a *App) redirectHandler(w http.ResponseWriter, r *http.Request) {
	short := mux.Vars(r)["short"]

	baseUrl, _ := redis.String(a.Conn.Do("HGET", short, "baseUrl"))

	if baseUrl == "" {
		http.Error(w, "Not found :(", http.StatusBadRequest)
		return
	} else {
		if !(strings.HasPrefix(baseUrl, "http://") ||
			strings.HasPrefix(baseUrl, "https://")) {
			baseUrl = "http://" + baseUrl
		}
		http.Redirect(w, r, baseUrl, http.StatusMovedPermanently)
	}

}

func (a *App) RouterInit() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/shorten", a.shortenHandler).Methods("POST")
	router.HandleFunc("/r/{short}", a.redirectHandler).Methods("GET")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public"))).Methods("GET")
	http.Handle("/", router)
	return router
}

func (a *App) RedisInit() error {
	var err error
	a.Conn, err = redis.Dial("tcp", ":6379")
	if err != nil {
		return err
	}

	// This will set counter to 0 and
	// new urls will overwrite existing urls
	_, err = a.Conn.Do("SET", "counter", "0")
	if err != nil {
		return err
	}

	return err
}
