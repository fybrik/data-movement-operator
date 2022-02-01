// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"
)

const (
	stepNameHashLength       = 10
	hashPostfixLength        = 5
	k8sMaxConformNameLength  = 63
	helmMaxConformNameLength = 53
)

// IsDenied returns true if the data access is denied
func IsDenied(actionName string) bool {
	return (actionName == "Deny") // TODO FIX THIS
}

// StructToMap converts a struct to a map using JSON marshal
func StructToMap(data interface{}) (map[string]interface{}, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	mapData := make(map[string]interface{})
	err = json.Unmarshal(dataBytes, &mapData)
	if err != nil {
		return nil, err
	}
	return mapData, nil
}

// Hash generates a name based on the unique identifier
func Hash(value string, hashLength int) string {
	data := sha256.Sum256([]byte(value))
	hashedStr := hex.EncodeToString(data[:])
	if hashLength >= len(hashedStr) {
		return hashedStr
	}
	return hashedStr[:hashLength]
}

// Generating release name based on blueprint module
func GetReleaseName(applicationName, namespace, instanceName string) string {
	return GetReleaseNameByStepName(applicationName, namespace, instanceName)
}

// Generate release name from blueprint module name
func GetReleaseNameByStepName(applicationName, namespace, moduleInstanceName string) string {
	fullName := applicationName + "-" + namespace + "-" + moduleInstanceName
	return HelmConformName(fullName)
}

// Generate fqdn for a module
func GenerateModuleEndpointFQDN(releaseName, blueprintNamespace string) string {
	return releaseName + "." + blueprintNamespace + ".svc.cluster.local"
}

// Some k8s objects only allow for a length of 63 characters.
// This method shortens the name keeping a prefix and using the last 5 characters of the
// new name for the hash of the postfix.
func K8sConformName(name string) string {
	return ShortenedName(name, k8sMaxConformNameLength, hashPostfixLength)
}

// Helm has stricter restrictions than K8s and restricts release names to 53 characters
func HelmConformName(name string) string {
	return ShortenedName(name, helmMaxConformNameLength, hashPostfixLength)
}

// Create a name for a step in a blueprint.
// Since this is part of the name of a release, this should be done in a central location to make testing easier
func CreateStepName(moduleName, assetID string) string {
	return moduleName + "-" + Hash(assetID, stepNameHashLength)
}

// This function shortens a name to the maximum length given and uses rest of the string that is too long
// as hash that gets added to the valid name.
func ShortenedName(name string, maxLength, hashLength int) string {
	if len(name) > maxLength {
		// The new name is in the form prefix-suffix
		// The prefix is the prefix of the original name (so it's human readable)
		// The suffix is a deterministic hash of the suffix of the original name
		// Overall, the new name is deterministic given the original name
		cutOffIndex := maxLength - hashLength - 1
		prefix := name[:cutOffIndex]
		suffix := Hash(name[cutOffIndex:], hashLength)
		return prefix + "-" + suffix
	}
	return name
}

func ListeningAddress(port int) string {
	address := fmt.Sprintf(":%d", port)
	if runtime.GOOS == "darwin" {
		address = "localhost" + address
	}
	return address
}
