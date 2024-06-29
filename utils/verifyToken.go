package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/dgrijalva/jwt-go"
	"github.com/pmas98/go-auth-service/models"
)

func verifyToken(tokenString string) (bool, int, string, string) {
	var JWTKey []byte
	const encodedKey = "jubGgkvgopNQeq2NMzDt4EzwENu8EvgU+ed8V59OJOU="
	JWTKey, _ = base64.StdEncoding.DecodeString(encodedKey)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWTKey, nil
	})
	if err != nil {
		log.Printf("Error parsing JWT token: %v", err)
		response := models.TokenVerificationResponse{
			Valid:  false,
			UserID: 0,
			Email:  "",
			Name:   "",
		}
		responseJSON, _ := json.Marshal(response)
		SendMessageJSONToKafka("token_verification_responses", responseJSON, "fail")
		return false, 0, "", ""
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		idFloat, idOk := claims["id"].(float64)
		id := int(idFloat)
		name, nameOk := claims["name"].(string)
		email, emailOk := claims["email"].(string)

		if !idOk || !nameOk || !emailOk {
			log.Printf("Error: Invalid claim types in token")
			return false, 0, "", ""
		}

		return true, id, name, email
	}

	log.Printf("Error: Token is not valid")
	return false, 0, "", ""
}
