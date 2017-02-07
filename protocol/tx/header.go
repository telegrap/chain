package tx

import "chain/protocol/bc"

type Header struct {
	body struct {
		Version              uint64
		ResultRefs           []bc.Hash
		Data                 bc.Hash
		MinTimeMS, MaxTimeMS uint64
		ExtHash              extHash
	}
}

func (Header) Type() string         { return "txheader" }
func (h *Header) Body() interface{} { return h.body }

func (Header) Ordinal() int { return -1 }

func newHeader(version uint64, resultRefs []bc.Hash, data bc.Hash, minTimeMS, maxTimeMS uint64) *Header {
	h := new(Header)
	h.body.Version = version
	h.body.ResultRefs = resultRefs
	h.body.Data = data
	h.body.MinTimeMS = minTimeMS
	h.body.MaxTimeMS = maxTimeMS
	return h
}
