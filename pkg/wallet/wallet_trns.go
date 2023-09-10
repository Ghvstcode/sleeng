package wallet

import (
	"context"
	"encoding/binary"
	"fmt"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"sync"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

const (
	transferInstructionType uint32 = 2
	rpcTimeout                     = 10 * time.Second // 10 seconds
	maxConcurrentRequests          = 50
	//systemProgramIDStr represents the system program ID for the solana chain which tells us more about the nature of instruction.
	systemProgramIDStr = "11111111111111111111111111111111"
)

// Transaction represents a single transaction.
type Transaction struct {
	Amount    uint64
	From      solana.PublicKey
	To        solana.PublicKey
	Timestamp time.Time
	IsSender  bool
}

// decodeSystemTransfer decodes a system transfer instruction from a transaction.
func decodeSystemTransfer(tx *solana.Transaction, timestamp time.Time, publicKey string) ([]*Transaction, error) {
	systemProgramID := solana.MustPublicKeyFromBase58(systemProgramIDStr)
	var transactions []*Transaction

	for _, instruction := range tx.Message.Instructions {
		progKey, err := tx.ResolveProgramIDIndex(instruction.ProgramIDIndex)
		if err != nil {
			return nil, fmt.Errorf("resolve program ID index: %w", err)
		}

		if !progKey.Equals(systemProgramID) {
			continue
		}

		instructionType := binary.LittleEndian.Uint32(instruction.Data[0:4])
		if instructionType != transferInstructionType {
			continue
		}

		sender := tx.Message.AccountKeys[instruction.Accounts[0]]
		receiver := tx.Message.AccountKeys[instruction.Accounts[1]]
		amount := binary.LittleEndian.Uint64(instruction.Data[4:12])

		transactions = append(transactions, &Transaction{
			Amount:    amount,
			From:      sender,
			To:        receiver,
			Timestamp: timestamp,
			IsSender:  sender.String() == publicKey,
		})
	}

	return transactions, nil
}

// fetchSingleTransaction fetches a single transaction for the given signature.
func fetchSingleTransaction(client *rpc.Client, signature solana.Signature, publicKey string) ([]*Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
	defer cancel()

	txResponse, err := client.GetTransaction(ctx, signature, &rpc.GetTransactionOpts{Encoding: solana.EncodingBase64})
	if err != nil {
		return nil, fmt.Errorf("get transaction: %w", err)
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(txResponse.Transaction.GetBinary()))
	if err != nil {
		return nil, fmt.Errorf("transaction from decoder: %w", err)
	}

	blockTime, err := client.GetBlockTime(ctx, txResponse.Slot)
	if err != nil {
		return nil, fmt.Errorf("get block time: %w", err)
	}

	return decodeSystemTransfer(tx, blockTime.Time(), publicKey)
}

// fetchTransactions fetches all transactions for the given public key.
// It First fetches all signatures for the given public key
// and then fetches each transaction for each signature.
func fetchTransactions(publicKey string) ([]*Transaction, error) {
	client := rpc.New(rpc.DevNet_RPC)
	pub, err := solana.PublicKeyFromBase58(publicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
	defer cancel()

	signatures, err := client.GetSignaturesForAddress(ctx, pub)
	if err != nil {
		return nil, fmt.Errorf("get signatures for address: %w", err)
	}

	var transactions []*Transaction
	transactionsMutex := &sync.Mutex{}
	sem := semaphore.NewWeighted(maxConcurrentRequests)

	eg, ctx := errgroup.WithContext(ctx)

	for _, sig := range signatures {
		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, fmt.Errorf("failed to acquire semaphore: %w", err)
		}

		sig := sig // pin

		eg.Go(func() error {
			defer sem.Release(1)

			txList, err := fetchSingleTransaction(client, sig.Signature, publicKey)
			if err != nil {
				return fmt.Errorf("fetching transaction failed for signature %s: %w", sig.Signature, err)
			}

			transactionsMutex.Lock()
			defer transactionsMutex.Unlock()

			transactions = append(transactions, txList...)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return transactions, nil
}
