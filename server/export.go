package server

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"io/ioutil"
	"path"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

const (
	flagHeight         = "height"
	flagForZeroHeight  = "for-zero-height"
	flagKickValidators = "kick"
)

// ExportCmd dumps app state to JSON.
func ExportCmd(ctx *Context, cdc *codec.Codec, appExporter AppExporter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export state to JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := viper.GetString("home")
			traceWriterFile := viper.GetString(flagTraceStore)
			emptyState, err := isEmptyState(home)
			if err != nil {
				return err
			}

			if emptyState {
				fmt.Println("WARNING: State is not initialized. Returning genesis file.")
				genesisFile := path.Join(home, "config", "genesis.json")
				genesis, err := ioutil.ReadFile(genesisFile)
				if err != nil {
					return err
				}
				fmt.Println(string(genesis))
				return nil
			}

			db, err := openDB(home)
			if err != nil {
				return err
			}
			traceWriter, err := openTraceWriter(traceWriterFile)
			if err != nil {
				return err
			}
			height := viper.GetInt64(flagHeight)
			forZeroHeight := viper.GetBool(flagForZeroHeight)
			kickValidatorsString := viper.GetString(flagKickValidators)
			var kickValidators []types.ValAddress
			if kickValidatorsString != "" {
				for _, val := range strings.Split(kickValidatorsString, ",") {
					valAddr, err := types.ValAddressFromBech32(val)
					if err != nil {
						return errors.Errorf("error decoding kick address '%v': %v\n", val, err)
					}
					kickValidators = append(kickValidators, valAddr)
				}
			}

			appState, validators, err := appExporter(ctx.Logger, db, traceWriter, height, forZeroHeight, kickValidators)
			if err != nil {
				return errors.Errorf("error exporting state: %v\n", err)
			}

			doc, err := tmtypes.GenesisDocFromFile(ctx.Config.GenesisFile())
			if err != nil {
				return err
			}

			doc.AppState = appState
			doc.Validators = validators

			encoded, err := codec.MarshalJSONIndent(cdc, doc)
			if err != nil {
				return err
			}

			fmt.Println(string(encoded))
			return nil
		},
	}
	cmd.Flags().Int64(flagHeight, -1, "Export state from a particular height (-1 means latest height)")
	cmd.Flags().Bool(flagForZeroHeight, false, "Export state to start at height zero (perform preproccessing)")
	cmd.Flags().String(flagKickValidators, "", "Kick validators from the set and null their accounts (perform preproccessing)")
	return cmd
}

func isEmptyState(home string) (bool, error) {
	files, err := ioutil.ReadDir(path.Join(home, "data"))
	if err != nil {
		return false, err
	}

	return len(files) == 0, nil
}
