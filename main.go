package main

import (
	"html/template"
	"net/http"
	"strconv"
)

// Struct for results
type Result struct {
	Location        string
	WaterDist       float64
	ForestDist      float64
	ResidentialDist float64
	Score           int
	Recommendation  string
}

func main() {
	// Serve static files if needed
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/analyze", analyzeHandler)
	http.HandleFunc("/results", resultsHandler)

	// Start server
	http.ListenAndServe(":8080", nil)
}

// Home page
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/index.html")
	tmpl.Execute(w, nil)
}

// About page
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/about.html")
	tmpl.Execute(w, nil)
}

// Analyze page (form)
func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/analyze.html")
	tmpl.Execute(w, nil)
}

// Results page (POST)
func resultsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/analyze", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	location := r.FormValue("location")
	waterDist, _ := strconv.ParseFloat(r.FormValue("water_distance"), 64)
	forestDist, _ := strconv.ParseFloat(r.FormValue("forest_distance"), 64)
	residentialDist, _ := strconv.ParseFloat(r.FormValue("residential_distance"), 64)

	score := 0
	// Water scoring
	if waterDist >= 1 {
		score += 30
	} else if waterDist >= 0.5 {
		score += 15
	}
	// Forest scoring
	if forestDist >= 2 {
		score += 30
	} else if forestDist >= 1 {
		score += 15
	}
	// Residential scoring
	if residentialDist >= 3 {
		score += 40
	} else if residentialDist >= 1 {
		score += 20
	}

	recommendation := ""
	if score >= 80 {
		recommendation = "Safe"
	} else if score >= 50 {
		recommendation = "Moderate Risk"
	} else {
		recommendation = "Unsafe"
	}

	result := Result{
		Location:        location,
		WaterDist:       waterDist,
		ForestDist:      forestDist,
		ResidentialDist: residentialDist,
		Score:           score,
		Recommendation:  recommendation,
	}

	tmpl, _ := template.ParseFiles("templates/results.html")
	tmpl.Execute(w, result)
}
