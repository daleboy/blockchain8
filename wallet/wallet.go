package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"sd-chain/blockchain8/util/env"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ripemd160"
)

var conf = env.New()
var (
	checkSumlength = conf.WalletAddressChecksum
	version        = byte(0x00) // hexadecimal representation of zero
)

// https://golang.org/pkg/crypto/ecdsa/
type Wallet struct {
	//eliptic curve digital algorithm
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// Validate Wallet Address
func ValidateAddress(address string) bool {

	if len(address) != 34 {
		return false
	}
	//Convert the address to public key hash
	fullHash := Base58Decode([]byte(address))
	// Get the checkSum from Address
	checkSumFromHash := fullHash[len(fullHash)-checkSumlength:]
	//Get the version
	version := fullHash[0]
	pubKeyHash := fullHash[1 : len(fullHash)-checkSumlength]
	checkSum := CheckSum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(checkSum, checkSumFromHash) == 0
}
func (w *Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)
	versionedHash := append([]byte{version}, pubHash...)
	checksum := CheckSum(versionedHash)
	//version-publickeyHash-checksum
	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)

	return address
}

// Generate new Key Pair using ecdsa
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pub
}

func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	return &Wallet{private, public}
}

func PublicKeyHash(pubKey []byte) []byte {
	//generate a hash using sha256
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}

	// Re-hash the genrated sha256 using ripemd160
	publicRipMd := hasher.Sum(nil)
	return publicRipMd
}

func CheckSum(data []byte) []byte {
	firstHash := sha256.Sum256(data)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checkSumlength]
}

// HashPubKey 对公钥进行哈希
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}