// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/internal/obj/typekind.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/internal/obj/typekind.go:1
// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package obj

// Must match runtime and reflect.
// Included by cmd/gc.

const (
	KindBool = 1 + iota
	KindInt
	KindInt8
	KindInt16
	KindInt32
	KindInt64
	KindUint
	KindUint8
	KindUint16
	KindUint32
	KindUint64
	KindUintptr
	KindFloat32
	KindFloat64
	KindComplex64
	KindComplex128
	KindArray
	KindChan
	KindFunc
	KindInterface
	KindMap
	KindPtr
	KindSlice
	KindString
	KindStruct
	KindUnsafePointer
	KindDirectIface = 1 << 5
	KindGCProg      = 1 << 6
	KindNoPointers  = 1 << 7
	KindMask        = (1 << 5) - 1
)
