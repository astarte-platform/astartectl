package utils

import (
	"io/ioutil"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// GenerateAstarteJWTFromKeyFile generates an Astarte Token for a specific API out of a Private Key File
func GenerateAstarteJWTFromKeyFile(privateKeyFile string, servicesAndClaims map[AstarteService][]string,
	ttlSeconds int64) (jwtString string, err error) {
	keyPEM, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		return "", err
	}

	return GenerateAstarteJWTFromPEMKey(keyPEM, servicesAndClaims, ttlSeconds)
}

// GenerateAstarteJWTFromPEMKey generates an Astarte Token for a specific API out of a Private Key PEM bytearray
func GenerateAstarteJWTFromPEMKey(privateKeyPEM []byte, servicesAndClaims map[AstarteService][]string,
	ttlSeconds int64) (jwtString string, err error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC().Unix()
	mapClaims := jwt.MapClaims{
		"iat": now,
	}
	if ttlSeconds > 0 {
		exp := now + ttlSeconds
		mapClaims["exp"] = exp
	}

	for svc, claims := range servicesAndClaims {
		accessClaimKey := svc.JwtClaim()

		if len(claims) == 0 {
			switch svc {
			case Channels:
				mapClaims[accessClaimKey] = []string{"JOIN::.*", "WATCH::.*"}
			default:
				mapClaims[accessClaimKey] = []string{"^.*$::^.*$"}
			}
		} else {
			mapClaims[accessClaimKey] = claims
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, mapClaims)

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
