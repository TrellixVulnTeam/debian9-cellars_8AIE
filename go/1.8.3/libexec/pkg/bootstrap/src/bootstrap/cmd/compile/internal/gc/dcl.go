// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/gc/dcl.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/gc/dcl.go:1
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"bootstrap/cmd/internal/obj"
	"fmt"
	"sort"
	"strings"
)

// Declaration stack & operations

var externdcl []*Node

var blockgen int32 // max block number

var block int32 // current block number

// dclstack maintains a stack of shadowed symbol declarations so that
// popdcl can restore their declarations when a block scope ends.
// The stack is maintained as a linked list, using Sym's Link field.
//
// In practice, the "stack" actually ends up forming a tree: goto and label
// statements record the current state of dclstack so that checkgoto can
// validate that a goto statement does not jump over any declarations or
// into a new block scope.
//
// Finally, the Syms in this list are not "real" Syms as they don't actually
// represent object names. Sym is just a convenient type for saving shadowed
// Sym definitions, and only a subset of its fields are actually used.
var dclstack *Sym

func dcopy(a, b *Sym) {
	a.Pkg = b.Pkg
	a.Name = b.Name
	a.Def = b.Def
	a.Block = b.Block
	a.Lastlineno = b.Lastlineno
}

func push() *Sym {
	d := new(Sym)
	d.Lastlineno = lineno
	d.Link = dclstack
	dclstack = d
	return d
}

// pushdcl pushes the current declaration for symbol s (if any) so that
// it can be shadowed by a new declaration within a nested block scope.
func pushdcl(s *Sym) *Sym {
	d := push()
	dcopy(d, s)
	return d
}

// popdcl pops the innermost block scope and restores all symbol declarations
// to their previous state.
func popdcl() {
	d := dclstack
	for ; d != nil && d.Name != ""; d = d.Link {
		s := Pkglookup(d.Name, d.Pkg)
		lno := s.Lastlineno
		dcopy(s, d)
		d.Lastlineno = lno
	}

	if d == nil {
		Fatalf("popdcl: no mark")
	}

	dclstack = d.Link // pop mark
	block = d.Block
}

// markdcl records the start of a new block scope for declarations.
func markdcl() {
	d := push()
	d.Name = "" // used as a mark in fifo
	d.Block = block

	blockgen++
	block = blockgen
}

// keep around for debugging
func dumpdclstack() {
	i := 0
	for d := dclstack; d != nil; d = d.Link {
		fmt.Printf("%6d  %p", i, d)
		if d.Name != "" {
			fmt.Printf("  '%s'  %v\n", d.Name, Pkglookup(d.Name, d.Pkg))
		} else {
			fmt.Printf("  ---\n")
		}
		i++
	}
}

func testdclstack() {
	for d := dclstack; d != nil; d = d.Link {
		if d.Name == "" {
			if nerrors != 0 {
				errorexit()
			}
			yyerror("mark left on the stack")
		}
	}
}

// redeclare emits a diagnostic about symbol s being redeclared somewhere.
func redeclare(s *Sym, where string) {
	if s.Lastlineno == 0 {
		var tmp string
		if s.Origpkg != nil {
			tmp = s.Origpkg.Path
		} else {
			tmp = s.Pkg.Path
		}
		pkgstr := tmp
		yyerror("%v redeclared %s\n"+
			"\tprevious declaration during import %q", s, where, pkgstr)
	} else {
		line1 := lineno
		line2 := s.Lastlineno

		// When an import and a declaration collide in separate files,
		// present the import as the "redeclared", because the declaration
		// is visible where the import is, but not vice versa.
		// See issue 4510.
		if s.Def == nil {
			line2 = line1
			line1 = s.Lastlineno
		}

		yyerrorl(line1, "%v redeclared %s\n"+
			"\tprevious declaration at %v", s, where, linestr(line2))
	}
}

var vargen int

// declare individual names - var, typ, const

var declare_typegen int

// declare records that Node n declares symbol n.Sym in the specified
// declaration context.
func declare(n *Node, ctxt Class) {
	if ctxt == PDISCARD {
		return
	}

	if isblank(n) {
		return
	}

	if n.Name == nil {
		// named OLITERAL needs Name; most OLITERALs don't.
		n.Name = new(Name)
	}
	n.Lineno = lineno
	s := n.Sym

	// kludgy: typecheckok means we're past parsing. Eg genwrapper may declare out of package names later.
	if importpkg == nil && !typecheckok && s.Pkg != localpkg {
		yyerror("cannot declare name %v", s)
	}

	if ctxt == PEXTERN && s.Name == "init" {
		yyerror("cannot declare init - must be func")
	}

	gen := 0
	if ctxt == PEXTERN {
		externdcl = append(externdcl, n)
	} else {
		if Curfn == nil && ctxt == PAUTO {
			Fatalf("automatic outside function")
		}
		if Curfn != nil {
			Curfn.Func.Dcl = append(Curfn.Func.Dcl, n)
		}
		if n.Op == OTYPE {
			declare_typegen++
			gen = declare_typegen
		} else if n.Op == ONAME && ctxt == PAUTO && !strings.Contains(s.Name, "·") {
			vargen++
			gen = vargen
		}
		pushdcl(s)
		n.Name.Curfn = Curfn
	}

	if ctxt == PAUTO {
		n.Xoffset = 0
	}

	if s.Block == block {
		// functype will print errors about duplicate function arguments.
		// Don't repeat the error here.
		if ctxt != PPARAM && ctxt != PPARAMOUT {
			redeclare(s, "in this block")
		}
	}

	s.Block = block
	s.Lastlineno = lineno
	s.Def = n
	n.Name.Vargen = int32(gen)
	n.Name.Funcdepth = funcdepth
	n.Class = ctxt

	autoexport(n, ctxt)
}

