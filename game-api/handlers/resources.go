package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"PawTribalWars/db"
)

// GetResourcesHandler godoc
// @Summary Pobierz zasoby wioski
// @Description Zwraca aktualny stan zasobów w wiosce. Produkcja zasobów jest liczona dynamicznie na podstawie poziomu budynków (lumbermill, claypit, ironmine) i czasu, jaki minął od ostatniej aktualizacji.
// @Tags resources
// @Produce json
// @Param village_id query int true "ID wioski"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string "Missing or invalid village_id"
// @Failure 404 {string} string "Resources not found"
// @Router /resources [get]
func GetResourcesHandler(w http.ResponseWriter, r *http.Request) {
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

	var wood, clay, iron int
	var updatedAt time.Time

	// pobieramy aktualne zasoby i timestamp
	err = db.DB.QueryRow(
		"SELECT wood, clay, iron, updated_at FROM resources WHERE village_id=$1",
		villageID,
	).Scan(&wood, &clay, &iron, &updatedAt)
	if err != nil {
		http.Error(w, "Resources not found", http.StatusNotFound)
		return
	}

	// pobieramy poziomy budynków
	var lumberLvl, clayLvl, ironLvl int
	_ = db.DB.QueryRow("SELECT level FROM buildings WHERE village_id=$1 AND type='lumbermill'", villageID).Scan(&lumberLvl)
	_ = db.DB.QueryRow("SELECT level FROM buildings WHERE village_id=$1 AND type='claypit'", villageID).Scan(&clayLvl)
	_ = db.DB.QueryRow("SELECT level FROM buildings WHERE village_id=$1 AND type='ironmine'", villageID).Scan(&ironLvl)

	// ile minut minęło od ostatniego update
	elapsedMinutes := int(time.Since(updatedAt).Minutes())
	if elapsedMinutes > 0 {
		wood += lumberLvl * 5 * elapsedMinutes
		clay += clayLvl * 5 * elapsedMinutes
		iron += ironLvl * 5 * elapsedMinutes

		// zapisujemy nowy stan
		_, _ = db.DB.Exec(
			"UPDATE resources SET wood=$1, clay=$2, iron=$3, updated_at=NOW() WHERE village_id=$4",
			wood, clay, iron, villageID,
		)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"village_id": villageID,
		"wood":       wood,
		"clay":       clay,
		"iron":       iron,
	})
}
