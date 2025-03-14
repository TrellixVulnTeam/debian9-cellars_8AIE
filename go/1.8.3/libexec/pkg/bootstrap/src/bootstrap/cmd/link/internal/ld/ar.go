// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/link/internal/ld/ar.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/link/internal/ld/ar.go:1
// Inferno utils/include/ar.h
// https://bitbucket.org/inferno-os/inferno-os/src/default/utils/include/ar.h
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

import (
	"bootstrap/cmd/internal/bio"
	"bootstrap/cmd/internal/obj"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	SARMAG  = 8
	SAR_HDR = 16 + 44
)

const (
	ARMAG = "!<arch>\n"
)

type ArHdr struct {
	name string
	date string
	uid  string
	gid  string
	mode string
	size string
	fmag string
}

// hostArchive reads an archive file holding host objects and links in
// required objects. The general format is the same as a Go archive
// file, but it has an armap listing symbols and the objects that
// define them. This is used for the compiler support library
// libgcc.a.
func hostArchive(ctxt *Link, name string) {
	f, err := bio.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			// It's OK if we don't have a libgcc file at all.
			if ctxt.Debugvlog != 0 {
				ctxt.Logf("skipping libgcc file: %v\n", err)
			}
			return
		}
		Exitf("cannot open file %s: %v", name, err)
	}
	defer f.Close()

	var magbuf [len(ARMAG)]byte
	if _, err := io.ReadFull(f, magbuf[:]); err != nil {
		Exitf("file %s too short", name)
	}

	var arhdr ArHdr
	l := nextar(f, f.Offset(), &arhdr)
	if l <= 0 {
		Exitf("%s missing armap", name)
	}

	var armap archiveMap
	if arhdr.name == "/" || arhdr.name == "/SYM64/" {
		armap = readArmap(name, f, arhdr)
	} else {
		Exitf("%s missing armap", name)
	}

	loaded := make(map[uint64]bool)
	any := true
	for any {
		var load []uint64
		for _, s := range ctxt.Syms.Allsym {
			for _, r := range s.R {
				if r.Sym != nil && r.Sym.Type&obj.SMASK == obj.SXREF {
					if off := armap[r.Sym.Name]; off != 0 && !loaded[off] {
						load = append(load, off)
						loaded[off] = true
					}
				}
			}
		}

		for _, off := range load {
			l := nextar(f, int64(off), &arhdr)
			if l <= 0 {
				Exitf("%s missing archive entry at offset %d", name, off)
			}
			pname := fmt.Sprintf("%s(%s)", name, arhdr.name)
			l = atolwhex(arhdr.size)

			libgcc := Library{Pkg: "libgcc"}
			h := ldobj(ctxt, f, &libgcc, l, pname, name, ArchiveObj)
			f.Seek(h.off, 0)
			h.ld(ctxt, f, h.pkg, h.length, h.pn)
		}

		any = len(load) > 0
	}
}

// archiveMap is an archive symbol map: a mapping from symbol name to
// offset within the archive file.
type archiveMap map[string]uint64

// readArmap reads the archive symbol map.
func readArmap(filename string, f *bio.Reader, arhdr ArHdr) archiveMap {
	is64 := arhdr.name == "/SYM64/"
	wordSize := 4
	if is64 {
		wordSize = 8
	}

	contents := make([]byte, atolwhex(arhdr.size))
	if _, err := io.ReadFull(f, contents); err != nil {
		Exitf("short read from %s", filename)
	}

	var c uint64
	if is64 {
		c = binary.BigEndian.Uint64(contents)
	} else {
		c = uint64(binary.BigEndian.Uint32(contents))
	}
	contents = contents[wordSize:]

	ret := make(archiveMap)

	names := contents[c*uint64(wordSize):]
	for i := uint64(0); i < c; i++ {
		n := 0
		for names[n] != 0 {
			n++
		}
		name := string(names[:n])
		names = names[n+1:]

		// For Mach-O and PE/386 files we strip a leading
		// underscore from the symbol name.
		if obj.GOOS == "darwin" || (obj.GOOS == "windows" && obj.GOARCH == "386") {
			if name[0] == '_' && len(name) > 1 {
				name = name[1:]
			}
		}

		var off uint64
		if is64 {
			off = binary.BigEndian.Uint64(contents)
		} else {
			off = uint64(binary.BigEndian.Uint32(contents))
		}
		contents = contents[wordSize:]

		ret[name] = off
	}

	return ret
}
