# Sanity

## Description

**Sanity** is a tiny set of **zero‑allocation normalization helpers for Go**.
It standardizes the boring parts you write in every service:

* set a **default** when a value is **zero** or **nil**
* **clamp** numeric values into a safe `[min,max]` range (with bound‑swap safety)
* lightweight **predicates** and **string defaults**
* **pointer ergonomics** for optional fields
* **float sanitizers** (handle `NaN` / `±Inf`)
* simple **duration** wrappers

**New in v0.2.0:** **typed validation errors** with **category sentinels**, **introspection interfaces** for `errors.As`, and a **redaction build tag** to hide sensitive “got …” details in error strings (while keeping values programmatically accessible).

**No reflection. No magic.** Just explicit, type‑safe helpers designed for hot paths.

---

## Requirements

* **Go**: 1.24+

---

## Install

```bash
go get github.com/sessaidi/sanity@v0.2.0
```

---

## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/sessaidi/sanity"
)

type Config struct {
	Port       int
	Retry      int
	Timeout    time.Duration
	Mode       string
	MaxBitrate *int
}

func main() {
	// Defaults & clamping
	cfg := Config{
		Port:    0,               // missing
		Retry:   -1,              // invalid
		Timeout: 0,               // missing
		Mode:    "  ",            // blank
		// MaxBitrate: nil
	}

	// Default if zero
	sanity.SetIfZero(&cfg.Port, 8080) // -> 8080

	// Default if <= limit (normalize non-positive to 3)
	sanity.SetIfLE(&cfg.Retry, 0, 3) // -> 3

	// Default then clamp timeout into [500ms, 10s]
	sanity.SetIfZeroThenClamp(&cfg.Timeout, 2*time.Second, 500*time.Millisecond, 10*time.Second) // -> 2s

	// Clamp numeric ranges (swap if min>max)
	sanity.Clamp(&cfg.Port, 1, 65535)

	// String default if blank (trim-aware)
	cfg.Mode = sanity.DefaultIfBlank(cfg.Mode, "prod") // -> "prod"

	// Optional pointer ergonomics
	bitrate := sanity.POrDefault(cfg.MaxBitrate, 3_000_000) // -> 3,000,000

	fmt.Println(cfg.Port, cfg.Retry, cfg.Timeout, cfg.Mode, bitrate)
}
```

### Additional usage (v0.2.0 typed errors)

```go
package main

import (
	"errors"
	"fmt"

	"github.com/sessaidi/sanity"
)

func validatePort(name string, port int) error {
	if port < 1 || port > 65535 {
		return sanity.OutOfRangeError[int]{Field: name, Min: 1, Max: 65535, Got: port}
	}
	return nil
}

