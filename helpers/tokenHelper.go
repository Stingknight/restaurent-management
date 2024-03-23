package helpers

import (
	"errors"
	"log"
	// "os"

	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


var SECRET_KEY = "EHUIQHRUEWRUEWBR"

func GenerateAllToken(objectID primitive.ObjectID) (signedToken string) {

	claims := jwt.MapClaims{
		"sub": objectID,
		"exp": time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Fatalf("Error %v", err)
		return ""
	}

	return tokenString

}

func ValidateToken(signedToken string) (claims jwt.MapClaims, err error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, _ := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(SECRET_KEY), nil
	})
	
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			err =errors.New("token has expired")
			return claims,err
		}
		
		return claims,nil

	} else {
		err =errors.New("token has expired")
		return claims,err
	}

}
