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

package appengine

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/astarte-platform/astartectl/client"
	"github.com/astarte-platform/astartectl/common"
	"github.com/astarte-platform/astartectl/utils"
	"github.com/jedib0t/go-pretty/table"

	"github.com/araddon/dateparse"

	"github.com/spf13/cobra"
)

// DevicesCmd represents the devices command
var devicesCmd = &cobra.Command{
	Use:     "devices",
	Short:   "Interact with Devices",
	Long:    `Perform actions on Astarte Devices.`,
	Aliases: []string{"device"},
}

var devicesListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List devices",
	Long:    `List all devices in the realm.`,
	Example: `  astartectl appengine devices list`,
	RunE:    devicesListF,
	Aliases: []string{"ls"},
}

var devicesShowCmd = &cobra.Command{
	Use:   "show <device_id_or_alias>",
	Short: "Show a Device",
	Long: `Show a Device in the realm, printing all its known information.
<device_id_or_alias> can be either a valid Astarte Device ID, or a Device Alias. In most cases,
this is automatically determined - however, you can tweak this behavior by using --force-device-id or
--force-id-type={device-id,alias}.`,
	Example: `  astartectl appengine devices show 2TBn-jNESuuHamE2Zo1anA`,
	Args:    cobra.ExactArgs(1),
	RunE:    devicesShowF,
}

var devicesDataSnapshotCmd = &cobra.Command{
	Use:   "data-snapshot <device_id_or_alias> [<interface_name>]",
	Short: "Outputs a Data Snapshot of a given Device",
	Long: `data-snapshot retrieves the last received sample
(if it is a Datastream), or the currently known value (if it is a property).
If <interface_name> is specified, the snapshot is returned only for that specific interface,
otherwise it's returned for all Interfaces in the Device's introspection.
<device_id_or_alias> can be either a valid Astarte Device ID, or a Device Alias. In most cases,
this is automatically determined - however, you can tweak this behavior by using --force-device-id or
--force-id-type={device-id,alias}.`,
	Example: `  astartectl appengine devices data-snapshot 2TBn-jNESuuHamE2Zo1anA`,
	Args:    cobra.RangeArgs(1, 2),
	RunE:    devicesDataSnapshotF,
}

var devicesGetSamplesCmd = &cobra.Command{
	Use:   "get-samples <device_id_or_alias> <interface_name> <path>",
	Short: "Retrieves samples for a given Datastream path",
	Long: `Retrieves and prints samples for a given device. By default, the first 10000 samples
are returned. You can tweak this behavior by using --count.
By default, samples are returned in descending order (starting from most recent). You can use --ascending to
change this behavior.

<device_id_or_alias> can be either a valid Astarte Device ID, or a Device Alias. In most cases,
this is automatically determined - however, you can tweak this behavior by using --force-device-id or
--force-id-type={device-id,alias}.`,
	Example: `  astartectl appengine devices get-samples 2TBn-jNESuuHamE2Zo1anA com.my.interface /my/path`,
	Args:    cobra.ExactArgs(3),
	RunE:    devicesGetSamplesF,
}

var supportedOutputTypes = []string{"default", "csv", "json"}

func isASupportedOutputType(outputType string) bool {
	for _, s := range supportedOutputTypes {
		if s == outputType {
			return true
		}
	}
	return false
}

