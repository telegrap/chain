package tx

import "chain/protocol/bc"

type Spend struct {
	body struct {
		SpentOutput bc.OutputID
		Data        bc.Hash
		ExtHash     extHash
	}
	ordinal int
}

func (Spend) Type() string         { return "spend1" }
func (s *Spend) Body() interface{} { return s.body }

func (s Spend) Ordinal() int { return s.ordinal }

func newSpend(spentOutput bc.OutputID, data bc.Hash, ordinal int) *Spend {
	s := new(Spend)
	s.body.SpentOutput = spentOutput
	s.body.Data = data
	s.ordinal = ordinal
	return s
}
