package auth

import (
	"database/sql"
	"encoding/json"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var oauth2State = "random" // Change this to a random string for security

// Initialize OAuth2 configuration
func getOAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
		Scopes:       []string{"email"},
		Endpoint:     google.Endpoint,
	}
}

// Send OTP Handler
func SendOTPHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request OTPRequest
		// Decode request body
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate email domain
		if !strings.HasSuffix(request.Email, "@"+os.Getenv("EMAIL_DOMAIN")) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Access Denied</title>
  <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
  <style>
    body { background: linear-gradient(135deg, #1e3c72 0%, #2a5298 100%); min-height: 100vh; display: flex; align-items: center; justify-content: center; }
    .error-card { background: #fff; border-radius: 1.5rem; box-shadow: 0 0 32px 0 rgba(0,0,0,0.2); padding: 2.5rem 2rem; text-align: center; max-width: 400px; border: 3px solid #dc3545; }
    .error-title { color: #dc3545; font-size: 2rem; font-weight: bold; margin-bottom: 1rem; }
    .error-desc { color: #333; font-size: 1.1rem; margin-bottom: 2rem; }
  </style>
</head>
<body>
  <div class="error-card">
    <div class="error-title">Access Denied</div>
    <div class="error-desc">Only <span style="color:#dc3545;font-weight:bold;">BITS Pilani, Pilani Campus</span> people allowed for now.</div>
    <a href="/index.html" class="btn btn-primary">Back to Home</a>
  </div>
</body>
</html>`))
			return
		}

		// Send OTP
		err := SendOTP(db, request.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Respond with success message
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OTP sent successfully"))
	}
}

// Verify OTP Handler
func VerifyOTPHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request VerifyOTPRequest
		// Decode request body
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Verify OTP
		token, err := VerifyOTP(db, request.Email, request.Code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Respond with JWT token
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}

// Google OAuth Login Handler
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect user to Google OAuth consent screen
	url := getOAuth2Config().AuthCodeURL(oauth2State, oauth2.AccessTypeOffline)
	// Debug print for the generated OAuth URL
	fmt.Println("Generated Google OAuth URL:", url)
	http.Redirect(w, r, url, http.StatusFound)
}

// Google OAuth Callback Handler
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Exchange the code for a token
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	token, err := getOAuth2Config().Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Get the user's email from the Google token
	client := getOAuth2Config().Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v1/userinfo?alt=json")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Parse the response and extract email
	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	// Check if the email domain is valid
	if !strings.HasSuffix(userInfo.Email, "@pilani.bits-pilani.ac.in") {
		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Access Denied</title>
  <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
  <style>
    body { background: linear-gradient(135deg, #1e3c72 0%, #2a5298 100%); min-height: 100vh; display: flex; align-items: center; justify-content: center; }
    .error-card { background: #fff; border-radius: 1.5rem; box-shadow: 0 0 32px 0 rgba(0,0,0,0.2); padding: 2.5rem 2rem; text-align: center; max-width: 400px; border: 3px solid #dc3545; }
    .error-title { color: #dc3545; font-size: 2rem; font-weight: bold; margin-bottom: 1rem; }
    .error-desc { color: #333; font-size: 1.1rem; margin-bottom: 2rem; }
  </style>
</head>
<body>
  <div class="error-card">
    <div class="error-title">Access Denied</div>
    <div class="error-desc">Only <span style="color:#dc3545;font-weight:bold;">BITS Pilani, Pilani Campus</span> people allowed for now.</div>
    <a href="/index.html" class="btn btn-primary">Back to Home</a>
  </div>
</body>
</html>`))
		return
	}

	// User is authenticated, create a JWT token
	tokenString, err := createJWT(userInfo.Email)
	if err != nil {
		http.Error(w, "Failed to create JWT", http.StatusInternalServerError)
		return
	}

	// Redirect to dashboard with token in URL fragment
	http.Redirect(w, r, "/dashboard.html#token="+tokenString, http.StatusFound)
}

// Function to create a JWT token for the user
func createJWT(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// Profile completion handlers
func ProfileGetHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := getEmailFromJWT(r)
		if email == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var name, branch, batch string
		err := db.QueryRow("SELECT full_name, year, department FROM users WHERE email=$1", email).Scan(&name, &batch, &branch)
		if err != nil || name == "" || branch == "" || batch == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"profileComplete":true}`))
	}
}

func ProfilePostHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := getEmailFromJWT(r)
		if email == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var req struct {
			Name  string `json:"name"`
			Age   string `json:"age"`
			Branch string `json:"branch"`
			Batch string `json:"batch"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Println("Profile update attempt for email:", email, "data:", req)
		res, err := db.Exec(`
  INSERT INTO users (id, email, full_name, year, department, verified)
  VALUES (gen_random_uuid(), $4, $1, $2, $3, TRUE)
  ON CONFLICT (email) DO UPDATE SET full_name=$1, year=$2, department=$3
`, req.Name, req.Batch, req.Branch, email)
		if err != nil {
			fmt.Println("Profile update error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func getEmailFromJWT(r *http.Request) string {
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		return ""
	}
	token = strings.TrimPrefix(token, "Bearer ")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	email, _ := claims["email"].(string)
	return email
}
