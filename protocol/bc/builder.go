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
		h                 *Header
		m                 *Mux
		spends, issuances []*EntryRef
		outputs           []*pendingOutput
		retirements       []*pendingRetirement
	}
)

func NewBuilder(version, minTimeMS, maxTimeMS uint64, base *Transaction) *Builder {
	result := &Builder{
		h: newHeader(version, nil, Hash{}, minTimeMS, maxTimeMS),
		m: newMux(nil, Program{VMVersion: 1, Code: []byte{0x51}}), // 0x51 == OP_TRUE (without the circular dependency)
	}
	if base != nil {
		for _, issRef := range base.Issuances {
			iss := issRef.Entry.(*Issuance)
			result.AddIssuance(iss.Anchor(), AssetAmount{AssetID: iss.AssetID(), Amount: iss.Amount()}, iss.Data())
		}
		for _, spRef := range base.Spends {
			sp := spRef.Entry.(*Spend)
			if sp.body.SpentOutput.Entry != nil {
				result.AddFullSpend(sp.body.SpentOutput, sp.Data())
			} else {
				result.AddPrevoutSpend(sp.body.SpentOutput.Hash(), sp.prevout, sp.Data())
			}
		}
		for _, oRef := range base.Outputs {
			o := oRef.Entry.(*Output)
			result.AddOutput(AssetAmount{AssetID: o.AssetID(), Amount: o.Amount()}, o.ControlProgram(), o.Data())
		}
		for _, rRef := range base.Retirements {
			r := rRef.Entry.(*Retirement)
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

func (b *Builder) AddIssuance(nonce *EntryRef, value AssetAmount, data Hash) *EntryRef {
	issRef := &EntryRef{Entry: newIssuance(nonce, value, data)}
	b.issuances = append(b.issuances, issRef)
	s := ValueSource{
		Ref:   issRef,
		Value: value,
	}
	b.m.body.Sources = append(b.m.body.Sources, s)
	return issRef
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

func (b *Builder) AddFullSpend(spentOutput *EntryRef, data Hash) *EntryRef {
	sp := NewFullSpend(spentOutput, data)
	return b.addSpend(sp)
}

func (b *Builder) AddPrevoutSpend(outputID Hash, prevout *Prevout, data Hash) *EntryRef {
	sp := NewPrevoutSpend(outputID, prevout, data)
	return b.addSpend(sp)
}

func (b *Builder) addSpend(sp *Spend) *EntryRef {
	spRef := &EntryRef{Entry: sp}
	b.spends = append(b.spends, spRef)
	src := ValueSource{
		Ref:   spRef,
		Value: sp.AssetAmount(),
	}
	b.m.body.Sources = append(b.m.body.Sources, src)
	return spRef
}

func (b *Builder) Build() *Transaction {
	var n uint64
	muxRef := &EntryRef{Entry: b.m}
	tx := &Transaction{
		Header:    &EntryRef{Entry: b.h},
		Spends:    b.spends,
		Issuances: b.issuances,
	}
	for _, po := range b.outputs {
		s := ValueSource{
			Ref:      muxRef,
			Value:    po.value,
			Position: n,
		}
		n++
		o := NewOutput(s, po.controlProg, po.data)
		oRef := &EntryRef{Entry: o}
		b.h.body.Results = append(b.h.body.Results, oRef)
		tx.Outputs = append(tx.Outputs, oRef)
	}
	for _, pr := range b.retirements {
		s := ValueSource{
			Ref:      muxRef,
			Value:    pr.value,
			Position: n,
		}
		n++
		r := newRetirement(s, pr.data)
		rRef := &EntryRef{Entry: r}
		b.h.body.Results = append(b.h.body.Results, rRef)
		tx.Retirements = append(tx.Retirements, rRef)
	}
	return tx
}
