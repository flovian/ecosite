package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// Struct for results
type Result struct {
	Location        string
	WaterDist       float64
	ForestDist      float64
	ResidentialDist float64
	Score           int
	Recommendation  string
	AIExplanation   string
}

type GeminiResponse struct {
	Score          int    `json:"score"`
	Recommendation string `json:"recommendation"`
	Explanation    string `json:"explanation"`
}

func main() {
	// Load .env file (if it exists)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	// Serve static files if needed
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/analyze", analyzeHandler)
	http.HandleFunc("/results", resultsHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/signup", signupHandler)

	// Start server
	fmt.Println("Server successfully started! Navigate to http://localhost:8080 in your web browser.")
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

	apiKey := os.Getenv("GEMINI_API_KEY")

	result := Result{
		Location:        location,
		WaterDist:       waterDist,
		ForestDist:      forestDist,
		ResidentialDist: residentialDist,
	}

	if apiKey != "" {
		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if err == nil {
			defer client.Close()
			model := client.GenerativeModel("gemini-2.5-flash")
			model.ResponseMIMEType = "application/json"

			prompt := fmt.Sprintf(`You are an environmental analysis AI for EcoSite.
A proposal has been made to establish an industrial site at "%s" with the following characteristics:
- Distance to Water: %.2f km
- Distance to Forest: %.2f km
- Distance to Residential Area: %.2f km

Analyze the environmental and social risks. Consider air quality, water pollution, and ecological sensitivity.
Output your response in strictly valid JSON format with the following keys:
- "score": An integer from 0-100 indicating sustainability (100 is best).
- "recommendation": A short string (e.g., "Safe", "Moderate Risk", "Unsafe").
- "explanation": A detailed string explaining your reasoning.`, location, waterDist, forestDist, residentialDist)

			resp, err := model.GenerateContent(ctx, genai.Text(prompt))
			if err == nil && len(resp.Candidates) > 0 {
				var part genai.Part = resp.Candidates[0].Content.Parts[0]
				if text, ok := part.(genai.Text); ok {
					var geminiResp GeminiResponse
					if err := json.Unmarshal([]byte(text), &geminiResp); err == nil {
						result.Score = geminiResp.Score
						result.Recommendation = geminiResp.Recommendation
						result.AIExplanation = geminiResp.Explanation
					}
				}
			} else {
				log.Printf("Error from Gemini: %v", err)
			}
		} else {
			log.Printf("Failed to create GenAI client: %v", err)
		}
	}

	// Fallback if AI failed or API key missing
	if result.Recommendation == "" {
		score := 0
		if waterDist >= 1 {
			score += 30
		} else if waterDist >= 0.5 {
			score += 15
		}
		if forestDist >= 2 {
			score += 30
		} else if forestDist >= 1 {
			score += 15
		}
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

		result.Score = score
		result.Recommendation = recommendation
		result.AIExplanation = "This is a basic rule-based recommendation. To enable AI analysis, please provide a valid GEMINI_API_KEY environment variable."
	}

	tmpl, _ := template.ParseFiles("templates/results.html")
	tmpl.Execute(w, result)
}

// Login page
func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/login.html")
	tmpl.Execute(w, nil)
}

// Signup page
func signupHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/signup.html")
	tmpl.Execute(w, nil)
}
