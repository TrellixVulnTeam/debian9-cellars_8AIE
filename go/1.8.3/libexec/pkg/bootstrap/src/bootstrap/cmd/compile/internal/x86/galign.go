// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/x86/galign.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/x86/galign.go:1
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x86

import (
	"bootstrap/cmd/compile/internal/gc"
	"bootstrap/cmd/internal/obj"
	"bootstrap/cmd/internal/obj/x86"
	"fmt"
	"os"
)

func Init() {
	gc.Thearch.LinkArch = &x86.Link386
	gc.Thearch.REGSP = x86.REGSP
	switch v := obj.GO386; v {
	case "387":
		gc.Thearch.Use387 = true
	case "sse2":
	default:
		fmt.Fprintf(os.Stderr, "unsupported setting GO386=%s\n", v)
		gc.Exit(1)
	}
	gc.Thearch.MAXWIDTH = (1 << 32) - 1

	gc.Thearch.Defframe = defframe
	gc.Thearch.Proginfo = proginfo

	gc.Thearch.SSAMarkMoves = ssaMarkMoves
	gc.Thearch.SSAGenValue = ssaGenValue
	gc.Thearch.SSAGenBlock = ssaGenBlock
	gc.Thearch.ZeroAuto = zeroAuto
}
