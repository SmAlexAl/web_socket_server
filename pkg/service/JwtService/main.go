package JwtService

import (
	"database/sql"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/dgrijalva/jwt-go/v4"
	"log"
	"os"
	"time"
)

type Claims struct {
	StandardClaims jwt.StandardClaims
	PackageName    string
	Udid           string
	InstallId      string
	Platform       string
}

func (c Claims) Valid(helper *jwt.ValidationHelper) error {
	return c.StandardClaims.Valid(helper)
}

func GenerateToken(tokenData *TokenData) (string, error) {
	var jwtKey = []byte(os.Getenv("JWT_TOKEN_SIGNATURE"))

	atClaims := Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwt.At(time.Now().Add(time.Second * 500)),
			IssuedAt:  jwt.At(time.Now()),
		},
		PackageName: tokenData.ApplicationPackageName,
		Udid:        tokenData.Udid,
		InstallId:   tokenData.InstallId,
		Platform:    tokenData.Platform,
	}

	fmt.Println(atClaims.StandardClaims.ExpiresAt)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)

	tokenString, err := token.SignedString(jwtKey)
	spew.Dump(tokenString)
	return tokenString, err
}

func ParseToken(token string, conn *sql.DB) (bool, *ProfileDto) {
	var jwtKey = []byte(os.Getenv("JWT_TOKEN_SIGNATURE"))

	tokenNew, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		log.Println(err)
		return false, nil
	}

	claims, _ := tokenNew.Claims.(*Claims)

	expiresAt := claims.StandardClaims.ExpiresAt

	if time.Now().After(expiresAt.Time) {
		return false, nil
	}

	tokenData := TokenData{
		ApplicationPackageName: claims.PackageName,
		Udid:                   claims.Udid,
		InstallId:              claims.InstallId,
		Platform:               claims.Platform,
	}

	profileDto := GetProfileData(conn, tokenData)

	if profileDto == nil {
		return false, nil
	}

	return true, profileDto
}
