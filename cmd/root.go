package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Solana Wallet CLI",
	Long:  `A command-line interface to interact with Solana wallet.`,
}

func Execute() error {
	return RootCmd.Execute()
}
