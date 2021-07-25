package merkletrie

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"log"
)

type Account struct {
	Address string
	Balance int
}

func (acc *Account) CalHash() []byte {
	b4hash := bytes.Join(
		[][]byte{
			[]byte(acc.Address),
			ToHex(int64(acc.Balance)),
		},
		[]byte{},
	)

	hash := sha256.Sum256(b4hash)

	return hash[:]
}

func (acc *Account) Equal(otherAcc Account) bool {
	return acc.Address == otherAcc.Address
}

func JoinAccList(accList []*Account) []byte {
	var accListHashes [][]byte
	var accListHash []byte

	for _, acc := range accList {
		accListHashes = append(accListHashes, acc.Serialize())
	}

	accListHash = bytes.Join(accListHashes, []byte("sepAccList"))

	return accListHash[:]
}

func SplitAccList(joinedAccList []byte) []*Account {
	var accList []*Account

	accListHashes := bytes.Split(joinedAccList, []byte("sepAccList"))

	for _, accHash := range accListHashes {
		accList = append(accList, Deserialize(accHash))
	}

	return accList
}

func (acc *Account) Serialize() []byte {
	var res bytes.Buffer

	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(acc)
	Handle(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Account {
	var account Account

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&account)
	Handle(err)

	return &account
}

func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)

	err := binary.Write(buff, binary.BigEndian, num)
	Handle(err)

	return buff.Bytes()
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
