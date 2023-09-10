package wallet

import (
	"context"
	"errors"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

type MockClientInterface struct {
	GetBalanceFn func(ctx context.Context, publicKey solana.PublicKey, commitment rpc.CommitmentType) (*rpc.GetBalanceResult, error)
}

func (m *MockClientInterface) GetBalance(ctx context.Context, publicKey solana.PublicKey, commitment rpc.CommitmentType) (*rpc.GetBalanceResult, error) {
	return m.GetBalanceFn(ctx, publicKey, commitment)
}

type MockKeyStore struct {
	GetCurrentPrivateKeyFn func() (string, error)
	GetPrivateKeyByAliasFn func(string) (string, error)
	KeyStore
}

func (m *MockKeyStore) GetCurrentPrivateKey() (string, error) {
	return m.GetCurrentPrivateKeyFn()
}

func (m *MockKeyStore) GetPrivateKeyByAlias(alias string) (string, error) {
	return m.GetPrivateKeyByAliasFn(alias)
}

func TestFetchSolBalance(t *testing.T) {
	mockWallet := solana.NewWallet()

	tests := []struct {
		name          string
		walletConfig  *WalletConfig
		alias         string
		mockResponse  *rpc.GetBalanceResult
		mockError     error
		expectedError string
		expectedValue decimal.Decimal
	}{
		{
			name: "Success with wallet",
			walletConfig: &WalletConfig{
				Wallet: mockWallet,
			},
			alias: "",
			mockResponse: &rpc.GetBalanceResult{
				Value: 5000000000,
			},
			mockError:     nil,
			expectedError: "",
			expectedValue: decimal.NewFromInt(5).Truncate(2),
		},
		{
			name: "Success with alias",
			walletConfig: &WalletConfig{
				Wallet: nil,
			},
			alias: "validAlias",
			mockResponse: &rpc.GetBalanceResult{
				Value: 4000000000,
			},
			mockError:     nil,
			expectedError: "",
			expectedValue: decimal.NewFromInt(4),
		},
		{
			name: "Success without alias",
			walletConfig: &WalletConfig{
				Wallet: nil,
			},
			alias: "",
			mockResponse: &rpc.GetBalanceResult{
				Value: 3000000000,
			},
			mockError:     nil,
			expectedError: "",
			expectedValue: decimal.NewFromInt(3),
		},
		{
			name: "Failure due to RPC error",
			walletConfig: &WalletConfig{
				Wallet: nil,
			},
			alias:         "",
			mockResponse:  nil,
			mockError:     errors.New("RPC error"),
			expectedError: "failed to fetch balance: RPC error",
			expectedValue: decimal.NewFromInt(0),
		},
		{
			name: "Failure due to invalid alias",
			walletConfig: &WalletConfig{
				Wallet: nil,
			},
			alias:         "invalidAlias",
			mockResponse:  nil,
			mockError:     nil,
			expectedError: "failed to fetch public key: invalid alias",
			expectedValue: decimal.NewFromInt(0),
		},
		{
			name: "Failure due to GetCurrentPrivateKey error",
			walletConfig: &WalletConfig{
				Wallet: nil,
			},
			alias:         "",
			mockResponse:  nil,
			mockError:     nil,
			expectedError: "failed to fetch public key: GetCurrentPrivateKey error",
			expectedValue: decimal.NewFromInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock RPC client
			mockClient := &MockClientInterface{
				GetBalanceFn: func(ctx context.Context, publicKey solana.PublicKey, commitment rpc.CommitmentType) (*rpc.GetBalanceResult, error) {
					return tt.mockResponse, tt.mockError
				},
			}
			rpcClient = mockClient

			// Mock KeyStore
			mockKeyStore := &MockKeyStore{
				GetCurrentPrivateKeyFn: func() (string, error) {
					if tt.name == "Failure due to GetCurrentPrivateKey error" {
						return "", errors.New("GetCurrentPrivateKey error")
					}
					return "23YcmrXnN9C74zNP6pzkqfCqQKVTNk93rGu8C5fVyw4KPsXeQgqtC7YTPkx1vZJrg6mqYuEUgAFdoxXiU2UrBPZe", nil
				},
				GetPrivateKeyByAliasFn: func(alias string) (string, error) {
					if alias == "validAlias" {
						return "23YcmrXnN9C74zNP6pzkqfCqQKVTNk93rGu8C5fVyw4KPsXeQgqtC7YTPkx1vZJrg6mqYuEUgAFdoxXiU2UrBPZe", nil
					}
					return "", errors.New("invalid alias")
				},
			}

			// Execute the test
			got, err := tt.walletConfig.fetchSolBalance(tt.alias, mockKeyStore)
			log.Print("GOT", got)
			fmt.Println("GOT", got)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, fmt.Sprint(err))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue.String(), got.String())
			}
		})
	}
}