func init() {
	AppEngineCmd.AddCommand(devicesCmd)

	devicesGetSamplesCmd.Flags().IntP("count", "c", 10000, "Number of samples to be retrieved. Defaults to 10000. Setting this to 0 retrieves all samples.")
	devicesGetSamplesCmd.Flags().Bool("ascending", false, "When set, returns samples in ascending order rather than descending.")
	devicesGetSamplesCmd.Flags().String("since", "", "When set, returns only samples newer than the provided date.")
	devicesGetSamplesCmd.Flags().String("to", "", "When set, returns only samples older than the provided date.")
	devicesGetSamplesCmd.Flags().StringP("output", "o", "default", "The type of output (default,csv,json)")
	devicesGetSamplesCmd.Flags().String("force-id-type", "", "When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")
	devicesGetSamplesCmd.Flags().Bool("aggregate", false, "When set, if Realm Management checks are disabled, it forces resolution of the interface as an aggregate datastream.")
	devicesGetSamplesCmd.Flags().Bool("skip-realm-management-checks", false, "When set, it skips any consistency checks on Realm Management before performing the Query. This might lead to unexpected errors.")

	devicesDataSnapshotCmd.Flags().StringP("output", "o", "default", "The type of output (default,csv,json)")
	devicesDataSnapshotCmd.Flags().String("force-id-type", "", "When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")
	devicesDataSnapshotCmd.Flags().Bool("skip-realm-management-checks", false, "When set, it skips any consistency checks on Realm Management before performing the Query. This might lead to unexpected errors. This has effect only if data-snapshot is invoked for a specific interface.")
	devicesDataSnapshotCmd.Flags().String("interface-type", "", "When set, if Realm Management checks are disabled, it forces resolution of the interface as the specified type. Valid options are: properties, individual-datastream, aggregate-datastream, individual-parametric-datastream, aggregate-parametric-datastream.")

	devicesShowCmd.Flags().String("force-id-type", "", "When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")

	devicesCmd.AddCommand(
		devicesListCmd,
		devicesShowCmd,
		devicesDataSnapshotCmd,
		devicesGetSamplesCmd,
	)
}

func devicesListF(command *cobra.Command, args []string) error {
	devices, err := astarteAPIClient.AppEngine.ListDevices(realm, appEngineJwt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(devices)
	return nil
}

func prettyPrintDeviceDetails(deviceDetails client.DeviceDetails) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "Device ID:\t%v\n", deviceDetails.DeviceID)
	fmt.Fprintf(w, "Connected:\t%v\n", deviceDetails.Connected)
	fmt.Fprintf(w, "Last Connection:\t%v\n", deviceDetails.LastConnection)
	fmt.Fprintf(w, "Last Disconnection:\t%v\n", deviceDetails.LastDisconnection)
	if len(deviceDetails.Introspection) > 0 {
		fmt.Fprintf(w, "Introspection:")
		// Iterate the introspection
		for i, v := range deviceDetails.Introspection {
			fmt.Fprintf(w, "\t%v v%v.%v\n", i, v.Major, v.Minor)
		}
	}
	if len(deviceDetails.Aliases) > 0 {
		fmt.Fprintf(w, "Aliases:")
		// Iterate the aliases
		for i, v := range deviceDetails.Aliases {
			fmt.Fprintf(w, "\t%v: %v\n", i, v)
		}
	}
	fmt.Fprintf(w, "Received Messages:\t%v\n", deviceDetails.TotalReceivedMessages)
	fmt.Fprintf(w, "Data Received:\t%v\n", bytefmt.ByteSize(deviceDetails.TotalReceivedBytes))
	fmt.Fprintf(w, "Last Seen IP:\t%v\n", deviceDetails.LastSeenIP)
	fmt.Fprintf(w, "Last Credentials Request IP:\t%v\n", deviceDetails.LastCredentialsRequestIP)
	fmt.Fprintf(w, "First Registration:\t%v\n", deviceDetails.FirstRegistration)
	fmt.Fprintf(w, "First Credentials Request:\t%v\n", deviceDetails.FirstCredentialsRequest)
	w.Flush()
}

func devicesShowF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	forceIDType, err := command.Flags().GetString("force-id-type")
	if err != nil {
		return err
	}
	deviceIdentifierType, err := deviceIdentifierTypeFromFlags(deviceID, forceIDType)
	if err != nil {
		return err
	}

	deviceDetails, err := astarteAPIClient.AppEngine.GetDevice(realm, deviceID, deviceIdentifierType, appEngineJwt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	prettyPrintDeviceDetails(deviceDetails)
	return nil
}

func tableWriterForOutputType(outputType string) table.Writer {
	t := table.NewWriter()
	switch outputType {
	case "default":
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleLight)
	case "csv":
		t.SetOutputMirror(os.Stdout)
	case "json":
	default:
		return nil
	}
	return t
}