func addvar(n *Node, t *Type, ctxt Class) {
	if n == nil || n.Sym == nil || (n.Op != ONAME && n.Op != ONONAME) || t == nil {
		Fatalf("addvar: n=%v t=%v nil", n, t)
	}

	n.Op = ONAME
	declare(n, ctxt)
	n.Type = t
}

// declare variables from grammar
// new_name_list (type | [type] = expr_list)
func variter(vl []*Node, t *Node, el []*Node) []*Node {
	var init []*Node
	doexpr := len(el) > 0

	if len(el) == 1 && len(vl) > 1 {
		e := el[0]
		as2 := nod(OAS2, nil, nil)
		as2.List.Set(vl)
		as2.Rlist.Set1(e)
		for _, v := range vl {
			v.Op = ONAME
			declare(v, dclcontext)
			v.Name.Param.Ntype = t
			v.Name.Defn = as2
			if funcdepth > 0 {
				init = append(init, nod(ODCL, v, nil))
			}
		}

		return append(init, as2)
	}

	for _, v := range vl {
		var e *Node
		if doexpr {
			if len(el) == 0 {
				yyerror("missing expression in var declaration")
				break
			}
			e = el[0]
			el = el[1:]
		}

		v.Op = ONAME
		declare(v, dclcontext)
		v.Name.Param.Ntype = t

		if e != nil || funcdepth > 0 || isblank(v) {
			if funcdepth > 0 {
				init = append(init, nod(ODCL, v, nil))
			}
			e = nod(OAS, v, e)
			init = append(init, e)
			if e.Right != nil {
				v.Name.Defn = e
			}
		}
	}

	if len(el) != 0 {
		yyerror("extra expression in var declaration")
	}
	return init
}

// declare constants from grammar
// new_name_list [[type] = expr_list]
func constiter(vl []*Node, t *Node, cl []*Node) []*Node {
	lno := int32(0) // default is to leave line number alone in listtreecopy
	if len(cl) == 0 {
		if t != nil {
			yyerror("const declaration cannot have type without expression")
		}
		cl = lastconst
		t = lasttype
		lno = vl[0].Lineno
	} else {
		lastconst = cl
		lasttype = t
	}
	clcopy := listtreecopy(cl, lno)

	var vv []*Node
	for _, v := range vl {
		if len(clcopy) == 0 {
			yyerror("missing value in const declaration")
			break
		}

		c := clcopy[0]
		clcopy = clcopy[1:]

		v.Op = OLITERAL
		declare(v, dclcontext)

		v.Name.Param.Ntype = t
		v.Name.Defn = c

		vv = append(vv, nod(ODCLCONST, v, nil))
	}

	if len(clcopy) != 0 {
		yyerror("extra expression in const declaration")
	}
	iota_ += 1
	return vv
}

// newname returns a new ONAME Node associated with symbol s.
func newname(s *Sym) *Node {
	if s == nil {
		Fatalf("newname nil")
	}
	n := nod(ONAME, nil, nil)
	n.Sym = s
	n.Addable = true
	n.Ullman = 1
	n.Xoffset = 0
	return n
}

// newnoname returns a new ONONAME Node associated with symbol s.
func newnoname(s *Sym) *Node {
	if s == nil {
		Fatalf("newnoname nil")
	}
	n := nod(ONONAME, nil, nil)
	n.Sym = s
	n.Addable = true
	n.Ullman = 1
	n.Xoffset = 0
	return n
}

// newfuncname generates a new name node for a function or method.
// TODO(rsc): Use an ODCLFUNC node instead. See comment in CL 7360.
func newfuncname(s *Sym) *Node {
	n := newname(s)
	n.Func = new(Func)
	n.Func.IsHiddenClosure = Curfn != nil
	return n
}

// this generates a new name node for a name
// being declared.
func dclname(s *Sym) *Node {
	n := newname(s)
	n.Op = ONONAME // caller will correct it
	return n
}

