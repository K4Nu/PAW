package main

import (
	"PawTribalWars/db"
	"PawTribalWars/handlers"
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
	// User authentication
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST")

	// Vilages endpoints
	r.Handle("/villages", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetVillagesHandler))).Methods("GET")
	r.Handle("/villages", handlers.AuthMiddleware(http.HandlerFunc(handlers.CreateVillageHandler))).Methods("POST")
	r.Handle("/villages/{id}", handlers.AuthMiddleware(http.HandlerFunc(handlers.UpdateVillageHandler))).Methods("PUT")
	r.Handle("/villages/{id}", handlers.AuthMiddleware(http.HandlerFunc(handlers.DeleteVillageHandler))).Methods("DELETE")

	// Resources
	r.Handle("/resources", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetResourcesHandler))).Methods("GET")

	// Buildings
	r.Handle("/buildings", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetBuildingsHandler))).Methods("GET")
	r.Handle("/buildings/upgrade", handlers.AuthMiddleware(http.HandlerFunc(handlers.UpgradeBuildingHandler))).Methods("PUT")
	r.Handle("/buildings/cost", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetBuildingCostHandler))).Methods("GET")

	// Units
	r.Handle("/units", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetUnitsHandler))).Methods("GET")
	r.Handle("/units/recruit", handlers.AuthMiddleware(http.HandlerFunc(handlers.RecruitUnitsHandler))).Methods("POST")
	
	fmt.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
