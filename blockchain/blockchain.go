package blockchain

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/Triad-0112/BlockChain.git/utils"
	"github.com/dgraph-io/badger/v3"
)

const (
	dbPath                       = "./tmp/blocks"
	dbLastHashKey                = "lh"
	genesisCoinbaseData          = "The Times 16/Oct/2025 Chancellor on brink of second bailout for banks"
	difficultyAdjustmentInterval = 5
	targetBlockTime              = 15
	startDifficulty              = 18
)

type Blockchain struct {
	lastHash []byte
	db       *badger.DB
}

func NewBlockchain(address string) *Blockchain {
	if DbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
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
			genesis.Difficulty = startDifficulty

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

func OpenBlockchain() *Blockchain {
	if !DbExists() {
		fmt.Println("No existing blockchain found. Create one first with 'createblockchain'.")
		os.Exit(1)
	}
	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(dbLastHashKey))
		if err != nil {
			return err
		}
		lastHash, err = item.ValueCopy(nil)
		return err
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
	if i.currentHash == nil {
		return nil
	}
	err := i.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(i.currentHash)
		if err == badger.ErrKeyNotFound {
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
	if err == badger.ErrKeyNotFound {
		return nil
	}
	if err != nil {
		log.Panic(err)
	}
	i.currentHash = block.PrevBlockHash
	return block
}

func (bc *Blockchain) CloseDB() {
	_ = bc.db.Close()
}

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	unspentTxs := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTxs {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0
Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(pubKeyHash) && accumulated < amount {
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

func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
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
				if out.CanBeUnlockedWith(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(pubKeyHash) {
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
		lastHash, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		log.Panic(err)
	}
	difficulty := bc.GetDifficulty()
	newBlock := NewBlock(transactions, lastHash, difficulty)
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

func (bc *Blockchain) getLatestBlock() *Block {
	var lastBlock *Block
	err := bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(dbLastHashKey))
		if err != nil {
			return err
		}
		lastHash, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		item, err = txn.Get(lastHash)
		if err != nil {
			return err
		}
		blockData, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		lastBlock = DeserializeBlock(blockData)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return lastBlock
}

func (bc *Blockchain) getBlock(blockHash []byte) (*Block, error) {
	var block *Block
	err := bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(blockHash)
		if err != nil {
			return err
		}
		blockData, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		block = DeserializeBlock(blockData)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (bc *Blockchain) getBlockHeight() int {
	var height int = 0
	bci := bc.Iterator()
	for {
		block := bci.Next()
		if block == nil {
			break
		}
		height++
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return height
}

func (bc *Blockchain) GetDifficulty() int {
	lastBlock := bc.getLatestBlock()
	if lastBlock == nil {
		return startDifficulty
	}
	if (bc.getBlockHeight())%difficultyAdjustmentInterval != 0 {
		return lastBlock.Difficulty
	}
	firstBlockOfInterval := lastBlock
	for i := 1; i < difficultyAdjustmentInterval; i++ {
		block, err := bc.getBlock(firstBlockOfInterval.PrevBlockHash)
		if err != nil {
			return startDifficulty
		}
		firstBlockOfInterval = block
	}
	actualTime := lastBlock.Timestamp - firstBlockOfInterval.Timestamp
	expectedTime := int64(difficultyAdjustmentInterval * targetBlockTime)
	if actualTime < expectedTime/2 {
		fmt.Println("Block time too fast, increasing difficulty")
		return lastBlock.Difficulty + 1
	} else if actualTime > expectedTime*2 {
		fmt.Println("Block time too slow, decreasing difficulty")
		if lastBlock.Difficulty > 1 {
			return lastBlock.Difficulty - 1
		}
	}
	return lastBlock.Difficulty
}

func DbExists() bool {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return false
	}
	return true
}
