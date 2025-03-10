// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/ssa/decompose.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/ssa/decompose.go:1
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// decompose converts phi ops on compound builtin types into phi
// ops on simple types.
// (The remaining compound ops are decomposed with rewrite rules.)
func decomposeBuiltIn(f *Func) {
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			if v.Op != OpPhi {
				continue
			}
			decomposeBuiltInPhi(v)
		}
	}

	// Split up named values into their components.
	// NOTE: the component values we are making are dead at this point.
	// We must do the opt pass before any deadcode elimination or we will
	// lose the name->value correspondence.
	var newNames []LocalSlot
	for _, name := range f.Names {
		t := name.Type
		switch {
		case t.IsInteger() && t.Size() == 8 && f.Config.IntSize == 4:
			var elemType Type
			if t.IsSigned() {
				elemType = f.Config.fe.TypeInt32()
			} else {
				elemType = f.Config.fe.TypeUInt32()
			}
			hiName, loName := f.Config.fe.SplitInt64(name)
			newNames = append(newNames, hiName, loName)
			for _, v := range f.NamedValues[name] {
				hi := v.Block.NewValue1(v.Line, OpInt64Hi, elemType, v)
				lo := v.Block.NewValue1(v.Line, OpInt64Lo, f.Config.fe.TypeUInt32(), v)
				f.NamedValues[hiName] = append(f.NamedValues[hiName], hi)
				f.NamedValues[loName] = append(f.NamedValues[loName], lo)
			}
			delete(f.NamedValues, name)
		case t.IsComplex():
			var elemType Type
			if t.Size() == 16 {
				elemType = f.Config.fe.TypeFloat64()
			} else {
				elemType = f.Config.fe.TypeFloat32()
			}
			rName, iName := f.Config.fe.SplitComplex(name)
			newNames = append(newNames, rName, iName)
			for _, v := range f.NamedValues[name] {
				r := v.Block.NewValue1(v.Line, OpComplexReal, elemType, v)
				i := v.Block.NewValue1(v.Line, OpComplexImag, elemType, v)
				f.NamedValues[rName] = append(f.NamedValues[rName], r)
				f.NamedValues[iName] = append(f.NamedValues[iName], i)
			}
			delete(f.NamedValues, name)
		case t.IsString():
			ptrType := f.Config.fe.TypeBytePtr()
			lenType := f.Config.fe.TypeInt()
			ptrName, lenName := f.Config.fe.SplitString(name)
			newNames = append(newNames, ptrName, lenName)
			for _, v := range f.NamedValues[name] {
				ptr := v.Block.NewValue1(v.Line, OpStringPtr, ptrType, v)
				len := v.Block.NewValue1(v.Line, OpStringLen, lenType, v)
				f.NamedValues[ptrName] = append(f.NamedValues[ptrName], ptr)
				f.NamedValues[lenName] = append(f.NamedValues[lenName], len)
			}
			delete(f.NamedValues, name)
		case t.IsSlice():
			ptrType := f.Config.fe.TypeBytePtr()
			lenType := f.Config.fe.TypeInt()
			ptrName, lenName, capName := f.Config.fe.SplitSlice(name)
			newNames = append(newNames, ptrName, lenName, capName)
			for _, v := range f.NamedValues[name] {
				ptr := v.Block.NewValue1(v.Line, OpSlicePtr, ptrType, v)
				len := v.Block.NewValue1(v.Line, OpSliceLen, lenType, v)
				cap := v.Block.NewValue1(v.Line, OpSliceCap, lenType, v)
				f.NamedValues[ptrName] = append(f.NamedValues[ptrName], ptr)
				f.NamedValues[lenName] = append(f.NamedValues[lenName], len)
				f.NamedValues[capName] = append(f.NamedValues[capName], cap)
			}
			delete(f.NamedValues, name)
		case t.IsInterface():
			ptrType := f.Config.fe.TypeBytePtr()
			typeName, dataName := f.Config.fe.SplitInterface(name)
			newNames = append(newNames, typeName, dataName)
			for _, v := range f.NamedValues[name] {
				typ := v.Block.NewValue1(v.Line, OpITab, ptrType, v)
				data := v.Block.NewValue1(v.Line, OpIData, ptrType, v)
				f.NamedValues[typeName] = append(f.NamedValues[typeName], typ)
				f.NamedValues[dataName] = append(f.NamedValues[dataName], data)
			}
			delete(f.NamedValues, name)
		case t.IsFloat():
			// floats are never decomposed, even ones bigger than IntSize
		case t.Size() > f.Config.IntSize:
			f.Fatalf("undecomposed named type %v %v", name, t)
		default:
			newNames = append(newNames, name)
		}
	}
	f.Names = newNames
}

