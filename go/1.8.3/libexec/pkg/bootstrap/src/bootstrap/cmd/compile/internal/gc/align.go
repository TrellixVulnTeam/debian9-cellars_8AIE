// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/gc/align.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/gc/align.go:1
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

// machine size and rounding alignment is dictated around
// the size of a pointer, set in betypeinit (see ../amd64/galign.go).
var defercalc int

func Rnd(o int64, r int64) int64 {
	if r < 1 || r > 8 || r&(r-1) != 0 {
		Fatalf("rnd %d", r)
	}
	return (o + r - 1) &^ (r - 1)
}

func offmod(t *Type) {
	o := int32(0)
	for _, f := range t.Fields().Slice() {
		f.Offset = int64(o)
		o += int32(Widthptr)
		if int64(o) >= Thearch.MAXWIDTH {
			yyerror("interface too large")
			o = int32(Widthptr)
		}
	}
}

func widstruct(errtype *Type, t *Type, o int64, flag int) int64 {
	starto := o
	maxalign := int32(flag)
	if maxalign < 1 {
		maxalign = 1
	}
	lastzero := int64(0)
	var w int64
	for _, f := range t.Fields().Slice() {
		if f.Type == nil {
			// broken field, just skip it so that other valid fields
			// get a width.
			continue
		}

		dowidth(f.Type)
		if int32(f.Type.Align) > maxalign {
			maxalign = int32(f.Type.Align)
		}
		if f.Type.Width < 0 {
			Fatalf("invalid width %d", f.Type.Width)
		}
		w = f.Type.Width
		if f.Type.Align > 0 {
			o = Rnd(o, int64(f.Type.Align))
		}
		f.Offset = o
		if f.Nname != nil {
			// addrescapes has similar code to update these offsets.
			// Usually addrescapes runs after widstruct,
			// in which case we could drop this,
			// but function closure functions are the exception.
			// NOTE(rsc): This comment may be stale.
			// It's possible the ordering has changed and this is
			// now the common case. I'm not sure.
			if f.Nname.Name.Param.Stackcopy != nil {
				f.Nname.Name.Param.Stackcopy.Xoffset = o
				f.Nname.Xoffset = 0
			} else {
				f.Nname.Xoffset = o
			}
		}

		if w == 0 {
			lastzero = o
		}
		o += w
		if o >= Thearch.MAXWIDTH {
			yyerror("type %L too large", errtype)
			o = 8 // small but nonzero
		}
	}

	// For nonzero-sized structs which end in a zero-sized thing, we add
	// an extra byte of padding to the type. This padding ensures that
	// taking the address of the zero-sized thing can't manufacture a
	// pointer to the next object in the heap. See issue 9401.
	if flag == 1 && o > starto && o == lastzero {
		o++
	}

	// final width is rounded
	if flag != 0 {
		o = Rnd(o, int64(maxalign))
	}
	t.Align = uint8(maxalign)

	// type width only includes back to first field's offset
	t.Width = o - starto

	return o
}

