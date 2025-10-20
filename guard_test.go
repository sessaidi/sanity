package sanity_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sessaidi/sanity"
)

type twoBools struct{ A, B bool }
type fourBools struct{ A, B, C, D bool }
type clampInfo struct {
	Is      bool
	Kept    int
	Dropped int
}

type sentinelCounts struct {
	OK1 bool
	K1  int
	D1  int
	OK2 bool
	K2  int
	D2  int
}
type iterCounts struct{ All, First3 int }

func TestGuard(t *testing.T) {
	testCases := []struct {
		name     string
		function func() interface{}
		expected interface{}
	}{
		{
			name: "FirstError captures only the first failure",
			function: func() interface{} {
				g := sanity.NewGuard()              // max=1
				g.Check(sanity.NonEmpty("env", "")) // first
				g.Check(sanity.NonZero("port", 0))  // ignored
				err := g.Err()
				return twoBools{
					A: errors.Is(err, sanity.ErrNonEmpty),
					B: errors.Is(err, sanity.ErrNonZero), // should be false
				}
			},
			expected: twoBools{A: true, B: false},
		},
		{
			name: "Collector unlimited keeps categories and RangeError",
			function: func() interface{} {
				g := sanity.NewGuard(sanity.WithMaxErrors(0)) // unlimited
				g.Add(sanity.NonEmpty("env", ""))
				g.Add(sanity.NonZero("port", 0))
				g.Add(sanity.InRangeNum("timeout", 0, 1, 5))
				err := g.Err()
				var re sanity.RangeError
				return fourBools{
					A: errors.Is(err, sanity.ErrNonEmpty),
					B: errors.Is(err, sanity.ErrNonZero),
					C: errors.Is(err, sanity.ErrOutOfRange),
					D: errors.As(err, &re),
				}
			},
			expected: fourBools{A: true, B: true, C: true, D: true},
		},
		{
			name: "Cap=2 produces ErrClamped sentinel with counts (via Add)",
			function: func() interface{} {
				g := sanity.NewGuard(sanity.WithMaxErrors(2))
				g.Add(sanity.NonEmpty("a", ""))
				g.Add(sanity.NonZero("b", 0))
				g.Add(sanity.InRangeNum("c", -1, 0, 1)) // dropped due to cap
				err := g.Err()
				var ce sanity.ErrorsClampedError
				_ = errors.As(err, &ce)
				return clampInfo{
					Is:      errors.Is(err, sanity.ErrClamped),
					Kept:    ce.Kept,
					Dropped: ce.Dropped,
				}
			},
			expected: clampInfo{Is: true, Kept: 2, Dropped: 1},
		},
		{
			name: "Iter visits all (7) and early-stop after 3",
			function: func() interface{} {
				g := sanity.NewGuard(sanity.WithMaxErrors(0))
				for i := 0; i < 7; i++ {
					g.Add(sanity.NonEmpty(fmt.Sprintf("f%d", i), ""))
				}
				err := g.Err()
				var eg sanity.ErrorGroup
				ok := errors.As(err, &eg)
				if !ok {
					return iterCounts{All: -1, First3: -1}
				}
				all := 0
				eg.Iter(func(e error) bool { all++; return true })
				first3 := 0
				eg.Iter(func(e error) bool {
					first3++
					return first3 < 3
				})
				return iterCounts{All: all, First3: first3}
			},
			expected: iterCounts{All: 7, First3: 3},
		},
		{
			name: "Stats basic (Kept/Dropped) using cap=2 with Run (no drop, gating stops eval)",
			function: func() interface{} {
				g := sanity.NewGuard(sanity.WithMaxErrors(2))
				g.Run(
					func() error { return sanity.NonEmpty("a", "") },         // kept
					func() error { return sanity.NonZero("b", 0) },           // kept
					func() error { return sanity.InRangeNum("c", -1, 0, 1) }, // NOT evaluated (gated at cap)
				)
				_ = g.Err()
				st := g.Stats()
				return []int{st.Kept, st.Dropped}
			},
			expected: []int{2, 0}, // with Run, Dropped==0 by design
		},
		{
			name: "Stats Dropped increments when adding beyond cap (via Add)",
			function: func() interface{} {
				g := sanity.NewGuard(sanity.WithMaxErrors(2))
				g.Add(sanity.NonEmpty("a", ""))
				g.Add(sanity.NonZero("b", 0))
				g.Add(sanity.InRangeNum("c", -1, 0, 1)) // attempted -> dropped
				_ = g.Err()
				st := g.Stats()
				return []int{st.Kept, st.Dropped}
			},
			expected: []int{2, 1},
		}, {
			name: "Err idempotent: sentinel counts stable across repeated Err()",
			function: func() interface{} {
				g := sanity.NewGuard(sanity.WithMaxErrors(1))
				// One kept + one dropped => sentinel should appear in aggregate
				g.Add(sanity.NonEmpty("a", "")) // kept
				g.Add(sanity.NonZero("b", 0))   // dropped due to cap

				// First Err()
				err1 := g.Err()
				var ce1 sanity.ErrorsClampedError
				ok1 := errors.As(err1, &ce1)

				// Second Err() should yield the *same* sentinel counts (no accumulation).
				err2 := g.Err()
				var ce2 sanity.ErrorsClampedError
				ok2 := errors.As(err2, &ce2)

				return sentinelCounts{
					OK1: ok1, K1: ce1.Kept, D1: ce1.Dropped,
					OK2: ok2, K2: ce2.Kept, D2: ce2.Dropped,
				}
			},
			expected: sentinelCounts{
				OK1: true, K1: 1, D1: 1,
				OK2: true, K2: 1, D2: 1,
			},
		},
		{
			name: "CheckLazy increments Stats.Checks and keeps Failures/Kept coherent",
			function: func() interface{} {
				g := sanity.NewGuard(sanity.WithMaxErrors(2))
				g.CheckLazy(func() error { return sanity.NonEmpty("a", "") }) // fail
				g.CheckLazy(func() error { return sanity.NonZero("b", 0) })   // fail
				_ = g.Err()
				st := g.Stats()
				return []int{st.Checks, st.Failures, st.Kept, st.Dropped}
			},
			expected: []int{2, 2, 2, 0},
		},
		{
			name: "ThreadSafe concurrent adds respect cap and sentinel",
			function: func() interface{} {
				const workers, per = 8, 8
				g := sanity.NewGuard(sanity.WithMaxErrors(10), sanity.WithThreadSafe())
				var wg sync.WaitGroup
				wg.Add(workers)
				for w := 0; w < workers; w++ {
					go func(id int) {
						defer wg.Done()
						for i := 0; i < per; i++ {
							if i%2 == 0 {
								g.Add(sanity.NonZero(fmt.Sprintf("w%d.i%d", id, i), 0))
							} else {
								g.Add(sanity.InRangeNum(fmt.Sprintf("w%d.i%d", id, i), -1, 0, 1))
							}
						}
					}(w)
				}
				wg.Wait()
				err := g.Err()
				st := g.Stats()
				ok := st.Kept <= 10
				if st.Failures > st.Kept {
					ok = ok && errors.Is(err, sanity.ErrClamped)
				}
				return ok
			},
			expected: true,
		},
		{
			name: "Duration validator quick sanity",
			function: func() interface{} {
				return sanity.InRangeDuration("d", 1500*time.Millisecond, time.Second, 2*time.Second) == nil
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.function()
			assert.Equal(t, tc.expected, got)
		})
	}
}
