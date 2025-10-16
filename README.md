# BlockChain

A simple, educational blockchain implementation written in Go. This project demonstrates the core concepts of a blockchain, including blocks, hashing, Proof-of-Work, and data persistence.

This is a command-line application designed to be a clear and understandable example of how a blockchain works under the hood.

---

## ## Features

* **Immutable Ledger:** A chain of blocks linked cryptographically.
* **Proof-of-Work (PoW):** A consensus algorithm to secure the blockchain by requiring computational effort ("mining") to add new blocks.
* **Data Persistence:** Uses **BadgerDB** (a key-value store in Go) to save the blockchain's state, so your data persists between sessions.
* **Command-Line Interface (CLI):** Simple and intuitive commands to interact with the blockchain.
* **Modular Design:** The code is structured with a clear separation of concerns, making it easy to read, understand, and extend.

---

## ## Requirements

To run this project, you only need:
* **Go** (version 1.18 or newer is recommended).

All dependencies are managed by Go Modules.

---

## ## Installation & Setup

1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/YOUR_USERNAME/go-blockchain.git](https://github.com/YOUR_USERNAME/go-blockchain.git)
    ```

2.  **Navigate to the project directory:**
    ```bash
    cd go-blockchain
    ```

3.  **Install the dependencies:**
    Go will automatically download the necessary dependencies (like BadgerDB) when you build or run the project. You can also do it manually:
    ```bash
    go mod tidy
    ```

---

## ## Usage

This application is run from the command line and has two main commands.

### ### Adding a Block

To add a new block with some data to the chain, use the `addblock` command.

```bash
go run main.go addblock -data "Send 1 GO to "
```

**Output:**
The program will print "Mining a new block" and then the hash of the mined block, followed by a success message. The first time you run this, it will also create the Genesis block.

```
No existing blockchain found. Creating a new one...
Mining a new block
000000...
Mining a new block
000000...
Success! A new block has been added.
```

### ### Printing the Blockchain

To view all the blocks currently in your blockchain, use the `printchain` command. It will print the blocks from newest to oldest.

```bash
go run main.go printchain
```
**Example Output:**
```
Prev. hash: 247b...
Data: Send 5 more GO to 
Hash: 000000...
PoW: true

Prev. hash: 000000...
Data: Send 1 GO to 
Hash: 247b...
PoW: true

Prev. hash:
Data: Genesis Block
Hash: 000000...
PoW: true
```

---

## ## Project Structure

The project is organized to separate the core blockchain logic from the user interface.

```
go-blockchain/
├── blockchain/         # Contains all core blockchain logic
│   ├── block.go        # Defines the Block struct and serialization
│   ├── blockchain.go   # Manages the chain and database interaction
│   └── proofofwork.go  # Implements the Proof-of-Work algorithm
│
├── cli/                # Handles command-line interface logic
│   └── cli.go
│
├── tmp/                # Stores the BadgerDB database files (auto-generated)
│
├── go.mod              # Manages project dependencies
├── go.sum
└── main.go             # Entry point of the application
```
