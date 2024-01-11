package jwt

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"sso_go_grpc/internal/domain/models"
	"time"
)

func NewToken(user *models.User, secret string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["uid"] = user.UserId
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(time.Hour * 48).Unix()

	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		return "", fmt.Errorf("INTERNAL SERVER ERROR")
	}

	return tokenString, nil
}
