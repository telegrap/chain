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
		SpentOutput Hash
		Data        Hash
		ExtHash     Hash
	}
	witness struct {
		Destination ValueDestination
		Arguments   [][]byte
	}
	prevout     *Prevout
	SpentOutput *Output
	Destination Entry
}

const typeSpend = "spend1"

func (Spend) Type() string            { return typeSpend }
func (s *Spend) Body() interface{}    { return &s.body }
func (s *Spend) Witness() interface{} { return &s.witness }

func (s *Spend) Data() Hash {
	return s.body.Data
}

func (s *Spend) Arguments() [][]byte {
	return s.witness.Arguments
}

func (s *Spend) SetArguments(args [][]byte) {
	s.witness.Arguments = args
}

func (s *Spend) OutputID() Hash {
	return s.body.SpentOutput
}

func (s *Spend) AssetAmount() AssetAmount {
	if s.SpentOutput != nil {
		return s.SpentOutput.AssetAmount()
	}
	return s.prevout.AssetAmount
}

func (s *Spend) ControlProgram() Program {
	if s.SpentOutput != nil {
		return s.SpentOutput.ControlProgram()
	}
	return s.prevout.Program
}

func NewFullSpend(spentOutput *Output, data Hash) *Spend {
	s := new(Spend)
	oID := EntryID(spentOutput)
	s.body.SpentOutput = oID
	s.SpentOutput = spentOutput
	s.body.Data = data
	return s
}

func NewPrevoutSpend(outputID Hash, prevout *Prevout, data Hash) *Spend {
	s := new(Spend)
	s.body.SpentOutput = outputID
	s.body.Data = data
	s.prevout = prevout
	return s
}
