package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func handle(w http.ResponseWriter, r *http.Request) {
	sqlFQDN := os.Getenv("MYSQL_FQDN")
	if sqlFQDN == "" {
		http.Error(w, "couldn't find MYSQL FQDN", http.StatusInternalServerError)
	}

	w.Write([]byte(fmt.Sprintf("Hello, I'm a webserver that wants to connect to a MYSQL at %s", sqlFQDN)))
}

func main() {

	http.HandleFunc("/", handle)

	fmt.Printf("Server starting on port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
