package bc

type Header struct {
	body struct {
		Version              uint64
		Results              []Hash
		Data                 Hash
		MinTimeMS, MaxTimeMS uint64
		ExtHash              Hash
	}
	Results []Entry
}

const typeHeader = "txheader"

func (Header) Type() string            { return typeHeader }
func (h *Header) Body() interface{}    { return &h.body }
func (h *Header) Witness() interface{} { return nil }

func (h *Header) Version() uint64 {
	return h.body.Version
}

func (h *Header) MinTimeMS() uint64 {
	return h.body.MinTimeMS
}

func (h *Header) MaxTimeMS() uint64 {
	return h.body.MaxTimeMS
}

func (h *Header) Data() Hash {
	return h.body.Data
}

func (h *Header) Walk(visitor func(id Hash, entry Entry) error) error {
	visited := make(map[Hash]bool)
	visit := func(id Hash, entry Entry) error {
		if entry == nil {
			return nil
		}
		if visited[id] {
			return nil
		}
		visited[id] = true
		return visitor(id, entry)
	}
	for i, id := range h.body.Results {
		var entry Entry
		if i < len(h.Results) {
			entry = h.Results[i]
		}
		err := visit(id, entry)
		if err != nil {
			return err
		}
		switch e2 := entry.(type) {
		case *Issuance:
			err = visit(e2.body.Anchor, e2.Anchor)
			if err != nil {
				return err
			}
			err = visit(e2.witness.Destination.Ref, e2.Destination)
			if err != nil {
				return err
			}
		case *Mux:
			for j, vs := range e2.body.Sources {
				var s Entry
				if j < len(e2.Sources) {
					s = e2.Sources[j]
				}
				err = visit(vs.Ref, s)
				if err != nil {
					return err
				}
			}
		case *Nonce:
			err = visit(e2.body.TimeRange, e2.TimeRange)
			if err != nil {
				return err
			}
		case *Output:
			err = visit(e2.body.Source.Ref, e2.Source)
			if err != nil {
				return err
			}
		case *Retirement:
			err = visit(e2.body.Source.Ref, e2.Source)
			if err != nil {
				return err
			}
		case *Spend:
			err = visit(e2.body.SpentOutput, e2.SpentOutput)
			if err != nil {
				return err
			}
			err = visit(e2.witness.Destination.Ref, e2.Destination)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Inputs returns all input entries (as two lists: spends and
// issuances) reachable from a header's result entries.
func (h *Header) Inputs() (spends []*Spend, issuances []*Issuance) {
	h.Walk(func(_ Hash, e Entry) error {
		switch e2 := e.(type) {
		case *Spend:
			spends = append(spends, e2)
		case *Issuance:
			issuances = append(issuances, e2)
		}
		return nil
	})
	return
}

func newHeader(version uint64, results []Hash, data Hash, minTimeMS, maxTimeMS uint64) *Header {
	h := new(Header)
	h.body.Version = version
	h.body.Results = results
	h.body.Data = data
	h.body.MinTimeMS = minTimeMS
	h.body.MaxTimeMS = maxTimeMS
	return h
}
