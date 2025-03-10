// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/math/big/ratmarsh.go

//line /tmp/go-20170524-43289-mangl1/go/src/math/big/ratmarsh.go:1
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements encoding/decoding of Rats.

package big

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Gob codec version. Permits backward-compatible changes to the encoding.
const ratGobVersion byte = 1

// GobEncode implements the gob.GobEncoder interface.
func (x *Rat) GobEncode() ([]byte, error) {
	if x == nil {
		return nil, nil
	}
	buf := make([]byte, 1+4+(len(x.a.abs)+len(x.b.abs))*_S) // extra bytes for version and sign bit (1), and numerator length (4)
	i := x.b.abs.bytes(buf)
	j := x.a.abs.bytes(buf[:i])
	n := i - j
	if int(uint32(n)) != n {
		// this should never happen
		return nil, errors.New("Rat.GobEncode: numerator too large")
	}
	binary.BigEndian.PutUint32(buf[j-4:j], uint32(n))
	j -= 1 + 4
	b := ratGobVersion << 1 // make space for sign bit
	if x.a.neg {
		b |= 1
	}
	buf[j] = b
	return buf[j:], nil
}

// GobDecode implements the gob.GobDecoder interface.
func (z *Rat) GobDecode(buf []byte) error {
	if len(buf) == 0 {
		// Other side sent a nil or default value.
		*z = Rat{}
		return nil
	}
	b := buf[0]
	if b>>1 != ratGobVersion {
		return fmt.Errorf("Rat.GobDecode: encoding version %d not supported", b>>1)
	}
	const j = 1 + 4
	i := j + binary.BigEndian.Uint32(buf[j-4:j])
	z.a.neg = b&1 != 0
	z.a.abs = z.a.abs.setBytes(buf[j:i])
	z.b.abs = z.b.abs.setBytes(buf[i:])
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (x *Rat) MarshalText() (text []byte, err error) {
	// TODO(gri): get rid of the []byte/string conversion
	return []byte(x.RatString()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (z *Rat) UnmarshalText(text []byte) error {
	// TODO(gri): get rid of the []byte/string conversion
	if _, ok := z.SetString(string(text)); !ok {
		return fmt.Errorf("math/big: cannot unmarshal %q into a *big.Rat", text)
	}
	return nil
}
