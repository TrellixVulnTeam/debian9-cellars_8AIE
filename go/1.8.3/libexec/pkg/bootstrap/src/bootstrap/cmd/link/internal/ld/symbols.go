// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/link/internal/ld/symbols.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/link/internal/ld/symbols.go:1
// Derived from Inferno utils/6l/l.h and related files.
// https://bitbucket.org/inferno-os/inferno-os/src/default/utils/6l/l.h
//
//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
//	Portions Copyright © 2009 The Go Authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package ld

type Symbols struct {
	symbolBatch []Symbol

	// Symbol lookup based on name and indexed by version.
	hash []map[string]*Symbol

	Allsym []*Symbol
}

func (syms *Symbols) newsym(name string, v int) *Symbol {
	batch := syms.symbolBatch
	if len(batch) == 0 {
		batch = make([]Symbol, 1000)
	}
	s := &batch[0]
	syms.symbolBatch = batch[1:]

	s.Dynid = -1
	s.Plt = -1
	s.Got = -1
	s.Name = name
	s.Version = int16(v)
	syms.Allsym = append(syms.Allsym, s)

	return s
}

// Look up the symbol with the given name and version, creating the
// symbol if it is not found.
func (syms *Symbols) Lookup(name string, v int) *Symbol {
	m := syms.hash[v]
	s := m[name]
	if s != nil {
		return s
	}
	s = syms.newsym(name, v)
	s.Extname = s.Name
	m[name] = s
	return s
}

// Look up the symbol with the given name and version, returning nil
// if it is not found.
func (syms *Symbols) ROLookup(name string, v int) *Symbol {
	return syms.hash[v][name]
}

// Allocate a new version (i.e. symbol namespace).
func (syms *Symbols) IncVersion() int {
	syms.hash = append(syms.hash, make(map[string]*Symbol))
	return len(syms.hash) - 1
}
