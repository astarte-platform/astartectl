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
	"github.com/spf13/cobra"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"os"
)

// utilsCmd represents the utils command
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

func init() {
	UtilsCmd.AddCommand(genKeypairCmd)
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
	asn1Bytes, err := asn1.Marshal(pubkey)
	checkError(err)

	var pemkey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
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
