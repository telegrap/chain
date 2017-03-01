package tx

import (
	"fmt"

	"chain/crypto/sha3pool"
	"chain/errors"
	"chain/protocol/bc"
)

func init() {
	bc.TxHashesFunc = TxHashes
	bc.BlockHeaderHashFunc = func(old *bc.BlockHeader) bc.Hash {
		hash, _ := mapBlockHeader(old)
		return hash
	}
}

// TxHashes returns all hashes needed for validation and state updates.
func TxHashes(oldTx *bc.TxData) (hashes *bc.TxHashes, err error) {
	txid, header, entries, err := mapTx(oldTx)
	if err != nil {
		return nil, errors.Wrap(err, "mapping old transaction to new")
	}

	hashes = new(bc.TxHashes)
	hashes.ID = txid

	// ResultHashes
	hashes.ResultHashes = make([]bc.Hash, len(header.body.Results))
	for i, resultHash := range header.body.Results {
		hashes.ResultHashes[i] = resultHash
	}

	hashes.VMContexts = make([]*bc.VMContext, len(oldTx.Inputs))

	for entryID, ent := range entries {
	retry:
		switch ent2 := ent.(type) {
		case *idWrapper:
			ent = ent2.entry
			goto retry

		case *nonce:
			// TODO: check time range is within network-defined limits
			trID := ent2.body.TimeRange
			trEntry, ok := entries[trID]
			if !ok {
				return nil, fmt.Errorf("nonce entry refers to nonexistent timerange entry")
			}
			if w, ok := trEntry.(*idWrapper); ok {
				trEntry = w.entry
			}
			tr, ok := trEntry.(*timeRange)
			if !ok {
				return nil, fmt.Errorf("nonce entry refers to %s entry, should be timerange1", trEntry.Type())
			}
			iss := struct {
				ID           bc.Hash
				ExpirationMS uint64
			}{entryID, tr.body.MaxTimeMS}
			hashes.Issuances = append(hashes.Issuances, iss)

		case *issuance:
			vmc := newVMContext(entryID, hashes.ID, header.body.Data, ent2.body.Data)
			vmc.NonceID = &ent2.body.Anchor
			hashes.VMContexts[ent2.Ordinal()] = vmc

		case *spend:
			vmc := newVMContext(entryID, hashes.ID, header.body.Data, ent2.body.Data)
			vmc.OutputID = &ent2.body.SpentOutput
			hashes.VMContexts[ent2.Ordinal()] = vmc
		}
	}

	return hashes, nil
}

// populates the common fields of a VMContext for an Entry, regardless of whether
// that Entry is a Spend or an Issuance
func newVMContext(entryID, txid, txData, inpData bc.Hash) *bc.VMContext {
	vmc := new(bc.VMContext)

	// TxRefDataHash
	vmc.TxRefDataHash = txData

	// RefDataHash (input-specific)
	vmc.RefDataHash = inpData

	// EntryID
	vmc.EntryID = entryID

	// TxSigHash
	hasher := sha3pool.Get256()
	defer sha3pool.Put256(hasher)
	hasher.Write(entryID[:])
	hasher.Write(txid[:])
	hasher.Read(vmc.TxSigHash[:])

	return vmc
}
