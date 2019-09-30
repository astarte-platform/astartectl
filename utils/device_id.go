package utils

import (
	"encoding/base64"

	"github.com/google/uuid"
)

// IsValidAstarteDeviceID returns whether the provided Device ID is a valid Astarte Device ID or not.
func IsValidAstarteDeviceID(deviceID string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(deviceID)
	if err != nil {
		return false
	}

	// 16 bytes == 128 bit
	if len(decoded) != 16 {
		return false
	}

	return true
}

// GenerateRandomAstarteDeviceID returns a new Astarte Device ID on a fully Random basis
func GenerateRandomAstarteDeviceID() (string, error) {
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	deviceID, err := randomUUID.MarshalBinary()
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(deviceID), nil
}

// GetNamespacedAstarteDeviceID returns an Astarte Device ID generated from a namespaced arbitrary payload.
// It is guaranteed to be always the same for the same namespace and payload
func GetNamespacedAstarteDeviceID(uuidNamespace string, payloadData []byte) (string, error) {
	encodedUUIDNamespace, err := uuid.Parse(uuidNamespace)
	if err != nil {
		return "", err
	}

	deviceUUID := uuid.NewSHA1(encodedUUIDNamespace, payloadData)

	deviceID, err := deviceUUID.MarshalBinary()
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(deviceID), nil
}
