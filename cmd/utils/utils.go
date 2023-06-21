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
	"text/tabwriter"

	"github.com/astarte-platform/astarte-go/astarteservices"
	"github.com/astarte-platform/astarte-go/auth"
	"github.com/astarte-platform/astartectl/config"
	"github.com/spf13/cobra"

	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
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
	Long:  `utils includes commands to generate and handle keypairs, device ids and JWT tokens`,
}

var genKeypairCmd = &cobra.Command{
	Use:   "gen-keypair <realm_name>",
	Short: "Generate an ECDSA keypair",
	Long: `Generate an ECDSA keypair to use for realm authentication.

The keypair will be saved in the current directory with names <realm_name>_private.pem and <realm_name>_public.pem`,
	Example: `  astartectl utils gen-keypair myrealm`,
	Args:    cobra.ExactArgs(1),
	RunE:    genKeypairF,
}

var jwtTypes = []string{"housekeeping", "realm-management", "pairing", "appengine", "channels"}

var genJwtCmd = &cobra.Command{
	Use:   "gen-jwt <type> [type]...",
	Short: "Generate a JWT token",
	Long: `Generate a signed JWT to access Astarte APIs.

The token will be signed with the key provided through the -k parameter, and will be valid for the API
sets specified as arguments to gen-jwt. If a config context is set and it holds a valid private key to
either its Realm or Housekeeping and -k is not set, those keys will be used.

Supported API sets are:

appengine - Would generate a token valid for AppEngine API. Requires a Realm key for signing.
channels - Would generate a token valid for Astarte Channels. Requires a Realm key for signing.
flow - Would generate a token valid for Astarte Flow. Requires a Realm key for signing.
realm-management - Would generate a token valid for Realm Management API. Requires a Realm key for signing.
pairing - Would generate a token valid for Pairing API (and, potentially, Device registration). Requires a Realm key for signing.
housekeeping - Would generate a token valid for Housekeeping API. Requires the Housekeeping key for signing.

It is possible to specify more than one API set at a time. If API sets bear incompatible keys (e.g. appengine and housekeeping),
generation will fail. For example, to generate a token which supports both appengine and channels, you would do:

	astartectl utils gen-jwt appengine channels -k test-realm.key

The "all-realm-apis" metatype is provided for convenience: this would generate a token for appengine, realm-management,
channels and pairing. When specified, it should not be combined with anything else.

Claims can also be specified on a per-API set basis when requesting multiple API sets. Prefix a claim with "<api set>:" to apply
the specified claim only to the requested API set. For example:

	astartectl utils gen-jwt all-realm-apis -k test-realm.key -c appengine:GET::* -c pairing:POST::*

Would generate a token with only the desired claims for appengine and pairing.
	`,
	Example:   `  astartectl utils gen-jwt realm-management -k test-realm.key`,
	Args:      cobra.MinimumNArgs(1),
	ValidArgs: jwtTypes,
	RunE:      genJwtF,
}

var showJwtClaimsCmd = &cobra.Command{
	Use:   "show-jwt-claims <token>",
	Short: "Show an Astarte JWT token's claims",
	Long: `Show the set of claims contained in an Astarte JWT token.

Besides standard JWT fields, Astarte JWT tokens contain a set of Astarte claims.
Astarte's data access APIs match the devices' topology like a tree, allowing to declare the
authorization in terms of path allow-listing.
Every claim is an array of regular expressions, which act as a logical OR.
Astarte specific token claims are:
	AppEngine API ("a_aea")
	Realm Management API ("a_rma")
	Housekeeping API ("a_ha")
	Pairing API ("a_pa")
	Channels ("a_ch")
	Flow ("a_f")
`,
	Example: `  astartectl utils show-jwt-claims $TOKEN`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    showJwtClaimsF,
}

func init() {
	genJwtCmd.Flags().StringP("private-key", "k", "", `Path to PEM encoded private key.
Should be Housekeeping key to generate an housekeeping token, Realm key for everything else.`)
	_ = genJwtCmd.MarkFlagFilename("private-key")
	genJwtCmd.Flags().StringSliceP("claims", "c", nil, `The list of claims to be added in the JWT. Defaults to all-access claims.
You can specify the flag multiple times or separate the claims with a comma.`)
	genJwtCmd.Flags().Int64P("expiry", "e", 28800, "Expiration time of the token in seconds. Defaults to 8h. 0 means the token will never expire.")

	UtilsCmd.AddCommand(genKeypairCmd)
	UtilsCmd.AddCommand(genJwtCmd)

	showJwtClaimsCmd.Flags().BoolP("pretty", "p", false, "Whether the output should be pretty-printed.")
	UtilsCmd.AddCommand(showJwtClaimsCmd)
}

func genKeypairF(command *cobra.Command, args []string) error {
	realm := args[0]

	reader := rand.Reader

	key, err := ecdsa.GenerateKey(elliptic.P256(), reader)
	checkError(err)

	publicKey := key.PublicKey

	fmt.Println("Keypair generated successfully")

	savePEMKey(realm+"_private.pem", key)
	savePublicPEMKey(realm+"_public.pem", publicKey)

	return nil
}

