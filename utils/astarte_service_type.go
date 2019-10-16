package utils

import (
	"errors"
)

// AstarteService represents one of Astarte's Services
type AstarteService int

const (
	// Unknown Astarte Service
	Unknown AstarteService = 0
	// Housekeeping is Astarte's service for managing Realms
	Housekeeping AstarteService = 1
	// RealmManagement is Astarte's service for managing configuration of a Realm
	RealmManagement AstarteService = 2
	// Pairing is Astarte's service for managing device provisioning and access
	Pairing AstarteService = 3
	// AppEngine is Astarte's service for interacting with Devices, Groups and more
	AppEngine AstarteService = 4
	// Channels is Astarte's service for WebSockets
	Channels AstarteService = 5
)

var astarteServiceToJwtClaim = map[AstarteService]string{
	Housekeeping:    "a_ha",
	RealmManagement: "a_rma",
	Pairing:         "a_pa",
	AppEngine:       "a_aea",
	Channels:        "a_ch",
}

var astarteServiceValidNames = map[string]AstarteService{
	"housekeeping":     Housekeeping,
	"hk":               Housekeeping,
	"realm-management": RealmManagement,
	"realmmanagement":  RealmManagement,
	"realm":            RealmManagement,
	"pairing":          Pairing,
	"appengine":        AppEngine,
	"app":              AppEngine,
	"channels":         Channels,
}

func (astarteService AstarteService) String() string {
	names := [...]string{
		"",
		"housekeeping",
		"realm-management",
		"pairing",
		"appengine",
		"channels"}

	if astarteService < Housekeeping || astarteService > Channels {
		return ""
	}

	return names[astarteService]
}

// JwtClaim returns the corresponding JWT claim associated to the Service (if any)
func (astarteService AstarteService) JwtClaim() string {
	return astarteServiceToJwtClaim[astarteService]
}

// AstarteServiceFromString returns a valid AstarteService out of a string
func AstarteServiceFromString(astarteServiceString string) (AstarteService, error) {
	if value, exist := astarteServiceValidNames[astarteServiceString]; exist {
		return value, nil
	}

	return Unknown, errors.New("Invalid type")
}
