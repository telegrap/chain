package bc

type ValueSource struct {
	Ref      Hash
	Value    AssetAmount
	Position uint64 // zero unless Ref is a Mux
}

type ValueDestination struct {
	Ref      Hash
	Position uint64
}