func decomposeBuiltInPhi(v *Value) {
	switch {
	case v.Type.IsInteger() && v.Type.Size() == 8 && v.Block.Func.Config.IntSize == 4:
		if v.Block.Func.Config.arch == "amd64p32" {
			// Even though ints are 32 bits, we have 64-bit ops.
			break
		}
		decomposeInt64Phi(v)
	case v.Type.IsComplex():
		decomposeComplexPhi(v)
	case v.Type.IsString():
		decomposeStringPhi(v)
	case v.Type.IsSlice():
		decomposeSlicePhi(v)
	case v.Type.IsInterface():
		decomposeInterfacePhi(v)
	case v.Type.IsFloat():
		// floats are never decomposed, even ones bigger than IntSize
	case v.Type.Size() > v.Block.Func.Config.IntSize:
		v.Fatalf("undecomposed type %s", v.Type)
	}
}

func decomposeStringPhi(v *Value) {
	fe := v.Block.Func.Config.fe
	ptrType := fe.TypeBytePtr()
	lenType := fe.TypeInt()

	ptr := v.Block.NewValue0(v.Line, OpPhi, ptrType)
	len := v.Block.NewValue0(v.Line, OpPhi, lenType)
	for _, a := range v.Args {
		ptr.AddArg(a.Block.NewValue1(v.Line, OpStringPtr, ptrType, a))
		len.AddArg(a.Block.NewValue1(v.Line, OpStringLen, lenType, a))
	}
	v.reset(OpStringMake)
	v.AddArg(ptr)
	v.AddArg(len)
}

func decomposeSlicePhi(v *Value) {
	fe := v.Block.Func.Config.fe
	ptrType := fe.TypeBytePtr()
	lenType := fe.TypeInt()

	ptr := v.Block.NewValue0(v.Line, OpPhi, ptrType)
	len := v.Block.NewValue0(v.Line, OpPhi, lenType)
	cap := v.Block.NewValue0(v.Line, OpPhi, lenType)
	for _, a := range v.Args {
		ptr.AddArg(a.Block.NewValue1(v.Line, OpSlicePtr, ptrType, a))
		len.AddArg(a.Block.NewValue1(v.Line, OpSliceLen, lenType, a))
		cap.AddArg(a.Block.NewValue1(v.Line, OpSliceCap, lenType, a))
	}
	v.reset(OpSliceMake)
	v.AddArg(ptr)
	v.AddArg(len)
	v.AddArg(cap)
}

func decomposeInt64Phi(v *Value) {
	fe := v.Block.Func.Config.fe
	var partType Type
	if v.Type.IsSigned() {
		partType = fe.TypeInt32()
	} else {
		partType = fe.TypeUInt32()
	}

	hi := v.Block.NewValue0(v.Line, OpPhi, partType)
	lo := v.Block.NewValue0(v.Line, OpPhi, fe.TypeUInt32())
	for _, a := range v.Args {
		hi.AddArg(a.Block.NewValue1(v.Line, OpInt64Hi, partType, a))
		lo.AddArg(a.Block.NewValue1(v.Line, OpInt64Lo, fe.TypeUInt32(), a))
	}
	v.reset(OpInt64Make)
	v.AddArg(hi)
	v.AddArg(lo)
}

func decomposeComplexPhi(v *Value) {
	fe := v.Block.Func.Config.fe
	var partType Type
	switch z := v.Type.Size(); z {
	case 8:
		partType = fe.TypeFloat32()
	case 16:
		partType = fe.TypeFloat64()
	default:
		v.Fatalf("decomposeComplexPhi: bad complex size %d", z)
	}

	real := v.Block.NewValue0(v.Line, OpPhi, partType)
	imag := v.Block.NewValue0(v.Line, OpPhi, partType)
	for _, a := range v.Args {
		real.AddArg(a.Block.NewValue1(v.Line, OpComplexReal, partType, a))
		imag.AddArg(a.Block.NewValue1(v.Line, OpComplexImag, partType, a))
	}
	v.reset(OpComplexMake)
	v.AddArg(real)
	v.AddArg(imag)
}

