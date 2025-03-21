package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// Email configuration struct
type Config struct {
	Port          string
	EmailFrom     string
	EmailPassword string
	EmailTo       []string
	SmtpHost      string
	SmtpPort      string
}

// Required environment variables
var requiredEnvVars = []string{
	"EMAIL_FROM",
	"EMAIL_PASSWORD",
	"EMAIL_TO",
	"SMTP_HOST",
	"SMTP_PORT",
}

// Load configuration from environment variables
func loadConfig() (Config, error) {
	// Check for required environment variables
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			return Config{}, fmt.Errorf("required environment variable %s is not set", envVar)
		}
	}

	// Get comma-separated list of recipients
	emailToEnv := os.Getenv("EMAIL_TO")
	emailToList := strings.Split(emailToEnv, ",")

	// Trim whitespace from each email
	for i, email := range emailToList {
		emailToList[i] = strings.TrimSpace(email)
	}

	return Config{
		Port:          getEnvWithDefault("PORT", "3333"),
		EmailFrom:     os.Getenv("EMAIL_FROM"),
		EmailPassword: os.Getenv("EMAIL_PASSWORD"),
		EmailTo:       emailToList,
		SmtpHost:      os.Getenv("SMTP_HOST"),
		SmtpPort:      os.Getenv("SMTP_PORT"),
	}, nil
}

// Helper function to get environment variable with default
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func mailDelivered(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("got /mail-delivered request\n")

		// Get current time for the subject with Pacific Time timezone
		loc, err := time.LoadLocation("America/Los_Angeles")
		if err != nil {
			// Fallback to UTC if timezone loading fails
			fmt.Printf("Error loading timezone: %s, falling back to UTC\n", err)
			loc = time.UTC
		}
		currentTime := time.Now().In(loc).Format("Jan 2, 2006 at 3:04 PM MST")
		fmt.Printf("Using time: %s\n", currentTime)

		// Parse email configuration
		from := config.EmailFrom
		password := config.EmailPassword
		to := config.EmailTo
		smtpHost := config.SmtpHost
		smtpPort := config.SmtpPort

		// Updated message with better headers and content
		subject := fmt.Sprintf("Mail Delivered - %s", currentTime)
		body := "You've got mail in your mailbox!\n\nThis notification was sent from your home mailbox system.\n\nBest regards,\nYour Mailbox"

		// Extract sender email for headers
		var senderEmail string
		fromParts := strings.Split(from, "<")
		if len(fromParts) > 1 {
			senderEmail = strings.TrimSuffix(fromParts[1], ">")
		} else {
			senderEmail = from
		}

		// Basic email headers
		message := []byte("From: " + from + "\r\n" +
			"To: " + strings.Join(to, ",") + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body + "\r\n")

		// Create a custom TLS configuration that skips verification
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         smtpHost,
		}

		// Connect to the server
		client, err := smtp.Dial(smtpHost + ":" + smtpPort)
		if err != nil {
			fmt.Printf("Error connecting to mail server: %s\n", err)
			returnJSONError(w, "Failed to connect to mail server")
			return
		}
		defer client.Close()

		// Start TLS with our custom config
		if err = client.StartTLS(tlsConfig); err != nil {
			fmt.Printf("Error starting TLS: %s\n", err)
			returnJSONError(w, "Failed to start TLS")
			return
		}

		// Authenticate
		auth := smtp.PlainAuth("", senderEmail, password, smtpHost)
		if err = client.Auth(auth); err != nil {
			fmt.Printf("Error authenticating: %s\n", err)
			returnJSONError(w, "Failed to authenticate")
			return
		}

		// Set sender
		if err = client.Mail(senderEmail); err != nil {
			fmt.Printf("Error setting sender: %s\n", err)
			returnJSONError(w, "Failed to set sender")
			return
		}

		// Set recipients
		for _, recipient := range to {
			if err = client.Rcpt(recipient); err != nil {
				fmt.Printf("Error setting recipient %s: %s\n", recipient, err)
				returnJSONError(w, "Failed to set recipient")
				return
			}
		}

		// Send the message
		w1, err := client.Data()
		if err != nil {
			fmt.Printf("Error getting data writer: %s\n", err)
			returnJSONError(w, "Failed to get data writer")
			return
		}

		_, err = w1.Write(message)
		if err != nil {
			fmt.Printf("Error writing message: %s\n", err)
			returnJSONError(w, "Failed to write message")
			return
		}

		err = w1.Close()
		if err != nil {
			fmt.Printf("Error closing data writer: %s\n", err)
			returnJSONError(w, "Failed to close data writer")
			return
		}

		// Quit the connection
		client.Quit()

		// Return success JSON response
		returnJSONSuccess(w)
	}
}

// Helper function to return a JSON success response
func returnJSONSuccess(w http.ResponseWriter) {
	response := map[string]bool{"success": true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to return a JSON error response
func returnJSONError(w http.ResponseWriter, errMsg string) {
	response := map[string]interface{}{
		"success": false,
		"error":   errMsg,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(response)
}

func main() {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %s\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mail-delivered", mailDelivered(config))

	fmt.Printf("Starting server on port %s...\n", config.Port)
	err = http.ListenAndServe(":"+config.Port, mux)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
