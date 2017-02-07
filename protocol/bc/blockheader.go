package bc

import (
	"bytes"
	"chain/encoding/bufpool"
	"database/sql/driver"
	"encoding/hex"
	"errors"
)

type BlockHeader struct {
	body struct {
		Version              uint64
		Height               uint64
		PreviousBlockID      Hash
		TimestampMS          uint64
		TransactionsRoot     Hash
		AssetsRoot           Hash
		NextConsensusProgram []byte
		ExtHash              Hash
	}
	witness struct {
		Arguments [][]byte
		ExtHash   Hash
	}
}

func (BlockHeader) Type() string             { return "blockheader" }
func (bh *BlockHeader) Body() interface{}    { return &bh.body }
func (bh *BlockHeader) Witness() interface{} { return &bh.witness }

func (bh *BlockHeader) Version() uint64 {
	return bh.body.Version
}

func (bh *BlockHeader) PreviousBlockID() Hash {
	return bh.body.PreviousBlockID
}

func (bh *BlockHeader) TimestampMS() uint64 {
	return bh.body.TimestampMS
}

func (bh *BlockHeader) TransactionsRoot() Hash {
	return bh.body.TransactionsRoot
}

func (bh *BlockHeader) AssetsRoot() Hash {
	return bh.body.AssetsRoot
}

func (bh *BlockHeader) NextConsensusProgram() []byte {
	return bh.body.NextConsensusProgram
}

func (bh *BlockHeader) Height() uint64 {
	return bh.body.Height
}

func (bh *BlockHeader) Arguments() [][]byte {
	return bh.witness.Arguments
}

func (bh *BlockHeader) SetArguments(args [][]byte) {
	bh.witness.Arguments = args
}

func (bh *BlockHeader) Scan(val interface{}) error {
	buf, ok := val.([]byte)
	if !ok {
		return errors.New("Scan must receive a byte slice")
	}
	return deserializeEntry(bytes.NewReader(buf), bh)
}

func (bh *BlockHeader) Value() (driver.Value, error) {
	buf := new(bytes.Buffer) // not bufpool.Get(), because buf.Bytes() escapes this function
	err := serializeEntry(buf, bh)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalText fulfills the json.Marshaler interface.
// This guarantees that block headers will get deserialized correctly
// when being parsed from HTTP requests.
func (bh *BlockHeader) MarshalText() ([]byte, error) {
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	err := serializeEntry(buf, bh)
	if err != nil {
		return nil, err
	}

	enc := make([]byte, hex.EncodedLen(buf.Len()))
	hex.Encode(enc, buf.Bytes())
	return enc, nil
}

// UnmarshalText fulfills the encoding.TextUnmarshaler interface.
func (bh *BlockHeader) UnmarshalText(text []byte) error {
	decoded := make([]byte, hex.DecodedLen(len(text)))
	_, err := hex.Decode(decoded, text)
	if err != nil {
		return err
	}
	return deserializeEntry(bytes.NewReader(decoded), bh)
}

func NewBlockHeader(version, height uint64, previousBlockID Hash, timestampMS uint64, transactionsRoot, assetsRoot Hash, nextConsensusProgram []byte) *BlockHeader {
	bh := new(BlockHeader)
	bh.body.Version = version
	bh.body.Height = height
	bh.body.PreviousBlockID = previousBlockID
	bh.body.TimestampMS = timestampMS
	bh.body.TransactionsRoot = transactionsRoot
	bh.body.AssetsRoot = assetsRoot
	bh.body.NextConsensusProgram = nextConsensusProgram
	return bh
}
