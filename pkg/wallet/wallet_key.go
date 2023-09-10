// Package wallet provides functionalities to manage wallet keys.
package wallet

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mr-tron/base58"
	"github.com/shopspring/decimal"
	"os"
)

// FileReader is an interface that wraps the ReadFile method.
type FileReader interface {
	ReadFile(filename string) ([]byte, error)
}

type FileWriter interface {
	WriteFile(filename string, data []byte) error
}

// KeyOps performs operations related to wallet keys.
type KeyOps struct {
	FileReader FileReader
	FileWriter FileWriter
}

const KeyFilePath = "standard.solana-keygen.json"

var ErrActiveWalletNotFound = errors.New("no active wallet found")

// readWalletData reads and unmarshals wallet data from a given file path.
func (k *KeyOps) readWalletData(filePath string) (WalletData, error) {
	var data WalletData

	fileData, err := k.FileReader.ReadFile(filePath)
	if err != nil {
		return data, fmt.Errorf("error reading file: %w", err)
	}

	if err = json.Unmarshal(fileData, &data); err != nil {
		return data, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return data, nil
}

// GetCurrentPrivateKey retrieves the current active wallet's private key.
func (k *KeyOps) GetCurrentPrivateKey() (string, error) {
	data, err := k.readWalletData(KeyFilePath)
	if err != nil {
		return "", err
	}

	activeWallet, exists := data.Wallets[data.ActiveAlias]
	if !exists {
		return "", ErrActiveWalletNotFound
	}

	return activeWallet.Key, nil
}

// GetPrivateKeyByAlias retrieves a wallet's private key by its alias.
func (k *KeyOps) GetPrivateKeyByAlias(alias string) (string, error) {
	data, err := k.readWalletData(KeyFilePath)
	if err != nil {
		return "", err
	}

	wallet, exists := data.Wallets[alias]
	if !exists {
		return "", fmt.Errorf("no wallet found for alias: %s", alias)
	}

	return wallet.Key, nil
}

// IsKeyFilePresent checks if there is a file containing some keys already in place.
func (k *KeyOps) IsKeyFilePresent() (bool, error) {
	// Try to read the file
	_, err := k.FileReader.ReadFile(KeyFilePath)
	if err != nil {
		// If the file doesn't exist, it's okay to return false
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		// For any other kind of error, return it
		return false, fmt.Errorf("error accessing file: %w", err)
	}

	// If the function reached this point, it means the file exists and is readable
	return true, nil
}

// SetActiveKey sets the active key to the alias specified.
func (k *KeyOps) SetActiveKey(aliasToActivate string) error {
	data, err := k.readWalletData(KeyFilePath)
	if err != nil {
		return err
	}

	// Check if the alias exists
	if _, exists := data.Wallets[aliasToActivate]; !exists {
		return fmt.Errorf("alias does not exist: %s", aliasToActivate)
	}

	// Set the alias to active
	data.ActiveAlias = aliasToActivate

	updatedData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Assuming you'll introduce FileWriter interface for writing to a file
	return k.FileWriter.WriteFile(KeyFilePath, updatedData)
}

// GetCurrentPublicKey retrieves the current active wallet's public key.
func (k *KeyOps) GetCurrentPublicKey() (string, error) {
	data, err := k.readWalletData(KeyFilePath)
	if err != nil {
		return "", err
	}

	activeWallet, exists := data.Wallets[data.ActiveAlias]
	if !exists {
		return "", ErrActiveWalletNotFound
	}

	return activeWallet.PublicKey, nil
}

// GetPublicKeyByAlias retrieves a wallet's public key by its alias.
func (k *KeyOps) GetPublicKeyByAlias(alias string) (string, error) {
	data, err := k.readWalletData(KeyFilePath)
	if err != nil {
		return "", err
	}

	wallet, exists := data.Wallets[alias]
	if !exists {
		return "", fmt.Errorf("no wallet found for alias: %s", alias)
	}

	return wallet.PublicKey, nil
}

func (k *KeyOps) WriteKeyToFile(alias string, key ed25519.PrivateKey, walletAddress string) error {
	var data WalletData
	fileExists, err := k.IsKeyFilePresent()

	if err != nil {
		return fmt.Errorf("error checking if keys are already present: %w", err)
	}

	if fileExists {
		data, err = k.readWalletData(KeyFilePath)
		if err != nil {
			return err
		}
	} else {
		data.Wallets = make(map[string]Wallet)
	}

	if _, exists := data.Wallets[alias]; exists {
		return fmt.Errorf("alias already exists: %s", alias)
	}

	data.Wallets[alias] = Wallet{Key: base58.Encode(key), Balance: decimal.Zero, PublicKey: walletAddress}
	data.ActiveAlias = alias // Setting the new key as active

	updatedData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Assuming FileWriter interface has been introduced for writing to a file
	return k.FileWriter.WriteFile(KeyFilePath, updatedData)
}

func (k *KeyOps) PrintAllKeys() ([]string, map[string]string, error) {
	data, err := k.readWalletData(KeyFilePath)
	if err != nil {
		return nil, nil, err
	}

	shouldPrintBalance := true
	rate, err := fetchSOLEURRate() // Assuming this function exists elsewhere in your code
	if err != nil {
		shouldPrintBalance = false
		return nil, nil, fmt.Errorf("error fetching SOL to EUR rate: %v", err)
	}

	aliases := make([]string, 0, len(data.Wallets))
	keyMap := make(map[string]string, len(data.Wallets))

	for alias, wallet := range data.Wallets {
		displayAlias := alias
		if alias == data.ActiveAlias {
			displayAlias += " (Active)"
		}

		if shouldPrintBalance {
			eurBalance := wallet.Balance.Mul(rate)
			displayAlias += fmt.Sprintf(" // BAL - (€ %s)", eurBalance.StringFixed(2))
		}

		aliases = append(aliases, displayAlias)
		keyMap[alias] = wallet.PublicKey
	}

	return aliases, keyMap, nil
}