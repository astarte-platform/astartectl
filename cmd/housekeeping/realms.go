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

package housekeeping

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/astarte-platform/astarte-go/misc"
	"github.com/astarte-platform/astartectl/config"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// realmsCmd represents the realms command
var realmsCmd = &cobra.Command{
	Use:     "realms",
	Short:   "Manage realms",
	Long:    `List, show or create realms in your Astarte instance.`,
	Aliases: []string{"realm"},
}

var realmsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List realms",
	Long:    "List realms present in your Astarte instance.",
	Example: `  astartectl housekeeping realms list`,
	RunE:    realmsListF,
	Aliases: []string{"ls"},
}

var realmsShowCmd = &cobra.Command{
	Use:     "show <realm_name>",
	Short:   "Show realm",
	Long:    "Show a realm in your Astarte instance.",
	Example: `  astartectl housekeeping realms show myrealm`,
	Args:    cobra.ExactArgs(1),
	RunE:    realmsShowF,
}

var realmsCreateCmd = &cobra.Command{
	Use:     "create <realm_name>",
	Short:   "Create realm",
	Long:    "Create a realm in your Astarte instance.",
	Example: `  astartectl housekeeping realms create myrealm --realm-public-key /path/to/public_key`,
	Args:    cobra.ExactArgs(1),
	RunE:    realmsCreateF,
}

