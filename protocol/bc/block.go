package bc

import (
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"io"

	"chain/encoding/blockchain"
	"chain/encoding/bufpool"
	"chain/errors"
)

// Block describes a complete block, including its header
// and the transactions it contains.
type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
}

// MarshalText fulfills the json.Marshaler interface.
// This guarantees that blocks will get deserialized correctly
// when being parsed from HTTP requests.
func (b *Block) MarshalText() ([]byte, error) {
	buf := bufpool.Get()
	defer bufpool.Put(buf)
	_, err := b.WriteTo(buf)
	if err != nil {
		return nil, err
	}

	enc := make([]byte, hex.EncodedLen(buf.Len()))
	hex.Encode(enc, buf.Bytes())
	return enc, nil
}

// UnmarshalText fulfills the encoding.TextUnmarshaler interface.
func (b *Block) UnmarshalText(text []byte) error {
	decoded := make([]byte, hex.DecodedLen(len(text)))
	_, err := hex.Decode(decoded, text)
	if err != nil {
		return err
	}
	return b.readFrom(bytes.NewReader(decoded))
}

// Scan fulfills the sql.Scanner interface.
func (b *Block) Scan(val interface{}) error {
	buf, ok := val.([]byte)
	if !ok {
		return errors.New("Scan must receive a byte slice")
	}
	return b.readFrom(bytes.NewReader(buf))
}

// Value fulfills the sql.driver.Valuer interface.
func (b *Block) Value() (driver.Value, error) {
	buf := new(bytes.Buffer)
	_, err := b.WriteTo(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (b *Block) readFrom(r io.Reader) error {
	// xxx do we still need serflags?
	bh := new(BlockHeader)
	err := deserializeEntry(r, bh)
	if err != nil {
		return errors.Wrap(err, "reading block header")
	}
	b.Header = bh
	n, _, err := blockchain.ReadVarint31(r)
	if err != nil {
		return errors.Wrap(err, "reading number of transactions")
	}
	for ; n > 0; n-- {
		var tx Transaction
		err = tx.readFrom(r)
		if err != nil {
			return errors.Wrapf(err, "reading transaction %d", len(b.Transactions))
		}
		b.Transactions = append(b.Transactions, &tx)
	}
	return nil
}

func (b *Block) WriteTo(w io.Writer) (int64, error) {
	ew := errors.NewWriter(w)
	b.writeTo(ew)
	return ew.Written(), ew.Err()
}

func (b *Block) writeTo(w io.Writer) error {
	// xxx do we still need serflags?
	err := serializeEntry(w, b.Header)
	if err != nil {
		return errors.Wrap(err, "writing blockheader")
	}
	_, err = blockchain.WriteVarint31(w, uint64(len(b.Transactions)))
	if err != nil {
		return errors.Wrap(err, "writing number of transactions")
	}
	for i, tx := range b.Transactions {
		err = tx.writeTo(w)
		if err != nil {
			return errors.Wrapf(err, "writing transaction %d", i)
		}
	}
	return nil
}

func (b *Block) Hash() Hash {
	return EntryID(b.Header)
}

func (b *Block) Version() uint64 {
	return b.Header.Version()
}

func (b *Block) PreviousBlockID() Hash {
	return b.Header.PreviousBlockID()
}

func (b *Block) TimestampMS() uint64 {
	return b.Header.TimestampMS()
}

func (b *Block) TransactionsRoot() Hash {
	return b.Header.TransactionsRoot()
}

func (b *Block) AssetsRoot() Hash {
	return b.Header.AssetsRoot()
}

func (b *Block) NextConsensusProgram() []byte {
	return b.Header.NextConsensusProgram()
}

func (b *Block) Height() uint64 {
	return b.Header.Height()
}

func (b *Block) Arguments() [][]byte {
	return b.Header.Arguments()
}

func (b *Block) SetArguments(args [][]byte) {
	b.Header.SetArguments(args)
}
