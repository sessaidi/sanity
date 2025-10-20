package sanity

import (
	"errors"
	"fmt"
	"sync"
)

type Guard struct {
	// SSO (small-struct optimization) for first 4 errors
	e0, e1, e2, e3 error
	more           []error
	n              int // kept errors

	// Controls
	max          int         // 0 -> unlimited; 1 -> first-error (default)
	compactRatio int         // 0 -> default(2); used only when not thread-safe
	mu           sync.Locker // nil => no locking; else a real mutex

	// Stats
	checks   int // closures evaluated via AddCheck/Run/CheckLazy
	failures int // non-nil errors seen (kept + dropped)
	dropped  int // errors dropped due to cap
}

// GuardOption configures Guard behavior.
type GuardOption func(*Guard)

// WithMaxErrors sets the maximum kept errors.
// n==1 (default) => first-error; n==0 => unlimited; n>=2 => keep up to n.
func WithMaxErrors(n int) GuardOption {
	if n < 0 {
		n = 0
	}
	return func(g *Guard) { g.max = n }
}

// WithCompactRatio sets the compaction ratio for the 'more' slice
// when not thread-safe. r <= 0 defaults to 2. Ignored if thread-safe.
func WithCompactRatio(r int) GuardOption {
	return func(g *Guard) { g.compactRatio = r }
}

// WithThreadSafe enables internal locking. Err() returns a snapshot.
func WithThreadSafe() GuardOption {
	return func(g *Guard) { g.mu = &sync.Mutex{} }
}

// NewGuard constructs a Guard. Default is first-error (max=1).
func NewGuard(opts ...GuardOption) Guard {
	g := Guard{max: 1}
	for _, opt := range opts {
		opt(&g)
	}
	return g
}

func (g *Guard) lock() {
	if g.mu != nil {
		g.mu.Lock()
	}
}
func (g *Guard) unlock() {
	if g.mu != nil {
		g.mu.Unlock()
	}
}

// MGStats
// Stats for inspection/telemetry.
type MGStats struct {
	Checks   int
	Failures int
	Kept     int
	Dropped  int
}

func (g *Guard) Stats() MGStats {
	g.lock()
	defer g.unlock()
	return MGStats{Checks: g.checks, Failures: g.failures, Kept: g.n, Dropped: g.dropped}
}

// Reset clears all state for reuse.
func (g *Guard) Reset() {
	g.lock()
	g.e0, g.e1, g.e2, g.e3 = nil, nil, nil, nil
	g.more = nil
	g.n = 0
	g.checks, g.failures, g.dropped = 0, 0, 0
	g.unlock()
}

// Ok reports whether no error has been recorded.
func (g *Guard) Ok() bool {
	g.lock()
	ok := g.n == 0
	g.unlock()
	return ok
}

// Add records err if non-nil; respects cap (max).
func (g *Guard) Add(err error) {
	if err == nil {
		return
	}
	g.lock()
	g.failures++
	if g.max > 0 && g.n >= g.max {
		g.dropped++
		g.unlock()
		return
	}
	switch g.n {
	case 0:
		g.e0 = err
	case 1:
		g.e1 = err
	case 2:
		g.e2 = err
	case 3:
		g.e3 = err
	default:
		g.more = append(g.more, err)
	}
	g.n++
	g.unlock()
}

// Check is a convenience alias for Add.
func (g *Guard) Check(err error) {
	if err == nil {
		return
	}
	g.Add(err)
}

// Check is a function that returns an error when evaluated.
type Check func() error

// CheckLazy evaluates makeErr only if not at cap; increments Checks when evaluated.
func (g *Guard) CheckLazy(makeErr func() error) {
	if makeErr == nil {
		return
	}
	g.lock()
	if g.max > 0 && g.n >= g.max {
		g.unlock()
		return
	}
	g.checks++
	g.unlock()

	if err := makeErr(); err != nil {
		g.Add(err)
	}
}

// AddCheck increments Checks and evaluates f unless cap reached.
func (g *Guard) AddCheck(f Check) {
	if f == nil {
		return
	}
	g.lock()
	if g.max > 0 && g.n >= g.max {
		g.unlock()
		return
	}
	g.checks++
	g.unlock()

	if err := f(); err != nil {
		g.Add(err)
	}
}

// Run evaluates checks in order, stopping once cap is reached.
func (g *Guard) Run(checks ...Check) {
	for _, f := range checks {
		g.lock()
		reached := g.max > 0 && g.n >= g.max
		g.unlock()
		if reached {
			return
		}
		g.AddCheck(f)
	}
}

// ErrClamped indicates some errors were dropped due to cap.
var ErrClamped = errors.New("sanity:errors_clamped")

// ErrorsClampedError reports how many errors were kept vs dropped.
type ErrorsClampedError struct {
	Kept, Dropped int
}

func (e ErrorsClampedError) Unwrap() error { return ErrClamped }
func (e ErrorsClampedError) Error() string {
	return fmt.Sprintf("validation: %d additional errors omitted (kept %d)", e.Dropped, e.Kept)
}

