package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"PawTribalWars/db"
	"github.com/gorilla/mux"
)

type Village struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// =============================
// GET /villages
// =============================
func GetVillagesHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	rows, err := db.DB.Query(`
		SELECT v.id, v.name, v.created_at
		FROM villages v
		JOIN users u ON v.user_id = u.id
		WHERE u.username = $1
	`, username)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var villages []Village
	for rows.Next() {
		var v Village
		rows.Scan(&v.ID, &v.Name, &v.CreatedAt)
		villages = append(villages, v)
	}

	json.NewEncoder(w).Encode(villages)
}

// =============================
// POST /villages
// =============================
func CreateVillageHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Missing village name", http.StatusBadRequest)
		return
	}

	var userID int
	err := db.DB.QueryRow("SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	// 🔹 policz ile gracz ma już wiosek
	var villageCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM villages WHERE user_id=$1", userID).Scan(&villageCount)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	// 🔹 znajdź najwyższy poziom ratusza gracza
	var maxTownhall int
	err = db.DB.QueryRow(`
		SELECT COALESCE(MAX(b.level), 0)
		FROM buildings b
		JOIN villages v ON b.village_id = v.id
		WHERE v.user_id=$1 AND b.type='townhall'
	`, userID).Scan(&maxTownhall)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	// 🔹 wylicz maksymalną liczbę wiosek
	maxVillages := (maxTownhall / 10) + 1
	if villageCount >= maxVillages {
		http.Error(w,
			fmt.Sprintf("You need higher Townhall level to create more villages (current max: %d)", maxVillages),
			http.StatusForbidden,
		)
		return
	}

	// Tworzymy nową wioskę
	var villageID int
	err = db.DB.QueryRow(
		"INSERT INTO villages (user_id, name) VALUES ($1, $2) RETURNING id",
		userID, name,
	).Scan(&villageID)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	// Inicjalizujemy zasoby
	_, _ = db.DB.Exec("INSERT INTO resources (village_id, wood, clay, iron) VALUES ($1, 100, 100, 100)", villageID)

	// Inicjalizujemy budynki (lvl 1)
	buildings := []string{"townhall", "lumbermill", "claypit", "ironmine", "warehouse", "barracks"}
	for _, b := range buildings {
		_, _ = db.DB.Exec("INSERT INTO buildings (village_id, type, level) VALUES ($1, $2, 1)", villageID, b)
	}

	// Inicjalizujemy jednostki (0)
	units := []string{"spearman", "swordsman", "archer"}
	for _, u := range units {
		_, _ = db.DB.Exec("INSERT INTO units (village_id, type, count) VALUES ($1, $2, 0)", villageID, u)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "New village created",
		"village_id":  villageID,
		"max_allowed": maxVillages,
		"current":     villageCount + 1,
	})
}

// =============================
// PUT /villages/{id}
// =============================
func UpdateVillageHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	idStr := mux.Vars(r)["id"]
	name := r.FormValue("name")

	villageID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid village ID", http.StatusBadRequest)
		return
	}

	// sprawdzamy czy ta wioska należy do usera
	var exists bool
	err = db.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM villages v
			JOIN users u ON v.user_id = u.id
			WHERE v.id=$1 AND u.username=$2
		)
	`, villageID, username).Scan(&exists)

	if !exists {
		http.Error(w, "Village not found or not yours", http.StatusForbidden)
		return
	}

	_, err = db.DB.Exec("UPDATE villages SET name=$1 WHERE id=$2", name, villageID)
	if err != nil {
		http.Error(w, "DB error on update", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Village updated"})
}

// =============================
// DELETE /villages/{id}
// =============================
func DeleteVillageHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	idStr := mux.Vars(r)["id"]

	villageID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid village ID", http.StatusBadRequest)
		return
	}

	// sprawdzamy czy ta wioska należy do usera
	var exists bool
	err = db.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM villages v
			JOIN users u ON v.user_id = u.id
			WHERE v.id=$1 AND u.username=$2
		)
	`, villageID, username).Scan(&exists)

	if !exists {
		http.Error(w, "Village not found or not yours", http.StatusForbidden)
		return
	}

	_, err = db.DB.Exec("DELETE FROM villages WHERE id=$1", villageID)
	if err != nil {
		http.Error(w, "DB error on delete", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Village deleted"})
}
