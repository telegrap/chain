package bc

type (
	// output contains a ValueSource that must refer to the Mux by its
	// entryID, but the Mux may not be complete at the time AddOutput is
	// called, so we hold outputs in a pending structure until Build is
	// called
	pendingOutput struct {
		value       AssetAmount
		controlProg Program
		data        Hash
	}

	pendingRetirement struct {
		value AssetAmount
		data  Hash
	}

	Builder struct {
		h           *Header
		m           *Mux
		spends      []*Spend
		issuances   []*Issuance
		outputs     []*pendingOutput
		retirements []*pendingRetirement
	}
)

func NewBuilder(version, minTimeMS, maxTimeMS uint64, base *Transaction) *Builder {
	result := &Builder{
		h: newHeader(version, nil, Hash{}, minTimeMS, maxTimeMS),
		m: newMux(nil, Program{VMVersion: 1, Code: []byte{0x51}}), // 0x51 == OP_TRUE (without the circular dependency)
	}
	if base != nil {
		for _, iss := range base.Issuances {
			result.AddIssuance(iss.Anchor, AssetAmount{AssetID: iss.AssetID(), Amount: iss.Amount()}, iss.Data())
		}
		for _, sp := range base.Spends {
			if sp.SpentOutput != nil {
				result.AddFullSpend(sp.SpentOutput, sp.Data())
			} else {
				result.AddPrevoutSpend(sp.body.SpentOutput, sp.prevout, sp.Data())
			}
		}
		for _, o := range base.Outputs {
			result.AddOutput(AssetAmount{AssetID: o.AssetID(), Amount: o.Amount()}, o.ControlProgram(), o.Data())
		}
		for _, r := range base.Retirements {
			result.AddRetirement(AssetAmount{AssetID: r.AssetID(), Amount: r.Amount()}, r.Data())
		}
	}
	return result
}

func (b *Builder) Data() Hash {
	return b.h.Data()
}

func (b *Builder) SetData(h Hash) {
	b.h.body.Data = h
}

func (b *Builder) RestrictMinTimeMS(minTimeMS uint64) {
	if minTimeMS > b.h.MinTimeMS() {
		b.h.body.MinTimeMS = minTimeMS
	}
}

func (b *Builder) RestrictMaxTimeMS(maxTimeMS uint64) {
	if maxTimeMS < b.h.MaxTimeMS() {
		b.h.body.MaxTimeMS = maxTimeMS
	}
}

func (b *Builder) MaxTimeMS() uint64 {
	return b.h.MaxTimeMS()
}

func (b *Builder) AddIssuance(nonce Entry, value AssetAmount, data Hash) *Issuance {
	iss := newIssuance(nonce, value, data)
	b.issuances = append(b.issuances, iss)
	issID := EntryID(iss)
	s := ValueSource{
		Ref:   issID,
		Value: value,
	}
	b.m.body.Sources = append(b.m.body.Sources, s)
	b.m.Sources = append(b.m.Sources, iss)
	return iss
}

// AddOutput does not return an entry, unlike other Add
// functions, since output objects aren't created until Build
func (b *Builder) AddOutput(value AssetAmount, controlProg Program, data Hash) {
	b.outputs = append(b.outputs, &pendingOutput{
		value:       value,
		controlProg: controlProg,
		data:        data,
	})
}

// AddRetirement does not return an entry, unlike most other Add
// functions, since retirement objects aren't created until Build
func (b *Builder) AddRetirement(value AssetAmount, data Hash) {
	b.retirements = append(b.retirements, &pendingRetirement{
		value: value,
		data:  data,
	})
}

func (b *Builder) AddFullSpend(spentOutput *Output, data Hash) *Spend {
	sp := NewFullSpend(spentOutput, data)
	return b.addSpend(sp)
}

func (b *Builder) AddPrevoutSpend(outputID Hash, prevout *Prevout, data Hash) *Spend {
	sp := NewPrevoutSpend(outputID, prevout, data)
	return b.addSpend(sp)
}

func (b *Builder) addSpend(sp *Spend) *Spend {
	b.spends = append(b.spends, sp)
	src := ValueSource{
		Ref:   EntryID(sp),
		Value: sp.AssetAmount(),
	}
	b.m.body.Sources = append(b.m.body.Sources, src)
	b.m.Sources = append(b.m.Sources, sp)
	return sp
}

func (b *Builder) Build() *Transaction {
	var n uint64
	tx := &Transaction{
		Header:    b.h,
		Spends:    b.spends,
		Issuances: b.issuances,
	}
	muxID := EntryID(b.m)
	for _, po := range b.outputs {
		s := ValueSource{
			Ref:      muxID,
			Value:    po.value,
			Position: n,
		}
		n++
		o := NewOutput(s, po.controlProg, po.data)
		o.Source = b.m
		oID := EntryID(o)
		b.h.body.Results = append(b.h.body.Results, oID)
		b.h.Results = append(b.h.Results, o)
		tx.Outputs = append(tx.Outputs, o)
	}
	for _, pr := range b.retirements {
		s := ValueSource{
			Ref:      muxID,
			Value:    pr.value,
			Position: n,
		}
		n++
		r := newRetirement(s, pr.data)
		r.Source = b.m
		rID := EntryID(r)
		b.h.body.Results = append(b.h.body.Results, rID)
		b.h.Results = append(b.h.Results, r)
		tx.Retirements = append(tx.Retirements, r)
	}
	return tx
}
