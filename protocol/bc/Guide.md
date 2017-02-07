A guide to “entries” changes in protocol/bc
===========================================

Beginning with
https://github.com/chain/chain/commit/b817512a72b085f4c9440fafd27bbc220713df44,
transactions and block headers are converted on-the-fly to new data
structures when their hashes are needed. The eventual goal is to
replace the old `protocol/bc` structures with the new ones
entirely. On-the-fly conversion will no longer be needed, and other
benefits of the new data model will become available.

The new data model is referred to as “tx entries” or alternately
“txgraph” because transactions are reorganized as a collection of
entries of different types, such as issuance, retirement, spend,
output, etc., plus a couple of important new types: “mux,” which can
split and combine values; and “header,” which serves as the root of
the graph formed by the entries. The header includes (among other
things) a list of “results,” which is equivalent to the old
`Transaction.Outputs` list, containing both “outputs” and
“retirements.” From those entries it is possible to traverse the graph
to other reachable entries (spends and issuances, typically via an
intervening mux) that form the transitive closure and therefore the
complete transaction.

(Blocks are also transformed by txgraph-style changes, but to a much
lesser degree.)

Each entry contains a body and a witness, and is addressable by a hash
that’s computed from the body. The hash of a transaction is simply the
hash of its singleton “header” entry. Each body, and each witness,
ends with an “exthash” value that, for now, is always all-zeroes.  In
the future, extensions to these entry types will add new fields beyond
those exthashes, and then the value of the exthash must be the hash of
the additional fields.

Typically, the body addresses other entries that are _sources_ of
value for the entry (plus essential metadata); the witness addresses
other entries that are _destinations_ for that entry’s value.  When
the source is a mux, it is necessary to specify not only the mux’s
hash, but also a particular output slot of the mux (like an old-style
txout index).

Completing the txgraph work
===========================

The description above applies to what landed on `main` prior to the
Feb 24th deadline, with the new data model in `protocol/tx` and
revised hashing functions in `protocol/bc` (initialized via function
pointers from `protocol/tx` to avoid circular dependencies). The code
for the new data model excludes witnesses and other details not
required specifically for computing hashes.

Completing the txgraph work involves replacing `Tx`, `TxData`,
`TxInput`, `TxOutput` and related types with the ones in `protocol/tx`
and propagating those changes through the codebase. This is what I’ve
begun to do on
[the `entries` branch](https://github.com/chain/chain/tree/entries). The
remainder of this document gives a high-level overview of the changes
this has entailed so far.

Move the contents of protocol/tx to protocol/bc
===============================================

Since we’re replacing most of the `protocol/bc` types, and since the
new types involve more than just transactions, and since there were
some dependency issues, it makes sense to move the contents of
`protocol/tx` into `protocol/bc`.

Exporting more types
====================

Many entries and supporting types and functions must be exported to
the rest of the codebase. I’ve been doing this on an as-needed
basis.

At this writing, I recently exported the `ValueSource` type but the
need to export it has disappeared; perhaps it can be re-unexported.

Getting rid of the entryRef, extHash, and OutputID types
========================================================

These types were aliases for (or in the case of `OutputID`, a simple
wrapper of) `bc.Hash`. I’ve collapsed them all down to just `bc.Hash`
for a variety of reasons:
- We’re not doing anything with exthashes yet;
- In light of `EntryRef` (see below), `entryRef` did not supply any additional type safety;
- With the advent of entries all having their own IDs, distinguishing `OutputID` from the types of other entry IDs was just wrong and confusing.

EntryRefs
=========

For hashing and serialization purposes, entries refer to one another
by their hashes; but for other logic traversing live txgraph objects
it’s handier and more efficient to traverse Go pointers. `EntryRef`
(not to be confused with `entryRef`, above) is a smart pointer type
that holds both a pointer to an entry and the entry’s hash. Either one
may be omitted. If the entry is present, the hash can be computed (and
cached) on demand.

Serializing an `EntryRef` serializes only the hash. See the section
below about the new `Transaction` type to learn about serialization
that must traverse the entries of a txgraph.

In practice, much of the new code using `EntryRef`s throughout the Core
expects the entry pointer to be present, with one main exception: a
`Spend` has special logic for coping with the absence of its previous
`Output` object. (This is described below.)

Encapsulate transactions in a new Transaction type
==================================================

Although a `Header` contains all the information needed to reach all
parts of a transaction, it is convenient for both performance,
notation, and semantics to wrap this up in another data type,
`Transaction`, representing a complete transaction. In addition to the
`Header`, a `Transaction` lists issuance, spend, output, and
retirement entries separately for easy access.

It also defines `writeTo` and `readFrom` methods written in terms of
entry serialization. The format of a written transaction is:
- Serialized header;
- Varint number of additional entries;
- All other reachable entries serialized one after the other.

This format is a placeholder until we settle on something permanent
and sufficiently extensible (e.g. protobufs).

It’s my plan to rename `Transaction` to `Tx` for brevity once all
traces of the old `bc.Tx` and `bc.TxData` are purged.

Add an optional prevout structure to Spends
===========================================

Thanks to prevouts, a txgraph conceptually extends all the way back to
the beginning of blockchain time. To prevent us having to reconstitute
all that history when spending utxos, the `Spend` type has an
auxiliary member (“auxiliary” because it’s neither body nor witness):
`prevout`. This structure, if present, contains the prevout data
required for validation: assetID, amount, and control program. Those
data are not present in the `Spend` itself, so we need either the
complete previous `Output` or this abbreviated structure in order to
do most of the things we need to do with `Spend`s.

When the complete previous `Output` is absent, a `Spend` must still
contain its outputID in its `SpentOutput` `EntryRef`.

A ripple from this change produced a simplification in the `utxo`
type, which is now able to reuse the new `bc.Prevout` data structure.

Remove deprecated outpoint logic
================================

Outpoints - specifically, txout indexes - no longer have any real
meaning in a txgraph world. However, for the moment at least, I’ve
preserved an output-position value (in the form of an index into a
`Header` entry’s `Results` list) for use in query cursors (the
`OutputsAfter` type in `core/query/outputs.go`).

bc.Builder
==========

Now that a transaction is a collection of smaller, more numerous
objects than before, it’s harder now to write one the way we’re used
to doing with e.g. `TxData` literals. (Especially since order of
creation matters when it comes to entry hashes being final so that
other entries can refer to them.)  So I’ve created a simple `Builder`
type in `protocol/bc`, and incidentally rewritten the builder in
`core/txbuilder` in terms of the one in `protocol/bc`.

What’s left
===========

Plenty.

There are xxx comments throughout the code with open questions and
unimplemented changes.

Almost no test code has been updated yet.

Transactions do not yet JSON encode or decode. Doing this in a way
that is compatible with our existing API may be a challenge.

Much of the validation logic of the new spec is not yet written.

EntryRefs aren’t as typesafe as we might like and there are panic-able
type assertions in many new places. The assumption by much of the
codebase that an entry is present in an `EntryRef` might be a problem.

We need to decide on whether query cursors continue using `Results`
indexes, or whether they should change to order by entry hash.

The blockchain no longer stores bare reference data, only hashes; so
where our API now accepts or delivers bare refdata, we need to store
it to or retrieve it from the db. Among other things, this raises
questions about how to expire such data from the db.

I’m on record as being not much of a protobuf fan, but the current _ad
hoc_ serialization format is not future-proof: old cores could not
correctly deserialize future transactions, because they’re not
sufficiently self-describing.  On the other hand, it’s nice to have
the same serialization logic used for both hashing and for persisting
objects.
