// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/math/big/floatconv.go

//line /tmp/go-20170524-43289-mangl1/go/src/math/big/floatconv.go:1
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements string-to-Float conversion functions.

package big

import (
	"fmt"
	"io"
	"strings"
)

var floatZero Float

// SetString sets z to the value of s and returns z and a boolean indicating
// success. s must be a floating-point number of the same format as accepted
// by Parse, with base argument 0. The entire string (not just a prefix) must
// be valid for success. If the operation failed, the value of z is undefined
// but the returned value is nil.
func (z *Float) SetString(s string) (*Float, bool) {
	if f, _, err := z.Parse(s, 0); err == nil {
		return f, true
	}
	return nil, false
}

// scan is like Parse but reads the longest possible prefix representing a valid
// floating point number from an io.ByteScanner rather than a string. It serves
// as the implementation of Parse. It does not recognize ±Inf and does not expect
// EOF at the end.
func (z *Float) scan(r io.ByteScanner, base int) (f *Float, b int, err error) {
	prec := z.prec
	if prec == 0 {
		prec = 64
	}

	// A reasonable value in case of an error.
	z.form = zero

	// sign
	z.neg, err = scanSign(r)
	if err != nil {
		return
	}

	// mantissa
	var fcount int // fractional digit count; valid if <= 0
	z.mant, b, fcount, err = z.mant.scan(r, base, true)
	if err != nil {
		return
	}

	// exponent
	var exp int64
	var ebase int
	exp, ebase, err = scanExponent(r, true)
	if err != nil {
		return
	}

	// special-case 0
	if len(z.mant) == 0 {
		z.prec = prec
		z.acc = Exact
		z.form = zero
		f = z
		return
	}
	// len(z.mant) > 0

	// The mantissa may have a decimal point (fcount <= 0) and there
	// may be a nonzero exponent exp. The decimal point amounts to a
	// division by b**(-fcount). An exponent means multiplication by
	// ebase**exp. Finally, mantissa normalization (shift left) requires
	// a correcting multiplication by 2**(-shiftcount). Multiplications
	// are commutative, so we can apply them in any order as long as there
	// is no loss of precision. We only have powers of 2 and 10, and
	// we split powers of 10 into the product of the same powers of
	// 2 and 5. This reduces the size of the multiplication factor
	// needed for base-10 exponents.

	// normalize mantissa and determine initial exponent contributions
	exp2 := int64(len(z.mant))*_W - fnorm(z.mant)
	exp5 := int64(0)

	// determine binary or decimal exponent contribution of decimal point
	if fcount < 0 {
		// The mantissa has a "decimal" point ddd.dddd; and
		// -fcount is the number of digits to the right of '.'.
		// Adjust relevant exponent accordingly.
		d := int64(fcount)
		switch b {
		case 10:
			exp5 = d
			fallthrough // 10**e == 5**e * 2**e
		case 2:
			exp2 += d
		case 16:
			exp2 += d * 4 // hexadecimal digits are 4 bits each
		default:
			panic("unexpected mantissa base")
		}
		// fcount consumed - not needed anymore
	}

	// take actual exponent into account
	switch ebase {
	case 10:
		exp5 += exp
		fallthrough
	case 2:
		exp2 += exp
	default:
		panic("unexpected exponent base")
	}
	// exp consumed - not needed anymore

	// apply 2**exp2
	if MinExp <= exp2 && exp2 <= MaxExp {
		z.prec = prec
		z.form = finite
		z.exp = int32(exp2)
		f = z
	} else {
		err = fmt.Errorf("exponent overflow")
		return
	}

	if exp5 == 0 {
		// no decimal exponent contribution
		z.round(0)
		return
	}
	// exp5 != 0

	// apply 5**exp5
	p := new(Float).SetPrec(z.Prec() + 64) // use more bits for p -- TODO(gri) what is the right number?
	if exp5 < 0 {
		z.Quo(z, p.pow5(uint64(-exp5)))
	} else {
		z.Mul(z, p.pow5(uint64(exp5)))
	}

	return
}

// These powers of 5 fit into a uint64.
//
//	for p, q := uint64(0), uint64(1); p < q; p, q = q, q*5 {
//		fmt.Println(q)
//	}
//
var pow5tab = [...]uint64{
	1,
	5,
	25,
	125,
	625,
	3125,
	15625,
	78125,
	390625,
	1953125,
	9765625,
	48828125,
	244140625,
	1220703125,
	6103515625,
	30517578125,
	152587890625,
	762939453125,
	3814697265625,
	19073486328125,
	95367431640625,
	476837158203125,
	2384185791015625,
	11920928955078125,
	59604644775390625,
	298023223876953125,
	1490116119384765625,
	7450580596923828125,
}

