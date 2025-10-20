package sanity_test

import (
	"testing"
	"time"

	"github.com/sessaidi/sanity"
)

func BenchmarkValidators(b *testing.B) {
	// Varied inputs to defeat constant-propagation
	nonEmpty := []string{"a", "bb", "ccc", "dddd"}
	empty := []string{"", "", "", ""}
	okNums := []int{1, 2, 3, 4, 5, 6, 7, 8}
	badNums := []int{-5, -4, -3, -2, -1, -6, -7, -8}
	okDur := []time.Duration{1500 * time.Millisecond, 1100 * time.Millisecond, 1999 * time.Millisecond}
	badDur := []time.Duration{500 * time.Millisecond, 0, 900 * time.Millisecond}
	set2 := map[string]struct{}{"auto": {}, "manual": {}}
	okModes := []string{"auto", "manual", "auto", "manual"}
	badModes := []string{"x", "q", "z", "zz"}

	b.Run("NonEmpty/OK", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s := nonEmpty[i&3]
			if err := sanity.NonEmpty("s", s); err != nil {
				b.Fatal("unexpected")
			}
		}
	})

	b.Run("NonEmpty/Fail", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s := empty[i&3]
			sinkErr = sanity.NonEmpty("s", s)
			blackbox(sinkErr)
		}
	})

	b.Run("NotNilPtr/OK", func(b *testing.B) {
		xs := [8]int{1, 2, 3, 4, 5, 6, 7, 8}
		for i := 0; i < b.N; i++ {
			p := &xs[i&7]
			if err := sanity.NotNilPtr("p", p); err != nil {
				b.Fatal("unexpected")
			}
		}
	})

	b.Run("NotNilPtr/Fail", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var p *int // nil but still not compile-time varied
			sinkErr = sanity.NotNilPtr("p", p)
			blackbox(sinkErr)
		}
	})

	b.Run("InRangeNum/OK", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v := okNums[i&7]
			if err := sanity.InRangeNum("n", v, 0, 10); err != nil {
				b.Fatal("unexpected")
			}
		}
	})

	b.Run("InRangeNum/Fail", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v := badNums[i&7]
			sinkErr = sanity.InRangeNum("n", v, 0, 10)
			blackbox(sinkErr)
		}
	})

	b.Run("InSet/OK", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mode := okModes[i&3]
			if err := sanity.InSet("mode", mode, set2); err != nil {
				b.Fatal("unexpected")
			}
		}
	})

	b.Run("InSet/Fail", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mode := badModes[i&3]
			sinkErr = sanity.InSet("mode", mode, set2)
			blackbox(sinkErr)
		}
	})

	b.Run("InRangeDuration/OK", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			d := okDur[i%len(okDur)]
			if err := sanity.InRangeDuration("d", d, time.Second, 2*time.Second); err != nil {
				b.Fatal("unexpected")
			}
		}
	})

	b.Run("InRangeDuration/Fail", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			d := badDur[i%len(badDur)]
			sinkErr = sanity.InRangeDuration("d", d, time.Second, 2*time.Second)
			blackbox(sinkErr)
		}
	})
}
