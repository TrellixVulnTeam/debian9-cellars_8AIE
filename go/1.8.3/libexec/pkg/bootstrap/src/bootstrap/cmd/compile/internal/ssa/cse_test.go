// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/ssa/cse_test.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/ssa/cse_test.go:1
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import "testing"

type tstAux struct {
	s string
}

// This tests for a bug found when partitioning, but not sorting by the Aux value.
func TestCSEAuxPartitionBug(t *testing.T) {
	c := testConfig(t)
	arg1Aux := &tstAux{"arg1-aux"}
	arg2Aux := &tstAux{"arg2-aux"}
	arg3Aux := &tstAux{"arg3-aux"}

	// construct lots of values with args that have aux values and place
	// them in an order that triggers the bug
	fun := Fun(c, "entry",
		Bloc("entry",
			Valu("start", OpInitMem, TypeMem, 0, nil),
			Valu("sp", OpSP, TypeBytePtr, 0, nil),
			Valu("r7", OpAdd64, TypeInt64, 0, nil, "arg3", "arg1"),
			Valu("r1", OpAdd64, TypeInt64, 0, nil, "arg1", "arg2"),
			Valu("arg1", OpArg, TypeInt64, 0, arg1Aux),
			Valu("arg2", OpArg, TypeInt64, 0, arg2Aux),
			Valu("arg3", OpArg, TypeInt64, 0, arg3Aux),
			Valu("r9", OpAdd64, TypeInt64, 0, nil, "r7", "r8"),
			Valu("r4", OpAdd64, TypeInt64, 0, nil, "r1", "r2"),
			Valu("r8", OpAdd64, TypeInt64, 0, nil, "arg3", "arg2"),
			Valu("r2", OpAdd64, TypeInt64, 0, nil, "arg1", "arg2"),
			Valu("raddr", OpAddr, TypeInt64Ptr, 0, nil, "sp"),
			Valu("raddrdef", OpVarDef, TypeMem, 0, nil, "start"),
			Valu("r6", OpAdd64, TypeInt64, 0, nil, "r4", "r5"),
			Valu("r3", OpAdd64, TypeInt64, 0, nil, "arg1", "arg2"),
			Valu("r5", OpAdd64, TypeInt64, 0, nil, "r2", "r3"),
			Valu("r10", OpAdd64, TypeInt64, 0, nil, "r6", "r9"),
			Valu("rstore", OpStore, TypeMem, 8, nil, "raddr", "r10", "raddrdef"),
			Goto("exit")),
		Bloc("exit",
			Exit("rstore")))

	CheckFunc(fun.f)
	cse(fun.f)
	deadcode(fun.f)
	CheckFunc(fun.f)

	s1Cnt := 2
	// r1 == r2 == r3, needs to remove two of this set
	s2Cnt := 1
	// r4 == r5, needs to remove one of these
	for k, v := range fun.values {
		if v.Op == OpInvalid {
			switch k {
			case "r1":
				fallthrough
			case "r2":
				fallthrough
			case "r3":
				if s1Cnt == 0 {
					t.Errorf("cse removed all of r1,r2,r3")
				}
				s1Cnt--

			case "r4":
				fallthrough
			case "r5":
				if s2Cnt == 0 {
					t.Errorf("cse removed all of r4,r5")
				}
				s2Cnt--
			default:
				t.Errorf("cse removed %s, but shouldn't have", k)
			}
		}
	}

	if s1Cnt != 0 || s2Cnt != 0 {
		t.Errorf("%d values missed during cse", s1Cnt+s2Cnt)
	}
}

// TestZCSE tests the zero arg cse.
func TestZCSE(t *testing.T) {
	c := testConfig(t)

	fun := Fun(c, "entry",
		Bloc("entry",
			Valu("start", OpInitMem, TypeMem, 0, nil),
			Valu("sp", OpSP, TypeBytePtr, 0, nil),
			Valu("sb1", OpSB, TypeBytePtr, 0, nil),
			Valu("sb2", OpSB, TypeBytePtr, 0, nil),
			Valu("addr1", OpAddr, TypeInt64Ptr, 0, nil, "sb1"),
			Valu("addr2", OpAddr, TypeInt64Ptr, 0, nil, "sb2"),
			Valu("a1ld", OpLoad, TypeInt64, 0, nil, "addr1", "start"),
			Valu("a2ld", OpLoad, TypeInt64, 0, nil, "addr2", "start"),
			Valu("c1", OpConst64, TypeInt64, 1, nil),
			Valu("r1", OpAdd64, TypeInt64, 0, nil, "a1ld", "c1"),
			Valu("c2", OpConst64, TypeInt64, 1, nil),
			Valu("r2", OpAdd64, TypeInt64, 0, nil, "a2ld", "c2"),
			Valu("r3", OpAdd64, TypeInt64, 0, nil, "r1", "r2"),
			Valu("raddr", OpAddr, TypeInt64Ptr, 0, nil, "sp"),
			Valu("raddrdef", OpVarDef, TypeMem, 0, nil, "start"),
			Valu("rstore", OpStore, TypeMem, 8, nil, "raddr", "r3", "raddrdef"),
			Goto("exit")),
		Bloc("exit",
			Exit("rstore")))

	CheckFunc(fun.f)
	zcse(fun.f)
	deadcode(fun.f)
	CheckFunc(fun.f)

	if fun.values["c1"].Op != OpInvalid && fun.values["c2"].Op != OpInvalid {
		t.Errorf("zsce should have removed c1 or c2")
	}
	if fun.values["sb1"].Op != OpInvalid && fun.values["sb2"].Op != OpInvalid {
		t.Errorf("zsce should have removed sb1 or sb2")
	}
}
