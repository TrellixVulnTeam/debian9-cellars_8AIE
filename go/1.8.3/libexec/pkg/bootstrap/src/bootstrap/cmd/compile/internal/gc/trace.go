// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/gc/trace.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/gc/trace.go:1
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.7

package gc

import (
	"os"
	tracepkg "runtime/trace"
)

func init() {
	traceHandler = traceHandlerGo17
}

func traceHandlerGo17(traceprofile string) {
	f, err := os.Create(traceprofile)
	if err != nil {
		Fatalf("%v", err)
	}
	if err := tracepkg.Start(f); err != nil {
		Fatalf("%v", err)
	}
	atExit(tracepkg.Stop)
}
