package tx

import "chain/protocol/bc"

type mux struct {
	body struct {
		Sources []valueSource
		Program program
		ExtHash bc.Hash
	}

	// Sources contains (pointers to) the manifested entries for each
	// body.Sources[i].Ref.
	Sources []entry
}

func (mux) Type() string         { return "mux1" }
func (m *mux) Body() interface{} { return m.body }

func (mux) Ordinal() int { return -1 }

func newMux(program program) *mux {
	m := new(mux)
	m.body.Program = program
	return m
}

func (m *mux) addSource(e entry, value bc.AssetAmount, position uint64) {
	w := newIDWrapper(e, nil)
	src := valueSource{
		Ref:      w.Hash,
		Value:    value,
		Position: position,
	}
	m.body.Sources = append(m.body.Sources, src)
	m.Sources = append(m.Sources, w)
}
