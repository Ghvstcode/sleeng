package wallet

import (
	"github.com/gagliardetto/solana-go"
	"github.com/shopspring/decimal"
)

type WalletConfig struct {
	PrivateKey   string `json:"private_key"`
	Alias        string `json:"alias,omitempty"`
	IsPaperBased bool   `json:"is_paper_based,omitempty"`
	SeedPhrase   string `json:"seed_phrase,omitempty"`
	Wallet       *solana.Wallet
}

func (c *WalletConfig) GenerateNewPaperWallet() (interface{}, interface{}, interface{}) {
	// TODO: implement
}

func (c *WalletConfig) ImportWalletFromSeed(phrase string) (interface{}, interface{}) {
	// TODO: implement
}

func (c *WalletConfig) HasWallets() (interface{}, interface{}) {
	// TODO: implement
}

type Wallet struct {
	Key       string          `json:"key"`
	Balance   decimal.Decimal `json:"balance"`
	PublicKey string          `json:"publicKey"`
}

type WalletData struct {
	ActiveAlias string            `json:"activeAlias"`
	Wallets     map[string]Wallet `json:"wallets"`
}

// NewWalletConfig initializes a new WalletConfig.
func NewWalletConfig() *WalletConfig {
	return &WalletConfig{}
}