func main() {
	err := validatePort("port", 0)
	if err != nil {
		// Category check (errors.Is):
		if errors.Is(err, sanity.ErrOutOfRange) {
			fmt.Println("category: out_of_range")
		}
		// Introspection (errors.As):
		var r sanity.RangeError
		if errors.As(err, &r) {
			min, max := r.Bounds()
			fmt.Printf("%s must be in [%v,%v], got=%v\n", r.FieldName(), min, max, r.Value())
		}
	}
}
```

**Redacted strings build:**

```bash
# removes “got …” from error strings in LenAtLeastError / OutOfRangeError[T]
go test -tags=redact ./...
```

---

## Documentation

> **Conventions**
>
> * Mutating helpers take a **pointer** (e.g., `*int`).
> * Numeric operations use a `Numeric` **constraint** (numbers only—no strings).
> * **Bounds safety:** if `min > max`, helpers **swap** to `[max,min]` instead of panicking.
> * **Booleans:** avoid `SetIfZero` for `bool` unless `false` truly means “unset”; prefer `*bool` + `SetIfNil`.

---

## v0.2.0 — Typed Errors, Sentinels, Introspection & Redaction

### Overview

v0.2.0 adds structured error types designed to compose with standard Go errors:

* **Types**:
  `NotNilError`, `NonZeroError`, `NonEmptyError`, `LenAtLeastError`, `OutOfRangeError[T]`, `NotInSetError`

* **Category sentinels** (for `errors.Is`):
  `ErrNotNil`, `ErrNonZero`, `ErrNonEmpty`, `ErrLenAtLeast`, `ErrOutOfRange`, `ErrNotInSet`

* **Introspection interfaces** (for `errors.As`):

   * `FieldError` → `FieldName() string`
   * `RangeError` → `FieldName() string`, `Bounds() (min any, max any)`, `Value() any`
     *(implemented by `OutOfRangeError[T]`)*

* **Redaction build tag**:
  Build with `-tags=redact` to hide **“got …”** in error strings for:

   * `LenAtLeastError`
   * `OutOfRangeError[T]`
     Programmatic access remains available via `RangeError`.

---

### Error types

#### NotNilError

**Type**

```go
type NotNilError struct{ Field string }
```

**Implements**

* `error`
* `FieldError` (`FieldName() string`)
* `Unwrap() error` → `ErrNotNil`

**String format (verbose & redacted)**

```
<field>: must not be nil
```

**Example**

```go
err := sanity.NotNilError{Field: "client"}
errors.Is(err, sanity.ErrNotNil) // true
```

---

#### NonZeroError

**Type**

```go
type NonZeroError struct{ Field string }
```

**Implements**

* `error`
* `FieldError`
* `Unwrap() error` → `ErrNonZero`

**String format**

```
<field>: must be non-zero
```

---

#### NonEmptyError

**Type**

```go
type NonEmptyError struct{ Field string }
```

**Implements**

* `error`
* `FieldError`
* `Unwrap() error` → `ErrNonEmpty`

**String format**

```
<field>: must be non-empty
```

---

#### NotInSetError

**Type**

```go
type NotInSetError struct{ Field string }
```

**Implements**

* `error`
* `FieldError`
* `Unwrap() error` → `ErrNotInSet`

**String format**

```
<field>: invalid value
```

---

#### LenAtLeastError

**Type**

```go
type LenAtLeastError struct {
    Field   string
    Want    int
	Got     int
}
```

**Implements**

* `error`
* `FieldError`
* `Unwrap() error` → `ErrLenAtLeast`

**String format**

* **verbose (`!redact`)**:
  `<field>: len must be >= <want> (got <got>)`
* **redacted (`redact`)**:
  `<field>: len must be >= <want>`

---

#### OutOfRangeError[T]

**Type**

```go
type OutOfRangeError[T any] struct {
    Field   string
    Min     T
	Max     T
    Got     T
}
```

**Implements**

* `error`
* `FieldError`
* `RangeError` (`Bounds() (any, any)`, `Value() any`)
* `Unwrap() error` → `ErrOutOfRange`

**String format**

* **verbose (`!redact`)**:
  `<field>: must be in [<min>,<max>], got <got>`
* **redacted (`redact`)**:
  `<field>: must be in [<min>,<max>]`

**Example**

```go
e := sanity.OutOfRangeError[int]{Field: "port", Min: 1, Max: 10, Got: 0}
if errors.Is(e, sanity.ErrOutOfRange) {
    var r sanity.RangeError
    if errors.As(e, &r) {
        min, max := r.Bounds()
        got := r.Value()
        _ = min
		_ = max
		_ = got
    }
}
```

---

## v0.1.0 — Defaults & Clamping

### Defaults & Clamping

#### SetIfZero

**Synopsis**

```go
func SetIfZero[T comparable](p *T, def T)
```

**Description**
If `*p` equals the zero value of `T`, assign `def` to `*p`. Otherwise do nothing.
Zero value follows Go semantics: `0`, `""`, `false`, `time.Duration(0)`, zeroed struct, etc.

**Parameters**

* `p`: pointer to the value to normalize.
* `def`: default value used only when `*p` is zero.

**Returns**
None (mutates `*p` in place).

**Edge cases / Notes**

* Using with `bool` is allowed but often ambiguous; prefer `*bool` + `SetIfNil`.
* `T` must be `comparable` (slices/maps/functions are not).

**Example**

```go
port := 0
sanity.SetIfZero(&port, 8080) // port = 8080
```

---

#### SetIfNil

**Synopsis**

```go
func SetIfNil[T any](p **T, def *T)
```

**Description**
If `*p == nil`, set `*p = def`. Useful for optional pointers.

**Parameters**

* `p`: pointer to a pointer you want to default.
* `def`: default pointer used only when `*p` is nil.

**Returns**
None (mutates `*p`).

**Edge cases / Notes**

* `def` is not copied; you pass ownership of the pointer reference.
* Safe for all `T`, including large structs.

**Example**

```go
type Limits struct{ Max int }
var lim *Limits
def := &Limits{Max: 100}
sanity.SetIfNil(&lim, def) // lim -> def
```

---

#### SetIfLE / SetIfLT / SetIfGE / SetIfGT

**Synopsis**

```go
func SetIfLE[T Numeric](p *T, limit, def T) // if *p <= limit
func SetIfLT[T Numeric](p *T, limit, def T) // if *p <  limit
func SetIfGE[T Numeric](p *T, limit, def T) // if *p >= limit
func SetIfGT[T Numeric](p *T, limit, def T) // if *p >  limit
```

**Description**
Conditionally set `*p = def` when the corresponding comparison holds. No change otherwise.

**Parameters**

* `p`: pointer to numeric value.
* `limit`: comparison threshold.
* `def`: default value assigned when condition is true.

**Returns**
None (mutates `*p`).

**Edge cases / Notes**

* `Numeric` excludes strings to prevent lexicographic mistakes.
* For floats, comparisons work as usual; note `NaN` never compares true. Sanitize with `ClampFinite` if needed.

**Example**

```go
retries := 0
sanity.SetIfLE(&retries, 0, 3) // retries = 3
```

---

#### SetIfZeroThenClamp

**Synopsis**

```go
func SetIfZeroThenClamp[T Numeric](p *T, def, min, max T)
```

**Description**
Two‑step normalization:

1. If `*p` is zero → set to `def`.
2. Clamp `*p` into inclusive `[min,max]`. If `min > max`, bounds are swapped.

**Parameters**

* `p`: pointer to numeric value.
* `def`: default used only if zero.
* `min`, `max`: inclusive clamp bounds (auto‑swapped if inverted).

**Returns**
None (mutates `*p`).

**Edge cases / Notes**

* Floats: `NaN`/`Inf` may bypass comparisons; use `ClampFinite` first if needed.
* No panic on misordered bounds; they’re swapped.

**Example**

```go
v := 0
sanity.SetIfZeroThenClamp(&v, 100, 10, 1) // swap -> [1,10]; v becomes 10
```

---

#### Clamp

**Synopsis**

```go
func Clamp[T Numeric](p *T, min, max T)
```

**Description**
Clamp `*p` into inclusive `[min,max]`. If `min > max`, swap to `[max,min]`.

**Parameters**

* `p`: pointer to numeric value.
* `min`, `max`: inclusive bounds.

**Returns**
None (mutates `*p`).

**Edge cases / Notes**

* Floats with `NaN`/`Inf` won’t compare normally; sanitize first if required.

**Example**

```go
x := 99
sanity.Clamp(&x, 1, 10) // x = 10
```

---

#### DefaultIf

**Synopsis**

```go
func DefaultIf[T comparable](v, def T) T
```

**Description**
Return `def` if `v` equals the zero value of `T`; otherwise return `v`. (Does not mutate inputs.)

**Parameters**

* `v`: value to check.
* `def`: default returned when `v` is zero.

**Returns**
`T`: either `def` (if zero) or `v`.

**Edge cases / Notes**

* Same caveats as `SetIfZero` (e.g., `bool` ambiguity).

**Example**

```go
mode := sanity.DefaultIf("", "prod") // "prod"
```

---

#### DefaultIfClamp

**Synopsis**

```go
func DefaultIfClamp[T Numeric](v, def, min, max T) T
```

**Description**
Return a normalized value by applying:

1. default‑if‑zero → `def`, then
2. clamp into `[min,max]` (auto‑swap if `min > max`).
   (Does not mutate inputs.)

**Parameters**

* `v`: input value.
* `def`: used only if `v` is zero.
* `min`, `max`: inclusive clamp bounds.

**Returns**
`T`: normalized value.

**Edge cases / Notes**

* Floats with `NaN`/`Inf` may need `ClampFinite` first.

**Example**

```go
n := sanity.DefaultIfClamp(0, 5, 1, 3) // -> 3
```

---

#### InRange

**Synopsis**

```go
func InRange[T Numeric](v, min, max T) bool
```

**Description**
Return `true` if `v` lies in inclusive `[min,max]`. If `min > max`, bounds are swapped.

**Parameters**

* `v`: value to test.
* `min`, `max`: intended bounds.

**Returns**
`bool`: `true` if `v ∈ [min,max]`, else `false`.

**Edge cases / Notes**

* Floats: beware of `NaN`—comparisons are always false.

**Example**

```go
ok := sanity.InRange(5, 1, 10) // true
```

---

#### DefaultIfBlank

**Synopsis**

```go
func DefaultIfBlank(v, def string) string
```

**Description**
Trim leading/trailing Unicode whitespace; if the result is empty, return `def`, else return the original `v`.

**Parameters**

* `v`: input string.
* `def`: default string used if `v` is blank after `TrimSpace`.

**Returns**
`string`: either `v` or `def`.

**Edge cases / Notes**

* Uses `strings.TrimSpace` (Unicode‑aware).

**Example**

```go
s := sanity.DefaultIfBlank(" \t", "prod") // "prod"
```

---

### Duration helpers

#### ClampDuration

**Synopsis**

```go
func ClampDuration(p *time.Duration, min, max time.Duration)
```

**Description**
Readability wrapper around `Clamp` specialized for `time.Duration`.

**Parameters**

* `p`, `min`, `max`: duration value and bounds (inclusive; auto‑swap if inverted).

**Returns**
None.

**Example**

```go
d := 3 * time.Second
sanity.ClampDuration(&d, 1*time.Second, 2*time.Second) // d = 2s
```

---

#### DefaultDurationClamp

**Synopsis**

```go
func DefaultDurationClamp(v, def, min, max time.Duration) time.Duration
```

**Description**
Return‑by‑value `Duration` version of `DefaultIfClamp`: default if zero, then clamp.

**Parameters**

* `v`, `def`, `min`, `max`: as above.

**Returns**
`time.Duration`: normalized value.

**Example**

```go
to := sanity.DefaultDurationClamp(0, 2*time.Second, time.Second, 3*time.Second) // 2s
```

---

### Pointer ergonomics

#### P

**Synopsis**

```go
func P[T any](ptr *T) T
```

**Description**
Dereference `ptr` if it’s non‑nil; otherwise return the zero value of `T`.

**Parameters**

* `ptr`: pointer to read.

**Returns**
`T`: dereferenced value or zero value.

**Example**

```go
var p *int
v := sanity.P(p) // 0
```

---

#### Ptr

**Synopsis**

```go
func Ptr[T any](value T) *T
```

**Description**
Return the address of `value`. Handy for literal pointers in config builders.

**Parameters**

* `value`: any value.

**Returns**
`*T`: pointer to a heap‑escaped copy of `value`.

**Edge cases / Notes**

* This **allocates** (addressable value must live long enough), so avoid in tight loops.

**Example**

```go
p := sanity.Ptr(42) // *int
```

---

#### POrDefault

**Synopsis**

```go
func POrDefault[T any](ptr *T, defaultVal T) T
```

**Description**
Dereference `ptr` if non‑nil; otherwise return `defaultVal`.

**Parameters**

* `ptr`: pointer to read (may be nil).
* `defaultVal`: value to return if `ptr` is nil.

**Returns**
`T`: `*ptr` or `defaultVal`.

**Example**

```go
var s *string
v := sanity.POrDefault(s, "prod") // "prod"
```

---

### Float sanitizers

> These accept both `float32` and `float64` via a `Float` constraint.

#### ZeroIfNaN

**Synopsis**

```go
func ZeroIfNaN[T ~float32 | ~float64](v T) T
```

**Description**
Return `0` if `v` is `NaN`; otherwise return `v`.

**Parameters**

* `v`: input float.

**Returns**
`T`: `0` if `NaN`, else `v`.

**Notes**

* Uses `math.IsNaN(float64(v))`; works for `float32` and `float64`.

**Example**

```go
x := sanity.ZeroIfNaN[float64](math.NaN()) // 0
```

---

#### DefaultIfNaN

**Synopsis**

```go
func DefaultIfNaN[T ~float32 | ~float64](v, def T) T
```

**Description**
Return `def` if `v` is `NaN`; otherwise return `v`.

**Parameters**

* `v`: input float.
* `def`: default when `v` is `NaN`.

**Returns**
`T`: normalized value.

**Example**

```go
x := sanity.DefaultIfNaN[float64](math.NaN(), 7) // 7
```

---

#### ClampFinite

**Synopsis**

```go
func ClampFinite[T ~float32 | ~float64](p *T, def T)
```

**Description**
If `*p` is `NaN` or `±Inf`, set `*p = def`. Otherwise leave `*p` unchanged.

**Parameters**

* `p`: pointer to float to sanitize.
* `def`: replacement for non‑finite values.

**Returns**
None (mutates `*p`).

**Example**

```go
v := math.Inf(1)
sanity.ClampFinite(&v, 0) // v = 0
```

---

## Notes & caveats

* **Numeric vs string semantics**: clamping/range helpers accept only numeric types (no strings), preventing lexicographic surprises.
* **Bounds safety**: if `min > max`, bounds are **swapped** to keep the call safe in production.
* **Booleans**: `SetIfZero` works for `bool` but can be ambiguous—prefer `*bool` + `SetIfNil` for optional flags.
* **Zero‑alloc**: all helpers are allocation‑free in normal use (except `Ptr`, which allocates to create an address).

---

## Tests & Benchmarks

```bash
go test ./...
go test -tags=redact ./...  # redacted error strings
```

---
