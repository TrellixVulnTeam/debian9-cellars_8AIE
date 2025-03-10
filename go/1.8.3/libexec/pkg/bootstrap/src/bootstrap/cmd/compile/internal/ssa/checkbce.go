// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/ssa/checkbce.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/ssa/checkbce.go:1
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// checkbce prints all bounds checks that are present in the function.
// Useful to find regressions. checkbce is only activated when with
// corresponding debug options, so it's off by default.
// See test/checkbce.go
func checkbce(f *Func) {
	if f.pass.debug <= 0 {
		return
	}

	for _, b := range f.Blocks {
		for _, v := range b.Values {
			if v.Op == OpIsInBounds || v.Op == OpIsSliceInBounds {
				f.Config.Warnl(v.Line, "Found %v", v.Op)
			}
		}
	}
}
