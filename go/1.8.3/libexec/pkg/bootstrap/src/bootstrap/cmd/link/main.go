// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/link/main.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/link/main.go:1
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bootstrap/cmd/internal/obj"
	"bootstrap/cmd/link/internal/amd64"
	"bootstrap/cmd/link/internal/arm"
	"bootstrap/cmd/link/internal/arm64"
	"bootstrap/cmd/link/internal/ld"
	"bootstrap/cmd/link/internal/mips"
	"bootstrap/cmd/link/internal/mips64"
	"bootstrap/cmd/link/internal/ppc64"
	"bootstrap/cmd/link/internal/s390x"
	"bootstrap/cmd/link/internal/x86"
	"fmt"
	"os"
)

// The bulk of the linker implementation lives in cmd/link/internal/ld.
// Architecture-specific code lives in cmd/link/internal/GOARCH.
//
// Program initialization:
//
// Before any argument parsing is done, the Init function of relevant
// architecture package is called. The only job done in Init is
// configuration of the ld.Thearch and ld.SysArch variables.
//
// Then control flow passes to ld.Main, which parses flags, makes
// some configuration decisions, and then gives the architecture
// packages a second chance to modify the linker's configuration
// via the ld.Thearch.Archinit function.

func main() {
	switch obj.GOARCH {
	default:
		fmt.Fprintf(os.Stderr, "link: unknown architecture %q\n", obj.GOARCH)
		os.Exit(2)
	case "386":
		x86.Init()
	case "amd64", "amd64p32":
		amd64.Init()
	case "arm":
		arm.Init()
	case "arm64":
		arm64.Init()
	case "mips", "mipsle":
		mips.Init()
	case "mips64", "mips64le":
		mips64.Init()
	case "ppc64", "ppc64le":
		ppc64.Init()
	case "s390x":
		s390x.Init()
	}
	ld.Main()
}
