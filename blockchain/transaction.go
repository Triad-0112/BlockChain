package blockchain

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/Triad-0112/BlockChain.git/utils"
	"github.com/Triad-0112/BlockChain.git/wallet"
)

type TXOutput struct {
	Value        int
	ScriptPubKey []byte
}

type TXInput struct {
	Txid   []byte
	Vout   int
	PubKey []byte
}

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randData)
	}

	txin := TXInput{[]byte{}, -1, []byte(data)}
	txout := TXOutput{100, nil}
	txout.Lock([]byte(to))
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}

func (in *TXInput) CanUnlockOutputWith(pubKeyHash []byte) bool {
	lockingHash := wallet.HashPubKey(in.PubKey)
	return bytes.Equal(lockingHash, pubKeyHash)
}

func (out *TXOutput) CanBeUnlockedWith(pubKeyHash []byte) bool {
	return bytes.Equal(out.ScriptPubKey, pubKeyHash)
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) (*Transaction, error) {
	var inputs []TXInput
	var outputs []TXOutput

	wallets, err := wallet.NewWallets()
	if err != nil {
		return nil, err
	}
	w := wallets.GetWallet(from)
	pubKeyHash := wallet.HashPubKey(w.PublicKey)

	acc, validOutputs := bc.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			return nil, err
		}
		for _, out := range outs {
			input := TXInput{txID, out, w.PublicKey}
			inputs = append(inputs, input)
		}
	}
	out := TXOutput{amount, nil}
	out.Lock([]byte(to))
	outputs = append(outputs, out)
	if acc > amount {
		changeOut := TXOutput{acc - amount, nil}
		changeOut.Lock([]byte(from))
		outputs = append(outputs, changeOut)
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx, nil
}

func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := utils.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.ScriptPubKey = pubKeyHash
}

func (tx *Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       PublicKey: %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.ScriptPubKey))
	}

	return strings.Join(lines, "\n")
}
