package auth

import "time"

type OTPRequest struct {
	Email string `json:"email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type OTP struct {
	Email     string
	Code      string
	ExpiresAt time.Time
}
