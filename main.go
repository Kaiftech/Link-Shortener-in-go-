package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var shortToURL = make(map[string]string)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Generate a random short ID
func generateShortID(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// Handler for shortening URLs
func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	type request struct {
		URL string `json:"url"`
	}

	var body request
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil || body.URL == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate a short ID for the URL
	id := generateShortID(6)
	shortToURL[id] = body.URL

	// Get the base URL from the environment variable
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // fallback for local dev
	}

	// Create the short URL
	shortLink := fmt.Sprintf("%s/short/%s", baseURL, id)

	// Return the short link as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"short": shortLink,
	})
}

// Handler for redirecting to the original URL
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	// Get the short ID from the URL path (remove "/short/")
	id := strings.TrimPrefix(r.URL.Path, "/short/")
	if originalURL, ok := shortToURL[id]; ok {
		// If short link exists, redirect to the original URL
		http.Redirect(w, r, originalURL, http.StatusFound)
		return
	}
	// Return 404 if the short link doesn't exist
	http.NotFound(w, r)
}

// Home page handler to show the URL shortening form
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintln(w, `<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>URL Shortener</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				background-color: #f5f5f5;
				margin: 0;
				padding: 0;
			}
			.container {
				max-width: 600px;
				margin: 50px auto;
				padding: 20px;
				background: #fff;
				box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
			}
			h1 {
				text-align: center;
				color: #333;
			}
			input[type="text"] {
				width: 100%;
				padding: 10px;
				margin: 10px 0;
				border: 1px solid #ccc;
				border-radius: 4px;
			}
			button {
				width: 100%;
				padding: 10px;
				background-color: #4CAF50;
				border: none;
				color: white;
				font-size: 16px;
				border-radius: 4px;
			}
			.result {
				margin-top: 20px;
				text-align: center;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>URL Shortener</h1>
			<form id="urlForm">
				<input type="text" id="urlInput" placeholder="Enter URL here" required>
				<button type="submit">Shorten URL</button>
			</form>
			<div class="result" id="result"></div>
		</div>
		<script>
			document.getElementById('urlForm').addEventListener('submit', async function(event) {
				event.preventDefault();
				const urlInput = document.getElementById('urlInput').value;
				const response = await fetch('/shorten', {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
					},
					body: JSON.stringify({ url: urlInput }),
				});
				const data = await response.json();
				if (data.short) {
					document.getElementById('result').innerHTML = '<p><strong>Shortened URL: </strong><a href="' + data.short + '" target="_blank">' + data.short + '</a></p>';
				} else {
					document.getElementById('result').innerHTML = '<p>Error: ' + data.error + '</p>';
				}
			});
		</script>
	</body>
	</html>`)
}

func main() {
	// Handle routes
	http.HandleFunc("/", homeHandler)           // Home page
	http.HandleFunc("/shorten", shortenHandler) // Shorten URL handler
	http.HandleFunc("/short/", redirectHandler) // Redirect handler for shortened URLs

	// Get the port to run the server on (default to 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback port
	}

	fmt.Println("🚀 Server running on port", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
