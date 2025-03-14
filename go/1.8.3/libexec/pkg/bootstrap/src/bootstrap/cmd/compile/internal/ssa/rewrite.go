// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/ssa/rewrite.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/ssa/rewrite.go:1
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
)

func applyRewrite(f *Func, rb func(*Block, *Config) bool, rv func(*Value, *Config) bool) {
	// repeat rewrites until we find no more rewrites
	var curb *Block
	var curv *Value
	defer func() {
		if curb != nil {
			curb.Fatalf("panic during rewrite of block %s\n", curb.LongString())
		}
		if curv != nil {
			curv.Fatalf("panic during rewrite of value %s\n", curv.LongString())
			// TODO(khr): print source location also
		}
	}()
	config := f.Config
	for {
		change := false
		for _, b := range f.Blocks {
			if b.Control != nil && b.Control.Op == OpCopy {
				for b.Control.Op == OpCopy {
					b.SetControl(b.Control.Args[0])
				}
			}
			curb = b
			if rb(b, config) {
				change = true
			}
			curb = nil
			for _, v := range b.Values {
				change = phielimValue(v) || change

				// Eliminate copy inputs.
				// If any copy input becomes unused, mark it
				// as invalid and discard its argument. Repeat
				// recursively on the discarded argument.
				// This phase helps remove phantom "dead copy" uses
				// of a value so that a x.Uses==1 rule condition
				// fires reliably.
				for i, a := range v.Args {
					if a.Op != OpCopy {
						continue
					}
					v.SetArg(i, copySource(a))
					change = true
					for a.Uses == 0 {
						b := a.Args[0]
						a.reset(OpInvalid)
						a = b
					}
				}

				// apply rewrite function
				curv = v
				if rv(v, config) {
					change = true
				}
				curv = nil
			}
		}
		if !change {
			break
		}
	}
	// remove clobbered values
	for _, b := range f.Blocks {
		j := 0
		for i, v := range b.Values {
			if v.Op == OpInvalid {
				f.freeValue(v)
				continue
			}
			if i != j {
				b.Values[j] = v
			}
			j++
		}
		if j != len(b.Values) {
			tail := b.Values[j:]
			for j := range tail {
				tail[j] = nil
			}
			b.Values = b.Values[:j]
		}
	}
}

// Common functions called from rewriting rules

func is64BitFloat(t Type) bool {
	return t.Size() == 8 && t.IsFloat()
}

func is32BitFloat(t Type) bool {
	return t.Size() == 4 && t.IsFloat()
}

func is64BitInt(t Type) bool {
	return t.Size() == 8 && t.IsInteger()
}

func is32BitInt(t Type) bool {
	return t.Size() == 4 && t.IsInteger()
}

func is16BitInt(t Type) bool {
	return t.Size() == 2 && t.IsInteger()
}

func is8BitInt(t Type) bool {
	return t.Size() == 1 && t.IsInteger()
}

func isPtr(t Type) bool {
	return t.IsPtrShaped()
}

func isSigned(t Type) bool {
	return t.IsSigned()
}

func typeSize(t Type) int64 {
	return t.Size()
}

// mergeSym merges two symbolic offsets. There is no real merging of
// offsets, we just pick the non-nil one.
func mergeSym(x, y interface{}) interface{} {
	if x == nil {
		return y
	}
	if y == nil {
		return x
	}
	panic(fmt.Sprintf("mergeSym with two non-nil syms %s %s", x, y))
}
func canMergeSym(x, y interface{}) bool {
	return x == nil || y == nil
}

