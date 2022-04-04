// Copyright © 2020 Vulcanize, Inc
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/vulcanize/ipld-eth-state-snapshot/pkg/snapshot"
)

// stateSnapshotCmd represents the stateSnapshot command
var stateSnapshotCmd = &cobra.Command{
	Use:   "stateSnapshot",
	Short: "Extract the entire Ethereum state from leveldb and publish into PG-IPFS",
	Long: `Usage

./ipld-eth-state-snapshot stateSnapshot --config={path to toml config file}`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *logrus.WithField("SubCommand", subCommand)
		stateSnapshot()
	},
}

func stateSnapshot() {
	modeStr := viper.GetString("snapshot.mode")
	mode := snapshot.SnapshotMode(modeStr)
	config, err := snapshot.NewConfig(mode)
	if err != nil {
		logWithCommand.Fatal("unable to initialize config: %v", err)
	}
	logWithCommand.Infof("opening levelDB and ancient data at %s and %s",
		config.Eth.LevelDBPath, config.Eth.AncientDBPath)
	edb, err := snapshot.NewLevelDB(config.Eth)
	if err != nil {
		logWithCommand.Fatal(err)
	}
	height := viper.GetInt64("snapshot.blockHeight")
	recoveryFile := viper.GetString("snapshot.recoveryFile")
	if recoveryFile == "" {
		recoveryFile = fmt.Sprintf("./%d_snapshot_recovery", height)
		logWithCommand.Infof("no recovery file set, creating default: %s", recoveryFile)
	}

	pub, err := snapshot.NewPublisher(mode, config)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	snapshotService, err := snapshot.NewSnapshotService(edb, pub, recoveryFile)
	if err != nil {
		logWithCommand.Fatal(err)
	}
	workers := viper.GetUint("snapshot.workers")

	if height < 0 {
		if err := snapshotService.CreateLatestSnapshot(workers); err != nil {
			logWithCommand.Fatal(err)
		}
	} else {
		params := snapshot.SnapshotParams{Workers: workers, Height: uint64(height)}
		if err := snapshotService.CreateSnapshot(params); err != nil {
			logWithCommand.Fatal(err)
		}
	}
	logWithCommand.Infof("state snapshot at height %d is complete", height)
}

func init() {
	rootCmd.AddCommand(stateSnapshotCmd)

	stateSnapshotCmd.PersistentFlags().String("leveldb-path", "", "path to primary datastore")
	stateSnapshotCmd.PersistentFlags().String("ancient-path", "", "path to ancient datastore")
	stateSnapshotCmd.PersistentFlags().String("block-height", "", "blockheight to extract state at")
	stateSnapshotCmd.PersistentFlags().Int("workers", 1, "number of concurrent workers to use")
	stateSnapshotCmd.PersistentFlags().String("recovery-file", "", "file to recover from a previous iteration")
	stateSnapshotCmd.PersistentFlags().String("snapshot-mode", "postgres", "output mode for snapshot ('file' or 'postgres')")
	stateSnapshotCmd.PersistentFlags().String("output-dir", "", "directory for writing ouput to while operating in 'file' mode")

	viper.BindPFlag(snapshot.LVL_DB_PATH_TOML, stateSnapshotCmd.PersistentFlags().Lookup("leveldb-path"))
	viper.BindPFlag(snapshot.ANCIENT_DB_PATH_TOML, stateSnapshotCmd.PersistentFlags().Lookup("ancient-path"))
	viper.BindPFlag(snapshot.SNAPSHOT_BLOCK_HEIGHT_TOML, stateSnapshotCmd.PersistentFlags().Lookup("block-height"))
	viper.BindPFlag(snapshot.SNAPSHOT_WORKERS_TOML, stateSnapshotCmd.PersistentFlags().Lookup("workers"))
	viper.BindPFlag(snapshot.SNAPSHOT_RECOVERY_FILE_TOML, stateSnapshotCmd.PersistentFlags().Lookup("recovery-file"))
	viper.BindPFlag(snapshot.SNAPSHOT_MODE_TOML, stateSnapshotCmd.PersistentFlags().Lookup("snapshot-mode"))
	viper.BindPFlag(snapshot.FILE_OUTPUT_DIR_TOML, stateSnapshotCmd.PersistentFlags().Lookup("output-dir"))
}
