package auth

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ameyg910/unidate/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

func generateOTP() (string, error) {
	bytes := make([]byte, 3)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", int(bytes[0])%1000000), nil
}

func SendOTP(db *sql.DB, email string) error {
	if !strings.HasSuffix(email, "@"+os.Getenv("EMAIL_DOMAIN")) {
		return fmt.Errorf("only university emails allowed")
	}

	otp, err := generateOTP()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(10 * time.Minute)
	_, err = db.Exec("INSERT INTO otps (email, code, expires_at) VALUES ($1, $2, $3) ON CONFLICT (email) DO UPDATE SET code = $2, expires_at = $3",
		email, otp, expiresAt)
	if err != nil {
		return err
	}

	return utils.SendOTPEmail(email, otp)
}

func VerifyOTP(db *sql.DB, email, code string) (string, error) {
	var dbCode string
	var expires time.Time

	err := db.QueryRow("SELECT code, expires_at FROM otps WHERE email=$1", email).Scan(&dbCode, &expires)
	if err != nil {
		return "", fmt.Errorf("OTP not found")
	}

	if code != dbCode || time.Now().After(expires) {
		return "", fmt.Errorf("invalid or expired OTP")
	}

	// Insert or update user
	_, _ = db.Exec("INSERT INTO users (id, email, verified) VALUES (gen_random_uuid(), $1, TRUE) ON CONFLICT (email) DO UPDATE SET verified = TRUE", email)

	// Create JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
