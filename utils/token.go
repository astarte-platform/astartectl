package utils

import (
	"io/ioutil"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// GenerateAstarteJWTFromKeyFile generates an Astarte Token for a specific API out of a Private Key File
func GenerateAstarteJWTFromKeyFile(privateKeyFile string, astarteService AstarteService,
	authorizationClaims []string, ttlSeconds int64) (jwtString string, err error) {
	keyPEM, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		return "", err
	}

	return GenerateAstarteJWTFromPEMKey(keyPEM, astarteService, authorizationClaims, ttlSeconds)
}

// GenerateAstarteJWTFromPEMKey generates an Astarte Token for a specific API out of a Private Key PEM bytearray
func GenerateAstarteJWTFromPEMKey(privateKeyPEM []byte, astarteService AstarteService,
	authorizationClaims []string, ttlSeconds int64) (jwtString string, err error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return "", err
	}

	accessClaimKey := astarteService.JwtClaim()

	if len(authorizationClaims) == 0 {
		switch astarteService {
		case Channels:
			authorizationClaims = []string{"JOIN::.*", "WATCH::.*"}
		default:
			authorizationClaims = []string{"^.*$::^.*$"}
		}
	}

	now := time.Now().UTC().Unix()
	mapClaims := jwt.MapClaims{
		accessClaimKey: authorizationClaims,
		"iat":          now,
	}
	if ttlSeconds > 0 {
		exp := now + ttlSeconds
		mapClaims["exp"] = exp
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, mapClaims)

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