func typenod(t *Type) *Node {
	// if we copied another type with *t = *u
	// then t->nod might be out of date, so
	// check t->nod->type too
	if t.nod == nil || t.nod.Type != t {
		t.nod = nod(OTYPE, nil, nil)
		t.nod.Type = t
		t.nod.Sym = t.Sym
	}

	return t.nod
}

func anonfield(typ *Type) *Node {
	return nod(ODCLFIELD, nil, typenod(typ))
}

func namedfield(s string, typ *Type) *Node {
	return nod(ODCLFIELD, newname(lookup(s)), typenod(typ))
}

// oldname returns the Node that declares symbol s in the current scope.
// If no such Node currently exists, an ONONAME Node is returned instead.
func oldname(s *Sym) *Node {
	n := s.Def
	if n == nil {
		// Maybe a top-level declaration will come along later to
		// define s. resolve will check s.Def again once all input
		// source has been processed.
		n = newnoname(s)
		n.SetIota(iota_) // save current iota value in const declarations
		return n
	}

	if Curfn != nil && n.Op == ONAME && n.Name.Funcdepth > 0 && n.Name.Funcdepth != funcdepth {
		// Inner func is referring to var in outer func.
		//
		// TODO(rsc): If there is an outer variable x and we
		// are parsing x := 5 inside the closure, until we get to
		// the := it looks like a reference to the outer x so we'll
		// make x a closure variable unnecessarily.
		c := n.Name.Param.Innermost
		if c == nil || c.Name.Funcdepth != funcdepth {
			// Do not have a closure var for the active closure yet; make one.
			c = nod(ONAME, nil, nil)
			c.Sym = s
			c.Class = PAUTOHEAP
			c.setIsClosureVar(true)
			c.Isddd = n.Isddd
			c.Name.Defn = n
			c.Addable = false
			c.Ullman = 2
			c.Name.Funcdepth = funcdepth

			// Link into list of active closure variables.
			// Popped from list in func closurebody.
			c.Name.Param.Outer = n.Name.Param.Innermost
			n.Name.Param.Innermost = c

			c.Xoffset = 0
			Curfn.Func.Cvars.Append(c)
		}

		// return ref to closure var, not original
		return c
	}

	return n
}

// := declarations
func colasname(n *Node) bool {
	switch n.Op {
	case ONAME,
		ONONAME,
		OPACK,
		OTYPE,
		OLITERAL:
		return n.Sym != nil
	}

	return false
}

func colasdefn(left []*Node, defn *Node) {
	for _, n := range left {
		if n.Sym != nil {
			n.Sym.Flags |= SymUniq
		}
	}

	var nnew, nerr int
	for i, n := range left {
		if isblank(n) {
			continue
		}
		if !colasname(n) {
			yyerrorl(defn.Lineno, "non-name %v on left side of :=", n)
			nerr++
			continue
		}

		if n.Sym.Flags&SymUniq == 0 {
			yyerrorl(defn.Lineno, "%v repeated on left side of :=", n.Sym)
			n.Diag = true
			nerr++
			continue
		}

		n.Sym.Flags &^= SymUniq
		if n.Sym.Block == block {
			continue
		}

		nnew++
		n = newname(n.Sym)
		declare(n, dclcontext)
		n.Name.Defn = defn
		defn.Ninit.Append(nod(ODCL, n, nil))
		left[i] = n
	}

	if nnew == 0 && nerr == 0 {
		yyerrorl(defn.Lineno, "no new variables on left side of :=")
	}
}

// declare the arguments in an
// interface field declaration.
func ifacedcl(n *Node) {
	if n.Op != ODCLFIELD || n.Right == nil {
		Fatalf("ifacedcl")
	}

	if isblank(n.Left) {
		yyerror("methods must have a unique non-blank name")
	}
}

// declare the function proper
// and declare the arguments.
// called in extern-declaration context
// returns in auto-declaration context.
func funchdr(n *Node) {
	// change the declaration context from extern to auto
	if funcdepth == 0 && dclcontext != PEXTERN {
		Fatalf("funchdr: dclcontext = %d", dclcontext)
	}

	if Ctxt.Flag_dynlink && importpkg == nil && n.Func.Nname != nil {
		makefuncsym(n.Func.Nname.Sym)
	}

	dclcontext = PAUTO
	funcstart(n)

	if n.Func.Nname != nil {
		funcargs(n.Func.Nname.Name.Param.Ntype)
	} else if n.Func.Ntype != nil {
		funcargs(n.Func.Ntype)
	} else {
		funcargs2(n.Type)
	}
}

