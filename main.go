package main

import (
	"log"
	"net/http"
	"os"

	"gitlab.com/zhangkui/orbit-relay/internal/handler"
	"gitlab.com/zhangkui/orbit-relay/internal/repository"
	"gitlab.com/zhangkui/orbit-relay/internal/service"
	"gitlab.com/zhangkui/orbit-relay/internal/store"
)

func main() {
	path := os.Getenv("ORBIT_RELAY_DB")
	if path == "" {
		path = "data/orbit-relay.db"
	}
	db, err := store.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	app := service.New(repository.New(db))
	defer app.Close()
	addr := os.Getenv("ORBIT_RELAY_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("orbit-relay listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler.New(app)))
}
