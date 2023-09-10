package wallet

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/mr-tron/base58"
	"github.com/shopspring/decimal"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ed25519"
	"io/ioutil"
	"math/rand"
	"strings"
)

type WalletConfig struct {
	PrivateKey   string `json:"private_key"`
	Alias        string `json:"alias,omitempty"`
	IsPaperBased bool   `json:"is_paper_based,omitempty"`
	SeedPhrase   string `json:"seed_phrase,omitempty"`
	Wallet       *solana.Wallet
	KeyOps       KeyStore
}

type Wallet struct {
	PrivateKey string          `json:"key"`
	Balance    decimal.Decimal `json:"balance"`
	PublicKey  string          `json:"publicKey"`
}

type WalletData struct {
	ActiveAlias string            `json:"activeAlias"`
	Wallets     map[string]Wallet `json:"wallets"`
}

// KeyStore represents key file operations.
type KeyStore interface {
	GetCurrentPrivateKey() (string, error)
	GetPrivateKeyByAlias(alias string) (string, error)
	IsKeyFilePresent() (bool, error)
	SetActiveKey(aliasToActivate string) error
	GetCurrentPublicKey() (string, error)
	GetPublicKeyByAlias(alias string) (string, error)
	WriteKeyToFile(alias string, key ed25519.PrivateKey, walletAddress string) error
	PrintAllKeys() ([]string, map[string]string, error)
}

// NewWalletConfig initializes a new WalletConfig.
func NewWalletConfig() *WalletConfig {
	return &WalletConfig{
		KeyOps: &KeyOps{
			FileReader: &IOUtilFileReader{},
			FileWriter: &IOUtilFileWriter{},
		},
	}
}

// GenerateNewPaperWallet generates a new paper wallet.
func (w *WalletConfig) GenerateNewPaperWallet() (string, string, error) {
	seed, privateKey, err := createKeyPairWithMnemonic("")
	if err != nil {
		return "", "", err
	}
	wallet, err := solana.WalletFromPrivateKeyBase58(base58.Encode(privateKey))
	if err != nil {
		return "", "", err
	}

	w.IsPaperBased = true
	w.Wallet = wallet
	return seed, wallet.PublicKey().String(), nil
}

func (w *WalletConfig) ImportWalletFromSeed(mnemonic string) (string, error) {
	_, privateKey, err := createKeyPairWithMnemonic(mnemonic)
	if err != nil {
		return "", err
	}
	wallet, err := solana.WalletFromPrivateKeyBase58(base58.Encode(privateKey))
	if err != nil {
		return "", err
	}

	w.IsPaperBased = true
	w.Wallet = wallet
	return wallet.PublicKey().String(), nil
}

func (w *WalletConfig) CreateNewWallet(alias string) (string, error) {
	account := solana.NewWallet()

	if alias == "" {
		alias = getRandomAlias() + "-" + "wallet"
	}

	err := w.KeyOps.WriteKeyToFile(alias, ed25519.PrivateKey(account.PrivateKey), account.PublicKey().String())
	if err != nil {
		return "", err
	}

	return account.PublicKey().String(), nil
}

func (w *WalletConfig) GetCurrentWalletBalanceInEUR(alias string) (string, error) {
	solBalance, err := w.fetchSolBalance(alias, w.KeyOps)
	if err != nil {
		return "", err
	}

	//Get the SOL to EUR exchange rate
	rate, err := fetchSOLEURRate()
	if err != nil {
		return "", err
	}

	//Convert SOL to EUR
	eurBalance := solBalance.Mul(rate)

	// Convert to string with 2 decimal places, e.g. 123.45 as this is the standard for displaying currencies
	return eurBalance.StringFixed(2), nil
}

func (w *WalletConfig) SwitchWallet(alias string) error {
	return w.KeyOps.SetActiveKey(alias)
}

func (w *WalletConfig) RetrieveWallets() ([]string, map[string]string, error) {
	return w.KeyOps.PrintAllKeys()
}

func (w *WalletConfig) RetrieveCurrentWalletAddress() (string, error) {
	if w.Wallet != nil {
		return w.Wallet.PublicKey().String(), nil
	}
	return w.KeyOps.GetCurrentPublicKey()
}

func (w *WalletConfig) RetrieveWalletAddressByAlias(alias string) (string, error) {
	return w.KeyOps.GetPublicKeyByAlias(alias)
}

func (w *WalletConfig) HasWallets() (bool, error) {
	return w.KeyOps.IsKeyFilePresent()
}

// createKeyPairWithMnemonic creates a key pair with an optional mnemonic.
func createKeyPairWithMnemonic(mnemonic string) (string, ed25519.PrivateKey, error) {
	if mnemonic == "" {
		entropy, err := bip39.NewEntropy(128)
		if err != nil {
			return "", nil, fmt.Errorf("error generating entropy: %w", err)
		}

		mnemonic, err = bip39.NewMnemonic(entropy)
		if err != nil {
			return "", nil, fmt.Errorf("error generating mnemonic: %w", err)
		}
	}

	entropy, err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		return "", nil, fmt.Errorf("mnemonic not valid: %w", err)
	}

	privateKey := ed25519.NewKeyFromSeed([]byte(hex.EncodeToString(entropy)))
	return mnemonic, privateKey, nil
}

