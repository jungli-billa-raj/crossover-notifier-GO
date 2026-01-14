package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Your Credentials (In a real app, these come from your SQLite DB or Fyne Preferences)
const (
	API_KEY    = "your_api_key_here"
	API_SECRET = "your_api_secret_here"
	USER_ID    = "your_user_id_here"
)

func main() {
	myApp := app.New()
	window := myApp.NewWindow("Flattrade Auth Engine")

	statusLabel := widget.NewLabel("Status: Waiting for Login...")

	// 1. Create a channel to receive the request_code from our local server
	codeChan := make(chan string)

	// 2. Start the local server to listen for the redirect
	go func() {
		http.HandleFunc("/flattrade/callback", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("request_code")
			if code != "" {
				fmt.Fprintf(w, "Login Successful! You can close this tab.")
				codeChan <- code
			}
		})
		http.ListenAndServe(":5000", nil)
	}()

	// 3. UI Button to trigger login
	loginBtn := widget.NewButton("Login to Flattrade", func() {
		authURL := fmt.Sprintf("https://auth.flattrade.in/?app_key=%s", API_KEY)
		u, _ := url.Parse(authURL)
		myApp.OpenURL(u) // Opens the browser for the user
		statusLabel.SetText("Status: Please log in via browser...")
	})

	// 4. Background listener for the request_code
	go func() {
		for code := range codeChan {
			statusLabel.SetText("Status: Received code, exchanging for token...")
			token, err := exchangeToken(code)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Error: %v", err))
			} else {
				statusLabel.SetText(fmt.Sprintf("Logged in! Token: %s...", token[:10]))
				// SAVE THIS TOKEN: You'll need it for every API call today
			}
		}
	}()

	window.SetContent(container.NewVBox(
		widget.NewLabel("Flattrade API Connector"),
		loginBtn,
		statusLabel,
	))

	window.Resize(fyne.NewSize(400, 200))
	window.ShowAndRun()
}

// 5. The Token Exchange Logic
func exchangeToken(code string) (string, error) {
	// Generate SHA256 Checksum: api_key + request_code + api_secret
	rawString := API_KEY + code + API_SECRET
	hash := sha256.Sum256([]byte(rawString))
	apiHash := hex.EncodeToString(hash[:])

	payload := map[string]string{
		"api_key":      API_KEY,
		"request_code": code,
		"api_secret":   apiHash,
	}

	jsonPayload, _ := json.Marshal(payload)
	resp, err := http.Post("https://authapi.flattrade.in/trade/apitoken", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["status"] == "Ok" {
		return result["token"].(string), nil
	}
	return "", fmt.Errorf("auth failed: %v", result["emsg"])
}
