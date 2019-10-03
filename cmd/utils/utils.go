// Copyright © 2019 Ispirata Srl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
)

// UtilsCmd represents the utils command
var UtilsCmd = &cobra.Command{
	Use:   "utils",
	Short: "Various utilities to interact with Astarte",
	Long:  `utils includes commands to generate keypairs and device ids`,
}

var genKeypairCmd = &cobra.Command{
	Use:   "gen-keypair <realm_name>",
	Short: "Generate an RSA keypair",
	Long: `Generate an RSA keypair to use for realm authentication.

The keypair will be saved in the current directory with names <realm_name>_private.pem and <realm_name>_public.pem`,
	Example: `  astartectl utils gen-keypair myrealm`,
	Args:    cobra.ExactArgs(1),
	RunE:    genKeypairF,
}

var jwtTypesToClaim = map[string]string{
	"housekeeping":     "a_ha",
	"realm-management": "a_rma",
	"pairing":          "a_pa",
	"appengine":        "a_aea",
	"channels":         "a_ch",
}

var jwtTypes = []string{"housekeeping", "realm-management", "pairing", "appengine", "channels"}

var genJwtCmd = &cobra.Command{
	Use:       "gen-jwt <type>",
	Short:     "Generate a JWT",
	Long:      `Generate a JWT to access one of astarte APIs.`,
	Example:   `  astartectl utils gen-jwt realm-management -p test-realm.key`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: jwtTypes,
	RunE:      genJwtF,
}

func init() {
	genJwtCmd.Flags().StringP("private-key", "p", "", `Path to PEM encoded private key.
Should be Housekeeping key to generate an housekeeping token, Realm key for everything else.`)
	genJwtCmd.MarkFlagRequired("private-key")
	genJwtCmd.MarkFlagFilename("private-key")
	genJwtCmd.Flags().StringSliceP("claims", "c", nil, `The list of claims to be added in the JWT. Defaults to all-access claims.
You can specify the flag multiple times or separate the claims with a comma.`)
	genJwtCmd.Flags().Int64P("expiry", "e", 300, "Expiration time of the token in seconds. 0 means the token will never expire.")

	UtilsCmd.AddCommand(genKeypairCmd)
	UtilsCmd.AddCommand(genJwtCmd)
}

func genKeypairF(command *cobra.Command, args []string) error {
	realm := args[0]

	reader := rand.Reader
	bitSize := 4096

	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return err
	}
	checkError(err)

	publicKey := key.PublicKey

	fmt.Println("Keypair generated successfully")

	savePEMKey(realm+"_private.pem", key)
	savePublicPEMKey(realm+"_public.pem", publicKey)

	return nil
}

func validJwtType(t string) bool {
	for _, validType := range jwtTypes {
		if t == validType {
			return true
		}
	}
	return false
}

func genJwtF(command *cobra.Command, args []string) error {
	jwtType := args[0]
	astarteService, err := utils.AstarteServiceFromString(jwtType)
	if err != nil {
		errorString := fmt.Sprintf("Invalid type. Valid types are: %s", strings.Join(jwtTypes, ", "))

		return errors.New(errorString)
	}

	privateKey, err := command.Flags().GetString("private-key")
	if err != nil {
		return err
	}

	expiryOffset, err := command.Flags().GetInt64("expiry")
	if err != nil {
		return err
	}

	accessClaims, err := command.Flags().GetStringSlice("claims")
	if err != nil {
		return err
	}

	tokenString, err := utils.GenerateAstarteJWTFromKeyFile(privateKey, astarteService, accessClaims, expiryOffset)
	if err != nil {
		return err
	}

	fmt.Println(tokenString)

	return nil
}

func savePEMKey(fileName string, key *rsa.PrivateKey) {
	outFile, err := os.Create(fileName)
	checkError(err)
	defer outFile.Close()

	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	err = pem.Encode(outFile, privateKey)
	checkError(err)

	fmt.Println("Wrote " + fileName)
}

func savePublicPEMKey(fileName string, pubkey rsa.PublicKey) {
	pkixBytes, err := x509.MarshalPKIXPublicKey(&pubkey)
	checkError(err)
	var pemkey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pkixBytes,
	}

	pemfile, err := os.Create(fileName)
	checkError(err)
	defer pemfile.Close()

	err = pem.Encode(pemfile, pemkey)
	checkError(err)

	fmt.Println("Wrote " + fileName)
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
