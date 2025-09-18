package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"PawTribalWars/db"
// Units
r.Handle("/units", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetUnitsHandler))).Methods("GET")
r.Handle("/units/recruit", handlers.AuthMiddleware(http.HandlerFunc(handlers.RecruitUnitsHandler))).Methods("POST"))

// =============================
// GET /units?village_id=1
// =============================
func GetUnitsHandler(w http.ResponseWriter, r *http.Request) {
	villageIDStr := r.URL.Query().Get("village_id")
	if villageIDStr == "" {
		http.Error(w, "Missing village_id", http.StatusBadRequest)
		return
	}
	villageID, err := strconv.Atoi(villageIDStr)
	if err != nil {
		http.Error(w, "Invalid village_id", http.StatusBadRequest)
		return
	}

	rows, err := db.DB.Query(
		"SELECT type, count FROM units WHERE village_id=$1",
		villageID,
	)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	units := []map[string]interface{}{}
	for rows.Next() {
		var uType string
		var count int
		rows.Scan(&uType, &count)
		units = append(units, map[string]interface{}{
			"type":  uType,
			"count": count,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"village_id": villageID,
		"units":      units,
	})
}

// =============================
// POST /units/recruit?village_id=1&type=spearman&count=5
// =============================
func RecruitUnitsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	villageIDStr := r.URL.Query().Get("village_id")
	unitType := r.URL.Query().Get("type")
	countStr := r.URL.Query().Get("count")

	if villageIDStr == "" || unitType == "" || countStr == "" {
		http.Error(w, "Missing params", http.StatusBadRequest)
		return
	}
	villageID, _ := strconv.Atoi(villageIDStr)
	count, _ := strconv.Atoi(countStr)

	// 🔹 sprawdź, czy wioska należy do usera
	var belongs bool
	err := db.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM villages v
			JOIN users u ON v.user_id = u.id
			WHERE v.id=$1 AND u.username=$2
		)`, villageID, username).Scan(&belongs)
	if err != nil || !belongs {
		http.Error(w, "Village not found or not yours", http.StatusForbidden)
		return
	}

	// 🔹 koszty jednostek
	costs := map[string][3]int{
		"spearman":  {50, 30, 20},
		"swordsman": {30, 50, 40},
		"archer":    {40, 40, 30},
	}
	cost, ok := costs[unitType]
	if !ok {
		http.Error(w, "Invalid unit type", http.StatusBadRequest)
		return
	}
	totalWood := cost[0] * count
	totalClay := cost[1] * count
	totalIron := cost[2] * count

	// 🔹 sprawdź zasoby
	var wood, clay, iron int
	err = db.DB.QueryRow("SELECT wood, clay, iron FROM resources WHERE village_id=$1", villageID).
		Scan(&wood, &clay, &iron)
	if err != nil {
		http.Error(w, "Resources not found", http.StatusInternalServerError)
		return
	}
	if wood < totalWood || clay < totalClay || iron < totalIron {
		http.Error(w, "Not enough resources", http.StatusForbidden)
		return
	}

	// 🔹 odejmij zasoby
	_, err = db.DB.Exec(`
		UPDATE resources
		SET wood = wood - $1, clay = clay - $2, iron = iron - $3, updated_at=NOW()
		WHERE village_id=$4
	`, totalWood, totalClay, totalIron, villageID)
	if err != nil {
		http.Error(w, "DB error on resources", http.StatusInternalServerError)
		return
	}

	// 🔹 dodaj jednostki
	_, err = db.DB.Exec(`
		UPDATE units SET count = count + $1
		WHERE village_id=$2 AND type=$3
	`, count, villageID, unitType)
	if err != nil {
		http.Error(w, "DB error on units", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Units recruited",
		"type":    unitType,
		"count":   count,
		"cost": map[string]int{
			"wood": totalWood,
			"clay": totalClay,
			"iron": totalIron,
		},
	})
}
