

# Development Notes
* I Decided to make the wallet both a **file-based wallet** and a **_paper-based wallet_**. The user can specify which they want to use by using the -p flag.
  * If a user decides to use the paper-based wallet, the seed phrase is displayed to the user and the user is prompted to write it down.
  * If a user decides to use the file-based wallet, the private key is saved to a file in the users local machine.
  * The user can specify a private key using the -k flag.
* I decided to make it a multi-wallet CLI, so the user can create multiple wallets and manage them.
  * The first wallet created is the default wallet, and the user can specify which wallet they want to use by using the -a flag.
  * The decision to make it a multi-wallet CLI affected the structure of the JSON file where the private keys are stored. Regardless I wanted the files to be sort of solana CLI compatible, so a user can copy the private key from the file and use it with the solana CLI or any other solana wallet.
  * The JSON file looks like this "{"activeAlias":"tester","wallets":{"tester":{"PrivateKey":"[92,15,204,133,134,51,239,138,85,241,239,241,4,222,128,85,174,71,118,164,209,211,36,177,104,134,209,129,117,109,239,73,218,29,176,154,4,216,255,8,183,212,169,166,234,195,223,93,32,19,59,140,197,76,42,59,114,79,212,204,90,38,195,171]","balance":"0","publicKey":"FgS8tPasZJW7TkwxpHdj5UeSrYrCT6mSw9jTx5aY8CNv"}}}"
  * With fields for the active alias, the wallets, and the wallet fields.
  * The active alias is the alias of the wallet that is currently being used.
  * The wallets field is a map of all the wallets created by the user.
  * The wallet fields are the private key, balance, and public key of the wallet.
  * Note that the balance field although it exists is barely useful, I did not get around to implementing updating it and all that.
* I did not get around to implementing logic around rate limiting request made by this app both to the Kraken API for exchange rate data and the Solana API for transaction history.
* A very interesting bit of this project was the retrieveing of transaction history from the Solana API. To get a single transaction we make about 3 RPC calls, 1 to get the signatures of all the public key, then we have all the transaction signatures we iterate over the array & make another API call to get confirmed transaction details for the signature, then another one to get the block time of the slot data. Expensive stuff. I make use of concurrency to try to imrpove the processs aloowing up to 50 transactions to be fetched concurrently once we get the signature.
* I only got around a 33.4% test coverage!
* Possible Improvements - (Better handling of transaction related errors, )
* 
# Solana Wallet CLI

## Overview

This command-line interface (CLI) is designed to interact with Solana wallets. The CLI allows you to perform a diverse set of tasks such as creating new wallets, checking balances, fetching exchange rates, sending funds, and viewing transaction history.

## Table of Contents

- [Commands](#commands)
    - [Root](#root)
    - [Initialize Wallet](#initialize-wallet)
    - [Send Funds](#send-funds)
    - [Transaction History](#transaction-history)
    - [Get Wallet Address](#get-wallet-address)
    - [Get Wallet Balance](#get-wallet-balance)
    - [Get Exchange Rate](#get-exchange-rate)
- [Options](#options)
    - [Persistent Flags](#persistent-flags)

---

## Commands

### Root

The root command initializes the CLI application.

Usage:
```bash
wallet
```
This command accepts persistent flags for specifying a base58 encoded private key and an optional alias for the wallet.

---

### Initialize Wallet

The `init` command creates a new Solana wallet and saves its private key to disk.

Usage:
```bash
wallet init
```
Flags:
- `--paper` or `-p`: Creates a paper-based wallet and displays the seed phrase.

> Note: The wallet address is copied to your clipboard after successful initialization.

---

### Send Funds

The `send` command allows you to send funds to another Solana wallet.

Usage:
```bash
wallet send [EUR amount] [destination]
```
Arguments:
- `EUR amount`: The amount of money you wish to send in EUR.
- `destination`: The destination Solana wallet address.

Upon successfully sending funds, a transaction signature will be displayed.

---

### Transaction History

The `transactions` command displays your transaction history, sorted by most recent transactions first.

Usage:
```bash
wallet transactions
```

> Note: If you have no transactions, "No transactions to display" will be shown.

---

### Get Wallet Address

The `address` command retrieves your Solana wallet address.

Usage:
```bash
wallet address
```

This command outputs the wallet address tied to the private key you've initialized or specified.

---

### Get Wallet Balance

The `balance` command provides the current balance of your Solana wallet in SOL and EUR.

Usage:
```bash
wallet balance
```

The balance is displayed in both SOL and its equivalent in EUR based on the current exchange rate.

---

### Get Exchange Rate

The `rate` command fetches the current exchange rate between SOL and EUR.

Usage:
```bash
wallet rate
```

The output shows the exchange rate at the moment of fetching and may include a timestamp.

---

## Options

### Persistent Flags

The following flags can be used with any command:

- `--key` or `-k`: A base58 encoded private key.
- `--alias` or `-a`: An optional alias for easier wallet management.

> Example: `wallet --key=<your_base58_encoded_key> --alias=<wallet_alias>`

---