func funcargs(nt *Node) {
	if nt.Op != OTFUNC {
		Fatalf("funcargs %v", nt.Op)
	}

	// re-start the variable generation number
	// we want to use small numbers for the return variables,
	// so let them have the chunk starting at 1.
	vargen = nt.Rlist.Len()

	// declare the receiver and in arguments.
	// no n->defn because type checking of func header
	// will not fill in the types until later
	if nt.Left != nil {
		n := nt.Left
		if n.Op != ODCLFIELD {
			Fatalf("funcargs receiver %v", n.Op)
		}
		if n.Left != nil {
			n.Left.Op = ONAME
			n.Left.Name.Param.Ntype = n.Right
			declare(n.Left, PPARAM)
			if dclcontext == PAUTO {
				vargen++
				n.Left.Name.Vargen = int32(vargen)
			}
		}
	}

	for _, n := range nt.List.Slice() {
		if n.Op != ODCLFIELD {
			Fatalf("funcargs in %v", n.Op)
		}
		if n.Left != nil {
			n.Left.Op = ONAME
			n.Left.Name.Param.Ntype = n.Right
			declare(n.Left, PPARAM)
			if dclcontext == PAUTO {
				vargen++
				n.Left.Name.Vargen = int32(vargen)
			}
		}
	}

	// declare the out arguments.
	gen := nt.List.Len()
	var i int = 0
	for _, n := range nt.Rlist.Slice() {
		if n.Op != ODCLFIELD {
			Fatalf("funcargs out %v", n.Op)
		}

		if n.Left == nil {
			// Name so that escape analysis can track it. ~r stands for 'result'.
			n.Left = newname(lookupN("~r", gen))
			gen++
		}

		// TODO: n->left->missing = 1;
		n.Left.Op = ONAME

		if isblank(n.Left) {
			// Give it a name so we can assign to it during return. ~b stands for 'blank'.
			// The name must be different from ~r above because if you have
			//	func f() (_ int)
			//	func g() int
			// f is allowed to use a plain 'return' with no arguments, while g is not.
			// So the two cases must be distinguished.
			// We do not record a pointer to the original node (n->orig).
			// Having multiple names causes too much confusion in later passes.
			nn := *n.Left
			nn.Orig = &nn
			nn.Sym = lookupN("~b", gen)
			gen++
			n.Left = &nn
		}

		n.Left.Name.Param.Ntype = n.Right
		declare(n.Left, PPARAMOUT)
		if dclcontext == PAUTO {
			i++
			n.Left.Name.Vargen = int32(i)
		}
	}
}

// Same as funcargs, except run over an already constructed TFUNC.
// This happens during import, where the hidden_fndcl rule has
// used functype directly to parse the function's type.
func funcargs2(t *Type) {
	if t.Etype != TFUNC {
		Fatalf("funcargs2 %v", t)
	}

	for _, ft := range t.Recvs().Fields().Slice() {
		if ft.Nname == nil || ft.Nname.Sym == nil {
			continue
		}
		n := ft.Nname // no need for newname(ft->nname->sym)
		n.Type = ft.Type
		declare(n, PPARAM)
	}

	for _, ft := range t.Params().Fields().Slice() {
		if ft.Nname == nil || ft.Nname.Sym == nil {
			continue
		}
		n := ft.Nname
		n.Type = ft.Type
		declare(n, PPARAM)
	}

	for _, ft := range t.Results().Fields().Slice() {
		if ft.Nname == nil || ft.Nname.Sym == nil {
			continue
		}
		n := ft.Nname
		n.Type = ft.Type
		declare(n, PPARAMOUT)
	}
}

var funcstack []*Node // stack of previous values of Curfn
var funcdepth int32   // len(funcstack) during parsing, but then forced to be the same later during compilation

// start the function.
// called before funcargs; undone at end of funcbody.
func funcstart(n *Node) {
	markdcl()
	funcstack = append(funcstack, Curfn)
	funcdepth++
	Curfn = n
}

// finish the body.
// called in auto-declaration context.
// returns in extern-declaration context.
func funcbody(n *Node) {
	// change the declaration context from auto to extern
	if dclcontext != PAUTO {
		Fatalf("funcbody: unexpected dclcontext %d", dclcontext)
	}
	popdcl()
	funcstack, Curfn = funcstack[:len(funcstack)-1], funcstack[len(funcstack)-1]
	funcdepth--
	if funcdepth == 0 {
		dclcontext = PEXTERN
	}
}

// new type being defined with name s.
func typedcl0(s *Sym) *Node {
	n := newname(s)
	n.Op = OTYPE
	declare(n, dclcontext)
	return n
}

// node n, which was returned by typedcl0
// is being declared to have uncompiled type t.
// return the ODCLTYPE node to use.
func typedcl1(n *Node, t *Node, local bool) *Node {
	n.Name.Param.Ntype = t
	n.Local = local
	return nod(ODCLTYPE, n, nil)
}

// structs, functions, and methods.
// they don't belong here, but where do they belong?
func checkembeddedtype(t *Type) {
	if t == nil {
		return
	}

	if t.Sym == nil && t.IsPtr() {
		t = t.Elem()
		if t.IsInterface() {
			yyerror("embedded type cannot be a pointer to interface")
		}
	}

	if t.IsPtr() || t.IsUnsafePtr() {
		yyerror("embedded type cannot be a pointer")
	} else if t.Etype == TFORW && t.ForwardType().Embedlineno == 0 {
		t.ForwardType().Embedlineno = lineno
	}
}