// canMergeLoad reports whether the load can be merged into target without
// invalidating the schedule.
func canMergeLoad(target, load *Value) bool {
	if target.Block.ID != load.Block.ID {
		// If the load is in a different block do not merge it.
		return false
	}
	mem := load.Args[len(load.Args)-1]

	// We need the load's memory arg to still be alive at target. That
	// can't be the case if one of target's args depends on a memory
	// state that is a successor of load's memory arg.
	//
	// For example, it would be invalid to merge load into target in
	// the following situation because newmem has killed oldmem
	// before target is reached:
	//     load = read ... oldmem
	//   newmem = write ... oldmem
	//     arg0 = read ... newmem
	//   target = add arg0 load
	//
	// If the argument comes from a different block then we can exclude
	// it immediately because it must dominate load (which is in the
	// same block as target).
	var args []*Value
	for _, a := range target.Args {
		if a != load && a.Block.ID == target.Block.ID {
			args = append(args, a)
		}
	}

	// memPreds contains memory states known to be predecessors of load's
	// memory state. It is lazily initialized.
	var memPreds map[*Value]bool
search:
	for i := 0; len(args) > 0; i++ {
		const limit = 100
		if i >= limit {
			// Give up if we have done a lot of iterations.
			return false
		}
		v := args[len(args)-1]
		args = args[:len(args)-1]
		if target.Block.ID != v.Block.ID {
			// Since target and load are in the same block
			// we can stop searching when we leave the block.
			continue search
		}
		if v.Op == OpPhi {
			// A Phi implies we have reached the top of the block.
			continue search
		}
		if v.Type.IsTuple() && v.Type.FieldType(1).IsMemory() {
			// We could handle this situation however it is likely
			// to be very rare.
			return false
		}
		if v.Type.IsMemory() {
			if memPreds == nil {
				// Initialise a map containing memory states
				// known to be predecessors of load's memory
				// state.
				memPreds = make(map[*Value]bool)
				m := mem
				const limit = 50
				for i := 0; i < limit; i++ {
					if m.Op == OpPhi {
						break
					}
					if m.Block.ID != target.Block.ID {
						break
					}
					if !m.Type.IsMemory() {
						break
					}
					memPreds[m] = true
					if len(m.Args) == 0 {
						break
					}
					m = m.Args[len(m.Args)-1]
				}
			}

			// We can merge if v is a predecessor of mem.
			//
			// For example, we can merge load into target in the
			// following scenario:
			//      x = read ... v
			//    mem = write ... v
			//   load = read ... mem
			// target = add x load
			if memPreds[v] {
				continue search
			}
			return false
		}
		if len(v.Args) > 0 && v.Args[len(v.Args)-1] == mem {
			// If v takes mem as an input then we know mem
			// is valid at this point.
			continue search
		}
		for _, a := range v.Args {
			if target.Block.ID == a.Block.ID {
				args = append(args, a)
			}
		}
	}
	return true
}

// isArg returns whether s is an arg symbol
func isArg(s interface{}) bool {
	_, ok := s.(*ArgSymbol)
	return ok
}

// isAuto returns whether s is an auto symbol
func isAuto(s interface{}) bool {
	_, ok := s.(*AutoSymbol)
	return ok
}

// isSameSym returns whether sym is the same as the given named symbol
func isSameSym(sym interface{}, name string) bool {
	s, ok := sym.(fmt.Stringer)
	return ok && s.String() == name
}

// nlz returns the number of leading zeros.
func nlz(x int64) int64 {
	// log2(0) == 1, so nlz(0) == 64
	return 63 - log2(x)
}

// ntz returns the number of trailing zeros.
func ntz(x int64) int64 {
	return 64 - nlz(^x&(x-1))
}

// nlo returns the number of leading ones.
func nlo(x int64) int64 {
	return nlz(^x)
}

// nto returns the number of trailing ones.
func nto(x int64) int64 {
	return ntz(^x)
}

// log2 returns logarithm in base of uint64(n), with log2(0) = -1.
// Rounds down.
func log2(n int64) (l int64) {
	l = -1
	x := uint64(n)
	for ; x >= 0x8000; x >>= 16 {
		l += 16
	}
	if x >= 0x80 {
		x >>= 8
		l += 8
	}
	if x >= 0x8 {
		x >>= 4
		l += 4
	}
	if x >= 0x2 {
		x >>= 2
		l += 2
	}
	if x >= 0x1 {
		l++
	}
	return
}

// isPowerOfTwo reports whether n is a power of 2.
func isPowerOfTwo(n int64) bool {
	return n > 0 && n&(n-1) == 0
}

// is32Bit reports whether n can be represented as a signed 32 bit integer.
func is32Bit(n int64) bool {
	return n == int64(int32(n))
}

// is16Bit reports whether n can be represented as a signed 16 bit integer.
func is16Bit(n int64) bool {
	return n == int64(int16(n))
}

// isU16Bit reports whether n can be represented as an unsigned 16 bit integer.
func isU16Bit(n int64) bool {
	return n == int64(uint16(n))
}

// isU32Bit reports whether n can be represented as an unsigned 32 bit integer.
func isU32Bit(n int64) bool {
	return n == int64(uint32(n))
}

// is20Bit reports whether n can be represented as a signed 20 bit integer.
func is20Bit(n int64) bool {
	return -(1<<19) <= n && n < (1<<19)
}

