// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/link/internal/ld/main.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/link/internal/ld/main.go:1
// Inferno utils/6l/obj.c
// https://bitbucket.org/inferno-os/inferno-os/src/default/utils/6l/obj.c
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
	"bufio"
	"bootstrap/cmd/internal/obj"
	"bootstrap/cmd/internal/sys"
	"flag"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
)

var (
	pkglistfornote []byte
)

func init() {
	flag.Var(&Linkmode, "linkmode", "set link `mode`")
	flag.Var(&Buildmode, "buildmode", "set build `mode`")
	flag.Var(&Headtype, "H", "set header `type`")
	flag.Var(&rpath, "r", "set the ELF dynamic linker search `path` to dir1:dir2:...")
}

// Flags used by the linker. The exported flags are used by the architecture-specific packages.
var (
	flagBuildid = flag.String("buildid", "", "record `id` as Go toolchain build id")

	flagOutfile    = flag.String("o", "", "write output to `file`")
	flagPluginPath = flag.String("pluginpath", "", "full path name for plugin")
	FlagLinkshared = flag.Bool("linkshared", false, "link against installed Go shared libraries")

	flagInstallSuffix = flag.String("installsuffix", "", "set package directory `suffix`")
	flagDumpDep       = flag.Bool("dumpdep", false, "dump symbol dependency graph")
	flagRace          = flag.Bool("race", false, "enable race detector")
	flagMsan          = flag.Bool("msan", false, "enable MSan interface")

	flagFieldTrack = flag.String("k", "", "set field tracking `symbol`")
	flagLibGCC     = flag.String("libgcc", "", "compiler support lib for internal linking; use \"none\" to disable")
	flagTmpdir     = flag.String("tmpdir", "", "use `directory` for temporary files")

	flagExtld      = flag.String("extld", "", "use `linker` when linking in external mode")
	flagExtldflags = flag.String("extldflags", "", "pass `flags` to external linker")
	flagExtar      = flag.String("extar", "", "archive program for buildmode=c-archive")

	flagA           = flag.Bool("a", false, "disassemble output")
	FlagC           = flag.Bool("c", false, "dump call graph")
	FlagD           = flag.Bool("d", false, "disable dynamic executable")
	flagF           = flag.Bool("f", false, "ignore version mismatch")
	flagG           = flag.Bool("g", false, "disable go package data checks")
	flagH           = flag.Bool("h", false, "halt on error")
	flagN           = flag.Bool("n", false, "dump symbol table")
	FlagS           = flag.Bool("s", false, "disable symbol table")
	flagU           = flag.Bool("u", false, "reject unsafe packages")
	FlagW           = flag.Bool("w", false, "disable DWARF generation")
	Flag8           bool // use 64-bit addresses in symbol table
	flagInterpreter = flag.String("I", "", "use `linker` as ELF dynamic linker")
	FlagDebugTramp  = flag.Int("debugtramp", 0, "debug trampolines")

	FlagRound       = flag.Int("R", -1, "set address rounding `quantum`")
	FlagTextAddr    = flag.Int64("T", -1, "set text segment `address`")
	FlagDataAddr    = flag.Int64("D", -1, "set data segment `address`")
	flagEntrySymbol = flag.String("E", "", "set `entry` symbol name")

	cpuprofile     = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile     = flag.String("memprofile", "", "write memory profile to `file`")
	memprofilerate = flag.Int64("memprofilerate", 0, "set runtime.MemProfileRate to `rate`")
)

