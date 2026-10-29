package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Lead struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Company string `json:"company"`
}

func main() {
	// Use PORT environment variable if available (required by Render.com), otherwise 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Handle API for leads
	http.HandleFunc("/api/lead", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var lead Lead
		if err := json.NewDecoder(r.Body).Decode(&lead); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Only send to Google Sheets
		scriptURL := "https://script.google.com/macros/s/AKfycbxl5C7hm8zDQKeM2E7-SJ1F3o0lvGYk_PIpPAO4Q5WR1Kye05xD_IVV0KcDJsB8ip3C/exec"
		formData := url.Values{
			"name":    {lead.Name},
			"phone":   {lead.Phone},
			"company": {lead.Company},
		}
		go func() {
			resp, err := http.PostForm(scriptURL, formData)
			if err != nil {
				log.Printf("Error sending to Google Sheets: %v\n", err)
			} else {
				defer resp.Body.Close()
				log.Printf("Google Sheets response status: %s\n", resp.Status)
			}
		}()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Lead saved successfully"})
	})

	// Admin route to download leads
	http.HandleFunc("/admin/download-leads", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", "attachment; filename=leads.csv")
		w.Header().Set("Content-Type", "text/csv")
		http.ServeFile(w, r, "leads.csv")
	})

	// Create a file server for the "static" directory
	fs := http.FileServer(http.Dir("./static"))
	
	// Handle all incoming requests by serving the static files
	http.Handle("/", fs)

	log.Printf("Server listening on port %s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
