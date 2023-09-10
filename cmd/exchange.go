package cmd

import (
	"fmt"
	"github.com/Ghvstcode/sleeng/pkg/wallet"
	"github.com/spf13/cobra"
)

// exchangeCmd represents the exchange command
var exchangeCmd = &cobra.Command{
	Use:   "exchange",
	Short: "Print the current exchange rate of SOL to EUR",
	Long:  `This command fetches and prints the current exchange rate of SOL to EUR.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return PrintExchangeRate()
	},
}

func PrintExchangeRate() error {
	wc := wallet.NewWalletConfig()
	rate, err := wc.FetchSOLEURRate()
	if err != nil {
		return err
	}
	fmt.Printf("Current exchange rate of SOL to EUR: %v\n", rate)

	return nil
}
