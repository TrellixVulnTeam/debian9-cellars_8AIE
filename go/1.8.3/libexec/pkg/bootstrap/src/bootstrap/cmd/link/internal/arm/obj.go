// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/link/internal/arm/obj.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/link/internal/arm/obj.go:1
// Inferno utils/5l/obj.c
// https://bitbucket.org/inferno-os/inferno-os/src/default/utils/5l/obj.c
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

package arm

import (
	"bootstrap/cmd/internal/obj"
	"bootstrap/cmd/internal/sys"
	"bootstrap/cmd/link/internal/ld"
	"fmt"
)

func Init() {
	ld.SysArch = sys.ArchARM

	ld.Thearch.Funcalign = funcAlign
	ld.Thearch.Maxalign = maxAlign
	ld.Thearch.Minalign = minAlign
	ld.Thearch.Dwarfregsp = dwarfRegSP
	ld.Thearch.Dwarfreglr = dwarfRegLR

	ld.Thearch.Adddynrel = adddynrel
	ld.Thearch.Archinit = archinit
	ld.Thearch.Archreloc = archreloc
	ld.Thearch.Archrelocvariant = archrelocvariant
	ld.Thearch.Trampoline = trampoline
	ld.Thearch.Asmb = asmb
	ld.Thearch.Elfreloc1 = elfreloc1
	ld.Thearch.Elfsetupplt = elfsetupplt
	ld.Thearch.Gentext = gentext
	ld.Thearch.Machoreloc1 = machoreloc1
	ld.Thearch.Lput = ld.Lputl
	ld.Thearch.Wput = ld.Wputl
	ld.Thearch.Vput = ld.Vputl
	ld.Thearch.Append16 = ld.Append16l
	ld.Thearch.Append32 = ld.Append32l
	ld.Thearch.Append64 = ld.Append64l

	ld.Thearch.Linuxdynld = "/lib/ld-linux.so.3" // 2 for OABI, 3 for EABI
	ld.Thearch.Freebsddynld = "/usr/libexec/ld-elf.so.1"
	ld.Thearch.Openbsddynld = "/usr/libexec/ld.so"
	ld.Thearch.Netbsddynld = "/libexec/ld.elf_so"
	ld.Thearch.Dragonflydynld = "XXX"
	ld.Thearch.Solarisdynld = "XXX"
}

func archinit(ctxt *ld.Link) {
	switch ld.Headtype {
	default:
		ld.Exitf("unknown -H option: %v", ld.Headtype)

	case obj.Hplan9: /* plan 9 */
		ld.HEADR = 32

		if *ld.FlagTextAddr == -1 {
			*ld.FlagTextAddr = 4128
		}
		if *ld.FlagDataAddr == -1 {
			*ld.FlagDataAddr = 0
		}
		if *ld.FlagRound == -1 {
			*ld.FlagRound = 4096
		}

	case obj.Hlinux, /* arm elf */
		obj.Hfreebsd,
		obj.Hnetbsd,
		obj.Hopenbsd:
		*ld.FlagD = false
		// with dynamic linking
		ld.Elfinit(ctxt)
		ld.HEADR = ld.ELFRESERVE
		if *ld.FlagTextAddr == -1 {
			*ld.FlagTextAddr = 0x10000 + int64(ld.HEADR)
		}
		if *ld.FlagDataAddr == -1 {
			*ld.FlagDataAddr = 0
		}
		if *ld.FlagRound == -1 {
			*ld.FlagRound = 0x10000
		}

	case obj.Hnacl:
		ld.Elfinit(ctxt)
		ld.HEADR = 0x10000
		ld.Funcalign = 16
		if *ld.FlagTextAddr == -1 {
			*ld.FlagTextAddr = 0x20000
		}
		if *ld.FlagDataAddr == -1 {
			*ld.FlagDataAddr = 0
		}
		if *ld.FlagRound == -1 {
			*ld.FlagRound = 0x10000
		}

	case obj.Hdarwin: /* apple MACH */
		*ld.FlagW = true // disable DWARF generation
		ld.Machoinit()
		ld.HEADR = ld.INITIAL_MACHO_HEADR
		if *ld.FlagTextAddr == -1 {
			*ld.FlagTextAddr = 4096 + int64(ld.HEADR)
		}
		if *ld.FlagDataAddr == -1 {
			*ld.FlagDataAddr = 0
		}
		if *ld.FlagRound == -1 {
			*ld.FlagRound = 4096
		}
	}

	if *ld.FlagDataAddr != 0 && *ld.FlagRound != 0 {
		fmt.Printf("warning: -D0x%x is ignored because of -R0x%x\n", uint64(*ld.FlagDataAddr), uint32(*ld.FlagRound))
	}
}