func structfield(n *Node) *Field {
	lno := lineno
	lineno = n.Lineno

	if n.Op != ODCLFIELD {
		Fatalf("structfield: oops %v\n", n)
	}

	f := newField()
	f.Isddd = n.Isddd

	if n.Right != nil {
		n.Right = typecheck(n.Right, Etype)
		n.Type = n.Right.Type
		if n.Left != nil {
			n.Left.Type = n.Type
		}
		if n.Embedded != 0 {
			checkembeddedtype(n.Type)
		}
	}

	n.Right = nil

	f.Type = n.Type
	if f.Type == nil {
		f.Broke = true
	}

	switch u := n.Val().U.(type) {
	case string:
		f.Note = u
	default:
		yyerror("field annotation must be string")
	case nil:
		// noop
	}

	if n.Left != nil && n.Left.Op == ONAME {
		f.Nname = n.Left
		f.Embedded = n.Embedded
		f.Sym = f.Nname.Sym
	}

	lineno = lno
	return f
}

// checkdupfields emits errors for duplicately named fields or methods in
// a list of struct or interface types.
func checkdupfields(what string, ts ...*Type) {
	lno := lineno

	seen := make(map[*Sym]bool)
	for _, t := range ts {
		for _, f := range t.Fields().Slice() {
			if f.Sym == nil || f.Nname == nil || isblank(f.Nname) {
				continue
			}
			if seen[f.Sym] {
				lineno = f.Nname.Lineno
				yyerror("duplicate %s %s", what, f.Sym.Name)
				continue
			}
			seen[f.Sym] = true
		}
	}

	lineno = lno
}

// convert a parsed id/type list into
// a type for struct/interface/arglist
func tostruct(l []*Node) *Type {
	t := typ(TSTRUCT)
	tostruct0(t, l)
	return t
}

func tostruct0(t *Type, l []*Node) {
	if t == nil || !t.IsStruct() {
		Fatalf("struct expected")
	}

	fields := make([]*Field, len(l))
	for i, n := range l {
		f := structfield(n)
		if f.Broke {
			t.Broke = true
		}
		fields[i] = f
	}
	t.SetFields(fields)

	checkdupfields("field", t)

	if !t.Broke {
		checkwidth(t)
	}
}

func tofunargs(l []*Node, funarg Funarg) *Type {
	t := typ(TSTRUCT)
	t.StructType().Funarg = funarg

	fields := make([]*Field, len(l))
	for i, n := range l {
		f := structfield(n)
		f.Funarg = funarg

		// esc.go needs to find f given a PPARAM to add the tag.
		if n.Left != nil && n.Left.Class == PPARAM {
			n.Left.Name.Param.Field = f
		}
		if f.Broke {
			t.Broke = true
		}
		fields[i] = f
	}
	t.SetFields(fields)
	return t
}

func tofunargsfield(fields []*Field, funarg Funarg) *Type {
	t := typ(TSTRUCT)
	t.StructType().Funarg = funarg

	for _, f := range fields {
		f.Funarg = funarg

		// esc.go needs to find f given a PPARAM to add the tag.
		if f.Nname != nil && f.Nname.Class == PPARAM {
			f.Nname.Name.Param.Field = f
		}
	}
	t.SetFields(fields)
	return t
}

func interfacefield(n *Node) *Field {
	lno := lineno
	lineno = n.Lineno

	if n.Op != ODCLFIELD {
		Fatalf("interfacefield: oops %v\n", n)
	}

	if n.Val().Ctype() != CTxxx {
		yyerror("interface method cannot have annotation")
	}

	f := newField()
	f.Isddd = n.Isddd

	if n.Right != nil {
		if n.Left != nil {
			// queue resolution of method type for later.
			// right now all we need is the name list.
			// avoids cycles for recursive interface types.
			n.Type = typ(TINTERMETH)
			n.Type.SetNname(n.Right)
			n.Left.Type = n.Type
			queuemethod(n)

			if n.Left.Op == ONAME {
				f.Nname = n.Left
				f.Embedded = n.Embedded
				f.Sym = f.Nname.Sym
			}
		} else {
			n.Right = typecheck(n.Right, Etype)
			n.Type = n.Right.Type

			if n.Embedded != 0 {
				checkembeddedtype(n.Type)
			}

			if n.Type != nil {
				switch n.Type.Etype {
				case TINTER:
					break

				case TFORW:
					yyerror("interface type loop involving %v", n.Type)
					f.Broke = true

				default:
					yyerror("interface contains embedded non-interface %v", n.Type)
					f.Broke = true
				}
			}
		}
	}

	n.Right = nil

	f.Type = n.Type
	if f.Type == nil {
		f.Broke = true
	}

	lineno = lno
	return f
}

