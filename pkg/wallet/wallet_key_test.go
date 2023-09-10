package wallet

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type MockFileReader struct {
	mockFileData []byte
	mockError    error
}

func (m *MockFileReader) ReadFile(filename string) ([]byte, error) {
	return m.mockFileData, m.mockError
}

type MockFileWriter struct {
	mockWriteError error
}

func (m *MockFileWriter) WriteFile(filename string, data []byte) error {
	return m.mockWriteError
}
func TestGetCurrentPrivateKey(t *testing.T) {
	tests := []struct {
		name         string
		mockFileData WalletData
		mockError    error
		expectedKey  string
		expectedErr  error
	}{
		{
			name: "Success",
			mockFileData: WalletData{
				ActiveAlias: "active",
				Wallets: map[string]Wallet{
					"active": {Key: "somekey"},
				},
			},
			expectedKey: "somekey",
		},
		{
			name:        "File Read Error",
			mockError:   errors.New("read error"),
			expectedErr: errors.New("error reading file: read error"),
		},
		{
			name:         "No Active Wallet",
			mockFileData: WalletData{},
			expectedErr:  ErrActiveWalletNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFileReader := &MockFileReader{
				mockFileData: jsonMarshal(t, tt.mockFileData),
				mockError:    tt.mockError,
			}

			ops := &KeyOps{FileReader: mockFileReader}

			got, err := ops.GetCurrentPrivateKey()

			assert.Equal(t, tt.expectedKey, got)
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestGetPrivateKeyByAlias(t *testing.T) {
	tests := []struct {
		name         string
		alias        string
		mockFileData WalletData
		mockError    error
		expectedKey  string
		expectedErr  error
	}{
		{
			name:  "Success",
			alias: "active",
			mockFileData: WalletData{
				Wallets: map[string]Wallet{
					"active": {Key: "somekey"},
				},
			},
			expectedKey: "somekey",
		},
		{
			name:        "File Read Error",
			alias:       "active",
			mockError:   errors.New("read error"),
			expectedErr: errors.New("error reading file: read error"),
		},
		{
			name:  "Alias Not Found",
			alias: "missing",
			mockFileData: WalletData{
				Wallets: map[string]Wallet{
					"active": {Key: "somekey"},
				},
			},
			expectedErr: errors.New("no wallet found for alias: missing"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFileReader := &MockFileReader{
				mockFileData: jsonMarshal(t, tt.mockFileData),
				mockError:    tt.mockError,
			}

			ops := &KeyOps{FileReader: mockFileReader}

			got, err := ops.GetPrivateKeyByAlias(tt.alias)

			assert.Equal(t, tt.expectedKey, got)
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestKeysAlreadyPresent(t *testing.T) {
	tests := []struct {
		name         string
		mockFileData []byte // Not used but could be for completeness
		mockError    error
		expectedBool bool
		expectedErr  error
	}{
		{
			name:         "File Exists",
			mockFileData: []byte("some data"),
			expectedBool: true,
			expectedErr:  nil,
		},
		{
			name:         "File Does Not Exist",
			mockError:    os.ErrNotExist,
			expectedBool: false,
			expectedErr:  nil,
		},
		{
			name:         "File Read Error",
			mockError:    errors.New("read error"),
			expectedBool: false,
			expectedErr:  errors.New("error accessing file: read error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFileReader := &MockFileReader{
				mockFileData: tt.mockFileData,
				mockError:    tt.mockError,
			}

			ops := &KeyOps{FileReader: mockFileReader}

			got, err := ops.IsKeyFilePresent()

			assert.Equal(t, tt.expectedBool, got)
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestSetActiveKey(t *testing.T) {
	tests := []struct {
		name         string
		aliasToSet   string
		mockFileData WalletData
		mockError    error
		expectedErr  error
	}{
		{
			name:       "Success",
			aliasToSet: "active",
			mockFileData: WalletData{
				Wallets: map[string]Wallet{
					"active": {Key: "somekey"},
				},
			},
		},
		{
			name:        "File Read Error",
			aliasToSet:  "active",
			mockError:   errors.New("read error"),
			expectedErr: errors.New("error reading file: read error"),
		},
		{
			name:       "Alias Not Found",
			aliasToSet: "missing",
			mockFileData: WalletData{
				Wallets: map[string]Wallet{
					"active": {Key: "somekey"},
				},
			},
			expectedErr: errors.New("alias does not exist: missing"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFileReader := &MockFileReader{mockFileData: jsonMarshal(t, tt.mockFileData), mockError: tt.mockError}
			mockFileWriter := &MockFileWriter{}
			ops := &KeyOps{FileReader: mockFileReader, FileWriter: mockFileWriter}

			err := ops.SetActiveKey(tt.aliasToSet)
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestWriteKeyToFile(t *testing.T) {
	tests := []struct {
		name        string
		alias       string
		mockError   error
		fileExists  bool
		expectedErr error
	}{
		{
			name:        "Success",
			alias:       "newkey",
			mockError:   nil,
			fileExists:  true,
			expectedErr: nil,
		},
		{
			name:        "File Write Error",
			alias:       "newkey",
			mockError:   errors.New("error marshaling JSON: write error"),
			fileExists:  true,
			expectedErr: errors.New("error marshaling JSON: write error"),
		},
		{
			name:        "Alias Already Exists",
			alias:       "existing",
			mockError:   nil,
			fileExists:  true,
			expectedErr: errors.New("alias already exists: existing"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFileReader := &MockFileReader{}
			mockFileWriter := &MockFileWriter{mockWriteError: tt.mockError}

			ops := &KeyOps{
				FileReader: mockFileReader,
				FileWriter: mockFileWriter,
			}

			if tt.fileExists {
				mockFileReader.mockFileData = jsonMarshal(t, WalletData{
					Wallets: map[string]Wallet{
						"existing": {Key: "existingkey"},
					},
				})
			}

			key := ed25519.PrivateKey("23YcmrXnN9C74zNP6pzkqfCqQKVTNk93rGu8C5fVyw4KPsXeQgqtC7YTPkx1vZJrg6mqYuEUgAFdoxXiU2UrBPZe")
			err := ops.WriteKeyToFile(tt.alias, key, "walletAddress")

			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestPrintAllKeys(t *testing.T) {
	tests := []struct {
		name         string
		mockFileData WalletData
		mockError    error
		expectedErr  error
	}{
		{
			name: "Success",
			mockFileData: WalletData{
				ActiveAlias: "active",
				Wallets: map[string]Wallet{
					"active":   {Key: "activekey", Balance: decimal.NewFromInt(10)},
					"inactive": {Key: "inactivekey", Balance: decimal.NewFromInt(20)},
				},
			},
			expectedErr: nil,
		},
		{
			name:        "File Read Error",
			mockError:   errors.New("read error"),
			expectedErr: errors.New("error reading file: read error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFileReader := &MockFileReader{
				mockFileData: jsonMarshal(t, tt.mockFileData),
				mockError:    tt.mockError,
			}

			ops := &KeyOps{
				FileReader: mockFileReader,
			}

			// Mocking fetchSOLEURRate function should be done here, to return tt.fetchRateError

			_, _, err := ops.PrintAllKeys()

			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

// Helper function to marshal a WalletData instance into a JSON byte array.
// Panics on failure, which will cause the test to fail.
func jsonMarshal(t *testing.T, data WalletData) []byte {
	t.Helper()

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("could not marshal data: %v", err)
	}

	return b
}