// Main is the main entry point for the linker code.
func Main() {
	ctxt := linknew(SysArch)
	ctxt.Bso = bufio.NewWriter(os.Stdout)

	// For testing behavior of go command when tools crash silently.
	// Undocumented, not in standard flag parser to avoid
	// exposing in usage message.
	for _, arg := range os.Args {
		if arg == "-crash_for_testing" {
			os.Exit(2)
		}
	}

	// TODO(matloob): define these above and then check flag values here
	if SysArch.Family == sys.AMD64 && obj.GOOS == "plan9" {
		flag.BoolVar(&Flag8, "8", false, "use 64-bit addresses in symbol table")
	}
	obj.Flagfn1("B", "add an ELF NT_GNU_BUILD_ID `note` when using ELF", addbuildinfo)
	obj.Flagfn1("L", "add specified `directory` to library path", func(a string) { Lflag(ctxt, a) })
	obj.Flagfn0("V", "print version and exit", doversion)
	obj.Flagfn1("X", "add string value `definition` of the form importpath.name=value", func(s string) { addstrdata1(ctxt, s) })
	obj.Flagcount("v", "print link trace", &ctxt.Debugvlog)

	obj.Flagparse(usage)

	startProfile()
	if Buildmode == BuildmodeUnset {
		Buildmode = BuildmodeExe
	}

	if Buildmode != BuildmodeShared && flag.NArg() != 1 {
		usage()
	}

	if *flagOutfile == "" {
		*flagOutfile = "a.out"
		if Headtype == obj.Hwindows || Headtype == obj.Hwindowsgui {
			*flagOutfile += ".exe"
		}
	}

	interpreter = *flagInterpreter

	libinit(ctxt) // creates outfile

	if Headtype == obj.Hunknown {
		Headtype.Set(obj.GOOS)
	}

	ctxt.computeTLSOffset()
	Thearch.Archinit(ctxt)

	if *FlagLinkshared && !Iself {
		Exitf("-linkshared can only be used on elf systems")
	}

	if ctxt.Debugvlog != 0 {
		ctxt.Logf("HEADER = -H%d -T0x%x -D0x%x -R0x%x\n", Headtype, uint64(*FlagTextAddr), uint64(*FlagDataAddr), uint32(*FlagRound))
	}

	switch Buildmode {
	case BuildmodeShared:
		for i := 0; i < flag.NArg(); i++ {
			arg := flag.Arg(i)
			parts := strings.SplitN(arg, "=", 2)
			var pkgpath, file string
			if len(parts) == 1 {
				pkgpath, file = "main", arg
			} else {
				pkgpath, file = parts[0], parts[1]
			}
			pkglistfornote = append(pkglistfornote, pkgpath...)
			pkglistfornote = append(pkglistfornote, '\n')
			addlibpath(ctxt, "command line", "command line", file, pkgpath, "")
		}
	case BuildmodePlugin:
		addlibpath(ctxt, "command line", "command line", flag.Arg(0), *flagPluginPath, "")
	default:
		addlibpath(ctxt, "command line", "command line", flag.Arg(0), "main", "")
	}
	ctxt.loadlib()

	ctxt.checkstrdata()
	deadcode(ctxt)
	fieldtrack(ctxt)
	ctxt.callgraph()

	ctxt.doelf()
	if Headtype == obj.Hdarwin {
		ctxt.domacho()
	}
	ctxt.dostkcheck()
	if Headtype == obj.Hwindows || Headtype == obj.Hwindowsgui {
		ctxt.dope()
	}
	ctxt.addexport()
	Thearch.Gentext(ctxt) // trampolines, call stubs, etc.
	ctxt.textbuildid()
	ctxt.textaddress()
	ctxt.pclntab()
	ctxt.findfunctab()
	ctxt.typelink()
	ctxt.symtab()
	ctxt.dodata()
	ctxt.address()
	ctxt.reloc()
	Thearch.Asmb(ctxt)
	ctxt.undef()
	ctxt.hostlink()
	ctxt.archive()
	if ctxt.Debugvlog != 0 {
		ctxt.Logf("%5.2f cpu time\n", obj.Cputime())
		ctxt.Logf("%d symbols\n", len(ctxt.Syms.Allsym))
		ctxt.Logf("%d liveness data\n", liveness)
	}

	ctxt.Bso.Flush()

	errorexit()
}

type Rpath struct {
	set bool
	val string
}

func (r *Rpath) Set(val string) error {
	r.set = true
	r.val = val
	return nil
}

func (r *Rpath) String() string {
	return r.val
}

func startProfile() {
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatalf("%v", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("%v", err)
		}
		AtExit(pprof.StopCPUProfile)
	}
	if *memprofile != "" {
		if *memprofilerate != 0 {
			runtime.MemProfileRate = int(*memprofilerate)
		}
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatalf("%v", err)
		}
		AtExit(func() {
			runtime.GC() // profile all outstanding allocations
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatalf("%v", err)
			}
		})
	}
}
