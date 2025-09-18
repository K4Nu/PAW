package main

import (
	"PawTribalWars/db"
	_ "PawTribalWars/docs"
	"PawTribalWars/handlers"
	"fmt"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
)

// @title Game API
// @version 1.0
// @description Strategy game backend with users, villages, resources, buildings, and units.
// @host localhost:8080
// @BasePath /
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

	// Swagger UI
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	fmt.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