// pow5 sets z to 5**n and returns z.
// n must not be negative.
func (z *Float) pow5(n uint64) *Float {
	const m = uint64(len(pow5tab) - 1)
	if n <= m {
		return z.SetUint64(pow5tab[n])
	}
	// n > m

	z.SetUint64(pow5tab[m])
	n -= m

	// use more bits for f than for z
	// TODO(gri) what is the right number?
	f := new(Float).SetPrec(z.Prec() + 64).SetUint64(5)

	for n > 0 {
		if n&1 != 0 {
			z.Mul(z, f)
		}
		f.Mul(f, f)
		n >>= 1
	}

	return z
}

// Parse parses s which must contain a text representation of a floating-
// point number with a mantissa in the given conversion base (the exponent
// is always a decimal number), or a string representing an infinite value.
//
// It sets z to the (possibly rounded) value of the corresponding floating-
// point value, and returns z, the actual base b, and an error err, if any.
// The entire string (not just a prefix) must be consumed for success.
// If z's precision is 0, it is changed to 64 before rounding takes effect.
// The number must be of the form:
//
//	number   = [ sign ] [ prefix ] mantissa [ exponent ] | infinity .
//	sign     = "+" | "-" .
//	prefix   = "0" ( "x" | "X" | "b" | "B" ) .
//	mantissa = digits | digits "." [ digits ] | "." digits .
//	exponent = ( "E" | "e" | "p" ) [ sign ] digits .
//	digits   = digit { digit } .
//	digit    = "0" ... "9" | "a" ... "z" | "A" ... "Z" .
//	infinity = [ sign ] ( "inf" | "Inf" ) .
//
// The base argument must be 0, 2, 10, or 16. Providing an invalid base
// argument will lead to a run-time panic.
//
// For base 0, the number prefix determines the actual base: A prefix of
// "0x" or "0X" selects base 16, and a "0b" or "0B" prefix selects
// base 2; otherwise, the actual base is 10 and no prefix is accepted.
// The octal prefix "0" is not supported (a leading "0" is simply
// considered a "0").
//
// A "p" exponent indicates a binary (rather then decimal) exponent;
// for instance "0x1.fffffffffffffp1023" (using base 0) represents the
// maximum float64 value. For hexadecimal mantissae, the exponent must
// be binary, if present (an "e" or "E" exponent indicator cannot be
// distinguished from a mantissa digit).
//
// The returned *Float f is nil and the value of z is valid but not
// defined if an error is reported.
//
func (z *Float) Parse(s string, base int) (f *Float, b int, err error) {
	// scan doesn't handle ±Inf
	if len(s) == 3 && (s == "Inf" || s == "inf") {
		f = z.SetInf(false)
		return
	}
	if len(s) == 4 && (s[0] == '+' || s[0] == '-') && (s[1:] == "Inf" || s[1:] == "inf") {
		f = z.SetInf(s[0] == '-')
		return
	}

	r := strings.NewReader(s)
	if f, b, err = z.scan(r, base); err != nil {
		return
	}

	// entire string must have been consumed
	if ch, err2 := r.ReadByte(); err2 == nil {
		err = fmt.Errorf("expected end of string, found %q", ch)
	} else if err2 != io.EOF {
		err = err2
	}

	return
}

// ParseFloat is like f.Parse(s, base) with f set to the given precision
// and rounding mode.
func ParseFloat(s string, base int, prec uint, mode RoundingMode) (f *Float, b int, err error) {
	return new(Float).SetPrec(prec).SetMode(mode).Parse(s, base)
}

var _ fmt.Scanner = &floatZero // *Float must implement fmt.Scanner

// Scan is a support routine for fmt.Scanner; it sets z to the value of
// the scanned number. It accepts formats whose verbs are supported by
// fmt.Scan for floating point values, which are:
// 'b' (binary), 'e', 'E', 'f', 'F', 'g' and 'G'.
// Scan doesn't handle ±Inf.
func (z *Float) Scan(s fmt.ScanState, ch rune) error {
	s.SkipSpace()
	_, _, err := z.scan(byteReader{s}, 0)
	return err
}
