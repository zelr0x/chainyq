# Changelog

## [0.2.3] - 2026-03-25

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
