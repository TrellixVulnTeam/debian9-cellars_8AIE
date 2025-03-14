// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/mips64/galign.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/mips64/galign.go:1
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mips64

import (
	"bootstrap/cmd/compile/internal/gc"
	"bootstrap/cmd/compile/internal/ssa"
	"bootstrap/cmd/internal/obj"
	"bootstrap/cmd/internal/obj/mips"
)

func Init() {
	gc.Thearch.LinkArch = &mips.Linkmips64
	if obj.GOARCH == "mips64le" {
		gc.Thearch.LinkArch = &mips.Linkmips64le
	}
	gc.Thearch.REGSP = mips.REGSP
	gc.Thearch.MAXWIDTH = 1 << 50

	gc.Thearch.Defframe = defframe
	gc.Thearch.Proginfo = proginfo

	gc.Thearch.SSAMarkMoves = func(s *gc.SSAGenState, b *ssa.Block) {}
	gc.Thearch.SSAGenValue = ssaGenValue
	gc.Thearch.SSAGenBlock = ssaGenBlock
	gc.Thearch.ZeroAuto = zeroAuto
}
