package blockchain8

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"sd-chain/blockchain8/wallet"








	log "github.com/sirupsen/logrus"
)

const subsidy = 10.0 //挖矿奖励

type Transaction struct {
	ID   []byte     //交易ID
	Vin  []TxInput  //交易输入，由上次交易输入（可能多个）
	Vout []TxOutput //交易输出，由本次交易产生（可能多个）
}

func (tx *Transaction) Serializer() []byte {
	var encoded bytes.Buffer

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	Handle(err)

	return encoded.Bytes()
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serializer())
	return hash[:]
}

//IsCoinbase 检查交易是否是创始区块交易
//创始区块交易没有输入，详细见NewCoinbaseTX
//tx.Vin只有一个输入，数组长度为1
//tx.Vin[0].Txid为[]byte{}，因此长度为0
//Vin[0].Vout设置为-1
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

//NewCoinbaseTX 创建一个区块链创始交易，不需要签名
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("奖励给%s", to) //fmt.Sprintf将数据格式化后赋值给变量data
	}

	//初始交易输入结构：引用输出的交易为空:引用交易的ID为空，交易引用的输出值为设为-1
	txin := TxInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTxOutput(subsidy, to)                           //本次交易的输出结构：奖励值为subsidy，奖励给地址to（当然也只有地址to可以解锁使用这笔钱）
	tx := Transaction{nil, []TxInput{txin}, []TxOutput{*txout}} //交易ID设为nil
	tx.ID = tx.Hash()

	return &tx
}
func (tx *Transaction) IsMinerTx() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {

	if tx.IsMinerTx() {
		return
	}

	for _, in := range tx.Vin {
		if prevTXs[hex.EncodeToString(in.Txid)].ID == nil {
			log.Fatal("ERROR: Previous Transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inId, in := range txCopy.Vin {
		prevTX := prevTXs[hex.EncodeToString(in.Txid)]
		txCopy.Vin[inId].Signature = nil
		//look for the transaction output that produced this input, then sign it with
		// the rest of the data
		txCopy.Vin[inId].PubKey = prevTX.Vout[in.Vout].PubKeyHash
		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		Handle(err)
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inId].Signature = signature
		txCopy.Vin[inId].PubKey = nil
	}
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, in := range tx.Vin {
		inputs = append(inputs, TxInput{in.Txid, in.Vout, nil, nil})
	}

	for _, out := range tx.Vout {
		outputs = append(outputs, TxOutput{out.Value, out.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

// Create new Transaction
func NewTransaction(w *wallet.Wallet, to string, amount float64, utxo *UTXOSet) (*Transaction, error) {
	var inputs []TxInput
	var outputs []TxOutput

	publicKeyHash := wallet.PublicKeyHash(w.PublicKey)

	acc, validoutputs := utxo.FindSpendableOutputs(publicKeyHash, amount)
	if acc < amount {
		err := errors.New("You dont have Enough Amount...")
		return nil, err
	}

	from := fmt.Sprintf("%s", w.Address())

	for txId, outs := range validoutputs {
		txID, err := hex.DecodeString(txId)

		Handle(err)
		for _, out := range outs {
			input := TxInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	var txo=NewTxOutput(amount, to)
	outputs = append(outputs, *txo)
	if acc > amount {
		txo=NewTxOutput(acc-amount, from)
		outputs = append(outputs, *txo)
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()

	// Sign the new transaction with wallet Private Key
	utxo.Blockchain.SignTransaction(&tx,w.PrivateKey)

	return &tx, nil
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsMinerTx() {
		return true
	}

	for _, in := range tx.Vin {
		if prevTXs[hex.EncodeToString(in.Txid)].ID == nil {
			log.Fatal("ERROR: Previous Transaction is not valid")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()
	for inId, in := range tx.Vin {
		prevTX := prevTXs[hex.EncodeToString(in.Txid)]
		txCopy.Vin[inId].Signature = nil
		txCopy.Vin[inId].PubKey = prevTX.Vout[in.Vout].PubKeyHash

		r := big.Int{}
		s := big.Int{}
		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false
		}
		txCopy.Vin[inId].PubKey = nil
	}

	return true
}

// Helper function for displaying transaction data in the console
func (tx *Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("---Transaction: %x", tx.ID))

	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("	Input (%d):", i))
		lines = append(lines, fmt.Sprintf(" 	 	TXID: %x", input.Txid))
		lines = append(lines, fmt.Sprintf("		Out: %d", input.Vout))
		lines = append(lines, fmt.Sprintf(" 	 	Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("		PubKey: %x", input.PubKey))
	}

	for i, out := range tx.Vout {
		lines = append(lines, fmt.Sprintf("	Output (%d):", i))
		lines = append(lines, fmt.Sprintf(" 	 	Value: %f", out.Value))
		lines = append(lines, fmt.Sprintf("		PubkeyHash: %x", out.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

// Miner Transaction with Input && Output credited with 20.000 token for the workdone
// No Signature is required for the miner transaction Input
func MinerTx(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 24)
		_, err := rand.Read(randData)
		Handle(err)
		data = fmt.Sprintf("%x", randData)
	}

	txIn := TxInput{[]byte{}, -1, nil, []byte(data)}
	txOut := NewTxOutput(20.000, to)

	tx := Transaction{nil, []TxInput{txIn}, []TxOutput{*txOut}}

	tx.ID = tx.Hash()

	return &tx
}
