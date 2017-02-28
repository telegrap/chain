package bc

import (
	"io"

	"chain/crypto/sha3pool"
	"chain/encoding/blockchain"
)

type Transaction struct {
	Header      *Header
	Issuances   []*Issuance
	Spends      []*Spend
	Outputs     []*Output
	Retirements []*Retirement
}

func NewTransaction(hdr *Header) *Transaction {
	spends, issuances := hdr.Inputs()
	tx := &Transaction{
		Header:    hdr,
		Issuances: issuances,
		Spends:    spends,
	}
	for _, r := range hdr.Results {
		switch r2 := r.(type) {
		case *Output:
			tx.Outputs = append(tx.Outputs, r2)
		case *Retirement:
			tx.Retirements = append(tx.Retirements, r2)
		}
	}
	return tx
}

func (tx *Transaction) ID() Hash {
	return EntryID(tx.Header)
}

func (tx *Transaction) SigHash(inpHash Hash) (hash Hash) {
	hasher := sha3pool.Get256()
	defer sha3pool.Put256(hasher)

	hasher.Write(inpHash[:])
	hash = tx.ID()
	hasher.Write(hash[:])
	hasher.Read(hash[:])
	return hash
}

func (tx *Transaction) Data() Hash {
	return tx.Header.Data()
}

func (tx *Transaction) Results() []Entry {
	return tx.Header.Results
}

func (tx *Transaction) Version() uint64 {
	return tx.Header.Version()
}

func (tx *Transaction) MinTimeMS() uint64 {
	return tx.Header.MinTimeMS()
}

func (tx *Transaction) MaxTimeMS() uint64 {
	return tx.Header.MaxTimeMS()
}

// writeTx writes the Header to w, followed by a varint31 count of
// entries, followed by all the entries reachable from the Header.
func (tx *Transaction) writeTo(w io.Writer) error {
	var entries []Entry
	tx.Header.Walk(func(_ Hash, entry Entry) error {
		if entry != tx.Header {
			entries = append(entries, entry)
		}
		return nil
	})
	err := serializeEntry(w, tx.Header)
	if err != nil {
		return err
	}
	_, err = blockchain.WriteVarint31(w, uint64(len(entries)))
	for _, e := range entries {
		err = writeEntry(w, e)
		if err != nil {
			return err
		}
	}
	return nil
}

// readTx reads the output of writeTx and populates entry pointers in
// the Header and the entries reachable from it.
func (tx *Transaction) readFrom(r io.Reader) error {
	var h Header
	err := deserialize(r, &h.body)
	if err != nil {
		return err
	}
	// xxx also deserialize into h.witness, eventually
	n, _, err := blockchain.ReadVarint31(r)
	if err != nil {
		return err
	}
	entries := make(map[Hash]Entry, n)
	for i := uint32(0); i < n; i++ {
		var e Entry
		err = readEntry(r, &e)
		if err != nil {
			return err
		}
		entries[EntryID(e)] = e
	}
	return h.Walk(func(id Hash, entry Entry) error {
		if _, ok := entries[id]; ok {
			// xxx
		}
		return nil
	})
	newTx := NewTransaction(&h)
	*tx = *newTx
	return nil
}
