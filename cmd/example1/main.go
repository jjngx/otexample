package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

func main() {
	run()
}

func run() {
	http.HandleFunc("/rolldice", rolldice)
	log.Fatal(http.ListenAndServe(":8088", nil))
}

func rolldice(w http.ResponseWriter, r *http.Request) {
	roll := 1 + rand.Intn(6)
	if _, err := fmt.Fprintf(w, "%d\n", roll); err != nil {
		log.Print(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}
