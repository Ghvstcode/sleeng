package wallet

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/shopspring/decimal"
	"log"
)

const LamportsInOneSol = 1000000000 // Lamports in one SOL

type ClientInterface interface {
	GetBalance(ctx context.Context, publicKey solana.PublicKey, commitment rpc.CommitmentType) (*rpc.GetBalanceResult, error)
}

var rpcClient ClientInterface = rpc.New(rpc.DevNet_RPC) // Create a global RPC client

func (w *WalletConfig) fetchSolBalance(alias string, keyStore KeyStore) (decimal.Decimal, error) {
	var publicKey solana.PublicKey
	var err error

	if w.Wallet != nil {
		publicKey = w.Wallet.PublicKey()
	} else if alias != "" {
		publicKey, err = fetchPublicKeyByAlias(alias, keyStore)
	} else {
		publicKey, err = fetchCurrentPublicKey(keyStore)
	}

	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("failed to fetch public key: %w", err)
	}

	balanceResp, err := rpcClient.GetBalance(context.TODO(), publicKey, rpc.CommitmentFinalized)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("failed to fetch balance: %w", err)
	}

	lamportValue := decimal.NewFromInt(int64(balanceResp.Value))
	fin := lamportValue.Div(decimal.NewFromInt(LamportsInOneSol))
	log.Print("BALANCE", fin)
	// Convert lamports to SOL
	return fin, nil
}

func fetchPublicKeyByAlias(alias string, keyStore KeyStore) (solana.PublicKey, error) {
	privateKey, err := keyStore.GetPrivateKeyByAlias(alias)
	if err != nil {
		return solana.PublicKey{}, err
	}

	return fetchPublicKeyFromPrivateKey(privateKey)
}

func fetchCurrentPublicKey(keyStore KeyStore) (solana.PublicKey, error) {
	privateKey, err := keyStore.GetCurrentPrivateKey()
	if err != nil {
		return solana.PublicKey{}, err
	}

	return fetchPublicKeyFromPrivateKey(privateKey)
}

// fetchPublicKeyFromPrivateKey fetches the public key given a private key.
func fetchPublicKeyFromPrivateKey(privateKey string) (solana.PublicKey, error) {
	account, err := solana.WalletFromPrivateKeyBase58(privateKey)
	if err != nil {
		return solana.PublicKey{}, err
	}
	return account.PublicKey(), nil
}
