package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"PawTribalWars/db"
)

// struktura kosztów budynków
type BuildingCost struct {
	Wood int `json:"wood"`
	Clay int `json:"clay"`
	Iron int `json:"iron"`
}

// bazowe koszty (dla level 1)
var baseCosts = map[string]BuildingCost{
	"lumbermill": {Wood: 50, Clay: 50, Iron: 20},
	"claypit":    {Wood: 50, Clay: 50, Iron: 20},
	"ironmine":   {Wood: 50, Clay: 50, Iron: 20},
	"warehouse":  {Wood: 100, Clay: 60, Iron: 40},
	"barracks":   {Wood: 120, Clay: 100, Iron: 80},
}

// oblicz koszt ulepszenia
func calculateUpgradeCost(buildingType string, nextLevel int) BuildingCost {
	base := baseCosts[buildingType]
	multiplier := math.Pow(2.5, float64(nextLevel-1))
	return BuildingCost{
		Wood: int(float64(base.Wood) * multiplier),
		Clay: int(float64(base.Clay) * multiplier),
		Iron: int(float64(base.Iron) * multiplier),
	}
}

// GET /buildings?village_id=1
func GetBuildingsHandler(w http.ResponseWriter, r *http.Request) {
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

	rows, err := db.DB.Query("SELECT type, level FROM buildings WHERE village_id=$1", villageID)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	buildings := []map[string]interface{}{}
	for rows.Next() {
		var bType string
		var level int
		rows.Scan(&bType, &level)
		buildings = append(buildings, map[string]interface{}{
			"type":  bType,
			"level": level,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"village_id": villageID,
		"buildings":  buildings,
	})
}

// PUT /buildings/upgrade?village_id=1&type=lumbermill
func UpgradeBuildingHandler(w http.ResponseWriter, r *http.Request) {
	villageIDStr := r.URL.Query().Get("village_id")
	buildingType := r.URL.Query().Get("type")

	if villageIDStr == "" || buildingType == "" {
		http.Error(w, "Missing village_id or building type", http.StatusBadRequest)
		return
	}
	villageID, err := strconv.Atoi(villageIDStr)
	if err != nil {
		http.Error(w, "Invalid village_id", http.StatusBadRequest)
		return
	}

	// pobierz aktualny level
	var currentLevel int
	err = db.DB.QueryRow("SELECT level FROM buildings WHERE village_id=$1 AND type=$2",
		villageID, buildingType).Scan(&currentLevel)
	if err != nil {
		http.Error(w, "Building not found", http.StatusNotFound)
		return
	}

	nextLevel := currentLevel + 1
	cost := calculateUpgradeCost(buildingType, nextLevel)

	// pobierz zasoby
	var wood, clay, iron int
	err = db.DB.QueryRow("SELECT wood, clay, iron FROM resources WHERE village_id=$1", villageID).
		Scan(&wood, &clay, &iron)
	if err != nil {
		http.Error(w, "Resources not found", http.StatusInternalServerError)
		return
	}

	// sprawdź czy stać
	if wood < cost.Wood || clay < cost.Clay || iron < cost.Iron {
		http.Error(w, "Not enough resources", http.StatusForbidden)
		return
	}

	// odejmij koszt
	_, err = db.DB.Exec(
		"UPDATE resources SET wood=wood-$1, clay=clay-$2, iron=iron-$3 WHERE village_id=$4",
		cost.Wood, cost.Clay, cost.Iron, villageID,
	)
	if err != nil {
		http.Error(w, "DB error on resources update", http.StatusInternalServerError)
		return
	}

	// podnieś level
	_, err = db.DB.Exec(
		"UPDATE buildings SET level=level+1 WHERE village_id=$1 AND type=$2",
		villageID, buildingType,
	)
	if err != nil {
		http.Error(w, "DB error on building update", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Building upgraded",
		"building_type": buildingType,
		"new_level":     nextLevel,
		"cost":          cost,
	})
}

// GET /buildings/cost?village_id=1&type=lumbermill
func GetBuildingCostHandler(w http.ResponseWriter, r *http.Request) {
	villageIDStr := r.URL.Query().Get("village_id")
	buildingType := r.URL.Query().Get("type")

	if villageIDStr == "" || buildingType == "" {
		http.Error(w, "Missing village_id or building type", http.StatusBadRequest)
		return
	}
	villageID, err := strconv.Atoi(villageIDStr)
	if err != nil {
		http.Error(w, "Invalid village_id", http.StatusBadRequest)
		return
	}

	// pobierz aktualny level
	var currentLevel int
	err = db.DB.QueryRow("SELECT level FROM buildings WHERE village_id=$1 AND type=$2",
		villageID, buildingType).Scan(&currentLevel)
	if err != nil {
		http.Error(w, "Building not found", http.StatusNotFound)
		return
	}

	nextLevel := currentLevel + 1
	cost := calculateUpgradeCost(buildingType, nextLevel)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"building_type": buildingType,
		"current_level": currentLevel,
		"next_level":    nextLevel,
		"cost":          cost,
	})
}
