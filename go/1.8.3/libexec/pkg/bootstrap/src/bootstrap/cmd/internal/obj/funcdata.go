// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/internal/obj/funcdata.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/internal/obj/funcdata.go:1
// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package obj

// This file defines the IDs for PCDATA and FUNCDATA instructions
// in Go binaries. It is included by assembly sources, so it must
// be written using #defines.
//
// The Go compiler also #includes this file, for now.
//
// symtab.go also contains a copy of these constants.

// Pseudo-assembly statements.

// GO_ARGS, GO_RESULTS_INITIALIZED, and NO_LOCAL_POINTERS are macros
// that communicate to the runtime information about the location and liveness
// of pointers in an assembly function's arguments, results, and stack frame.
// This communication is only required in assembly functions that make calls
// to other functions that might be preempted or grow the stack.
// NOSPLIT functions that make no calls do not need to use these macros.

// GO_ARGS indicates that the Go prototype for this assembly function
// defines the pointer map for the function's arguments.
// GO_ARGS should be the first instruction in a function that uses it.
// It can be omitted if there are no arguments at all.
// GO_ARGS is inserted implicitly by the linker for any function
// that also has a Go prototype and therefore is usually not necessary
// to write explicitly.

// GO_RESULTS_INITIALIZED indicates that the assembly function
// has initialized the stack space for its results and that those results
// should be considered live for the remainder of the function.

// NO_LOCAL_POINTERS indicates that the assembly function stores
// no pointers to heap objects in its local stack variables.

// ArgsSizeUnknown is set in Func.argsize to mark all functions
// whose argument size is unknown (C vararg functions, and
// assembly code without an explicit specification).
// This value is generated by the compiler, assembler, or linker.
const (
	PCDATA_StackMapIndex       = 0
	FUNCDATA_ArgsPointerMaps   = 0
	FUNCDATA_LocalsPointerMaps = 1
	ArgsSizeUnknown            = -0x80000000
)
