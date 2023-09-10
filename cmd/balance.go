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

func displayBalance(_ *cobra.Command, _ []string) error {
	var balance string
	var err error
	wc := wallet.NewWalletConfig()
	if aliasFlag == "" {
		balance, err = wc.GetCurrentWalletBalanceInEUR("")
	} else {
		balance, err = wc.GetCurrentWalletBalanceInEUR(aliasFlag)
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
