package cmd

import (
	"fmt"
	"github.com/Ghvstcode/sleeng/pkg/wallet"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"strings"
)

var (
	listAll bool
)

var AddressCmd = &cobra.Command{
	Use:   "address",
	Short: "Prints the public key of the Solana wallet",
	Long: `By default, prints the public key of the current active Solana wallet.
Provide an alias to get the public key of a specific wallet.
Use the --all flag to list public keys of all wallets.`,
	RunE: displayAddress,
}

func init() {
	AddressCmd.Flags().BoolVar(&listAll, "all", false, "List all wallet addresses")
}

func displayAddress(_ *cobra.Command, _ []string) error {
	blue := color.New(color.FgBlue)
	boldBlue := blue.Add(color.Bold)

	wc := wallet.NewWalletConfig()

	if listAll {
		aliases, addressMap, err := wc.RetrieveWallets()
		if err != nil {
			return fmt.Errorf("failed to retrieve wallets: %v", err)
		}
		for _, ali := range aliases {
			actualAlias := strings.Split(ali, " ")[0]
			boldBlue.Printf("Public Key of %s: %s\n", actualAlias, addressMap[actualAlias])
		}
		return nil
	}

	if aliasFlag != "" {
		publicKey, err := wc.RetrieveWalletAddressByAlias(aliasFlag)
		if err != nil {
			return fmt.Errorf("failed to retrieve public key for alias %s: %v", aliasFlag, err)
		}
		boldBlue.Printf("Public Key of %s: %s\n", aliasFlag, publicKey)
		return nil
	}

	publicKey, err := wc.RetrieveCurrentWalletAddress()
	if err != nil {
		return fmt.Errorf("failed to retrieve public key: %v", err)
	}

	boldBlue.Printf("Public Key of The Active Wallet: %s\n", publicKey)
	return nil
}
