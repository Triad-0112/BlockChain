# GoChain ⛓️

A functional cryptocurrency model built from scratch in Go. This project is an educational deep dive into the core mechanics of a blockchain, demonstrating how wallets, addresses, transactions, and mining work together to create a decentralized ledger.

This command-line application simulates a complete cryptocurrency workflow, from creating the first block (genesis) to sending coins between user-generated wallets.

---

## ## Features

* **Wallet Generation:** Creates and manages wallets with ECDSA public/private key pairs.
* **Base58 Addresses:** Generates human-readable, checksummed public addresses from wallets, similar to Bitcoin.
* **UTXO Transaction Model:** Tracks coin ownership through Unspent Transaction Outputs, just like Bitcoin. No traditional accounts are used.
* **Proof-of-Work (PoW):** Secures the blockchain by requiring computational effort ("mining") to add new blocks of transactions.
* **Data Persistence:** Uses **BadgerDB** to save the blockchain's state, ensuring all data persists between sessions.
* **Command-Line Interface (CLI):** A complete set of commands to create wallets, manage the blockchain, check balances, and send coins.

---

## ## Requirements

* **Go** (version 1.18 or newer is recommended).

All other dependencies are managed by Go Modules.

---

## ## Quick Start & Workflow

Here’s a complete workflow demonstrating how to use GoChain from start to finish.

### ### 1. Create Wallets

First, you need a sender and a receiver. Let's create two wallets.

```bash
# Create the first wallet
go run main.go createwallet
# Output: Your new address: 1Abc... (Copy this address)

# Create the second wallet
go run main.go createwallet
# Output: Your new address: 1Xyz... (Copy this address too)
```

### ### 2. Create the Blockchain

Initialize the blockchain. The very first block (the "genesis block") will contain a coinbase transaction that sends a mining reward to the address you specify.

```bash
# Use your first address to receive the reward
go run main.go createblockchain -address 1Abc...
```

### ### 3. Check the Initial Balance

Verify that the first wallet received the mining reward.

```bash
go run main.go getbalance -address 1Abc...
# Expected Output: Balance of '1Abc...': 100
```

### ### 4. Send Coins

Now, send some coins from the first wallet to the second. This will create a new transaction, mine a new block to include it, and add it to the chain.

```bash
go run main.go send -from 1Abc... -to 1Xyz... -amount 25
```

### ### 5. Verify the Final Balances

Check the balances again to confirm the transaction was successful.

```bash
# Check the sender's new balance (100 - 25 = 75)
go run main.go getbalance -address 1Abc...
# Expected Output: Balance of '1Abc...': 75

# Check the receiver's new balance
go run main.go getbalance -address 1Xyz...
# Expected Output: Balance of '1Xyz...': 25
```

---

## ## All Commands

* `createwallet`
    * Generates a new wallet and saves it to `wallets.dat`.

* `listaddresses`
    * Lists all addresses stored in the wallet file.

* `createblockchain -address <ADDRESS>`
    * Creates a new blockchain and sends the genesis block reward to the specified address.

* `getbalance -address <ADDRESS>`
    * Gets the balance of an address by summing its unspent transaction outputs (UTXOs).

* `send -from <SENDER> -to <RECEIVER> -amount <AMOUNT>`
    * Creates a transaction, mines a new block, and sends coins.

* `printchain`
    * Prints all the blocks and the transactions within them.

---

## ## Project Structure

```
go-blockchain/
├── blockchain/         # Core blockchain logic (blocks, PoW, transactions)
├── cli/                # Command-line interface logic
├── utils/              # Helper functions like Base58 encoding
├── wallet/             # Wallet and address generation logic
├── tmp/                # Stores the BadgerDB database files (auto-generated)
├── wallets.dat         # Stores wallet data (auto-generated)
└── ...
```

---

## ## License

This project is licensed under the MIT License.

```
MIT License

Copyright (c) 2025 Panji Tri Wahyudi

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction...
```