// b2i translates a boolean value to 0 or 1 for assigning to auxInt.
func b2i(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// i2f is used in rules for converting from an AuxInt to a float.
func i2f(i int64) float64 {
	return math.Float64frombits(uint64(i))
}

// i2f32 is used in rules for converting from an AuxInt to a float32.
func i2f32(i int64) float32 {
	return float32(math.Float64frombits(uint64(i)))
}

// f2i is used in the rules for storing a float in AuxInt.
func f2i(f float64) int64 {
	return int64(math.Float64bits(f))
}

// uaddOvf returns true if unsigned a+b would overflow.
func uaddOvf(a, b int64) bool {
	return uint64(a)+uint64(b) < uint64(a)
}

// isSamePtr reports whether p1 and p2 point to the same address.
func isSamePtr(p1, p2 *Value) bool {
	if p1 == p2 {
		return true
	}
	if p1.Op != p2.Op {
		return false
	}
	switch p1.Op {
	case OpOffPtr:
		return p1.AuxInt == p2.AuxInt && isSamePtr(p1.Args[0], p2.Args[0])
	case OpAddr:
		// OpAddr's 0th arg is either OpSP or OpSB, which means that it is uniquely identified by its Op.
		// Checking for value equality only works after [z]cse has run.
		return p1.Aux == p2.Aux && p1.Args[0].Op == p2.Args[0].Op
	case OpAddPtr:
		return p1.Args[1] == p2.Args[1] && isSamePtr(p1.Args[0], p2.Args[0])
	}
	return false
}

// moveSize returns the number of bytes an aligned MOV instruction moves
func moveSize(align int64, c *Config) int64 {
	switch {
	case align%8 == 0 && c.IntSize == 8:
		return 8
	case align%4 == 0:
		return 4
	case align%2 == 0:
		return 2
	}
	return 1
}

// mergePoint finds a block among a's blocks which dominates b and is itself
// dominated by all of a's blocks. Returns nil if it can't find one.
// Might return nil even if one does exist.
func mergePoint(b *Block, a ...*Value) *Block {
	// Walk backward from b looking for one of the a's blocks.

	// Max distance
	d := 100

	for d > 0 {
		for _, x := range a {
			if b == x.Block {
				goto found
			}
		}
		if len(b.Preds) > 1 {
			// Don't know which way to go back. Abort.
			return nil
		}
		b = b.Preds[0].b
		d--
	}
	return nil // too far away
found:
	// At this point, r is the first value in a that we find by walking backwards.
	// if we return anything, r will be it.
	r := b

	// Keep going, counting the other a's that we find. They must all dominate r.
	na := 0
	for d > 0 {
		for _, x := range a {
			if b == x.Block {
				na++
			}
		}
		if na == len(a) {
			// Found all of a in a backwards walk. We can return r.
			return r
		}
		if len(b.Preds) > 1 {
			return nil
		}
		b = b.Preds[0].b
		d--

	}
	return nil // too far away
}

// clobber invalidates v.  Returns true.
// clobber is used by rewrite rules to:
//   A) make sure v is really dead and never used again.
//   B) decrement use counts of v's args.
func clobber(v *Value) bool {
	v.reset(OpInvalid)
	// Note: leave v.Block intact.  The Block field is used after clobber.
	return true
}

// noteRule is an easy way to track if a rule is matched when writing
// new ones.  Make the rule of interest also conditional on
//     noteRule("note to self: rule of interest matched")
// and that message will print when the rule matches.
func noteRule(s string) bool {
	fmt.Println(s)
	return true
}

// warnRule generates a compiler debug output with string s when
// cond is true and the rule is fired.
func warnRule(cond bool, v *Value, s string) bool {
	if cond {
		v.Block.Func.Config.Warnl(v.Line, s)
	}
	return true
}

// logRule logs the use of the rule s. This will only be enabled if
// rewrite rules were generated with the -log option, see gen/rulegen.go.
func logRule(s string) {
	if ruleFile == nil {
		// Open a log file to write log to. We open in append
		// mode because all.bash runs the compiler lots of times,
		// and we want the concatenation of all of those logs.
		// This means, of course, that users need to rm the old log
		// to get fresh data.
		// TODO: all.bash runs compilers in parallel. Need to synchronize logging somehow?
		w, err := os.OpenFile(filepath.Join(os.Getenv("GOROOT"), "src", "rulelog"),
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		ruleFile = w
	}
	_, err := fmt.Fprintf(ruleFile, "rewrite %s\n", s)
	if err != nil {
		panic(err)
	}
}

var ruleFile *os.File
