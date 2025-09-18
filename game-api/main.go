package main

import (
	"PawTribalWars/db"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	// Połącz się z bazą
	db.ConnectDB()

	// Router
	r := mux.NewRouter()
	fmt.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
