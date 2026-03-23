# Changelog

## [0.2.0] - 2026-03-24

### Added
- `Deque.Reserve`, `Deque.ReserveFront`, `Deque.ReserveBack`.

### Fixed
- `Deque` growth at the front and `Deque.EnsureFront` specifically, making it 0 B/op as expected.

### Removed
- `deque.Allocate`.
