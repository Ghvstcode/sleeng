package cmd

import (
	"fmt"
	"github.com/Ghvstcode/sleeng/pkg/wallet"
	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"sort"
	"strings"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a new Solana wallet and saves the private key to disk",
	RunE:  initializeWallet,
}

var isPaperBased bool

var templates = &promptui.SelectTemplates{
	Label:    "{{ . | cyan }} ",
	Selected: "{{ . | green }} ",
}

func init() {
	InitCmd.Flags().BoolVarP(&isPaperBased, "paper", "p", false, "Create a paper-based wallet with seed phrase instead of saving private key to disk")
}

func printBlue(msg string, args ...interface{}) {
	blue := color.New(color.FgBlue)
	blue.Printf(msg, args...)
}

func initializeWallet(cmd *cobra.Command, args []string) error {
	wc := wallet.NewWalletConfig()
	if isPaperBased {
		return handlePaperBasedWallet(wc)
	}
	return handleFileBasedWallet(wc)
}

func handlePaperBasedWallet(wc *wallet.WalletConfig) error {
	choice, err := promptForChoice("Do you want to create a new paper-based wallet or import an existing one?", []string{"New", "Import"})
	if err != nil {
		return fmt.Errorf("failed to get user choice: %w", err)
	}
	switch choice {
	case "New":
		return createNewPaperWallet(wc)
	case "Import":
		return importExistingPaperWallet(wc)
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}
}

func createNewPaperWallet(wc *wallet.WalletConfig) error {
	seed, walletAddr, err := wc.GenerateNewPaperWallet()
	if err != nil {
		return fmt.Errorf("failed to generate new paper wallet: %w", err)
	}
	clipboard.WriteAll(walletAddr)
	printBlue("New Wallet Created. Your Address Is: %s (copied to clipboard)\n", walletAddr)
	printBlue("Seed Phrase (keep this safe): %s\n", seed)
	return postWalletInitializationActions(wc)
}

func importExistingPaperWallet(wc *wallet.WalletConfig) error {
	seedPhrase, err := promptForInput("Please enter your existing seed phrase:", wc.IsValidSeed)
	if err != nil {
		return fmt.Errorf("failed to get seed phrase: %w", err)
	}
	address, err := wc.ImportWalletFromSeed(seedPhrase)
	if err != nil {
		return fmt.Errorf("failed to import wallet: %w", err)
	}
	clipboard.WriteAll(address)
	printBlue("New Wallet Created. Your Address Is: %s (copied to clipboard)\n", address)
	return postWalletInitializationActions(wc)
}

func handleFileBasedWallet(wc *wallet.WalletConfig) error {
	if privateKeyFlag != "" {
		return createNewFileBasedWallet(wc, aliasFlag, privateKeyFlag)
	}

	hasWallets, err := wc.HasWallets()
	if err != nil {
		return fmt.Errorf("error checking for existing wallets: %w", err)
	}

	if hasWallets {
		return handleExistingWallets(wc)
	}
	return createNewFileBasedWallet(wc, aliasFlag, "")
}

func handleExistingWallets(wc *wallet.WalletConfig) error {
	choice, err := promptForChoice("You already have an existing wallet with keys saved on this computer! Select an option below", []string{"Select Existing Wallet", "Create New Wallet"})
	if err != nil {
		return fmt.Errorf("failed to get user choice: %w", err)
	}

	switch choice {
	case "Select Existing Wallet":
		return selectExistingWallet(wc)
	case "Create New Wallet":
		return createNewFileBasedWallet(wc, "", "")
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}
}

