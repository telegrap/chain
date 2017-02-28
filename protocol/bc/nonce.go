package bc

type Nonce struct {
	body struct {
		Program   Program
		TimeRange Hash
		ExtHash   Hash
	}
	TimeRange *TimeRange
}

const typeNonce = "nonce1"

func (Nonce) Type() string            { return typeNonce }
func (n *Nonce) Body() interface{}    { return &n.body }
func (n *Nonce) Witness() interface{} { return nil }

func NewNonce(p Program, tr *TimeRange) *Nonce {
	n := new(Nonce)
	n.body.Program = p
	if tr != nil {
		id := EntryID(tr)
		n.body.TimeRange = id
		n.TimeRange = tr
	}
	return n
}
