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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/astarte-platform/astarte-go/client"
	"github.com/astarte-platform/astarte-go/interfaces"
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
	Use:   "get-samples <device_id_or_alias> <interface_name> [path]",
	Short: "Retrieves samples for a given Datastream path",
	Long: `Retrieves and prints samples for a given device. By default, the first 10000 samples
are returned. You can tweak this behavior by using --count.
By default, samples are returned in descending order (starting from most recent). You can use --ascending to
change this behavior.

When dealing with an aggregate, non parametric interface, path can be omitted. It is compulsory for
all other cases.

<device_id_or_alias> can be either a valid Astarte Device ID, or a Device Alias. In most cases,
this is automatically determined - however, you can tweak this behavior by using --force-device-id or
--force-id-type={device-id,alias}.`,
	Example: `  astartectl appengine devices get-samples 2TBn-jNESuuHamE2Zo1anA com.my.interface /my/path`,
	Args:    cobra.RangeArgs(2, 3),
	RunE:    devicesGetSamplesF,
}

var devicesSendDataCmd = &cobra.Command{
	Use:   "send-data <device_id_or_alias> <interface_name> <path> <data>",
	Short: "Sends data to a given interface path",
	Long: `Sends data to a given interface path. This works both for datastream with individual and properties.

When dealing with an aggregate, non parametric interface, path must still be provided, adhering to the
interface structure. In that case, <data> should be a JSON string which contains a key/value dictionary,
with key bearing the name (without trailing slashes) of the tip of the endpoint, and value being the
value of that specific endpoint, correctly typed.

<device_id_or_alias> can be either a valid Astarte Device ID, or a Device Alias. In most cases,
this is automatically determined - however, you can tweak this behavior by using --force-device-id or
--force-id-type={device-id,alias}.`,
	Example: `  astartectl appengine devices send-data 2TBn-jNESuuHamE2Zo1anA com.my.interface /my/path "value"`,
	Args:    cobra.ExactArgs(4),
	RunE:    devicesSendDataF,
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

	devicesListCmd.Flags().BoolP("details", "d", false, "When set, return the device list with all the DeviceDetails. Otherwise, just return the Device ID")

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

	devicesSendDataCmd.Flags().String("force-id-type", "", "When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")
	devicesSendDataCmd.Flags().Bool("skip-realm-management-checks", false, "When set, it skips any consistency checks on Realm Management before performing the Query. This might lead to unexpected errors. This has effect only if data-snapshot is invoked for a specific interface.")
	devicesSendDataCmd.Flags().String("interface-type", "", "When set, if Realm Management checks are disabled, it forces resolution of the interface as the specified type. Valid options are: properties, individual-datastream, aggregate-datastream, individual-parametric-datastream, aggregate-parametric-datastream.")
	devicesSendDataCmd.Flags().String("payload-type", "", "When set, forces the conversion of the given payload into the given type. Valid values are any value in Astarte interfaces.")

	devicesShowCmd.Flags().String("force-id-type", "", "When set, rather than autodetecting, it forces the device ID to be evaluated as a (device-id,alias).")

	devicesCmd.AddCommand(
		devicesListCmd,
		devicesShowCmd,
		devicesDataSnapshotCmd,
		devicesGetSamplesCmd,
		devicesSendDataCmd,
	)
}

func devicesListF(command *cobra.Command, args []string) error {
	details, err := command.Flags().GetBool("details")
	if err != nil {
		return err
	}

	if details {
		paginator, err := astarteAPIClient.AppEngine.GetDeviceListPaginator(realm, 50)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for hasNext := paginator.HasNextPage(); hasNext; hasNext = paginator.HasNextPage() {
			page, err := paginator.GetNextDeviceDetailsPage()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			for _, deviceDetails := range page {
				prettyPrintDeviceDetails(deviceDetails)
				fmt.Println()
			}
		}

	} else {
		devices, err := astarteAPIClient.AppEngine.ListDevices(realm)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(devices)
	}

	return nil
}