func selectExistingWallet(wc *wallet.WalletConfig) error {
	aliases, _, err := wc.RetrieveWallets()
	if err != nil {
		return fmt.Errorf("failed to retrieve existing wallets: %w", err)
	}

	selectedWallet, err := promptForChoice("Choose From Your List Of Existing Wallets", aliases)
	if err != nil {
		return fmt.Errorf("failed to get user choice: %w", err)
	}

	actualAlias := strings.Split(selectedWallet, " ")[0]
	err = wc.SwitchWallet(actualAlias)
	if err != nil {
		return fmt.Errorf("failed to switch to existing wallet: %w", err)
	}

	newAddr, err := wc.RetrieveCurrentWalletAddress()
	if err != nil {
		return fmt.Errorf("failed to get the current wallet address: %w", err)
	}

	clipboard.WriteAll(newAddr)
	printBlue("Switched To A New Wallet. Your Address Is: %s (copied to clipboard)\n", newAddr)
	return nil
}

func createNewFileBasedWallet(wc *wallet.WalletConfig, alias, privateKey string) error {
	// Prompt for alias if it's empty
	if alias == "" {
		var err error
		alias, err = promptForInput("Create An Alias For Your Wallet:", nil)
		if err != nil {
			return fmt.Errorf("failed to get wallet alias: %w", err)
		}
	}

	// Create or import the wallet based on whether a private key is provided
	var newWallet string
	var err error
	if privateKey == "" {
		newWallet, err = wc.CreateNewWallet(alias)
	} else {
		newWallet, err = wc.CreateNewWalletWithKey(alias, privateKey)
	}
	if err != nil {
		return fmt.Errorf("failed to create new wallet: %w", err)
	}

	// Copy the new wallet address to the clipboard and print it
	clipboard.WriteAll(newWallet)
	action := "Created"
	if privateKey != "" {
		action = "Imported"
	}
	printBlue("New Wallet %s. Your Address Is: %s (copied to clipboard)\n", action, newWallet)

	return nil
}

func promptForChoice(label string, items []string) (string, error) {
	prompt := promptui.Select{
		Label:     label,
		Items:     items,
		Templates: templates,
	}
	_, choice, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return choice, nil
}

func promptForInput(label string, validator func(input string) error) (string, error) {
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validator,
	}
	return prompt.Run()
}

func postWalletInitializationActions(wc *wallet.WalletConfig) error {
	for {
		choice, err := promptForChoice("What would you like to do next?", []string{"Check Balance(EUR)", "Get Current SOL/EUR Rate", "Retrieve Wallet Address", "Retrieve Transactions", "Exit"})
		if err != nil {
			return fmt.Errorf("failed to get user choice: %w", err)
		}

		switch choice {
		case "Exit":
			return nil
		default:
			err := processPostInitializationChoice(choice, wc)
			if err != nil {
				return fmt.Errorf("failed to process choice: %w", err)
			}
		}
	}
}

func processPostInitializationChoice(choice string, wc *wallet.WalletConfig) error {
	switch choice {
	case "Check Balance(EUR)":
		bal, err := wc.GetCurrentWalletBalanceInEUR("")
		if err != nil {
			return fmt.Errorf("failed to check balance: %w", err)
		}
		printBlue("Balance of the active wallet: €%s\n", bal)
	case "Retrieve Wallet Address":
		publicKey, err := wc.RetrieveCurrentWalletAddress()
		if err != nil {
			return fmt.Errorf("failed to retrieve wallet address: %w", err)
		}
		printBlue("Public Key of The Active Wallet: %s\n", publicKey)
	case "Get Current SOL/EUR Rate":
		rate, err := wc.FetchSOLEURRate()
		if err != nil {
			return fmt.Errorf("failed to retrieve rate: %w", err)
		}

		printBlue("Current SOL/EUR Rate: €%s\n", rate)
	case "Retrieve Transactions":
		transactions, err := wc.GetTransactionHistory()
		if err != nil {
			return fmt.Errorf("failed to retrieve transactions: %w", err)
		}

		sort.Slice(transactions, func(i, j int) bool {
			return transactions[i].Timestamp.After(transactions[j].Timestamp)
		})

		rate, err := wc.FetchSOLEURRate()
		if err != nil {
			return fmt.Errorf("error fetching SOL to EUR rate: %v", err)
		}

		printTransactions(transactions, rate)
	default:
		fmt.Println("Invalid choice. Returning to main menu.")
	}
	return nil
}