func devicesDataSnapshotF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	forceIDType, err := command.Flags().GetString("force-id-type")
	if err != nil {
		return err
	}
	deviceIdentifierType, err := deviceIdentifierTypeFromFlags(deviceID, forceIDType)
	if err != nil {
		return err
	}
	var snapshotInterface string
	if len(args) == 2 {
		snapshotInterface = args[1]
	}
	skipRealmManagementChecks, err := command.Flags().GetBool("skip-realm-management-checks")
	if err != nil {
		return err
	}
	if skipRealmManagementChecks && snapshotInterface == "" {
		return fmt.Errorf("When using --skip-realm-management-checks, an interface should always be specified")
	}
	interfaceTypeString, err := command.Flags().GetString("interface-type")
	if err != nil {
		return err
	}
	if skipRealmManagementChecks && interfaceTypeString == "" {
		return fmt.Errorf("When using --skip-realm-management-checks, --interface-type should always be specified")
	}

	outputType, err := command.Flags().GetString("output")
	if err != nil {
		return err
	}
	if !isASupportedOutputType(outputType) {
		return fmt.Errorf("%v is not a supported output type. Supported output types are %v", outputType, supportedOutputTypes)
	}

	// Go with the table header
	t := tableWriterForOutputType(outputType)
	if snapshotInterface == "" {
		t.AppendHeader(table.Row{"Interface", "Path", "Value", "Ownership", "Timestamp (Datastream only)"})
	}
	jsonOutput := make(map[string]interface{})

	if snapshotInterface != "" {
		var interfaceType common.AstarteInterfaceType
		var interfaceAggregation common.AstarteInterfaceAggregation
		isParametricInterface := false

		if skipRealmManagementChecks {
			switch interfaceTypeString {
			case "properties":
				interfaceType = common.PropertiesType
				interfaceAggregation = common.IndividualAggregation
				isParametricInterface = false
			case "individual-datastream":
				interfaceType = common.DatastreamType
				interfaceAggregation = common.IndividualAggregation
				isParametricInterface = false
			case "aggregate-datastream":
				interfaceType = common.DatastreamType
				interfaceAggregation = common.ObjectAggregation
				isParametricInterface = false
			case "individual-parametric-datastream":
				interfaceType = common.DatastreamType
				interfaceAggregation = common.IndividualAggregation
				isParametricInterface = true
			case "aggregate-parametric-datastream":
				interfaceType = common.DatastreamType
				interfaceAggregation = common.ObjectAggregation
				isParametricInterface = true
			default:
				return fmt.Errorf("%s is not a valid Interface Type. Valid interface types are: properties, individual-datastream, aggregate-datastream, individual-parametric-datastream, aggregate-parametric-datastream", interfaceTypeString)
			}
		} else {
			// Get the device introspection
			deviceDetails, err := astarteAPIClient.AppEngine.GetDevice(realm, deviceID, deviceIdentifierType, appEngineJwt)
			if err != nil {
				return err
			}

			var interfaceDescription common.AstarteInterface
			interfaceFound := false
			for astarteInterface, interfaceIntrospection := range deviceDetails.Introspection {
				if astarteInterface != snapshotInterface {
					continue
				}

				// Query Realm Management to get details on the interface
				interfaceDescription, err = astarteAPIClient.RealmManagement.GetInterface(realm, astarteInterface,
					interfaceIntrospection.Major, realmManagementJwt)
				if err != nil {
					fmt.Printf(err.Error())
					os.Exit(1)
				}
				interfaceFound = true
				break
			}

			if !interfaceFound {
				fmt.Printf("Interface %s does not exist in Device %s\n", snapshotInterface, deviceID)
				os.Exit(1)
			}

			interfaceType = interfaceDescription.Type
			interfaceAggregation = interfaceDescription.Aggregation
			isParametricInterface = interfaceDescription.IsParametric()
		}

		switch interfaceType {
		case common.DatastreamType:
			t.AppendHeader(table.Row{"Interface", "Path", "Value", "Timestamp"})
			if interfaceAggregation == common.ObjectAggregation {
				if isParametricInterface {
					val, err := astarteAPIClient.AppEngine.GetAggregateParametricDatastreamSnapshot(realm, deviceID, deviceIdentifierType, snapshotInterface, appEngineJwt)
					if err != nil {
						return err
					}
					for path, aggregate := range val {
						if outputType == "json" {
							jsonOutput[snapshotInterface] = val
						} else {
							for _, k := range aggregate.Values.Keys() {
								v, _ := aggregate.Values.Get(k)
								t.AppendRow([]interface{}{snapshotInterface, fmt.Sprintf("%s/%s", path, k), v, timestampForOutput(aggregate.Timestamp, outputType)})
							}
						}
					}
				} else {
					val, err := astarteAPIClient.AppEngine.GetAggregateDatastreamSnapshot(realm, deviceID, deviceIdentifierType, snapshotInterface, appEngineJwt)
					if err != nil {
						return err
					}
					if outputType == "json" {
						jsonOutput[snapshotInterface] = val
					} else {
						for _, k := range val.Values.Keys() {
							v, _ := val.Values.Get(k)
							t.AppendRow([]interface{}{snapshotInterface, fmt.Sprintf("/%s", k), v, timestampForOutput(val.Timestamp, outputType)})
						}
					}
				}
			} else {
				val, err := astarteAPIClient.AppEngine.GetDatastreamSnapshot(realm, deviceID, deviceIdentifierType, snapshotInterface, appEngineJwt)
				if err != nil {
					return err
				}
				jsonRepresentation := make(map[string]interface{})
				for k, v := range val {
					jsonRepresentation[k] = v
					t.AppendRow([]interface{}{snapshotInterface, k, v.Value, timestampForOutput(v.Timestamp, outputType)})
				}
				jsonOutput[snapshotInterface] = jsonRepresentation
			}
		case common.PropertiesType:
			t.AppendHeader(table.Row{"Interface", "Path", "Value"})
			val, err := astarteAPIClient.AppEngine.GetProperties(realm, deviceID, deviceIdentifierType, snapshotInterface, appEngineJwt)
			if err != nil {
				return err
			}
			jsonRepresentation := make(map[string]interface{})
			for k, v := range val {
				jsonRepresentation[k] = v
				t.AppendRow([]interface{}{snapshotInterface, k, v})
			}
			jsonOutput[snapshotInterface] = jsonRepresentation
		}
	} else {
		// Get the device introspection
		deviceDetails, err := astarteAPIClient.AppEngine.GetDevice(realm, deviceID, deviceIdentifierType, appEngineJwt)
		if err != nil {
			return err
		}

		for astarteInterface, interfaceIntrospection := range deviceDetails.Introspection {
			// Query Realm Management to get details on the interface
			interfaceDescription, err := astarteAPIClient.RealmManagement.GetInterface(realm, astarteInterface,
				interfaceIntrospection.Major, realmManagementJwt)
			if err != nil {
				return err
			}

			switch interfaceDescription.Type {
			case common.DatastreamType:
				if interfaceDescription.Aggregation == common.ObjectAggregation {
					if interfaceDescription.IsParametric() {
						val, err := astarteAPIClient.AppEngine.GetAggregateParametricDatastreamSnapshot(realm, deviceID, deviceIdentifierType, astarteInterface, appEngineJwt)
						if err != nil {
							return err
						}
						for path, aggregate := range val {
							if outputType == "json" {
								jsonOutput[astarteInterface] = val
							} else {
								for _, k := range aggregate.Values.Keys() {
									v, _ := aggregate.Values.Get(k)
									t.AppendRow([]interface{}{astarteInterface, fmt.Sprintf("%s/%s", path, k), v, interfaceDescription.Ownership.String(),
										timestampForOutput(aggregate.Timestamp, outputType)})
								}
							}
						}
					} else {
						val, err := astarteAPIClient.AppEngine.GetAggregateDatastreamSnapshot(realm, deviceID, deviceIdentifierType, astarteInterface, appEngineJwt)
						if err != nil {
							return err
						}
						if outputType == "json" {
							jsonOutput[astarteInterface] = val
						} else {
							for _, k := range val.Values.Keys() {
								v, _ := val.Values.Get(k)
								t.AppendRow([]interface{}{astarteInterface, fmt.Sprintf("/%s", k), v, interfaceDescription.Ownership.String(),
									timestampForOutput(val.Timestamp, outputType)})
							}
						}
					}
				} else {
					val, err := astarteAPIClient.AppEngine.GetDatastreamSnapshot(realm, deviceID, deviceIdentifierType, astarteInterface, appEngineJwt)
					if err != nil {
						return err
					}
					jsonRepresentation := make(map[string]interface{})
					for k, v := range val {
						jsonRepresentation[k] = v
						t.AppendRow([]interface{}{astarteInterface, k, v.Value, interfaceDescription.Ownership.String(),
							timestampForOutput(v.Timestamp, outputType)})
					}
					jsonOutput[astarteInterface] = jsonRepresentation
				}
			case common.PropertiesType:
				val, err := astarteAPIClient.AppEngine.GetProperties(realm, deviceID, deviceIdentifierType, astarteInterface, appEngineJwt)
				if err != nil {
					return err
				}
				jsonRepresentation := make(map[string]interface{})
				for k, v := range val {
					jsonRepresentation[k] = v
					t.AppendRow([]interface{}{astarteInterface, k, v, interfaceDescription.Ownership.String(), ""})
				}
				jsonOutput[astarteInterface] = jsonRepresentation
			}
		}
	}

	// Done
	renderOutput(t, jsonOutput, outputType)

	return nil
}