func prettyPrintDeviceDetails(deviceDetails client.DeviceDetails) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	if deviceDetails.CredentialsInhibited {
		fmt.Fprintf(w, "Credentials Inhibited:\t%v\n", deviceDetails.CredentialsInhibited)
	}
	fmt.Fprintf(w, "Device ID:\t%v\n", deviceDetails.DeviceID)
	fmt.Fprintf(w, "Connected:\t%v\n", deviceDetails.Connected)
	fmt.Fprintf(w, "Last Connection:\t%v\n", deviceDetails.LastConnection)
	fmt.Fprintf(w, "Last Disconnection:\t%v\n", deviceDetails.LastDisconnection)
	if len(deviceDetails.Introspection) > 0 {
		fmt.Fprintf(w, "Introspection:")
		// Iterate the introspection
		for i, v := range deviceDetails.Introspection {
			interfaceLine := fmt.Sprintf("\t%v v%v.%v", i, v.Major, v.Minor)
			if v.ExchangedMessages > 0 {
				interfaceLine += fmt.Sprintf(" exchanged messages: %v", v.ExchangedMessages)
			}
			if v.ExchangedBytes > 0 {
				interfaceLine += fmt.Sprintf(" exchanged bytes: %v", bytefmt.ByteSize(v.ExchangedBytes))
			}
			fmt.Fprintln(w, interfaceLine)
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
	if len(deviceDetails.PreviousInterfaces) > 0 {
		fmt.Fprintf(w, "Previous Interfaces:")
		// Iterate the previous introspection
		for _, v := range deviceDetails.PreviousInterfaces {
			interfaceLine := fmt.Sprintf("\t%v v%v.%v", v.Name, v.Major, v.Minor)
			if v.ExchangedMessages > 0 {
				interfaceLine += fmt.Sprintf(" exchanged messages: %v", v.ExchangedMessages)
			}
			if v.ExchangedBytes > 0 {
				interfaceLine += fmt.Sprintf(" exchanged bytes: %v", bytefmt.ByteSize(v.ExchangedBytes))
			}
			fmt.Fprintln(w, interfaceLine)
		}
	}
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

	deviceDetails, err := astarteAPIClient.AppEngine.GetDevice(realm, deviceID, deviceIdentifierType)
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
	skipRealmManagementChecks = skipRealmManagementChecks || astarteAPIClient.RealmManagement == nil
	if skipRealmManagementChecks && snapshotInterface == "" {
		return fmt.Errorf("When not using Realm Management checks, an interface should always be specified")
	}
	interfaceTypeString, err := command.Flags().GetString("interface-type")
	if err != nil {
		return err
	}
	if skipRealmManagementChecks && interfaceTypeString == "" {
		return fmt.Errorf("When not using Realm Management checks, --interface-type should always be specified")
	}

	outputType, err := command.Flags().GetString("output")
	if err != nil {
		return err
	}
	if !isASupportedOutputType(outputType) {
		return fmt.Errorf("%v is not a supported output type. Supported output types are %v", outputType, supportedOutputTypes)
	}

	interfacesToFetch := []interfaces.AstarteInterface{}

	// Go with the table header
	t := tableWriterForOutputType(outputType)

	// Distinguish here whether we're doing a full snapshot or just a single-interface snapshot, and act accordingly
	if snapshotInterface == "" {
		t.AppendHeader(table.Row{"Interface", "Path", "Value", "Ownership", "Timestamp (Datastream only)"})

		deviceDetails, err := astarteAPIClient.AppEngine.GetDevice(realm, deviceID, deviceIdentifierType)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		for astarteInterface, interfaceIntrospection := range deviceDetails.Introspection {
			// Query Realm Management to get details on the interface
			interfaceDescription, err := astarteAPIClient.RealmManagement.GetInterface(realm, astarteInterface,
				interfaceIntrospection.Major)
			if err != nil {
				// If we're requesting a full snapshot, do not fail but just warn the user
				fmt.Fprintf(os.Stderr, "warn: Could not fetch details for interface %s\n", astarteInterface)
				continue
			}
			interfacesToFetch = append(interfacesToFetch, interfaceDescription)
		}
	} else {
		// Get the proto interface
		iface, err := getProtoInterface(deviceID, deviceIdentifierType, snapshotInterface, interfaceTypeString, skipRealmManagementChecks)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// We're dealing with only one interface, so let's be as smart as possible.
		switch iface.Type {
		case interfaces.DatastreamType:
			t.AppendHeader(table.Row{"Interface", "Path", "Value", "Timestamp"})
		case interfaces.PropertiesType:
			t.AppendHeader(table.Row{"Interface", "Path", "Value"})
		}
	}
	jsonOutput := make(map[string]interface{})

	for _, i := range interfacesToFetch {
		switch i.Type {
		case interfaces.DatastreamType:
			if i.Aggregation == interfaces.ObjectAggregation {
				if i.IsParametric() {
					val, err := astarteAPIClient.AppEngine.GetAggregateParametricDatastreamSnapshot(realm, deviceID, deviceIdentifierType, i.Name)
					if err != nil {
						warnOrFail(snapshotInterface, i.Name, err)
					}
					for path, aggregate := range val {
						if outputType == "json" {
							jsonOutput[i.Name] = val
						} else {
							for _, k := range aggregate.Values.Keys() {
								v, _ := aggregate.Values.Get(k)
								if v == nil {
									v = "(null)"
								}
								if snapshotInterface == "" {
									t.AppendRow([]interface{}{i.Name, fmt.Sprintf("%s/%s", path, k), v, i.Ownership,
										timestampForOutput(aggregate.Timestamp, outputType)})
								} else {
									t.AppendRow([]interface{}{i.Name, fmt.Sprintf("%s/%s", path, k), v,
										timestampForOutput(aggregate.Timestamp, outputType)})
								}
							}
						}
					}
				} else {
					val, err := astarteAPIClient.AppEngine.GetAggregateDatastreamSnapshot(realm, deviceID, deviceIdentifierType, i.Name)
					if err != nil {
						warnOrFail(snapshotInterface, i.Name, err)
					}
					if outputType == "json" {
						jsonOutput[i.Name] = val
					} else {
						for _, k := range val.Values.Keys() {
							v, _ := val.Values.Get(k)
							if v == nil {
								v = "(null)"
							}
							if snapshotInterface == "" {
								t.AppendRow([]interface{}{i.Name, fmt.Sprintf("/%s", k), v, i.Ownership,
									timestampForOutput(val.Timestamp, outputType)})
							} else {
								t.AppendRow([]interface{}{i.Name, fmt.Sprintf("/%s", k), v,
									timestampForOutput(val.Timestamp, outputType)})
							}
						}
					}
				}
			} else {
				val, err := astarteAPIClient.AppEngine.GetDatastreamSnapshot(realm, deviceID, deviceIdentifierType, i.Name)
				if err != nil {
					warnOrFail(snapshotInterface, i.Name, err)
				}
				jsonRepresentation := make(map[string]interface{})
				for k, v := range val {
					jsonRepresentation[k] = v
					if v.Value == nil {
						v.Value = "(null)"
					}
					if snapshotInterface == "" {
						t.AppendRow([]interface{}{i.Name, k, v.Value, i.Ownership,
							timestampForOutput(v.Timestamp, outputType)})
					} else {
						t.AppendRow([]interface{}{i.Name, k, v.Value,
							timestampForOutput(v.Timestamp, outputType)})
					}
				}
				jsonOutput[i.Name] = jsonRepresentation
			}
		case interfaces.PropertiesType:
			val, err := astarteAPIClient.AppEngine.GetProperties(realm, deviceID, deviceIdentifierType, i.Name)
			if err != nil {
				warnOrFail(snapshotInterface, i.Name, err)
			}
			jsonRepresentation := make(map[string]interface{})
			for k, v := range val {
				jsonRepresentation[k] = v
				if v == nil {
					v = "(null)"
				}
				if snapshotInterface == "" {
					t.AppendRow([]interface{}{i.Name, k, v, i.Ownership, ""})
				} else {
					t.AppendRow([]interface{}{i.Name, k, v})
				}
			}
			jsonOutput[i.Name] = jsonRepresentation
		}
	}

	// Done
	renderOutput(t, jsonOutput, outputType)

	return nil
}

func warnOrFail(snapshotInterface, interfaceName string, err error) {
	if snapshotInterface != "" {
		// Fail only if we're parsing a single interface
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Just warn
	fmt.Fprintf(os.Stderr, "warn: Could not parse results for interface %s: %s\n", interfaceName, err)
}

func devicesGetSamplesF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	interfaceName := args[1]
	var interfacePath string
	if len(args) == 3 {
		interfacePath = args[2]
	}
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
	skipRealmManagementChecks = skipRealmManagementChecks || astarteAPIClient.RealmManagement == nil
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
		deviceDetails, err := astarteAPIClient.AppEngine.GetDevice(realm, deviceID, deviceIdentifierType)
		if err != nil {
			return err
		}
		for astarteInterface, interfaceIntrospection := range deviceDetails.Introspection {
			if astarteInterface != interfaceName {
				continue
			}

			// Query Realm Management to get details on the interface
			interfaceDescription, err := astarteAPIClient.RealmManagement.GetInterface(realm, astarteInterface,
				interfaceIntrospection.Major)
			if err != nil {
				return err
			}

			if interfaceDescription.Type != interfaces.DatastreamType {
				fmt.Printf("%s is not a Datastream interface. get-samples works only on Datastream interfaces\n", interfaceName)
				os.Exit(1)
			}

			interfaceFound = true
			isAggregate = interfaceDescription.Aggregation == interfaces.ObjectAggregation

			switch {
			case isAggregate && interfaceDescription.IsParametric() && interfacePath == "":
				fmt.Printf("%s is an aggregate parametric interface, a valid path should be specified\n", interfaceName)
				os.Exit(1)
			case !isAggregate && interfacePath == "":
				fmt.Printf("You need to specify a valid path for interface %s\n", interfaceName)
				os.Exit(1)
			default:
				if err := interfaces.ValidateQuery(interfaceDescription, interfacePath); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}

		if !interfaceFound {
			fmt.Printf("Device %s has no interface named %s\n", deviceID, interfaceName)
			os.Exit(1)
		}
	} else {
		isAggregate = forceAggregate
	}

	// We are good to go.
	t := tableWriterForOutputType(outputType)
	if !isAggregate {
		// Go with the table header
		t.AppendHeader(table.Row{"Timestamp", "Value"})
		printedValues := 0
		jsonOutput := []client.DatastreamValue{}
		datastreamPaginator, err := astarteAPIClient.AppEngine.GetDatastreamsTimeWindowPaginator(realm, deviceID,
			deviceIdentifierType, interfaceName, interfacePath, sinceTime, toTime, resultSetOrder)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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
		datastreamPaginator, err := astarteAPIClient.AppEngine.GetDatastreamsTimeWindowPaginator(realm, deviceID, deviceIdentifierType, interfaceName, interfacePath,
			sinceTime, toTime, resultSetOrder)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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
							headerRow = append(headerRow, path)
						}
						if value != nil {
							line = append(line, value)
						} else {
							line = append(line, "(null)")
						}
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

func devicesSendDataF(command *cobra.Command, args []string) error {
	deviceID := args[0]
	interfaceName := args[1]
	interfacePath := args[2]
	payloadData := args[3]
	forceIDType, err := command.Flags().GetString("force-id-type")
	if err != nil {
		return err
	}
	deviceIdentifierType, err := deviceIdentifierTypeFromFlags(deviceID, forceIDType)
	if err != nil {
		return err
	}
	skipRealmManagementChecks, err := command.Flags().GetBool("skip-realm-management-checks")
	if err != nil {
		return err
	}
	skipRealmManagementChecks = skipRealmManagementChecks || astarteAPIClient.RealmManagement == nil
	interfaceTypeString, err := command.Flags().GetString("interface-type")
	if err != nil {
		return err
	}
	if skipRealmManagementChecks && interfaceTypeString == "" {
		return fmt.Errorf("When not using Realm Management checks, --interface-type should always be specified")
	}

	iface, err := getProtoInterface(deviceID, deviceIdentifierType, interfaceName, interfaceTypeString, skipRealmManagementChecks)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if !skipRealmManagementChecks {
		if iface.Ownership != interfaces.ServerOwnership {
			fmt.Println("send-data makes sense only for server-owned interfaces")
			os.Exit(1)
		}
	}

	// Time to understand the payload type
	payloadTypeString, err := command.Flags().GetString("payload-type")
	if err != nil {
		return err
	}
	if skipRealmManagementChecks && payloadTypeString == "" {
		switch interfaceTypeString {
		case "aggregate-datastream", "aggregate-parametric-datastream":
			// In this case, it's ok not to pass anything as the type
		default:
			return fmt.Errorf("When not using Realm Management checks with interfaces with individual aggregation, --payload-type should always be specified")
		}
	}

	// Assign a payload Type only if it's not an aggregate
	var payloadType interfaces.AstarteMappingType
	if payloadTypeString != "" {
		payloadType = interfaces.AstarteMappingType(payloadTypeString)
		if err := payloadType.IsValid(); err != nil {
			// It's an input error, so return err
			return err
		}
	} else if !skipRealmManagementChecks && iface.Aggregation == interfaces.IndividualAggregation {
		if payloadType == "" {
			mapping, err := interfaces.InterfaceMappingFromPath(iface, interfacePath)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			payloadType = mapping.Type
		}
	}

	var parsedPayloadData interface{}
	if err := payloadType.IsValid(); err == nil {
		if parsedPayloadData, err = parseSendDataPayload(payloadData, payloadType); err != nil {
			return err
		}
	} else {
		// We have to treat it as an aggregate.
		aggrPayload := map[string]interface{}{}
		if err := json.Unmarshal([]byte(payloadData), &aggrPayload); err != nil {
			return err
		}

		// The json module parses all numbers into float64. To ensure Astarte will validate our payloads
		// correctly, we should convert to int every payload for which an integer conversion does not lose
		// in precision
		for k, v := range aggrPayload {
			switch val := v.(type) {
			case float64:
				if val == math.Trunc(val) {
					aggrPayload[k] = int(val)
				}
			}
		}

		parsedPayloadData = aggrPayload
	}

	if !skipRealmManagementChecks {
		// We can delegate the entirety of this to astarte-go
		err = astarteAPIClient.AppEngine.SendData(realm, deviceID, deviceIdentifierType, iface, interfacePath, parsedPayloadData)
	} else {
		// Don't risk it. Use raw functions and trust the server to fail, in case.
		switch interfaceTypeString {
		case "properties":
			err = astarteAPIClient.AppEngine.SetProperty(realm, deviceID, deviceIdentifierType, interfaceName, interfacePath, parsedPayloadData)
		case "individual-datastream", "individual-parametric-datastream":
			err = astarteAPIClient.AppEngine.SendDatastream(realm, deviceID, deviceIdentifierType, interfaceName, interfacePath, parsedPayloadData)
		case "aggregate-datastream", "aggregate-parametric-datastream":
			err = astarteAPIClient.AppEngine.SendDatastream(realm, deviceID, deviceIdentifierType, interfaceName, interfacePath, parsedPayloadData)
		default:
			err = fmt.Errorf("%s is not a valid Interface Type. Valid interface types are: properties, individual-datastream, aggregate-datastream, individual-parametric-datastream, aggregate-parametric-datastream", interfaceTypeString)
		}
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Done
	fmt.Println("ok")
	return nil
}

func getProtoInterface(deviceID string, deviceIdentifierType client.DeviceIdentifierType,
	interfaceName, interfaceTypeString string, skipRealmManagementChecks bool) (interfaces.AstarteInterface, error) {
	iface := interfaces.AstarteInterface{}

	if skipRealmManagementChecks {
		interfaceType := interfaces.DatastreamType
		interfaceAggregation := interfaces.IndividualAggregation
		isParametricInterface := false
		switch interfaceTypeString {
		case "properties":
			interfaceType = interfaces.PropertiesType
		case "individual-datastream":
		case "aggregate-datastream":
			interfaceAggregation = interfaces.ObjectAggregation
		case "individual-parametric-datastream":
			isParametricInterface = true
		case "aggregate-parametric-datastream":
			interfaceAggregation = interfaces.ObjectAggregation
			isParametricInterface = true
		default:
			return iface, fmt.Errorf("%s is not a valid Interface Type. Valid interface types are: properties, individual-datastream, aggregate-datastream, individual-parametric-datastream, aggregate-parametric-datastream", interfaceTypeString)
		}
		iface = interfaces.AstarteInterface{
			Name:        interfaceName,
			Type:        interfaceType,
			Aggregation: interfaceAggregation,
		}
		// Just a trick to trick the parser into doing the right thing.
		if isParametricInterface {
			iface.Mappings = []interfaces.AstarteInterfaceMapping{
				interfaces.AstarteInterfaceMapping{
					Endpoint: "/it/%{is}/parametric",
				},
			}
		}
	} else {
		// Get the device introspection
		deviceDetails, err := astarteAPIClient.AppEngine.GetDevice(realm, deviceID, deviceIdentifierType)
		if err != nil {
			return iface, err
		}

		for astarteInterface, interfaceIntrospection := range deviceDetails.Introspection {
			if astarteInterface != interfaceName {
				continue
			}

			// Query Realm Management to get details on the interface
			iface, err = astarteAPIClient.RealmManagement.GetInterface(realm, astarteInterface,
				interfaceIntrospection.Major)
			if err != nil {
				// Die here, given we really can't recover further
				return iface, err
			}
			break
		}
	}

	return iface, nil
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

func parseSendDataPayload(payload string, mappingType interfaces.AstarteMappingType) (interface{}, error) {
	// Default to string, as it will be ok for most cases
	var ret interface{} = payload
	var err error
	switch mappingType {
	case interfaces.Double:
		if ret, err = strconv.ParseFloat(payload, 64); err != nil {
			return nil, err
		}
	case interfaces.Integer, interfaces.LongInteger:
		if ret, err = strconv.ParseInt(payload, 10, 64); err != nil {
			return nil, err
		}
	case interfaces.Boolean:
		if ret, err = strconv.ParseBool(payload); err != nil {
			return nil, err
		}
	case interfaces.BinaryBlob:
		// We have to verify base64 decoding works
		if _, err := base64.StdEncoding.DecodeString(payload); err != nil {
			return nil, err
		}
	case interfaces.DateTime:
		if ret, err = dateparse.ParseAny(payload); err != nil {
			return nil, err
		}
	case interfaces.BinaryBlobArray, interfaces.BooleanArray, interfaces.DateTimeArray, interfaces.DoubleArray,
		interfaces.IntegerArray, interfaces.LongIntegerArray, interfaces.StringArray:
		var jsonOut []interface{}
		if err := json.Unmarshal([]byte(payload), &jsonOut); err != nil {
			return nil, err
		}
		retArray := []interface{}{}
		// Do a smarter conversion here.
		for _, v := range jsonOut {
			switch val := v.(type) {
			case string:
				p, err := parseSendDataPayload(val, interfaces.AstarteMappingType(strings.TrimSuffix(string(mappingType), "array")))
				if err != nil {
					return nil, err
				}
				retArray = append(retArray, p)
			default:
				retArray = append(retArray, val)
			}
		}
		ret = retArray
	}

	return ret, nil
}