func init() {
	HousekeepingCmd.AddCommand(realmsCmd)

	realmsCreateCmd.Flags().String("realm-private-key", "", "Path to PEM encoded public key used as realm key")
	if err := realmsCreateCmd.MarkFlagFilename("realm-private-key"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	realmsCreateCmd.Flags().String("realm-public-key", "", "Path to PEM encoded public key used as realm key")
	if err := realmsCreateCmd.MarkFlagFilename("realm-public-key"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	realmsCreateCmd.Flags().IntP("replication-factor", "r", 0, `Replication factor for the realm, used with SimpleStrategy replication.`)
	realmsCreateCmd.Flags().StringSliceP("datacenter-replication", "d", nil,
		`Replication factor for a datacenter, used with NetworkTopologyStrategy replication.

The format is <datacenter-name>:<replication-factor>,<other-datacenter-name>:<other-replication-factor>.
You can also specify the flag multiple times instead of separating it with a comma.`)

	realmsCmd.AddCommand(
		realmsListCmd,
		realmsShowCmd,
		realmsCreateCmd,
	)
}

func realmsListF(command *cobra.Command, args []string) error {
	realms, err := astarteAPIClient.Housekeeping.ListRealms()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(realms)
	return nil
}

func realmsShowF(command *cobra.Command, args []string) error {
	realm := args[0]

	realmDetails, err := astarteAPIClient.Housekeeping.GetRealm(realm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("%+v\n", realmDetails)
	return nil
}

func realmsCreateF(command *cobra.Command, args []string) error {
	realm := args[0]
	publicKey, err := command.Flags().GetString("realm-public-key")
	if err != nil {
		return err
	}
	privateKey, err := command.Flags().GetString("realm-private-key")
	if err != nil {
		return err
	}
	if privateKey != "" && publicKey != "" {
		return errors.New("when passing --realm-private-key, --realm-public-key should not be specified")
	}

	replicationFactor, err := command.Flags().GetInt("replication-factor")
	if err != nil {
		return err
	}

	datacenterReplications, err := command.Flags().GetStringSlice("datacenter-replication")
	if err != nil {
		return err
	}

	if replicationFactor > 0 && len(datacenterReplications) > 0 {
		return errors.New("replication-factor and datacenter-replication are mutually exclusive, you only have to specify one")
	}

	createContext := true
	clusterConfigurationName, err := getClusterNameFromURLs()
	if err != nil {
		createContext = false
	}
	if privateKey == "" && publicKey != "" {
		createContext = false
	}

	datacenterReplicationFactors := make(map[string]int)
	if len(datacenterReplications) > 0 {
		for _, datacenterString := range datacenterReplications {
			tokens := strings.Split(datacenterString, ":")
			if len(tokens) != 2 {
				errString := "Invalid datacenter replication: " + datacenterString + "."
				errString += "\nFormat must be <datacenter-name>:<replication-factor>"
				return errors.New(errString)
			}
			datacenter := tokens[0]
			datacenterReplicationFactor, err := strconv.Atoi(tokens[1])
			if err != nil {
				return errors.New("Invalid replication factor " + tokens[1])
			}
			datacenterReplicationFactors[datacenter] = datacenterReplicationFactor
		}
	}

	urlString := viper.GetString("url")
	if viper.GetString("individual-urls.housekeeping") != "" {
		urlString = viper.GetString("individual-urls.housekeeping")
	}
	astarteURL, err := url.Parse(urlString)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	contextName := ""
	if createContext {
		contextName = fmt.Sprintf("%s-realm-%s", astarteURL.Hostname(), realm)
	}

	fmt.Println("Will create Astarte Realm with following parameters:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "Astarte Cluster API Host:\t%s\n", astarteURL.Hostname())
	fmt.Fprintf(w, "Realm name:\t%s\n", realm)
	if len(datacenterReplicationFactors) > 0 {
		fmt.Fprint(w, "Datacenter replications:")
		for k, v := range datacenterReplicationFactors {
			fmt.Fprintf(w, "\t%s: %d\n", k, v)
		}
	} else {
		printedReplicationFactor := replicationFactor
		if replicationFactor == 0 {
			printedReplicationFactor = 1
		}
		fmt.Fprintf(w, "Replication factor:\t%d\n", printedReplicationFactor)
	}
	if createContext {
		fmt.Fprintf(w, "Astarte Context:\t%s\n", contextName)
	}
	w.Flush()
	fmt.Println()

	if !createContext {
		fmt.Println("Will not create an Astarte context - to do so, you need to have a matching Astarte Cluster configuration and supply a private key for the Realm.")
		fmt.Println()
	}

	if privateKey == "" && publicKey == "" {
		if createContext {
			fmt.Println("A new private key will be generated for this realm and saved in your configuration.")
			fmt.Printf("You will be able to access it with \"astartectl config contexts get-realm-key %s\".\n", contextName)
		} else {
			fmt.Println("A new private key will be generated for this realm and printed after successful Realm creation.")
		}
		fmt.Println()
	}

	if ok, err := utils.AskForConfirmation("Do you want to continue?"); !ok || err != nil {
		os.Exit(0)
	}

	var publicKeyContent []byte
	var privateKeyContent []byte
	switch {
	case publicKey != "":
		publicKeyContent, err = ioutil.ReadFile(publicKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case privateKey != "":
		privateKeyContent, err = ioutil.ReadFile(privateKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		key, err := misc.ParsePrivateKeyFromPEM(privateKeyContent)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if publicKeyContent, err = getPublicKeyPEMBytesFromPrivateKey(key); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		reader := rand.Reader
		key, err := ecdsa.GenerateKey(elliptic.P256(), reader)
		if err != nil {
			return err
		}

		if privateKeyContent, err = getPrivateKeyPEMBytes(key); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if publicKeyContent, err = getPublicKeyPEMBytesFromPrivateKey(key); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if replicationFactor > 0 {
		err = astarteAPIClient.Housekeeping.CreateRealmWithReplicationFactor(realm, string(publicKeyContent), replicationFactor)
	} else if len(datacenterReplicationFactors) > 0 {
		err = astarteAPIClient.Housekeeping.CreateRealmWithDatacenterReplication(realm, string(publicKeyContent),
			datacenterReplicationFactors)
	} else {
		err = astarteAPIClient.Housekeeping.CreateRealm(realm, string(publicKeyContent))
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Realm %s created successfully!\n", realm)

	// If we're not creating a context, return (and print the key for reference)
	if !createContext {
		if privateKey == "" && publicKey == "" {
			fmt.Println()
			fmt.Println("This is your Realm's private key. Make sure you store it somewhere safe.")
			fmt.Println()
			fmt.Println(string(privateKeyContent))
		}

		return nil
	}

	realmContext := config.RealmConfiguration{
		Name: realm,
	}

	if privateKeyContent != nil {
		realmContext.Key = base64.StdEncoding.EncodeToString(privateKeyContent)
	}

	configContext := config.ContextFile{
		Cluster: clusterConfigurationName,
		Realm:   realmContext,
	}

	configDir := config.GetConfigDir()
	if err := config.SaveContextConfiguration(configDir, contextName, configContext, true); err != nil {
		fmt.Fprintf(os.Stderr, "Could not save cluster configuration: %s\n", err)
		if privateKey == "" {
			// Dump the private key
			fmt.Fprintln(os.Stderr, "Dumping private key for reference")
			fmt.Println(string(privateKeyContent))
		}
	} else {
		fmt.Printf("Context %s created successfully\n", contextName)

		// Now set the current context to the new one
		baseConfig, err := config.LoadBaseConfiguration(configDir)
		if err != nil {
			// Shoot out a warning, but don't fail
			baseConfig = config.BaseConfigFile{}
			fmt.Fprintf(os.Stderr, "warn: Could not load configuration file: %s. Will proceed creating a new one\n", err.Error())
		}

		baseConfig.CurrentContext = contextName
		if err := config.SaveBaseConfiguration(configDir, baseConfig); err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, "warn: Context not switched")
		} else {
			fmt.Printf("Context switched to %s\n", contextName)
		}
	}

	return nil
}

func getPrivateKeyPEMBytes(key *ecdsa.PrivateKey) ([]byte, error) {
	marshaled, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}

	var pemkey = &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: marshaled,
	}
	var outbuf bytes.Buffer

	if err := pem.Encode(&outbuf, pemkey); err != nil {
		return nil, err
	}

	return outbuf.Bytes(), nil
}

func getPublicKeyPEMBytesFromPrivateKey(key interface{}) ([]byte, error) {
	var pkixBytes []byte
	var err error

	switch k := key.(type) {
	case *rsa.PrivateKey:
		pkixBytes, err = x509.MarshalPKIXPublicKey(k.Public())

	case *ecdsa.PrivateKey:
		pkixBytes, err = x509.MarshalPKIXPublicKey(k.Public())

	default:
		return nil, errors.New("Unsupported private key type")
	}

	if err != nil {
		return nil, err
	}
	var pemkey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pkixBytes,
	}
	var outbuf bytes.Buffer

	if err := pem.Encode(&outbuf, pemkey); err != nil {
		return nil, err
	}

	return outbuf.Bytes(), nil
}

func getClusterNameFromURLs() (string, error) {
	configDir := config.GetConfigDir()
	clusters, err := config.ListClusterConfigurations(configDir)
	if err == nil {
		for _, c := range clusters {
			cluster, err := config.LoadClusterConfiguration(configDir, c)
			if err != nil {
				continue
			}
			if (cluster.IndividualURLs.Housekeeping == viper.GetString("individual-urls.housekeeping") && viper.GetString("individual-urls.housekeeping") != "") ||
				(cluster.URL == viper.GetString("url") && viper.GetString("url") != "") {
				// yay
				return c, nil
			}
		}
	} else {
		return "", err
	}
	// Skip context creation
	return "", errors.New("Not found")
}
