package bc

// Prevout is not an entry type. It encapsulate those elements of a
// spent output required for validation: assetid, amount, and control
// program. It can be used in a Spend when the full previous Output is
// not available. In that case, the spend's body.SpentOutput.Entry
// must be nil, its body.SpentOutput.ID must contain the correct
// outputid, and its prevout must be non-nil.
type Prevout struct {
	AssetAmount
	Program
}

type Spend struct {
	body struct {
		SpentOutput *EntryRef
		Data        Hash
		ExtHash     Hash
	}
	witness struct {
		Destination ValueDestination
		Arguments   [][]byte
	}
	prevout *Prevout
}

const typeSpend = "spend1"

func (Spend) Type() string            { return typeSpend }
func (s *Spend) Body() interface{}    { return &s.body }
func (s *Spend) Witness() interface{} { return &s.witness }

func (s *Spend) Data() Hash {
	return s.body.Data
}

func (s *Spend) Destination() ValueDestination {
	return s.witness.Destination
}

func (s *Spend) Arguments() [][]byte {
	return s.witness.Arguments
}

func (s *Spend) SetArguments(args [][]byte) {
	s.witness.Arguments = args
}

func (s *Spend) OutputID() Hash {
	return s.body.SpentOutput.Hash()
}

func (s *Spend) AssetAmount() AssetAmount {
	if s.prevout != nil {
		return s.prevout.AssetAmount
	}
	return s.body.SpentOutput.Entry.(*Output).AssetAmount()
}

func (s *Spend) ControlProgram() Program {
	if s.prevout != nil {
		return s.prevout.Program
	}
	return s.body.SpentOutput.Entry.(*Output).ControlProgram()
}

func NewFullSpend(spentOutput *EntryRef, data Hash) *Spend {
	s := new(Spend)
	s.body.SpentOutput = spentOutput
	s.body.Data = data
	return s
}

func NewPrevoutSpend(outputID Hash, prevout *Prevout, data Hash) *Spend {
	s := new(Spend)
	s.body.SpentOutput = &EntryRef{ID: &outputID}
	s.body.Data = data
	s.prevout = prevout
	return s
}