func tointerface(l []*Node) *Type {
	t := typ(TINTER)
	tointerface0(t, l)
	return t
}

func tointerface0(t *Type, l []*Node) *Type {
	if t == nil || !t.IsInterface() {
		Fatalf("interface expected")
	}

	var fields []*Field
	for _, n := range l {
		f := interfacefield(n)

		if n.Left == nil && f.Type.IsInterface() {
			// embedded interface, inline methods
			for _, t1 := range f.Type.Fields().Slice() {
				f = newField()
				f.Type = t1.Type
				f.Broke = t1.Broke
				f.Sym = t1.Sym
				if f.Sym != nil {
					f.Nname = newname(f.Sym)
				}
				fields = append(fields, f)
			}
		} else {
			fields = append(fields, f)
		}
		if f.Broke {
			t.Broke = true
		}
	}
	sort.Sort(methcmp(fields))
	t.SetFields(fields)

	checkdupfields("method", t)
	checkwidth(t)

	return t
}

func embedded(s *Sym, pkg *Pkg) *Node {
	const (
		CenterDot = 0xB7
	)
	// Names sometimes have disambiguation junk
	// appended after a center dot. Discard it when
	// making the name for the embedded struct field.
	name := s.Name

	if i := strings.Index(s.Name, string(CenterDot)); i >= 0 {
		name = s.Name[:i]
	}

	var n *Node
	if exportname(name) {
		n = newname(lookup(name))
	} else if s.Pkg == builtinpkg {
		// The name of embedded builtins belongs to pkg.
		n = newname(Pkglookup(name, pkg))
	} else {
		n = newname(Pkglookup(name, s.Pkg))
	}
	n = nod(ODCLFIELD, n, oldname(s))
	n.Embedded = 1
	return n
}

// thisT is the singleton type used for interface method receivers.
var thisT *Type

func fakethis() *Node {
	if thisT == nil {
		thisT = ptrto(typ(TSTRUCT))
	}
	return nod(ODCLFIELD, nil, typenod(thisT))
}

func fakethisfield() *Field {
	if thisT == nil {
		thisT = ptrto(typ(TSTRUCT))
	}
	f := newField()
	f.Type = thisT
	return f
}

// Is this field a method on an interface?
// Those methods have thisT as the receiver.
// (See fakethis above.)
func isifacemethod(f *Type) bool {
	return f.Recv().Type == thisT
}

// turn a parsed function declaration into a type
func functype(this *Node, in, out []*Node) *Type {
	t := typ(TFUNC)
	functype0(t, this, in, out)
	return t
}

func functype0(t *Type, this *Node, in, out []*Node) {
	if t == nil || t.Etype != TFUNC {
		Fatalf("function type expected")
	}

	var rcvr []*Node
	if this != nil {
		rcvr = []*Node{this}
	}
	t.FuncType().Receiver = tofunargs(rcvr, FunargRcvr)
	t.FuncType().Results = tofunargs(out, FunargResults)
	t.FuncType().Params = tofunargs(in, FunargParams)

	checkdupfields("argument", t.Recvs(), t.Results(), t.Params())

	if t.Recvs().Broke || t.Results().Broke || t.Params().Broke {
		t.Broke = true
	}

	t.FuncType().Outnamed = false
	if len(out) > 0 && out[0].Left != nil && out[0].Left.Orig != nil {
		s := out[0].Left.Orig.Sym
		if s != nil && (s.Name[0] != '~' || s.Name[1] != 'r') { // ~r%d is the name invented for an unnamed result
			t.FuncType().Outnamed = true
		}
	}
}

func functypefield(this *Field, in, out []*Field) *Type {
	t := typ(TFUNC)
	functypefield0(t, this, in, out)
	return t
}

func functypefield0(t *Type, this *Field, in, out []*Field) {
	var rcvr []*Field
	if this != nil {
		rcvr = []*Field{this}
	}
	t.FuncType().Receiver = tofunargsfield(rcvr, FunargRcvr)
	t.FuncType().Results = tofunargsfield(out, FunargRcvr)
	t.FuncType().Params = tofunargsfield(in, FunargRcvr)

	t.FuncType().Outnamed = false
	if len(out) > 0 && out[0].Nname != nil && out[0].Nname.Orig != nil {
		s := out[0].Nname.Orig.Sym
		if s != nil && (s.Name[0] != '~' || s.Name[1] != 'r') { // ~r%d is the name invented for an unnamed result
			t.FuncType().Outnamed = true
		}
	}
}

var methodsym_toppkg *Pkg

