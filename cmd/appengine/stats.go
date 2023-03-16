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
	"fmt"

	"github.com/astarte-platform/astartectl/utils"
	"github.com/spf13/cobra"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:     "stats",
	Short:   "Show stats",
	Aliases: []string{"stat", "statistics"},
}

var statsDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Show devices stats",
	Long: `Show various devices stats, such as the total number of devices and the number of
connected devices`,
	Example: `  astartectl appengine stats devices`,
	Args:    cobra.NoArgs,
	Aliases: []string{"device"},
	RunE:    statsDevicesF,
}

func init() {
	statsCmd.AddCommand(statsDevicesCmd)

	AppEngineCmd.AddCommand(statsCmd)
}

func statsDevicesF(command *cobra.Command, args []string) error {
	devicesStatsReq, err := astarteAPIClient.GetDevicesStats(realm)
	if err != nil {
		return err
	}

	utils.MaybeCurlAndExit(devicesStatsReq, astarteAPIClient)

	devicesStatsRes, err := devicesStatsReq.Run(astarteAPIClient)
	if err != nil {
		return err
	}
	devicesStats, _ := devicesStatsRes.Parse()

	fmt.Printf("%+v\n", devicesStats)
	return nil
}
