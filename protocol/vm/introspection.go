package vm

import (
	"bytes"
	"fmt"
	"math"

	"chain/protocol/bc"
)

func opCheckOutput(vm *virtualMachine) error {
	if vm.tx == nil {
		return ErrContext
	}

	err := vm.applyCost(16)
	if err != nil {
		return err
	}

	code, err := vm.pop(true)
	if err != nil {
		return err
	}
	vmVersion, err := vm.popInt64(true)
	if err != nil {
		return err
	}
	if vmVersion < 0 {
		return ErrBadValue
	}
	assetID, err := vm.pop(true)
	if err != nil {
		return err
	}
	amount, err := vm.popInt64(true)
	if err != nil {
		return err
	}
	if amount < 0 {
		return ErrBadValue
	}
	refdatahash, err := vm.pop(true)
	if err != nil {
		return err
	}
	index, err := vm.popInt64(true)
	if err != nil {
		return err
	}
	if index < 0 {
		return ErrBadValue
	}

	// The following is per the discussion at
	// https://chainhq.slack.com/archives/txgraph/p1487964172000960
	var inpDest bc.Entry
	switch inp := vm.input.(type) {
	case *bc.Spend:
		inpDest = inp.Destination
	case *bc.Issuance:
		inpDest = inp.Destination
	default:
		// xxx error
	}
	mux, ok := inpDest.(*bc.Mux)
	if !ok {
		return vm.pushBool(false, true)
	}
	muxDests := mux.Destinations
	if index >= int64(len(muxDests)) {
		return vm.pushBool(false, true) // xxx or should this be a range/badvalue error?
	}

	someChecks := func(resAssetID bc.AssetID, resAmount uint64, resData bc.Hash) bool {
		if !bytes.Equal(resAssetID[:], assetID) {
			return false
		}
		if resAmount != uint64(amount) {
			return false
		}
		if len(refdatahash) > 0 && !bytes.Equal(refdatahash, resData[:]) {
			return false
		}
		return true
	}

	if vmVersion == 1 && bytes.Equal(code, []byte{byte(OP_FAIL)}) {
		// Special case alert! Old-style retirements were just outputs
		// with the control program [FAIL]. New-style retirements do not
		// have control programs, but for compatibility we allow
		// CHECKOUTPUT to test for when the [FAIL] program is specified.
		r, ok := muxDests[index].(*bc.Retirement)
		if !ok {
			return vm.pushBool(false, true)
		}
		ok = someChecks(r.AssetID(), r.Amount(), r.Data())
		return vm.pushBool(ok, true)
	}

	o, ok := muxDests[index].(*bc.Output)
	if !ok {
		return vm.pushBool(false, true)
	}

	if !someChecks(o.AssetID(), o.Amount(), o.Data()) {
		return vm.pushBool(false, true)
	}
	prog := o.ControlProgram()
	if prog.VMVersion != uint64(vmVersion) {
		return vm.pushBool(false, true)
	}
	if !bytes.Equal(prog.Code, code) {
		return vm.pushBool(false, true)
	}
	return vm.pushBool(true, true)
}

func opAsset(vm *virtualMachine) error {
	if vm.tx == nil {
		return ErrContext
	}

	err := vm.applyCost(1)
	if err != nil {
		return err
	}

	var assetID bc.AssetID

	switch e := vm.input.(type) {
	case *bc.Spend:
		assetID = e.AssetAmount().AssetID

	case *bc.Issuance:
		assetID = e.AssetID()

	default:
		// xxx error
	}

	return vm.push(assetID[:], true)
}

func opAmount(vm *virtualMachine) error {
	if vm.tx == nil {
		return ErrContext
	}

	err := vm.applyCost(1)
	if err != nil {
		return err
	}

	var amount uint64

	switch e := vm.input.(type) {
	case *bc.Spend:
		amount = e.AssetAmount().Amount

	case *bc.Issuance:
		amount = e.Amount()

	default:
		// xxx error
	}

	return vm.pushInt64(int64(amount), true)
}

func opProgram(vm *virtualMachine) error {
	if vm.tx == nil {
		return ErrContext
	}

	err := vm.applyCost(1)
	if err != nil {
		return err
	}

	return vm.push(vm.mainprog, true)
}

func opMinTime(vm *virtualMachine) error {
	if vm.tx == nil {
		return ErrContext
	}

	err := vm.applyCost(1)
	if err != nil {
		return err
	}

	return vm.pushInt64(int64(vm.tx.MinTimeMS()), true)
}

func opMaxTime(vm *virtualMachine) error {
	if vm.tx == nil {
		return ErrContext
	}

	err := vm.applyCost(1)
	if err != nil {
		return err
	}

	maxTime := vm.tx.MaxTimeMS()
	if maxTime == 0 || maxTime > math.MaxInt64 {
		maxTime = uint64(math.MaxInt64)
	}

	return vm.pushInt64(int64(maxTime), true)
}

func opRefDataHash(vm *virtualMachine) error {
	if vm.tx == nil {
		return ErrContext
	}

	err := vm.applyCost(1)
	if err != nil {
		return err
	}

	var h bc.Hash

	switch e := vm.input.(type) {
	case *bc.Spend:
		h = e.Data()
	case *bc.Issuance:
		h = e.Data()
	default:
		// xxx error
	}

	return vm.push(h[:], true)
}

func opTxRefDataHash(vm *virtualMachine) error {
	if vm.tx == nil {
		return ErrContext
	}

	err := vm.applyCost(1)
	if err != nil {
		return err
	}

	h := vm.tx.Data()
	return vm.push(h[:], true)
}

func opOutputID(vm *virtualMachine) error {
	if vm.tx == nil {
		return ErrContext
	}

	sp, ok := vm.input.(*bc.Spend)
	if !ok {
		return ErrContext
	}
	if sp == nil {
		// xxx error
	}
	outID := sp.OutputID()

	err := vm.applyCost(1)
	if err != nil {
		return err
	}

	return vm.push(outID[:], true)
}

func opNonce(vm *virtualMachine) error {
	if vm.tx == nil {
		return ErrContext
	}

	_, ok := vm.input.(*bc.Issuance)
	if !ok {
		return ErrContext
	}

	err := vm.applyCost(1)
	if err != nil {
		return err
	}

	var nonce []byte
	// xxx
	return vm.push(nonce, true)
}

func opNextProgram(vm *virtualMachine) error {
	if vm.block == nil {
		return ErrContext
	}
	err := vm.applyCost(1)
	if err != nil {
		return err
	}
	return vm.push(vm.block.NextConsensusProgram(), true)
}

func opBlockTime(vm *virtualMachine) error {
	if vm.block == nil {
		return ErrContext
	}
	err := vm.applyCost(1)
	if err != nil {
		return err
	}
	if vm.block.TimestampMS() > math.MaxInt64 {
		return fmt.Errorf("block timestamp out of range")
	}
	return vm.pushInt64(int64(vm.block.TimestampMS()), true)
}
