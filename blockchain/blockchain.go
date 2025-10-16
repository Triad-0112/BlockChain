// In blockchain/blockchain.go

package blockchain

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger/v3"
)

const (
	dbPath        = "./tmp/blocks"
	dbLastHashKey = "lh" // Key to store the hash of the last block
)

// Blockchain keeps a sequence of Blocks.
type Blockchain struct {
	lastHash []byte
	db       *badger.DB
}

// NewBlockchain creates a new Blockchain with a genesis Block if one doesn't exist.
func NewBlockchain() *Blockchain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	// Disabling the logger to keep the console clean
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	// The Update function gives us a read-write transaction
	err = db.Update(func(txn *badger.Txn) error {
		// Check if a blockchain already exists by looking for the last hash key
		if _, err := txn.Get([]byte(dbLastHashKey)); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found. Creating a new one...")
			genesis := NewGenesisBlock()
			err = txn.Set(genesis.Hash, genesis.Serialize())
			if err != nil {
				return err
			}
			err = txn.Set([]byte(dbLastHashKey), genesis.Hash)
			lastHash = genesis.Hash
			return err
		} else {
			// An existing blockchain was found
			item, err := txn.Get([]byte(dbLastHashKey))
			if err != nil {
				return err
			}
			err = item.Value(func(val []byte) error {
				lastHash = val
				return nil
			})
			return err
		}
	})

	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{lastHash, db}
}

// AddBlock adds a new block to the blockchain.
func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	// View gives a read-only transaction
	err := bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(dbLastHashKey))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		return err
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(data, lastHash)

	// Update gives a read-write transaction
	err = bc.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}
		err = txn.Set([]byte(dbLastHashKey), newBlock.Hash)
		bc.lastHash = newBlock.Hash
		return err
	})
	if err != nil {
		log.Panic(err)
	}
}

// Iterator returns a BlockchainIterator.
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.lastHash, bc.db}
}

// BlockchainIterator is used to iterate over blockchain blocks.
type BlockchainIterator struct {
	currentHash []byte
	db          *badger.DB
}

// Next returns the next block from the iterator (iterating backwards from the newest).
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(i.currentHash)
		if err != nil {
			return err
		}
		var encodedBlock []byte
		err = item.Value(func(val []byte) error {
			encodedBlock = val
			return nil
		})
		if err != nil {
			return err
		}
		block = DeserializeBlock(encodedBlock)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash
	return block
}

// CloseDB closes the database connection.
func (bc *Blockchain) CloseDB() {
	bc.db.Close()
}
