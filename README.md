# GoChain ⛓️ - An Educational Blockchain

A functional cryptocurrency model built from scratch in Go, repo at [https://github.com/Triad-0112/BlockChain](https://github.com/Triad-0112/BlockChain). This project is a deep dive into the core mechanics of a blockchain, demonstrating how wallets, addresses, transactions, Proof-of-Work, and dynamic difficulty adjustment work together.

This command-line application simulates a complete cryptocurrency workflow, from creating a genesis block to sending coins between wallets. It's an excellent learning tool for understanding the complexities of blockchain technology.

**Note:** This project was built as an educational exercise and contains known bugs that highlight common challenges in blockchain development.

---

## Features

* **Wallet Generation:** Creates and manages wallets with ECDSA public/private key pairs.
* **Base58 Addresses:** Generates human-readable, checksummed public addresses, similar to Bitcoin.
* **UTXO Transaction Model:** Tracks coin ownership through Unspent Transaction Outputs.
* **Dynamic Difficulty Adjustment:** The Proof-of-Work difficulty automatically adjusts to maintain a target block time, simulating one of Bitcoin's core features. 
* **Data Persistence:** Uses **BadgerDB** to save the blockchain's state.
* **CLI Block Explorer:** A `printchain` command that displays detailed information for every block and transaction.

---

## Known Issues & Bugs 🐞

This project contains a critical bug related to balance calculation that serves as an excellent case study for debugging UTXO-based systems.

### 1. Incorrect Sender Balance After Transaction

When a user with multiple unspent outputs (e.g., from mining several blocks) sends coins, the sender's final balance is calculated incorrectly.

* **Symptom:** A user with a balance of `400` sends `75` coins. Their expected new balance is `325`. The program incorrectly reports a balance of `125`.
* **Root Cause:** The `FindUTXO` function, which is responsible for calculating the balance, contains flawed accounting logic. When it scans the blockchain, it correctly identifies that an output has been spent. However, it then fails to correctly gather all of the *other remaining unspent outputs* that belong to the sender, leading to an inaccurate total.

### 2. Duplicate Coinbase Transaction IDs

If a user mines multiple blocks in quick succession, the coinbase transactions (which grant the mining reward) can end up with the **exact same transaction ID**.

* **Symptom:** The `printchain` command shows that several different blocks contain a coinbase transaction with an identical ID hash.
* **Root Cause:** A transaction ID is a hash of its contents. If the miner's reward address is the same and the blocks are mined so quickly that the timestamp doesn't change, the resulting hash is identical. This breaks the UTXO model, because when an output from one of these transactions is spent, the system incorrectly flags the identical outputs in other blocks as spent too. This is a primary contributor to the incorrect balance calculation bug.

---

## How to Use

1.  **Clone & Tidy:**
    ```bash
    git clone [https://github.com/Triad-0112/BlockChain.git](https://github.com/Triad-0112/BlockChain.git)
    cd BlockChain
    go mod tidy
    ```

2.  **Create Wallets:**
    ```bash
    go run main.go createwallet
    ```

3.  **Create the Blockchain:**
    ```bash
    go run main.go createblockchain -address <YOUR_ADDRESS>
    ```

4.  **Mine, Send, and Check Balances:**
    ```bash
    go run main.go mine -address <YOUR_ADDRESS>
    go run main.go getbalance -address <YOUR_ADDRESS>
    go run main.go send -from <SENDER> -to <RECEIVER> -amount <AMOUNT>
    go run main.go printchain
    ```