func methodsym(nsym *Sym, t0 *Type, iface int) *Sym {
	var s *Sym
	var p string
	var suffix string
	var spkg *Pkg

	t := t0
	if t == nil {
		goto bad
	}
	s = t.Sym
	if s == nil && t.IsPtr() {
		t = t.Elem()
		if t == nil {
			goto bad
		}
		s = t.Sym
	}

	spkg = nil
	if s != nil {
		spkg = s.Pkg
	}

	// if t0 == *t and t0 has a sym,
	// we want to see *t, not t0, in the method name.
	if t != t0 && t0.Sym != nil {
		t0 = ptrto(t)
	}

	suffix = ""
	if iface != 0 {
		dowidth(t0)
		if t0.Width < Types[Tptr].Width {
			suffix = "·i"
		}
	}

	if (spkg == nil || nsym.Pkg != spkg) && !exportname(nsym.Name) {
		if t0.Sym == nil && t0.IsPtr() {
			p = fmt.Sprintf("(%-S).%s.%s%s", t0, nsym.Pkg.Prefix, nsym.Name, suffix)
		} else {
			p = fmt.Sprintf("%-S.%s.%s%s", t0, nsym.Pkg.Prefix, nsym.Name, suffix)
		}
	} else {
		if t0.Sym == nil && t0.IsPtr() {
			p = fmt.Sprintf("(%-S).%s%s", t0, nsym.Name, suffix)
		} else {
			p = fmt.Sprintf("%-S.%s%s", t0, nsym.Name, suffix)
		}
	}

	if spkg == nil {
		if methodsym_toppkg == nil {
			methodsym_toppkg = mkpkg("go")
		}
		spkg = methodsym_toppkg
	}

	s = Pkglookup(p, spkg)

	return s

bad:
	yyerror("illegal receiver type: %v", t0)
	return nil
}

func methodname(n *Node, t *Node) *Node {
	star := false
	if t.Op == OIND {
		star = true
		t = t.Left
	}

	return methodname0(n.Sym, star, t.Sym)
}

func methodname0(s *Sym, star bool, tsym *Sym) *Node {
	if tsym == nil || isblanksym(s) {
		return newfuncname(s)
	}

	var p string
	if star {
		p = fmt.Sprintf("(*%v).%v", tsym, s)
	} else {
		p = fmt.Sprintf("%v.%v", tsym, s)
	}

	if exportname(tsym.Name) {
		s = lookup(p)
	} else {
		s = Pkglookup(p, tsym.Pkg)
	}

	return newfuncname(s)
}

// Add a method, declared as a function.
// - msym is the method symbol
// - t is function type (with receiver)
func addmethod(msym *Sym, t *Type, local, nointerface bool) {
	// get field sym
	if msym == nil {
		Fatalf("no method symbol")
	}

	// get parent type sym
	rf := t.Recv() // ptr to this structure
	if rf == nil {
		yyerror("missing receiver")
		return
	}

	mt := methtype(rf.Type)
	if mt == nil || mt.Sym == nil {
		pa := rf.Type
		t := pa
		if t != nil && t.IsPtr() {
			if t.Sym != nil {
				yyerror("invalid receiver type %v (%v is a pointer type)", pa, t)
				return
			}
			t = t.Elem()
		}

		switch {
		case t == nil || t.Broke:
			// rely on typecheck having complained before
		case t.Sym == nil:
			yyerror("invalid receiver type %v (%v is an unnamed type)", pa, t)
		case t.IsPtr():
			yyerror("invalid receiver type %v (%v is a pointer type)", pa, t)
		case t.IsInterface():
			yyerror("invalid receiver type %v (%v is an interface type)", pa, t)
		default:
			// Should have picked off all the reasons above,
			// but just in case, fall back to generic error.
			yyerror("invalid receiver type %v (%L / %L)", pa, pa, t)
		}
		return
	}

	if local && !mt.Local {
		yyerror("cannot define new methods on non-local type %v", mt)
		return
	}

	if isblanksym(msym) {
		return
	}

	if mt.IsStruct() {
		for _, f := range mt.Fields().Slice() {
			if f.Sym == msym {
				yyerror("type %v has both field and method named %v", mt, msym)
				return
			}
		}
	}

	for _, f := range mt.Methods().Slice() {
		if msym.Name != f.Sym.Name {
			continue
		}
		// eqtype only checks that incoming and result parameters match,
		// so explicitly check that the receiver parameters match too.
		if !eqtype(t, f.Type) || !eqtype(t.Recv().Type, f.Type.Recv().Type) {
			yyerror("method redeclared: %v.%v\n\t%v\n\t%v", mt, msym, f.Type, t)
		}
		return
	}

	f := newField()
	f.Sym = msym
	f.Nname = newname(msym)
	f.Type = t
	f.Nointerface = nointerface

	mt.Methods().Append(f)
}

