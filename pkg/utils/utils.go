package utils

import (
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

type ApiResponse struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Data       any    `json:"data,omitempty"`
	Error      string `json:"error,omitempty"`
}

func GetEnvAsInt() int {
	if valueStr := os.Getenv("PG_PORT"); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return 5433
}


func HashString(data string)(string, error){
	hash, err := bcrypt.GenerateFromPassword([]byte(data), bcrypt.DefaultCost)
	return string(hash), err
}

func CompareHash(hash, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}