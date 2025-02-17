package subcmd_loadledger

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/urfave/cli.v1"

	"github.com/vitelabs/go-vite/cmd/nodemanager"
	"github.com/vitelabs/go-vite/cmd/utils"
	"github.com/vitelabs/go-vite/ledger/pipeline"
	"github.com/vitelabs/go-vite/log15"
)

var (
	fromDirFlag = cli.StringFlag{
		Name:  "fromDir",
		Usage: "from directory",
	}
	LoadLedgerCommand = cli.Command{
		Action:      utils.MigrateFlags(exportLedgerAction),
		Name:        "load",
		Usage:       "load --fromDir /xxx/xxx",
		Flags:       append([]cli.Flag{fromDirFlag}, utils.ConfigFlags...),
		Category:    "LOCAL COMMANDS",
		Description: `Load ledger.`,
	}
	log = log15.New("module", "gvite/loadledger")
)

func exportLedgerAction(ctx *cli.Context) error {
	fromDir := ctx.String(fromDirFlag.GetName())

	if fromDir == "" {
		return errors.New("fromDir not set")
	}

	if _, err := os.Stat(fromDir); os.IsNotExist(err) {
		return fmt.Errorf("directory %s is not exist", fromDir)
	}

	node, err := nodemanager.LocalNodeMaker{}.MakeNode(ctx)
	if err != nil {
		return err
	}

	if err := node.Prepare(); err != nil {
		return err
	}

	if err := node.Start(); err != nil {
		return err
	}
	pipe, err := pipeline.NewBlocksPipeline(fromDir, node.Vite().Chain().GetLatestSnapshotBlock().Height)
	if err != nil {
		log.Error("create blocks pipeline fail", err)
		node.Stop()
		return err
	}

	log.Info("run blocks pipeline successful")
	node.Vite().Pool().AddPipeline(pipe)
	node.Wait()
	return nil
}
