package blockchain8

import (
	"bytes"
	"encoding/gob"
)

//TxOutput 交易的输出
type TxOutput struct {
	Value float64 //输出里面存储“币”

	//锁定输出的公钥（比特币里面是一个脚本，这里是公钥）
	PubKeyHash []byte
}

// Lock 对输出锁定，即反编码address后，获得实际的公钥哈希
func (out *TxOutput) Lock(address []byte) {
	expubKeyHash := Base58Decode(address)
	pubKeyHash := expubKeyHash[1 : len(expubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey 检查输出是否能够被公钥pubKeyHash拥有者使用
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTxOutput 创建一个新的 TXOutput
//注意，这里需要将address进行反编码成实际的地址
func NewTxOutput(value float64,address string) *TxOutput {
	txo := &TxOutput{value, nil} //构建TxOutput，PubKeyHash暂设为nil
	txo.Lock([]byte(address))    //接着设定TxOutput的PubKeyHash值

	return txo
}

// TxOutputs TxOutput集合
type TxOutputs struct {
	Outputs []TxOutput
}

// Serialize 序列化TxOutputs
func (outs TxOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	Handle(err)
	return buff.Bytes()
}

// DeserializeOutputs 反序列化TxOutputs
func DeserializeOutputs(data []byte) TxOutputs {
	var outputs TxOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	Handle(err)

	return outputs
}
