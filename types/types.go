package types

//GetInfoResponse 获取链信息返回消息结构
type GetInfoResponse struct {
	BlockHeight uint64
}

//SendTxResponse 发送交易返回消息结构
type SendTxResponse struct {
	Txid string
}

//GetNewAddressResponse 获得钱包地址返回消息结构
type GetNewAddressResponse struct {
	Address string
}
