package main

import (
	"log"
	"net/http"

	"github.com/221bye/urlshortener/handlers"
)

func main() {
	a := &handlers.App{}
	err := a.RedisInit()
	if err != nil {
		panic(err)
	}
	defer a.Conn.Close()

	router := a.RouterInit()

	log.Fatal(http.ListenAndServe(":8080", router))
}