func decomposeInterfacePhi(v *Value) {
	ptrType := v.Block.Func.Config.fe.TypeBytePtr()

	itab := v.Block.NewValue0(v.Line, OpPhi, ptrType)
	data := v.Block.NewValue0(v.Line, OpPhi, ptrType)
	for _, a := range v.Args {
		itab.AddArg(a.Block.NewValue1(v.Line, OpITab, ptrType, a))
		data.AddArg(a.Block.NewValue1(v.Line, OpIData, ptrType, a))
	}
	v.reset(OpIMake)
	v.AddArg(itab)
	v.AddArg(data)
}

func decomposeUser(f *Func) {
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			if v.Op != OpPhi {
				continue
			}
			decomposeUserPhi(v)
		}
	}
	// Split up named values into their components.
	// NOTE: the component values we are making are dead at this point.
	// We must do the opt pass before any deadcode elimination or we will
	// lose the name->value correspondence.
	i := 0
	var fnames []LocalSlot
	var newNames []LocalSlot
	for _, name := range f.Names {
		t := name.Type
		switch {
		case t.IsStruct():
			n := t.NumFields()
			fnames = fnames[:0]
			for i := 0; i < n; i++ {
				fnames = append(fnames, f.Config.fe.SplitStruct(name, i))
			}
			for _, v := range f.NamedValues[name] {
				for i := 0; i < n; i++ {
					x := v.Block.NewValue1I(v.Line, OpStructSelect, t.FieldType(i), int64(i), v)
					f.NamedValues[fnames[i]] = append(f.NamedValues[fnames[i]], x)
				}
			}
			delete(f.NamedValues, name)
			newNames = append(newNames, fnames...)
		case t.IsArray():
			if t.NumElem() == 0 {
				// TODO(khr): Not sure what to do here.  Probably nothing.
				// Names for empty arrays aren't important.
				break
			}
			if t.NumElem() != 1 {
				f.Fatalf("array not of size 1")
			}
			elemName := f.Config.fe.SplitArray(name)
			for _, v := range f.NamedValues[name] {
				e := v.Block.NewValue1I(v.Line, OpArraySelect, t.ElemType(), 0, v)
				f.NamedValues[elemName] = append(f.NamedValues[elemName], e)
			}

		default:
			f.Names[i] = name
			i++
		}
	}
	f.Names = f.Names[:i]
	f.Names = append(f.Names, newNames...)
}

func decomposeUserPhi(v *Value) {
	switch {
	case v.Type.IsStruct():
		decomposeStructPhi(v)
	case v.Type.IsArray():
		decomposeArrayPhi(v)
	}
}

// decomposeStructPhi replaces phi-of-struct with structmake(phi-for-each-field),
// and then recursively decomposes the phis for each field.
func decomposeStructPhi(v *Value) {
	t := v.Type
	n := t.NumFields()
	var fields [MaxStruct]*Value
	for i := 0; i < n; i++ {
		fields[i] = v.Block.NewValue0(v.Line, OpPhi, t.FieldType(i))
	}
	for _, a := range v.Args {
		for i := 0; i < n; i++ {
			fields[i].AddArg(a.Block.NewValue1I(v.Line, OpStructSelect, t.FieldType(i), int64(i), a))
		}
	}
	v.reset(StructMakeOp(n))
	v.AddArgs(fields[:n]...)

	// Recursively decompose phis for each field.
	for _, f := range fields[:n] {
		decomposeUserPhi(f)
	}
}

// decomposeArrayPhi replaces phi-of-array with arraymake(phi-of-array-element),
// and then recursively decomposes the element phi.
func decomposeArrayPhi(v *Value) {
	t := v.Type
	if t.NumElem() == 0 {
		v.reset(OpArrayMake0)
		return
	}
	if t.NumElem() != 1 {
		v.Fatalf("SSAable array must have no more than 1 element")
	}
	elem := v.Block.NewValue0(v.Line, OpPhi, t.ElemType())
	for _, a := range v.Args {
		elem.AddArg(a.Block.NewValue1I(v.Line, OpArraySelect, t.ElemType(), 0, a))
	}
	v.reset(OpArrayMake1)
	v.AddArg(elem)

	// Recursively decompose elem phi.
	decomposeUserPhi(elem)
}

// MaxStruct is the maximum number of fields a struct
// can have and still be SSAable.
const MaxStruct = 4

// StructMakeOp returns the opcode to construct a struct with the
// given number of fields.
func StructMakeOp(nf int) Op {
	switch nf {
	case 0:
		return OpStructMake0
	case 1:
		return OpStructMake1
	case 2:
		return OpStructMake2
	case 3:
		return OpStructMake3
	case 4:
		return OpStructMake4
	}
	panic("too many fields in an SSAable struct")
}