func dowidth(t *Type) {
	if Widthptr == 0 {
		Fatalf("dowidth without betypeinit")
	}

	if t == nil {
		return
	}

	if t.Width > 0 {
		if t.Align == 0 {
			// See issue 11354
			Fatalf("zero alignment with nonzero size %v", t)
		}
		return
	}

	if t.Width == -2 {
		if !t.Broke {
			t.Broke = true
			yyerrorl(t.Lineno, "invalid recursive type %v", t)
		}

		t.Width = 0
		return
	}

	// break infinite recursion if the broken recursive type
	// is referenced again
	if t.Broke && t.Width == 0 {
		return
	}

	// defer checkwidth calls until after we're done
	defercalc++

	lno := lineno
	lineno = t.Lineno
	t.Width = -2
	t.Align = 0

	et := t.Etype
	switch et {
	case TFUNC, TCHAN, TMAP, TSTRING:
		break

	// simtype == 0 during bootstrap
	default:
		if simtype[t.Etype] != 0 {
			et = simtype[t.Etype]
		}
	}

	w := int64(0)
	switch et {
	default:
		Fatalf("dowidth: unknown type: %v", t)

	// compiler-specific stuff
	case TINT8, TUINT8, TBOOL:
		// bool is int8
		w = 1

	case TINT16, TUINT16:
		w = 2

	case TINT32, TUINT32, TFLOAT32:
		w = 4

	case TINT64, TUINT64, TFLOAT64:
		w = 8
		t.Align = uint8(Widthreg)

	case TCOMPLEX64:
		w = 8
		t.Align = 4

	case TCOMPLEX128:
		w = 16
		t.Align = uint8(Widthreg)

	case TPTR32:
		w = 4
		checkwidth(t.Elem())

	case TPTR64:
		w = 8
		checkwidth(t.Elem())

	case TUNSAFEPTR:
		w = int64(Widthptr)

	case TINTER: // implemented as 2 pointers
		w = 2 * int64(Widthptr)

		t.Align = uint8(Widthptr)
		offmod(t)

	case TCHAN: // implemented as pointer
		w = int64(Widthptr)

		checkwidth(t.Elem())

		// make fake type to check later to
		// trigger channel argument check.
		t1 := typChanArgs(t)
		checkwidth(t1)

	case TCHANARGS:
		t1 := t.ChanArgs()
		dowidth(t1) // just in case
		if t1.Elem().Width >= 1<<16 {
			yyerror("channel element type too large (>64kB)")
		}
		t.Width = 1

	case TMAP: // implemented as pointer
		w = int64(Widthptr)
		checkwidth(t.Val())
		checkwidth(t.Key())

	case TFORW: // should have been filled in
		if !t.Broke {
			yyerror("invalid recursive type %v", t)
		}
		w = 1 // anything will do

	case TANY:
		// dummy type; should be replaced before use.
		Fatalf("dowidth any")

	case TSTRING:
		if sizeof_String == 0 {
			Fatalf("early dowidth string")
		}
		w = int64(sizeof_String)
		t.Align = uint8(Widthptr)

	case TARRAY:
		if t.Elem() == nil {
			break
		}
		if t.isDDDArray() {
			if !t.Broke {
				yyerror("use of [...] array outside of array literal")
				t.Broke = true
			}
			break
		}

		dowidth(t.Elem())
		if t.Elem().Width != 0 {
			cap := (uint64(Thearch.MAXWIDTH) - 1) / uint64(t.Elem().Width)
			if uint64(t.NumElem()) > cap {
				yyerror("type %L larger than address space", t)
			}
		}
		w = t.NumElem() * t.Elem().Width
		t.Align = t.Elem().Align

	case TSLICE:
		if t.Elem() == nil {
			break
		}
		w = int64(sizeof_Array)
		checkwidth(t.Elem())
		t.Align = uint8(Widthptr)

	case TSTRUCT:
		if t.IsFuncArgStruct() {
			Fatalf("dowidth fn struct %v", t)
		}
		w = widstruct(t, t, 0, 1)

	// make fake type to check later to
	// trigger function argument computation.
	case TFUNC:
		t1 := typFuncArgs(t)
		checkwidth(t1)
		w = int64(Widthptr) // width of func type is pointer

	// function is 3 cated structures;
	// compute their widths as side-effect.
	case TFUNCARGS:
		t1 := t.FuncArgs()
		w = widstruct(t1, t1.Recvs(), 0, 0)
		w = widstruct(t1, t1.Params(), w, Widthreg)
		w = widstruct(t1, t1.Results(), w, Widthreg)
		t1.Extra.(*FuncType).Argwid = w
		if w%int64(Widthreg) != 0 {
			Warn("bad type %v %d\n", t1, w)
		}
		t.Align = 1
	}

	if Widthptr == 4 && w != int64(int32(w)) {
		yyerror("type %v too large", t)
	}

	t.Width = w
	if t.Align == 0 {
		if w > 8 || w&(w-1) != 0 {
			Fatalf("invalid alignment for %v", t)
		}
		t.Align = uint8(w)
	}

	lineno = lno

	if defercalc == 1 {
		resumecheckwidth()
	} else {
		defercalc--
	}
}

// when a type's width should be known, we call checkwidth
// to compute it.  during a declaration like
//
//	type T *struct { next T }
//
// it is necessary to defer the calculation of the struct width
// until after T has been initialized to be a pointer to that struct.
// similarly, during import processing structs may be used
// before their definition.  in those situations, calling
// defercheckwidth() stops width calculations until
// resumecheckwidth() is called, at which point all the
// checkwidths that were deferred are executed.
// dowidth should only be called when the type's size
// is needed immediately.  checkwidth makes sure the
// size is evaluated eventually.

var deferredTypeStack []*Type

func checkwidth(t *Type) {
	if t == nil {
		return
	}

	// function arg structs should not be checked
	// outside of the enclosing function.
	if t.IsFuncArgStruct() {
		Fatalf("checkwidth %v", t)
	}

	if defercalc == 0 {
		dowidth(t)
		return
	}

	if t.Deferwidth {
		return
	}
	t.Deferwidth = true

	deferredTypeStack = append(deferredTypeStack, t)
}

func defercheckwidth() {
	// we get out of sync on syntax errors, so don't be pedantic.
	if defercalc != 0 && nerrors == 0 {
		Fatalf("defercheckwidth")
	}
	defercalc = 1
}

func resumecheckwidth() {
	if defercalc == 0 {
		Fatalf("resumecheckwidth")
	}
	for len(deferredTypeStack) > 0 {
		t := deferredTypeStack[len(deferredTypeStack)-1]
		deferredTypeStack = deferredTypeStack[:len(deferredTypeStack)-1]
		t.Deferwidth = false
		dowidth(t)
	}

	defercalc = 0
}
