// Package wallet provides functionalities to manage wallet keys.
package wallet

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mr-tron/base58/base58"
	"github.com/shopspring/decimal"
	"os"
	"strconv"
	"strings"
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
	ret, err := getPrivateKeyFromSolCLICompStr(activeWallet.PrivateKey)
	if err != nil {
		return "", err
	}

	return base58.Encode(ret), nil
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

	return wallet.PrivateKey, nil
}

// IsKeyFilePresent checks if there is a file containing some keys already in place.
func (k *KeyOps) IsKeyFilePresent() (bool, error) {
	_, err := k.FileReader.ReadFile(KeyFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("error accessing file: %w", err)
	}

	return true, nil
}

// SetActiveKey sets the active key to the alias specified.
func (k *KeyOps) SetActiveKey(aliasToActivate string) error {
	data, err := k.readWalletData(KeyFilePath)
	if err != nil {
		return err
	}

	if _, exists := data.Wallets[aliasToActivate]; !exists {
		return fmt.Errorf("alias does not exist: %s", aliasToActivate)
	}

	data.ActiveAlias = aliasToActivate

	updatedData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

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

// WriteKeyToFile writes a new key to the key file.
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

	solanaCliCompatiblekey := getSolCLIComptKey(key)
	data.Wallets[alias] = Wallet{PrivateKey: solanaCliCompatiblekey, Balance: decimal.Zero, PublicKey: walletAddress}
	data.ActiveAlias = alias

	updatedData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	return k.FileWriter.WriteFile(KeyFilePath, updatedData)
}

// PrintAllKeys prints all keys in the key file.
func (k *KeyOps) PrintAllKeys() ([]string, map[string]string, error) {
	data, err := k.readWalletData(KeyFilePath)
	if err != nil {
		return nil, nil, err
	}

	shouldPrintBalance := true
	rate, err := fetchSOLEURRate()
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

// getSolCLIComptKey converts a private key to a Solana CLI compatible string.
func getSolCLIComptKey(key ed25519.PrivateKey) string {
	intArr := make([]int, 0, len(key))

	for _, j := range key {
		intArr = append(intArr, int(j))
	}

	var builder strings.Builder
	builder.WriteString("[")

	for i, n := range intArr {
		builder.WriteString(strconv.Itoa(n))
		if i < len(intArr)-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString("]")

	return builder.String()
}

// getPrivateKeyFromSolCLICompStr converts a Solana CLI compatible string to a private key.
func getPrivateKeyFromSolCLICompStr(strKey string) (ed25519.PrivateKey, error) {
	strKey = strings.TrimPrefix(strKey, "[")
	strKey = strings.TrimSuffix(strKey, "]")

	strArr := strings.Split(strKey, ",")

	byteArr := make([]byte, len(strArr))

	for i, s := range strArr {
		num, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		byteArr[i] = byte(num)
	}

	return ed25519.PrivateKey(byteArr), nil
}
