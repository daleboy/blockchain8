package blockchain8

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

//Block 区块结构新版，增加了计数器nonce，主要目的是为了校验区块是否合法
//即挖出的区块是否满足工作量证明要求的条件
type Block struct {
	Timestamp    int64          `json:"Timestamp"`
	Hash         []byte         `json:"Hash"`
	PrevHash     []byte         `json:"PrevHash"`
	Transactions []*Transaction `json:"Transactions"`
	Nonce        int            `json:"Nonce"`
	Height       int            `json:"Height"`
	MerkleRoot   []byte         `json:"MerkleRoot"`
	Difficulty   int            `json:"Difficulty"`
	TxCount      int            `json:"TxCount"`
}

//NewBlock 创建普通区块
//一个block里面可以包含多个交易
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int) *Block {
	block := &Block{
		time.Now().Unix(),
		[]byte{},
		prevBlockHash,
		transactions,
		0,
		height,
		[]byte{},
		Difficulty,
		len(transactions),
	}
	//挖矿实质上是算出符合要求的哈希
	pow := NewProofOfWork(block) //注意传递block指针作为参数
	nonce, hash := pow.Run()

	//设置block的计数器和哈希
	block.Nonce = nonce
	block.Hash = hash[:]

	//设置 MerkleRoot
	block.MerkleRoot = block.HashTransactions()

	return block
}

// HashTransactions 计算交易组合的哈希值，最后得到的是Merkle tree的根节点
//获得每笔交易的哈希，将它们关联起来，然后获得一个连接后的组合哈希
//此方法只会被PoW使用
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serializer())
	}
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data //返回Merkle tree的根节点
}

//NewGenesisBlock 创建创始区块，包含创始交易。注意，创建创始区块也需要挖矿。
func NewGenesisBlock(coninbase *Transaction) *Block {
	return NewBlock([]*Transaction{coninbase}, []byte{}, 1)
}

//Serialize Block序列化
//特别注意，block对象的任何不以大写字母开头命令的变量，其值都不会被序列化到[]byte中
func (b *Block) Serialize() []byte {
	var result bytes.Buffer //定义一个buffer存储序列化后的数据

	//初始化一个encoder，gob是标准库的一部分
	//encoder根据参数的类型来创建，这里将编码为字节数组
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b) //编码
	Handle(err)              //错误处理

	return result.Bytes()
}

// DeserializeBlock 反序列化，注意返回的是Block的指针（引用）
func DeserializeBlock(d []byte) *Block {
	var block Block //一般都不会通过指针来创建一个struct。记住struct是一个值类型

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	Handle(err)

	return &block //返回block的引用
}

func (b *Block) IsGenesis() bool {
	return b.PrevHash == nil
}

// 通过确认区块中的信息检查区块是否合法
func (b *Block) IsBlockValid(oldBlock Block) bool {
	if oldBlock.Height+1 != b.Height {
		return false
	}
	res := bytes.Compare(oldBlock.Hash, b.PrevHash)
	if res != 0 {
		return false
	}
	// pow := NewProof(b)
	// validate := pow.Validate()

	return true
}

func ConstructJSON(buffer *bytes.Buffer, block *Block) {
	buffer.WriteString("{")
	buffer.WriteString(fmt.Sprintf("\"%s\":\"%d\",", "Timestamp", block.Timestamp))
	buffer.WriteString(fmt.Sprintf("\"%s\":\"%x\",", "PrevHash", block.PrevHash))

	buffer.WriteString(fmt.Sprintf("\"%s\":\"%x\",", "Hash", block.Hash))

	buffer.WriteString(fmt.Sprintf("\"%s\":%d,", "Difficulty", block.Difficulty))

	buffer.WriteString(fmt.Sprintf("\"%s\":%d,", "Nonce", block.Nonce))

	buffer.WriteString(fmt.Sprintf("\"%s\":\"%x\",", "MerkleRoot", block.MerkleRoot))
	buffer.WriteString(fmt.Sprintf("\"%s\":%d", "TxCount", block.TxCount))
	buffer.WriteString("}")
}

func (bs *Block) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("[")
	ConstructJSON(buffer, bs)
	buffer.WriteString("]")
	return buffer.Bytes(), nil
}
