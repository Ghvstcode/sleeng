package cmd

import (
	"context"
	"fmt"
	"github.com/Ghvstcode/sleeng/pkg/wallet"
	"github.com/spf13/cobra"
	"log"
)

var sendCmd = &cobra.Command{
	Use:   "send [EUR amount] [destination]",
	Short: "Sends <EUR amount>'s worth of SOL to the destination address",
	Args:  cobra.ExactArgs(2), // You expect exactly two arguments
	Run:   send,
}

func send(cmd *cobra.Command, args []string) {
	amount := args[0]
	destination := args[1]

	// Assuming you have a WalletConfig object named `walletConfig`
	walletConfig := wallet.NewWalletConfig()

	signature, err := walletConfig.SendFunds(context.Background(), amount, destination)
	if err != nil {
		log.Fatalf("Failed to send funds: %v", err.Error())
	}

	fmt.Printf("Successfully sent %s EUR to %s. Transaction Signature: %s\n", amount, destination, signature)
}
