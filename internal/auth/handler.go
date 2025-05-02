package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

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
			http.Error(w, "Only university emails allowed", http.StatusUnauthorized)
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
