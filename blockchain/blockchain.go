package blockchain

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/dgraph-io/badger/v3"
)

const (
	dbPath              = "./tmp/blocks"
	dbLastHashKey       = "lh"
	genesisCoinbaseData = "The Times 16/Oct/2025 Chancellor on brink of second bailout for banks"
)

type Blockchain struct {
	lastHash []byte
	db       *badger.DB
}

func NewBlockchain(address string) *Blockchain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte(dbLastHashKey)); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found. Creating a new one...")
			cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
			genesis := NewGenesisBlock(cbtx)

			err = txn.Set(genesis.Hash, genesis.Serialize())
			if err != nil {
				return err
			}
			err = txn.Set([]byte(dbLastHashKey), genesis.Hash)
			lastHash = genesis.Hash
			return err
		} else {
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

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.lastHash, bc.db}
}

type BlockchainIterator struct {
	currentHash []byte
	db          *badger.DB
}

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

func (bc *Blockchain) CloseDB() {
	bc.db.Close()
}

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()
		if block == nil {
			break
		}

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output already spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				// Is this an output that the address can unlock?
				if out.CanBeUnlockedWith([]byte(address)) {
					UTXOs = append(UTXOs, out)
				}
			}

			// Collect all inputs in a block (to identify spent outputs)
			if !tx.IsCoinbase() { // Coinbase transactions don't spend anything
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith([]byte(address)) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOutputs
}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()
		if block == nil {
			break
		}

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlockedWith([]byte(address)) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTXs
}

func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

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

	newBlock := NewBlock(transactions, lastHash)

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