// Err returns nil, a single error, or an aggregate snapshot.
// In thread-safe mode, it copies the backing slice under lock.
func (gd *Guard) Err() error {
	gd.lock()
	switch gd.n {
	case 0:
		gd.unlock()
		return nil
	case 1:
		e0, dropped := gd.e0, gd.dropped
		gd.unlock()
		if e0 == nil {
			return nil
		}
		if dropped == 0 {
			return e0
		}
		return multiError{e0: e0, more: []error{
			ErrorsClampedError{Kept: 1, Dropped: dropped},
		}}
	default:
		e0, e1, e2, e3, more, dropped := gd.snapshotErrorsLocked()
		gd.unlock()
		if dropped > 0 {
			kept := countNonNil4(e0, e1, e2, e3) + len(more)
			more = append(more, ErrorsClampedError{Kept: kept, Dropped: dropped})
		}
		return multiError{e0: e0, e1: e1, e2: e2, e3: e3, more: more}
	}
}

// snapshotErrorsLocked returns copies when needed while the lock is held.
// The returned 'more' is safe for the caller to append to.
func (gd *Guard) snapshotErrorsLocked() (error, error, error, error, []error, int) {
	e0, e1, e2, e3 := gd.e0, gd.e1, gd.e2, gd.e3
	m, dropped := gd.more, gd.dropped

	ratio := gd.compactRatio
	if ratio <= 0 {
		ratio = 2
	}
	threadsafe := gd.mu != nil
	needCopy := threadsafe || dropped > 0 ||
		(len(m) > 0 && cap(m) > ratio*len(m))

	if !needCopy {
		return e0, e1, e2, e3, m, dropped
	}

	want := len(m)
	if dropped > 0 {
		want++
	}
	out := make([]error, 0, want)
	if len(m) > 0 {
		out = append(out, m...)
	}
	return e0, e1, e2, e3, out, dropped
}

func countNonNil4(a, b, c, d error) (n int) {
	if a != nil {
		n++
	}
	if b != nil {
		n++
	}
	if c != nil {
		n++
	}
	if d != nil {
		n++
	}
	return
}

// ----- Group error: iterator + Is/As + Unwrap -----

// ErrorGroup is the aggregate interface.
type ErrorGroup interface {
	error
	Iter(func(error) bool) // return false to stop early
}

type multiError struct {
	e0, e1, e2, e3 error
	more           []error // immutable or safely copied
}

func (m multiError) Error() string { return "multiple errors" }

// Iter visits SSO (e0..e3) then all entries in 'more' in order.
func (m multiError) Iter(fn func(error) bool) {
	if m.e0 != nil && !fn(m.e0) {
		return
	}
	if m.e1 != nil && !fn(m.e1) {
		return
	}
	if m.e2 != nil && !fn(m.e2) {
		return
	}
	if m.e3 != nil && !fn(m.e3) {
		return
	}
	for i := 0; i < len(m.more); i++ {
		if !fn(m.more[i]) {
			return
		}
	}
}

// Is scans members; zero-alloc.
func (m multiError) Is(target error) bool {
	if target == nil {
		return false
	}
	if m.e0 != nil && errors.Is(m.e0, target) {
		return true
	}
	if m.e1 != nil && errors.Is(m.e1, target) {
		return true
	}
	if m.e2 != nil && errors.Is(m.e2, target) {
		return true
	}
	if m.e3 != nil && errors.Is(m.e3, target) {
		return true
	}
	for _, e := range m.more {
		if errors.Is(e, target) {
			return true
		}
	}
	return false
}

// As scans members; zero-alloc.
func (m multiError) As(target any) bool {
	if target == nil {
		return false
	}
	if m.e0 != nil && errors.As(m.e0, target) {
		return true
	}
	if m.e1 != nil && errors.As(m.e1, target) {
		return true
	}
	if m.e2 != nil && errors.As(m.e2, target) {
		return true
	}
	if m.e3 != nil && errors.As(m.e3, target) {
		return true
	}
	for _, e := range m.more {
		if errors.As(e, target) {
			return true
		}
	}
	return false
}

// Unwrap allocates a flat []error for interop with code expecting slices.
// errors.Is/As won't use this because multiError implements Is/As.
func (m multiError) Unwrap() []error {
	n := 0
	if m.e0 != nil {
		n++
	}
	if m.e1 != nil {
		n++
	}
	if m.e2 != nil {
		n++
	}
	if m.e3 != nil {
		n++
	}
	out := make([]error, 0, n+len(m.more))
	if m.e0 != nil {
		out = append(out, m.e0)
	}
	if m.e1 != nil {
		out = append(out, m.e1)
	}
	if m.e2 != nil {
		out = append(out, m.e2)
	}
	if m.e3 != nil {
		out = append(out, m.e3)
	}
	out = append(out, m.more...)
	return out
}