func genJwtF(command *cobra.Command, args []string) error {
	servicesAndClaims := map[astarteservices.AstarteService][]string{}

	shouldUseHousekeepingKey := false
	for _, t := range args {
		// Metatype
		if t == "all-realm-apis" {
			if len(args) != 1 {
				return errors.New("When specifying all-realm-apis, no other types can be specified")
			}

			// Add all types
			servicesAndClaims = map[astarteservices.AstarteService][]string{
				astarteservices.AppEngine:       {},
				astarteservices.Channels:        {},
				astarteservices.Flow:            {},
				astarteservices.Pairing:         {},
				astarteservices.RealmManagement: {},
			}

			break
		}

		astarteService, err := astarteservices.FromString(t)
		if err != nil {
			return fmt.Errorf("Invalid type. Valid types are: %s", strings.Join(jwtTypes, ", "))
		}

		if astarteService == astarteservices.Housekeeping {
			if len(args) != 1 {
				return errors.New("Conflicting API types specified. Specify only API sets which require the same key type for signing")
			}
			shouldUseHousekeepingKey = true
		}

		servicesAndClaims[astarteService] = []string{}
	}

	// Compute claims
	accessClaims, err := command.Flags().GetStringSlice("claims")
	if err != nil {
		return err
	}
	for _, claim := range accessClaims {
		// Does it specify an API set-specific claim?
		apiSetSpecific := false
		for _, svc := range jwtTypes {
			if strings.HasPrefix(claim, svc+":") {
				apiSetSpecific = true
				break
			}
		}

		if apiSetSpecific {
			tokens := strings.SplitN(claim, ":", 2)
			astarteService, err := astarteservices.FromString(tokens[0])
			if err != nil {
				return fmt.Errorf("Invalid type specified in claim. Valid types are: %s", strings.Join(jwtTypes, ", "))
			}
			servicesAndClaims[astarteService] = append(servicesAndClaims[astarteService], tokens[1])
		} else {
			for k := range servicesAndClaims {
				servicesAndClaims[k] = append(servicesAndClaims[k], claim)
			}
		}
	}

	expiryOffset, err := command.Flags().GetInt64("expiry")
	if err != nil {
		return err
	}

	var tokenString string

	privateKey, err := command.Flags().GetString("private-key")
	if err != nil {
		return err
	}

	if privateKey == "" {
		// In this case, retrieve the key from the context
		c, err := config.LoadBaseConfiguration(config.GetConfigDir())
		if err != nil {
			return err
		}

		context, err := config.LoadContextConfiguration(config.GetConfigDir(), c.CurrentContext)
		if err != nil {
			return err
		}

		var loadedKey string
		if !shouldUseHousekeepingKey {
			if context.Realm.Key == "" {
				return errors.New("private key not provided, and current context doesn't have a private realm key")
			}
			loadedKey = context.Realm.Key
		} else {
			cluster, err := config.LoadClusterConfiguration(config.GetConfigDir(), context.Cluster)
			if err != nil {
				return err
			}

			if cluster.Housekeeping.Key == "" {
				return errors.New("private key not provided, and current context doesn't have a private housekeeping key")
			}
			loadedKey = cluster.Housekeeping.Key
		}

		decoded, err := base64.StdEncoding.DecodeString(loadedKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		tokenString, err = auth.GenerateAstarteJWTFromPEMKey(decoded, servicesAndClaims, expiryOffset)
		if err != nil {
			return err
		}
	} else {
		tokenString, err = auth.GenerateAstarteJWTFromKeyFile(privateKey, servicesAndClaims, expiryOffset)
		if err != nil {
			return err
		}
	}

	fmt.Println(tokenString)

	return nil
}

func savePEMKey(fileName string, key *ecdsa.PrivateKey) {
	outFile, err := os.Create(fileName)
	checkError(err)
	defer outFile.Close()

	marshaled, err := x509.MarshalECPrivateKey(key)
	checkError(err)

	var privateKey = &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: marshaled,
	}

	err = pem.Encode(outFile, privateKey)
	checkError(err)

	fmt.Println("Wrote " + fileName)
}

func savePublicPEMKey(fileName string, pubkey ecdsa.PublicKey) {
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

func showJwtClaimsF(command *cobra.Command, args []string) error {
	claims, err := auth.GetJWTAstarteClaims(args[0])
	if err != nil {
		return err
	}

	pretty, err := command.Flags().GetBool("pretty")
	if err != nil {
		return err
	}

	if pretty {
		prettyPrintJwtClaims(claims)
	} else {
		out, _ := json.MarshalIndent(claims, "", "    ")
		fmt.Println(string(out))
	}

	return nil
}

func prettyPrintJwtClaims(claims auth.AstarteClaims) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "AppEngine API:\t%v\n", claims.AppEngineAPI)
	fmt.Fprintf(w, "RealmManagement API:\t%v\n", claims.RealmManagement)
	fmt.Fprintf(w, "Housekeeping API:\t%v\n", claims.Housekeeping)
	fmt.Fprintf(w, "Pairing API:\t%v\n", claims.Pairing)
	fmt.Fprintf(w, "Channels:\t%v\n", claims.Channels)
	fmt.Fprintf(w, "Flow:\t%v\n", claims.Flow)
	fmt.Fprintf(w, "ExpiresAt:\t%v\n", claims.ExpiresAt)
	fmt.Fprintf(w, "IssuedAt:\t%v\n", claims.IssuedAt)
	fmt.Fprintf(w, "NotBefore:\t%v\n", claims.NotBefore)
	fmt.Fprintf(w, "Issuer:\t%v\n", claims.Issuer)
	w.Flush()
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
