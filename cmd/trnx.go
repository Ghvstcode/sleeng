package cmd

import (
	"fmt"
	"github.com/Ghvstcode/sleeng/pkg/wallet"
	"github.com/shopspring/decimal"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

const (
	solToLamportConversion = 1e9 // 1 SOL = 1,000,000,000 lamports
)

var transactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "Prints the transaction history in EUR, from newest to oldest.",
	RunE:  executeTransactions,
}

func executeTransactions(cmd *cobra.Command, args []string) error {
	wc := wallet.NewWalletConfig()

	transactions, err := wc.GetTransactionHistory()
	if err != nil {
		return fmt.Errorf("error fetching transactions: %v", err)
	}

	// Sort transactions by timestamp from newest to oldest.
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Timestamp.After(transactions[j].Timestamp)
	})

	rate, err := wc.FetchSOLEURRate()
	if err != nil {
		return fmt.Errorf("error fetching SOL to EUR rate: %v", err)
	}

	printTransactions(transactions, rate)

	return nil
}

func printTransactions(transactions []*wallet.Transaction, rate decimal.Decimal) {
	if len(transactions) == 0 {
		fmt.Println("No transactions to display.")
		return
	}
	for _, tx := range transactions {
		printTransaction(tx, rate)
	}
}

func printTransaction(tx *wallet.Transaction, rate decimal.Decimal) {
	amountInLamports := decimal.NewFromInt(int64(tx.Amount))
	amountInSol := amountInLamports.Div(decimal.NewFromInt(solToLamportConversion))
	amountInEur := amountInSol.Mul(rate)

	action := "Received"
	if tx.IsSender {
		action = "Sent"
	}

	fmt.Printf(
		"Action: %s\nFrom: %s\nTo: %s\nAmount: %s EUR\nTimestamp: %s\n---\n",
		action,
		tx.From,
		tx.To,
		amountInEur.StringFixed(2),
		tx.Timestamp.Format(time.RFC3339),
	)
}
