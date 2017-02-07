package bc

type ValueSource struct {
	Ref      *EntryRef
	Value    AssetAmount
	Position uint64 // zero unless Ref is a Mux
}

type ValueDestination struct {
	Ref      *EntryRef
	Position uint64
}