func devicesGetSamplesF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	interfaceName := args[1]
	interfacePath := args[2]
	forceIDType, err := command.Flags().GetString("force-id-type")
	if err != nil {
		return err
	}
	deviceIdentifierType, err := deviceIdentifierTypeFromFlags(deviceID, forceIDType)
	if err != nil {
		return err
	}
	limit, err := command.Flags().GetInt("count")
	if err != nil {
		return err
	}
	ascending, err := command.Flags().GetBool("ascending")
	if err != nil {
		return err
	}
	resultSetOrder := client.DescendingOrder
	if ascending {
		resultSetOrder = client.AscendingOrder
	}
	skipRealmManagementChecks, err := command.Flags().GetBool("skip-realm-management-checks")
	if err != nil {
		return err
	}
	forceAggregate, err := command.Flags().GetBool("aggregate")
	if err != nil {
		return err
	}
	since, err := command.Flags().GetString("since")
	if err != nil {
		return err
	}
	sinceTime := time.Unix(0, 0)
	if since != "" {
		sinceTime, err = dateparse.ParseLocal(since)
		if err != nil {
			return err
		}
	}
	to, err := command.Flags().GetString("to")
	if err != nil {
		return err
	}
	toTime := time.Now()
	if to != "" {
		toTime, err = dateparse.ParseLocal(to)
		if err != nil {
			return err
		}
	}
	outputType, err := command.Flags().GetString("output")
	if err != nil {
		return err
	}
	if !isASupportedOutputType(outputType) {
		return fmt.Errorf("%v is not a supported output type. Supported output types are %v", outputType, supportedOutputTypes)
	}

	var isAggregate bool
	if !skipRealmManagementChecks {
		// Get the device introspection
		interfaceFound := false
		deviceDetails, err := astarteAPIClient.AppEngine.GetDevice(realm, deviceID, deviceIdentifierType, appEngineJwt)
		if err != nil {
			return err
		}
		for astarteInterface, interfaceIntrospection := range deviceDetails.Introspection {
			if astarteInterface != interfaceName {
				continue
			}

			// Query Realm Management to get details on the interface
			interfaceDescription, err := astarteAPIClient.RealmManagement.GetInterface(realm, astarteInterface,
				interfaceIntrospection.Major, realmManagementJwt)
			if err != nil {
				return err
			}

			if interfaceDescription.Type != common.DatastreamType {
				fmt.Printf("%s is not a Datastream interface. get-samples works only on Datastream interfaces\n", interfaceName)
				os.Exit(1)
			}

			// TODO: Check paths when we have a better parsing for interfaces
			interfaceFound = true
			err = utils.ValidateInterfacePath(interfaceDescription, interfacePath)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			interfacePathTokens := strings.Split(interfacePath, "/")
			validationEndpointTokens := strings.Split(interfaceDescription.Mappings[0].Endpoint, "/")
			isAggregate = interfaceDescription.Aggregation == common.ObjectAggregation
			// Special case
			if len(interfacePathTokens) == len(validationEndpointTokens) && interfacePath != "/" {
				isAggregate = false
			}
		}

		if !interfaceFound {
			fmt.Printf("Device %s has no interface named %s\n", deviceID, interfaceName)
			os.Exit(1)
		}
	} else {
		isAggregate = forceAggregate
	}

	if interfacePath == "/" {
		interfacePath = ""
	}

	// We are good to go.
	t := tableWriterForOutputType(outputType)
	if !isAggregate {
		// Go with the table header
		t.AppendHeader(table.Row{"Timestamp", "Value"})
		printedValues := 0
		jsonOutput := []client.DatastreamValue{}
		datastreamPaginator := astarteAPIClient.AppEngine.GetDatastreamsTimeWindowPaginator(realm, deviceID,
			deviceIdentifierType, interfaceName, interfacePath, sinceTime, toTime, resultSetOrder, appEngineJwt)
		for ok := true; ok; ok = datastreamPaginator.HasNextPage() {
			page, err := datastreamPaginator.GetNextPage()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if outputType == "json" {
				jsonOutput = append(jsonOutput, page...)
			} else {
				for _, v := range page {
					t.AppendRow([]interface{}{timestampForOutput(v.Timestamp, outputType), v.Value})
					printedValues++
					if printedValues >= limit && limit > 0 {
						renderOutput(t, jsonOutput, outputType)
						return nil
					}
				}
			}
		}
		renderOutput(t, jsonOutput, outputType)
	} else {
		headerRow := table.Row{"Timestamp"}
		headerPrinted := false

		jsonOutput := []client.DatastreamAggregateValue{}
		printedValues := 0
		datastreamPaginator := astarteAPIClient.AppEngine.GetDatastreamsTimeWindowPaginator(realm, deviceID, deviceIdentifierType, interfaceName, interfacePath,
			sinceTime, toTime, resultSetOrder, appEngineJwt)
		for ok := true; ok; ok = datastreamPaginator.HasNextPage() {
			page, err := datastreamPaginator.GetNextAggregatePage()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if outputType == "json" {
				jsonOutput = append(jsonOutput, page...)
			} else {
				for _, v := range page {
					// Iterate the aggregate
					line := []interface{}{}
					line = append(line, timestampForOutput(v.Timestamp, outputType))
					for _, path := range v.Values.Keys() {
						value, _ := v.Values.Get(path)
						if !headerPrinted {
							headerRow = append(headerRow, fmt.Sprintf("%s/%s", interfacePath, path))
						}
						line = append(line, value)
					}
					if !headerPrinted {
						t.AppendHeader(headerRow)
						headerPrinted = true
					}
					t.AppendRow(line)
					printedValues++
					if printedValues >= limit && limit > 0 {
						renderOutput(t, jsonOutput, outputType)
						return nil
					}
				}
			}
		}
		renderOutput(t, jsonOutput, outputType)
	}

	return nil
}

func timestampForOutput(timestamp time.Time, outputType string) string {
	switch outputType {
	case "default":
		return timestamp.String()
	case "csv":
		return timestamp.Format(time.RFC3339Nano)
	case "json":
	}

	return ""
}

func renderOutput(t table.Writer, jsonOutput interface{}, outputType string) {
	switch outputType {
	case "default":
		t.Render()
	case "csv":
		t.RenderCSV()
	case "json":
		respJSON, _ := json.MarshalIndent(jsonOutput, "", "  ")
		fmt.Println(string(respJSON))
	}
}