// FetchSOLEURRate fetches the current SOL to EUR exchange rate.
func (w *WalletConfig) FetchSOLEURRate() (decimal.Decimal, error) {
	return fetchSOLEURRate()
}

func (w *WalletConfig) GetTransactionHistory() ([]*Transaction, error) {
	var err error
	var publicKeyStr string

	// Check if the Wallet object is already available
	if w.Wallet != nil {
		publicKeyStr = w.Wallet.PublicKey().String()
	} else {
		publicKeyStr, err = w.KeyOps.GetCurrentPublicKey()
		if err != nil {
			return nil, fmt.Errorf("failed to get current private key: %w", err)
		}
	}

	// Fetch transactions using the public key
	transactions, err := fetchTransactions(publicKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	return transactions, nil
}

func (w *WalletConfig) CreateNewWalletWithKey(alias, privateKey string) (string, error) {
	privkey, err := solana.PrivateKeyFromBase58(privateKey)
	if err != nil {
		return "", fmt.Errorf("error generating new wallet with key: %w", err)
	}

	if alias == "" {
		alias = getRandomAlias() + "-" + "wallet"
	}

	err = w.KeyOps.WriteKeyToFile(alias, ed25519.PrivateKey(privkey), privkey.PublicKey().String())
	if err != nil {
		return "", err
	}

	return privkey.PublicKey().String(), nil
}

// getRandomAlias generates a random alias using words from the BIP-39 word list.
func getRandomAlias() string {
	// Get the English BIP-39 word list
	wordList := bip39.GetWordList()
	// Pick a random word from the list
	// TODO: Use a cryptographically secure random number generator or seed the random number generator
	return wordList[rand.Intn(len(wordList))]
}

// IOUtilFileReader is a file reader using ioutil.
type IOUtilFileReader struct{}

// ReadFile reads a file and returns its content.
func (r *IOUtilFileReader) ReadFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filename, err)
	}
	return data, nil
}

// IOUtilFileWriter is a file writer using ioutil.
type IOUtilFileWriter struct{}

// WriteFile writes data to a file.
func (w *IOUtilFileWriter) WriteFile(filename string, data []byte) error {
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("error writing to file %s: %w", filename, err)
	}
	return nil
}

func (w *WalletConfig) IsValidSeed(mnemonic string) error {
	// 1. Check if mnemonic is empty
	if mnemonic == "" {
		return fmt.Errorf("mnemonic is empty")
	}

	// 2. Split the mnemonic into words
	words := strings.Fields(mnemonic)
	wordCount := len(words)

	// 3. Mnemonic should be 12, 15, 18, 21, or 24 words long
	if wordCount != 12 && wordCount != 15 && wordCount != 18 && wordCount != 21 && wordCount != 24 {
		return fmt.Errorf("invalid mnemonic length. got %d words, expected 12, 15, 18, 21, or 24 words", wordCount)
	}

	// 5. Check if the mnemonic as a whole is valid (this includes checksum validation)
	if !bip39.IsMnemonicValid(mnemonic) {
		return fmt.Errorf("mnemonic is not valid")
	}

	return nil
}

func (w *WalletConfig) SendFunds(ctx context.Context, amount, recipient string) (string, error) {
	var privKeyStr string
	rpcClient := rpc.New(rpc.DevNet_RPC)
	wsClient, err := ws.Connect(ctx, rpc.DevNet_WS)
	if err != nil {
		return "", err
	}

	if w.Wallet != nil {
		privKeyStr = w.Wallet.PrivateKey.String()
	} else {
		privKeyStr, err = w.KeyOps.GetCurrentPrivateKey()
		if err != nil {
			return "", fmt.Errorf("failed to get current private key: %w", err)
		}
	}

	accountFrom, err := solana.PrivateKeyFromBase58(privKeyStr)
	if err != nil {
		return "", err
	}

	accountTo := solana.MustPublicKeyFromBase58(recipient)

	rate, err := fetchSOLEURRate()
	if err != nil {
		return "", err
	}

	amountToSend, err := ConvertEurToLamports(amount, rate)
	if err != nil {
		return "", err
	}

	recent, err := rpcClient.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", err
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				uint64(amountToSend),
				accountFrom.PublicKey(),
				accountTo,
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(accountFrom.PublicKey()),
	)
	if err != nil {
		return "", err
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if accountFrom.PublicKey().Equals(key) {
				return &accountFrom
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("unable to sign transaction: %w", err)
	}

	sig, err := confirm.SendAndConfirmTransaction(
		ctx,
		rpcClient,
		wsClient,
		tx,
	)
	if err != nil {
		return "", err
	}

	return sig.String(), nil
}

func ConvertEurToLamports(eurStr string, eurToSolRate decimal.Decimal) (int64, error) {
	eurAmount, err := decimal.NewFromString(eurStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse EUR string: %w", err)
	}

	solAmount := eurAmount.Div(eurToSolRate)

	lamports := solAmount.Mul(decimal.NewFromInt(1_000_000_000)).IntPart()

	return lamports, nil
}
