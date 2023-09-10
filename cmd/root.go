package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Solana Wallet CLI",
	Long:  `A command-line interface to interact with Solana wallet.`,
}

var (
	privateKeyFlag, aliasFlag string
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&privateKeyFlag, "key", "k", "", "A base58 encoded private key to use instead of the one saved on disk")
	RootCmd.PersistentFlags().StringVarP(&aliasFlag, "alias", "a", "", "Optional alias for the wallet")
	RootCmd.AddCommand(InitCmd, AddressCmd, BalanceCmd, exchangeCmd, transactionsCmd, sendCmd)
}

func Execute() error {
	return RootCmd.Execute()
}
