package cmd

import (
	"fmt"
	"github.com/Ghvstcode/sleeng/pkg/wallet"
	"github.com/spf13/cobra"
)

var BalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Prints the balance of a specific or the current active Solana wallet in EUR",
	RunE:  displayBalance,
}

func displayBalance(cmd *cobra.Command, args []string) error {
	var balance string
	var err error
	wc := wallet.NewWalletConfig()
	if aliasFlag == "" {
		balance, err = wc.GetCurrentWalletBalanceInEUR("") // Retrieve balance of the current active wallet
	} else {
		balance, err = wc.GetCurrentWalletBalanceInEUR(aliasFlag) // Retrieve balance of the wallet by alias
	}

	if err != nil {
		return fmt.Errorf("failed to retrieve wallet balance: %v", err)
	}

	if aliasFlag != "" {
		fmt.Printf("Balance of %s wallet: €%s\n", aliasFlag, balance)
	} else {
		fmt.Printf("Balance of the active wallet: €%s\n", balance)
	}

	return nil
}
