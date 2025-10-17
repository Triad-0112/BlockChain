package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"log"
	"math/big"
	"os"
)

const walletFile = "wallets.dat"

type Wallets struct {
	Wallets map[string]*Wallet
}

type serializableWallet struct {
	PrivateKey []byte
	PublicKey  []byte
}

func NewWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile()
	return &wallets, err
}

func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := string(wallet.GetAddress())
	ws.Wallets[address] = wallet
	return address
}

func (ws *Wallets) GetAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var serializableWallets map[string]serializableWallet
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&serializableWallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = make(map[string]*Wallet)
	for address, sWallet := range serializableWallets {
		wallet := Wallet{}
		wallet.PublicKey = sWallet.PublicKey

		curve := elliptic.P256()
		privKey := new(big.Int)
		privKey.SetBytes(sWallet.PrivateKey)

		wallet.PrivateKey.D = privKey
		wallet.PrivateKey.Curve = curve

		ws.Wallets[address] = &wallet
	}

	return nil
}

func (ws *Wallets) SaveToFile() {
	var content bytes.Buffer

	serializableWallets := make(map[string]serializableWallet)
	for address, wallet := range ws.Wallets {
		serializableWallets[address] = serializableWallet{
			PrivateKey: wallet.PrivateKey.D.Bytes(),
			PublicKey:  wallet.PublicKey,
		}
	}

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(serializableWallets)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