func funccompile(n *Node) {
	Stksize = BADWIDTH
	Maxarg = 0

	if n.Type == nil {
		if nerrors == 0 {
			Fatalf("funccompile missing type")
		}
		return
	}

	// assign parameter offsets
	checkwidth(n.Type)

	if Curfn != nil {
		Fatalf("funccompile %v inside %v", n.Func.Nname.Sym, Curfn.Func.Nname.Sym)
	}

	Stksize = 0
	dclcontext = PAUTO
	funcdepth = n.Func.Depth + 1
	compile(n)
	Curfn = nil
	pc = nil
	funcdepth = 0
	dclcontext = PEXTERN
	if nerrors != 0 {
		// If we have compile errors, ignore any assembler/linker errors.
		Ctxt.DiagFunc = func(string, ...interface{}) {}
	}
	obj.Flushplist(Ctxt) // convert from Prog list to machine code
}

func funcsym(s *Sym) *Sym {
	if s.Fsym != nil {
		return s.Fsym
	}

	s1 := Pkglookup(s.Name+"·f", s.Pkg)
	if !Ctxt.Flag_dynlink && s1.Def == nil {
		s1.Def = newfuncname(s1)
		s1.Def.Func.Shortname = newname(s)
		funcsyms = append(funcsyms, s1.Def)
	}
	s.Fsym = s1
	return s1
}

func makefuncsym(s *Sym) {
	if isblanksym(s) {
		return
	}
	if compiling_runtime && s.Name == "getg" {
		// runtime.getg() is not a real function and so does
		// not get a funcsym.
		return
	}
	s1 := funcsym(s)
	s1.Def = newfuncname(s1)
	s1.Def.Func.Shortname = newname(s)
	funcsyms = append(funcsyms, s1.Def)
}

type nowritebarrierrecChecker struct {
	curfn  *Node
	stable bool

	// best maps from the ODCLFUNC of each visited function that
	// recursively invokes a write barrier to the called function
	// on the shortest path to a write barrier.
	best map[*Node]nowritebarrierrecCall
}

type nowritebarrierrecCall struct {
	target *Node
	depth  int
	lineno int32
}

func checknowritebarrierrec() {
	c := nowritebarrierrecChecker{
		best: make(map[*Node]nowritebarrierrecCall),
	}
	visitBottomUp(xtop, func(list []*Node, recursive bool) {
		// Functions with write barriers have depth 0.
		for _, n := range list {
			if n.Func.WBLineno != 0 && n.Func.Pragma&Yeswritebarrierrec == 0 {
				c.best[n] = nowritebarrierrecCall{target: nil, depth: 0, lineno: n.Func.WBLineno}
			}
		}

		// Propagate write barrier depth up from callees. In
		// the recursive case, we have to update this at most
		// len(list) times and can stop when we an iteration
		// that doesn't change anything.
		for _ = range list {
			c.stable = false
			for _, n := range list {
				if n.Func.Pragma&Yeswritebarrierrec != 0 {
					// Don't propagate write
					// barrier up to a
					// yeswritebarrierrec function.
					continue
				}
				if n.Func.WBLineno == 0 {
					c.curfn = n
					c.visitcodelist(n.Nbody)
				}
			}
			if c.stable {
				break
			}
		}

		// Check nowritebarrierrec functions.
		for _, n := range list {
			if n.Func.Pragma&Nowritebarrierrec == 0 {
				continue
			}
			call, hasWB := c.best[n]
			if !hasWB {
				continue
			}

			// Build the error message in reverse.
			err := ""
			for call.target != nil {
				err = fmt.Sprintf("\n\t%v: called by %v%s", linestr(call.lineno), n.Func.Nname, err)
				n = call.target
				call = c.best[n]
			}
			err = fmt.Sprintf("write barrier prohibited by caller; %v%s", n.Func.Nname, err)
			yyerrorl(n.Func.WBLineno, err)
		}
	})
}

func (c *nowritebarrierrecChecker) visitcodelist(l Nodes) {
	for _, n := range l.Slice() {
		c.visitcode(n)
	}
}

func (c *nowritebarrierrecChecker) visitcode(n *Node) {
	if n == nil {
		return
	}

	if n.Op == OCALLFUNC || n.Op == OCALLMETH {
		c.visitcall(n)
	}

	c.visitcodelist(n.Ninit)
	c.visitcode(n.Left)
	c.visitcode(n.Right)
	c.visitcodelist(n.List)
	c.visitcodelist(n.Nbody)
	c.visitcodelist(n.Rlist)
}

func (c *nowritebarrierrecChecker) visitcall(n *Node) {
	fn := n.Left
	if n.Op == OCALLMETH {
		fn = n.Left.Sym.Def
	}
	if fn == nil || fn.Op != ONAME || fn.Class != PFUNC || fn.Name.Defn == nil {
		return
	}
	defn := fn.Name.Defn

	fnbest, ok := c.best[defn]
	if !ok {
		return
	}
	best, ok := c.best[c.curfn]
	if ok && fnbest.depth+1 >= best.depth {
		return
	}
	c.best[c.curfn] = nowritebarrierrecCall{target: defn, depth: fnbest.depth + 1, lineno: n.Lineno}
	c.stable = false
}
