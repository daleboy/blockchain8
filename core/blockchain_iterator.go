package blockchain8

import (
	"github.com/boltdb/bolt"
)

//BlockchainIterator 区块链迭代器，用于对区块链中的区块进行迭代
type BlockchainIterator struct {
	CurrentHash []byte
	Database    *bolt.DB
}

//Iterator 每当需要对链中的区块进行迭代时候，我们就通过Blockchain创建迭代器
//注意，迭代器初始状态为链中的tip，因此迭代是从最新到最旧的进行获取
func (bc *Blockchain) Iterator() *BlockchainIterator {
	if bc.Tip == nil {
		return nil
	}
	bci := &BlockchainIterator{bc.Tip, bc.Database}
	return bci
}

//Next 区块链迭代，返回当前区块，并更新迭代器的currentHash为当前区块的PrevBlockHash
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.Database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodeBlock := b.Get(i.CurrentHash)
		block = DeserializeBlock(encodeBlock)

		return nil
	})

	Handle(err)

	i.CurrentHash = block.PrevHash

	return block
}
