package utils

import (
	"fmt"
	"strings"

	"github.com/astarte-platform/astartectl/common"
)

// ValidateInterfacePath validates path against the structure of astarteInterface, and returns a meaningful error
// the path cannot be resolved.
func ValidateInterfacePath(astarteInterface common.AstarteInterface, interfacePath string) error {
	isAggregate := astarteInterface.Aggregation == common.ObjectAggregation
	// Ensure the path resolves the parameter
	interfacePathTokens := strings.Split(interfacePath, "/")
	if isAggregate {
		// Any mapping is fine here, so just take the first
		validationEndpointTokens := strings.Split(astarteInterface.Mappings[0].Endpoint, "/")
		if len(interfacePathTokens) > len(validationEndpointTokens) {
			return fmt.Errorf("Path %s does not exist on Interface %s", interfacePath, astarteInterface.Name)
		}
		if astarteInterface.IsParametric() {
			for index, token := range validationEndpointTokens {
				if len(interfacePathTokens) < index {
					return fmt.Errorf("%s is not a valid path for retrieving samples on Interface %s", interfacePath, astarteInterface.Name)
				}
				if strings.HasPrefix(token, "%{") {
					if len(interfacePathTokens) == index+1 {
						// The user requested the parametric endpoint of an aggregate, go for it.
						return nil
					}
				} else if token != interfacePathTokens[index] {
					// It could be the last one - in that case, it's ok.
					if len(validationEndpointTokens) != index+1 {
						return fmt.Errorf("Cannot resolve path %s on Interface %s", interfacePath, astarteInterface.Name)
					}
				}
				// This token is valid.
			}
		} else {
			// Just verify in which range we're moving
			if len(interfacePathTokens) == len(validationEndpointTokens) && interfacePath != "/" {
				// Is the path valid?
				return simpleMappingValidation(astarteInterface, interfacePath)
			}
		}
	} else {
		// Ensure we're matching exactly one of the mappings.
		if !astarteInterface.IsParametric() {
			return simpleMappingValidation(astarteInterface, interfacePath)
		}

		for _, mapping := range astarteInterface.Mappings {
			mappingTokens := strings.Split(mapping.Endpoint, "/")
			if len(mappingTokens) != len(interfacePathTokens) {
				continue
			}
			// Iterate
			matchFound := true
			for index, token := range mappingTokens {
				if interfacePathTokens[index] != token && !strings.HasPrefix(token, "%{") {
					matchFound = false
					break
				}
			}
			if matchFound {
				return nil
			}
		}
		return fmt.Errorf("Path %s does not exist on Interface %s", interfacePath, astarteInterface.Name)
	}

	// The path is valid
	return nil
}

func simpleMappingValidation(astarteInterface common.AstarteInterface, interfacePath string) error {
	// Is the path valid?
	for _, mapping := range astarteInterface.Mappings {
		if mapping.Endpoint == interfacePath {
			return nil
		}
	}
	return fmt.Errorf("Path %s does not exist on Interface %s", interfacePath, astarteInterface.Name)
}
