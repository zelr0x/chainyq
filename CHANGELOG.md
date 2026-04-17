# Changelog

## [0.2.5] - 2026-04-18

### Added
- Added `IterAll` (`iter.Seq` range support) for all unidirectional iterators

## [0.2.4] - 2026-04-17

### Added
- Added constructors `seq.Sized`, `seq.SizedU64`, `seq.ExactSized`,
`seq.ExactSizedU64`, and `Seq.WithSkip` for optimized sequences.
Methods of `Seq` have been rewritten to utilize size knowledge and fast
skipping
- Added `Seq.TakeWhile`, `Seq.Find`, `Seq.Any`, `Seq.All`, and `Seq.None`
- Added `Seq.Slice` that provides safe sequence slicing that short-circuits
to an empty sequence on invalid arguments to simplify `Skip` + `Take`
chaining
- Added `Seq.IterAll` and `Seq.IterSlice` that return `iter.Seq` to allow using
sequences in loops with the range syntax
- Documented sequence upstream interactions

### Fixed
- Fixed incorrect implementation of `Seq.SkipWhile` that skipped all matched
items in the sequence and not until the first mismatch
- Fixed incorrect implementations of `List.BidiIter.RevSeq`, `List.RevIter.Seq`,
`Deque.BidiIter.RevSeq`, and `Deque.RevIter.Seq` - earlier implementations
incorrectly reset the copy of the iterator before creating a reversed sequence,
resulting in it always starting from the end, which is inconsistent with
"all operations are on remaining items" semantics and with forward sequence
creation methods


## [0.2.3] - 2026-03-25

### Added
- `Deque.Clone`

### Fixed
- Off-by-one error in `Deque.BackPtr`


## [0.2.2] - 2026-03-24

### Removed
- Dependencies used for benchmarks, benchmarks moved to a separate module


## [0.2.1] - 2026-03-24

### Changed
- Refactor `chainyq.Deque` interface definition


## [0.2.0] - 2026-03-24

### Added
- `Deque.Reserve`, `Deque.ReserveFront`, `Deque.ReserveBack`.

### Fixed
- `Deque` growth at the front and `Deque.EnsureFront` specifically, making it 0 B/op as expected.

### Removed
- `deque.Allocate`.
