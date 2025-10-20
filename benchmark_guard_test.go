package sanity_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sessaidi/sanity"
)

var sinkErr error
var sinkInt int

func blackbox[T any](v T) { sinkInt++ }

func BenchmarkGuard(b *testing.B) {
	b.Run("FirstError/OK/Const", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			g := sanity.NewGuard()
			g.Check(sanity.NotNilPtr("client", new(int)))
			g.Check(sanity.NonEmpty("mode", "auto"))
			g.Check(sanity.InRangeNum("temp", 12.34, 0.0, 30.0))
			if g.Err() != nil {
				b.Fatal("unexpected error")
			}
		}
	})

	b.Run("FirstError/OK/Varied", func(b *testing.B) {
		modes := []string{"auto", "manual"}
		temps := []float64{12.34, 18.9, 27.1}
		for i := 0; i < b.N; i++ {
			g := sanity.NewGuard()
			mode := modes[i&1]
			tp := temps[i%len(temps)]
			g.Check(sanity.NotNilPtr("client", new(int)))
			g.Check(sanity.NonEmpty("mode", mode))
			g.Check(sanity.InRangeNum("temp", tp, 0.0, 30.0))
			sinkErr = g.Err()
			blackbox(sinkErr)
		}
	})

	b.Run("FirstError/Fail/Eager", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			g := sanity.NewGuard()
			g.Check(sanity.NonEmpty("env", "")) // fails immediately
			g.Check(sanity.NonZero("n", 0))     // wasted work after failure
			sinkErr = g.Err()
			blackbox(sinkErr)
		}
	})

	b.Run("FirstError/Fail/Lazy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			g := sanity.NewGuard()
			g.CheckLazy(func() error { return sanity.NonEmpty("env", "") }) // fails
			g.CheckLazy(func() error { return sanity.NonZero("n", 0) })     // not evaluated
			sinkErr = g.Err()
			blackbox(sinkErr)
		}
	})

	b.Run("Collector/OK/Unlimited", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			g := sanity.NewGuard(sanity.WithMaxErrors(0))
			g.Run(
				func() error { return sanity.NonEmpty("name", "x") },
				func() error { return sanity.InRangeNum("age", 42, 0, 130) },
				func() error { return sanity.InRangeDuration("d", 1500*time.Millisecond, time.Second, 2*time.Second) },
			)
			if g.Err() != nil {
				b.Fatal("unexpected error")
			}
		}
	})

	b.Run("Collector/Fail/Cap=4/Eager", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			g := sanity.NewGuard(sanity.WithMaxErrors(4))
			g.Check(sanity.NonEmpty("a", ""))
			g.Check(sanity.NonZero("b", 0))
			g.Check(sanity.InRangeNum("c", -1, 0, 1))
			g.Check(sanity.StrLenAtLeast("d", "ab", 3))
			g.Check(sanity.NonEmpty("e", "")) // built but dropped due to cap
			sinkErr = g.Err()
		}
	})

	b.Run("Collector/Fail/Cap=4/Lazy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			g := sanity.NewGuard(sanity.WithMaxErrors(4))
			g.Run(
				func() error { return sanity.NonEmpty("a", "") },
				func() error { return sanity.NonZero("b", 0) },
				func() error { return sanity.InRangeNum("c", -1, 0, 1) },
				func() error { return sanity.StrLenAtLeast("d", "ab", 3) },
				func() error { return sanity.NonEmpty("e", "") }, // not evaluated once cap is hit
			)
			sinkErr = g.Err()
		}
	})

	b.Run("Collector/IsAs", func(b *testing.B) {
		g := sanity.NewGuard(sanity.WithMaxErrors(0))
		g.Add(sanity.NonEmpty("env", ""))
		g.Add(sanity.NonZero("port", 0))
		g.Add(sanity.InRangeNum("timeout", 0, 1, 5))
		err := g.Err()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if !errors.Is(err, sanity.ErrNonEmpty) ||
				!errors.Is(err, sanity.ErrNonZero) ||
				!errors.Is(err, sanity.ErrOutOfRange) {
				b.Fatal("missing categories")
			}
			var re sanity.RangeError
			if !errors.As(err, &re) {
				b.Fatal("missing RangeError via As")
			}
		}
	})

	b.Run("Collector/Iter7", func(b *testing.B) {
		g := sanity.NewGuard(sanity.WithMaxErrors(0))
		for i := 0; i < 7; i++ {
			g.Add(sanity.NonEmpty("f", ""))
		}
		err := g.Err()
		eg, ok := err.(sanity.ErrorGroup)
		if !ok {
			b.Fatal("no ErrorGroup")
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			eg.Iter(func(e error) bool { count++; return true })
			sinkInt += count
		}
	})
}
