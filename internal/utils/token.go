package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/oauth2"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultTokenPath = "token.txt"

func WriteTokenToFile(token *oauth2.Token) error {
	tokenString := fmt.Sprintf("%s|%s|%s|%d",
		token.AccessToken, token.RefreshToken, token.TokenType, token.Expiry.Unix())

	err := os.WriteFile(defaultTokenPath, []byte(tokenString), 0644)
	if err != nil {
		return err
	}

	return nil
}

func ParseTokenString(tokenString string) (*oauth2.Token, error) {
	parts := strings.Split(tokenString, "|")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid token string")
	}

	// Parse the token fields from the parts array
	accessToken := parts[0]
	refreshToken := parts[1]
	tokenType := parts[2]
	expiryUnix, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return nil, err
	}

	expiryTime := time.Unix(expiryUnix, 0)

	token := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		Expiry:       expiryTime,
	}

	return token, nil
}

func GenerateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
