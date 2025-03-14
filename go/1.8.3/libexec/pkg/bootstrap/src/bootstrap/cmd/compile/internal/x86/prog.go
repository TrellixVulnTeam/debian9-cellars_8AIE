// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/x86/prog.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/x86/prog.go:1
// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x86

import (
	"bootstrap/cmd/compile/internal/gc"
	"bootstrap/cmd/internal/obj"
	"bootstrap/cmd/internal/obj/x86"
)

const (
	LeftRdwr  uint32 = gc.LeftRead | gc.LeftWrite
	RightRdwr uint32 = gc.RightRead | gc.RightWrite
)

// This table gives the basic information about instruction
// generated by the compiler and processed in the optimizer.
// See opt.h for bit definitions.
//
// Instructions not generated need not be listed.
// As an exception to that rule, we typically write down all the
// size variants of an operation even if we just use a subset.
//
// The table is formatted for 8-space tabs.
var progtable = [x86.ALAST & obj.AMask]gc.ProgInfo{
	obj.ATYPE:     {Flags: gc.Pseudo | gc.Skip},
	obj.ATEXT:     {Flags: gc.Pseudo},
	obj.AFUNCDATA: {Flags: gc.Pseudo},
	obj.APCDATA:   {Flags: gc.Pseudo},
	obj.AUNDEF:    {Flags: gc.Break},
	obj.AUSEFIELD: {Flags: gc.OK},
	obj.AVARDEF:   {Flags: gc.Pseudo | gc.RightWrite},
	obj.AVARKILL:  {Flags: gc.Pseudo | gc.RightWrite},
	obj.AVARLIVE:  {Flags: gc.Pseudo | gc.LeftRead},

	// NOP is an internal no-op that also stands
	// for USED and SET annotations, not the Intel opcode.
	obj.ANOP:                   {Flags: gc.LeftRead | gc.RightWrite},
	x86.AADCL & obj.AMask:      {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.SetCarry | gc.UseCarry},
	x86.AADCW & obj.AMask:      {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.SetCarry | gc.UseCarry},
	x86.AADDB & obj.AMask:      {Flags: gc.SizeB | gc.LeftRead | RightRdwr | gc.SetCarry},
	x86.AADDL & obj.AMask:      {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.SetCarry},
	x86.AADDW & obj.AMask:      {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.SetCarry},
	x86.AADDSD & obj.AMask:     {Flags: gc.SizeD | gc.LeftRead | RightRdwr},
	x86.AADDSS & obj.AMask:     {Flags: gc.SizeF | gc.LeftRead | RightRdwr},
	x86.AANDB & obj.AMask:      {Flags: gc.SizeB | gc.LeftRead | RightRdwr | gc.SetCarry},
	x86.AANDL & obj.AMask:      {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.SetCarry},
	x86.AANDW & obj.AMask:      {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.SetCarry},
	obj.ACALL:                  {Flags: gc.RightAddr | gc.Call | gc.KillCarry},
	x86.ACDQ & obj.AMask:       {Flags: gc.OK},
	x86.ACWD & obj.AMask:       {Flags: gc.OK},
	x86.ACLD & obj.AMask:       {Flags: gc.OK},
	x86.ASTD & obj.AMask:       {Flags: gc.OK},
	x86.ACMPB & obj.AMask:      {Flags: gc.SizeB | gc.LeftRead | gc.RightRead | gc.SetCarry},
	x86.ACMPL & obj.AMask:      {Flags: gc.SizeL | gc.LeftRead | gc.RightRead | gc.SetCarry},
	x86.ACMPW & obj.AMask:      {Flags: gc.SizeW | gc.LeftRead | gc.RightRead | gc.SetCarry},
	x86.ACOMISD & obj.AMask:    {Flags: gc.SizeD | gc.LeftRead | gc.RightRead | gc.SetCarry},
	x86.ACOMISS & obj.AMask:    {Flags: gc.SizeF | gc.LeftRead | gc.RightRead | gc.SetCarry},
	x86.ACVTSD2SL & obj.AMask:  {Flags: gc.SizeL | gc.LeftRead | gc.RightWrite | gc.Conv},
	x86.ACVTSD2SS & obj.AMask:  {Flags: gc.SizeF | gc.LeftRead | gc.RightWrite | gc.Conv},
	x86.ACVTSL2SD & obj.AMask:  {Flags: gc.SizeD | gc.LeftRead | gc.RightWrite | gc.Conv},
	x86.ACVTSL2SS & obj.AMask:  {Flags: gc.SizeF | gc.LeftRead | gc.RightWrite | gc.Conv},
	x86.ACVTSS2SD & obj.AMask:  {Flags: gc.SizeD | gc.LeftRead | gc.RightWrite | gc.Conv},
	x86.ACVTSS2SL & obj.AMask:  {Flags: gc.SizeL | gc.LeftRead | gc.RightWrite | gc.Conv},
	x86.ACVTTSD2SL & obj.AMask: {Flags: gc.SizeL | gc.LeftRead | gc.RightWrite | gc.Conv},
	x86.ACVTTSS2SL & obj.AMask: {Flags: gc.SizeL | gc.LeftRead | gc.RightWrite | gc.Conv},
	x86.ADECB & obj.AMask:      {Flags: gc.SizeB | RightRdwr},
	x86.ADECL & obj.AMask:      {Flags: gc.SizeL | RightRdwr},
	x86.ADECW & obj.AMask:      {Flags: gc.SizeW | RightRdwr},
	x86.ADIVB & obj.AMask:      {Flags: gc.SizeB | gc.LeftRead | gc.SetCarry},
	x86.ADIVL & obj.AMask:      {Flags: gc.SizeL | gc.LeftRead | gc.SetCarry},
	x86.ADIVW & obj.AMask:      {Flags: gc.SizeW | gc.LeftRead | gc.SetCarry},
	x86.ADIVSD & obj.AMask:     {Flags: gc.SizeD | gc.LeftRead | RightRdwr},
	x86.ADIVSS & obj.AMask:     {Flags: gc.SizeF | gc.LeftRead | RightRdwr},
	x86.AFLDCW & obj.AMask:     {Flags: gc.SizeW | gc.LeftAddr},
	x86.AFSTCW & obj.AMask:     {Flags: gc.SizeW | gc.RightAddr},
	x86.AFSTSW & obj.AMask:     {Flags: gc.SizeW | gc.RightAddr | gc.RightWrite},
	x86.AFADDD & obj.AMask:     {Flags: gc.SizeD | gc.LeftAddr | RightRdwr},
	x86.AFADDDP & obj.AMask:    {Flags: gc.SizeD | gc.LeftAddr | RightRdwr},
	x86.AFADDF & obj.AMask:     {Flags: gc.SizeF | gc.LeftAddr | RightRdwr},
	x86.AFCOMD & obj.AMask:     {Flags: gc.SizeD | gc.LeftAddr | gc.RightRead},
	x86.AFCOMDP & obj.AMask:    {Flags: gc.SizeD | gc.LeftAddr | gc.RightRead},
	x86.AFCOMDPP & obj.AMask:   {Flags: gc.SizeD | gc.LeftAddr | gc.RightRead},
	x86.AFCOMF & obj.AMask:     {Flags: gc.SizeF | gc.LeftAddr | gc.RightRead},
	x86.AFCOMFP & obj.AMask:    {Flags: gc.SizeF | gc.LeftAddr | gc.RightRead},
	// NOTE(khr): don't use FUCOMI* instructions, not available
	// on Pentium MMX.  See issue 13923.
	//x86.AFUCOMIP&obj.AMask:   {Flags: gc.SizeF | gc.LeftAddr | gc.RightRead},
	x86.AFUCOMP & obj.AMask:  {Flags: gc.SizeD | gc.LeftRead | gc.RightRead},
	x86.AFUCOMPP & obj.AMask: {Flags: gc.SizeD | gc.LeftRead | gc.RightRead},
	x86.AFCHS & obj.AMask:    {Flags: gc.SizeD | RightRdwr}, // also SizeF

	x86.AFDIVDP & obj.AMask:  {Flags: gc.SizeD | gc.LeftAddr | RightRdwr},
	x86.AFDIVF & obj.AMask:   {Flags: gc.SizeF | gc.LeftAddr | RightRdwr},
	x86.AFDIVD & obj.AMask:   {Flags: gc.SizeD | gc.LeftAddr | RightRdwr},
	x86.AFDIVRDP & obj.AMask: {Flags: gc.SizeD | gc.LeftAddr | RightRdwr},
	x86.AFDIVRF & obj.AMask:  {Flags: gc.SizeF | gc.LeftAddr | RightRdwr},
	x86.AFDIVRD & obj.AMask:  {Flags: gc.SizeD | gc.LeftAddr | RightRdwr},
	x86.AFXCHD & obj.AMask:   {Flags: gc.SizeD | LeftRdwr | RightRdwr},
	x86.AFSUBD & obj.AMask:   {Flags: gc.SizeD | gc.LeftAddr | RightRdwr},
	x86.AFSUBDP & obj.AMask:  {Flags: gc.SizeD | gc.LeftAddr | RightRdwr},
	x86.AFSUBF & obj.AMask:   {Flags: gc.SizeF | gc.LeftAddr | RightRdwr},
	x86.AFSUBRD & obj.AMask:  {Flags: gc.SizeD | gc.LeftAddr | RightRdwr},
	x86.AFSUBRDP & obj.AMask: {Flags: gc.SizeD | gc.LeftAddr | RightRdwr},
	x86.AFSUBRF & obj.AMask:  {Flags: gc.SizeF | gc.LeftAddr | RightRdwr},
	x86.AFMOVD & obj.AMask:   {Flags: gc.SizeD | gc.LeftAddr | gc.RightWrite},
	x86.AFMOVF & obj.AMask:   {Flags: gc.SizeF | gc.LeftAddr | gc.RightWrite},
	x86.AFMOVL & obj.AMask:   {Flags: gc.SizeL | gc.LeftAddr | gc.RightWrite},
	x86.AFMOVW & obj.AMask:   {Flags: gc.SizeW | gc.LeftAddr | gc.RightWrite},
	x86.AFMOVV & obj.AMask:   {Flags: gc.SizeQ | gc.LeftAddr | gc.RightWrite},

	// These instructions are marked as RightAddr
	// so that the register optimizer does not try to replace the
	// memory references with integer register references.
	// But they do not use the previous value at the address, so
	// we also mark them RightWrite.
	x86.AFMOVDP & obj.AMask:  {Flags: gc.SizeD | gc.LeftRead | gc.RightWrite | gc.RightAddr},
	x86.AFMOVFP & obj.AMask:  {Flags: gc.SizeF | gc.LeftRead | gc.RightWrite | gc.RightAddr},
	x86.AFMOVLP & obj.AMask:  {Flags: gc.SizeL | gc.LeftRead | gc.RightWrite | gc.RightAddr},
	x86.AFMOVWP & obj.AMask:  {Flags: gc.SizeW | gc.LeftRead | gc.RightWrite | gc.RightAddr},
	x86.AFMOVVP & obj.AMask:  {Flags: gc.SizeQ | gc.LeftRead | gc.RightWrite | gc.RightAddr},
	x86.AFMULD & obj.AMask:   {Flags: gc.SizeD | gc.LeftAddr | RightRdwr},
	x86.AFMULDP & obj.AMask:  {Flags: gc.SizeD | gc.LeftAddr | RightRdwr},
	x86.AFMULF & obj.AMask:   {Flags: gc.SizeF | gc.LeftAddr | RightRdwr},
	x86.AIDIVB & obj.AMask:   {Flags: gc.SizeB | gc.LeftRead | gc.SetCarry},
	x86.AIDIVL & obj.AMask:   {Flags: gc.SizeL | gc.LeftRead | gc.SetCarry},
	x86.AIDIVW & obj.AMask:   {Flags: gc.SizeW | gc.LeftRead | gc.SetCarry},
	x86.AIMULB & obj.AMask:   {Flags: gc.SizeB | gc.LeftRead | gc.SetCarry},
	x86.AIMULL & obj.AMask:   {Flags: gc.SizeL | gc.LeftRead | gc.ImulAXDX | gc.SetCarry},
	x86.AIMULW & obj.AMask:   {Flags: gc.SizeW | gc.LeftRead | gc.ImulAXDX | gc.SetCarry},
	x86.AINCB & obj.AMask:    {Flags: gc.SizeB | RightRdwr},
	x86.AINCL & obj.AMask:    {Flags: gc.SizeL | RightRdwr},
	x86.AINCW & obj.AMask:    {Flags: gc.SizeW | RightRdwr},
	x86.AJCC & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJCS & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJEQ & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJGE & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJGT & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJHI & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJLE & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJLS & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJLT & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJMI & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJNE & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJOC & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJOS & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJPC & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJPL & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	x86.AJPS & obj.AMask:     {Flags: gc.Cjmp | gc.UseCarry},
	obj.AJMP:                 {Flags: gc.Jump | gc.Break | gc.KillCarry},
	x86.ALEAW & obj.AMask:    {Flags: gc.LeftAddr | gc.RightWrite},
	x86.ALEAL & obj.AMask:    {Flags: gc.LeftAddr | gc.RightWrite},
	x86.AMOVBLSX & obj.AMask: {Flags: gc.SizeL | gc.LeftRead | gc.RightWrite | gc.Conv},
	x86.AMOVBLZX & obj.AMask: {Flags: gc.SizeL | gc.LeftRead | gc.RightWrite | gc.Conv},
	x86.AMOVBWSX & obj.AMask: {Flags: gc.SizeW | gc.LeftRead | gc.RightWrite | gc.Conv},
	x86.AMOVBWZX & obj.AMask: {Flags: gc.SizeW | gc.LeftRead | gc.RightWrite | gc.Conv},
	x86.AMOVWLSX & obj.AMask: {Flags: gc.SizeL | gc.LeftRead | gc.RightWrite | gc.Conv},
	x86.AMOVWLZX & obj.AMask: {Flags: gc.SizeL | gc.LeftRead | gc.RightWrite | gc.Conv},
	x86.AMOVB & obj.AMask:    {Flags: gc.SizeB | gc.LeftRead | gc.RightWrite | gc.Move},
	x86.AMOVL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RightWrite | gc.Move},
	x86.AMOVW & obj.AMask:    {Flags: gc.SizeW | gc.LeftRead | gc.RightWrite | gc.Move},
	x86.AMOVSB & obj.AMask:   {Flags: gc.OK},
	x86.AMOVSL & obj.AMask:   {Flags: gc.OK},
	x86.AMOVSW & obj.AMask:   {Flags: gc.OK},
	obj.ADUFFCOPY:            {Flags: gc.OK},
	x86.AMOVSD & obj.AMask:   {Flags: gc.SizeD | gc.LeftRead | gc.RightWrite | gc.Move},
	x86.AMOVSS & obj.AMask:   {Flags: gc.SizeF | gc.LeftRead | gc.RightWrite | gc.Move},

	// We use MOVAPD as a faster synonym for MOVSD.
	x86.AMOVAPD & obj.AMask:  {Flags: gc.SizeD | gc.LeftRead | gc.RightWrite | gc.Move},
	x86.AMULB & obj.AMask:    {Flags: gc.SizeB | gc.LeftRead | gc.SetCarry},
	x86.AMULL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.SetCarry},
	x86.AMULW & obj.AMask:    {Flags: gc.SizeW | gc.LeftRead | gc.SetCarry},
	x86.AMULSD & obj.AMask:   {Flags: gc.SizeD | gc.LeftRead | RightRdwr},
	x86.AMULSS & obj.AMask:   {Flags: gc.SizeF | gc.LeftRead | RightRdwr},
	x86.ANEGB & obj.AMask:    {Flags: gc.SizeB | RightRdwr | gc.SetCarry},
	x86.ANEGL & obj.AMask:    {Flags: gc.SizeL | RightRdwr | gc.SetCarry},
	x86.ANEGW & obj.AMask:    {Flags: gc.SizeW | RightRdwr | gc.SetCarry},
	x86.ANOTB & obj.AMask:    {Flags: gc.SizeB | RightRdwr},
	x86.ANOTL & obj.AMask:    {Flags: gc.SizeL | RightRdwr},
	x86.ANOTW & obj.AMask:    {Flags: gc.SizeW | RightRdwr},
	x86.AORB & obj.AMask:     {Flags: gc.SizeB | gc.LeftRead | RightRdwr | gc.SetCarry},
	x86.AORL & obj.AMask:     {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.SetCarry},
	x86.AORW & obj.AMask:     {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.SetCarry},
	x86.APOPL & obj.AMask:    {Flags: gc.SizeL | gc.RightWrite},
	x86.APUSHL & obj.AMask:   {Flags: gc.SizeL | gc.LeftRead},
	x86.APXOR & obj.AMask:    {Flags: gc.SizeD | gc.LeftRead | RightRdwr},
	x86.ARCLB & obj.AMask:    {Flags: gc.SizeB | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry | gc.UseCarry},
	x86.ARCLL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry | gc.UseCarry},
	x86.ARCLW & obj.AMask:    {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry | gc.UseCarry},
	x86.ARCRB & obj.AMask:    {Flags: gc.SizeB | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry | gc.UseCarry},
	x86.ARCRL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry | gc.UseCarry},
	x86.ARCRW & obj.AMask:    {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry | gc.UseCarry},
	x86.AREP & obj.AMask:     {Flags: gc.OK},
	x86.AREPN & obj.AMask:    {Flags: gc.OK},
	obj.ARET:                 {Flags: gc.Break | gc.KillCarry},
	x86.AROLB & obj.AMask:    {Flags: gc.SizeB | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.AROLL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.AROLW & obj.AMask:    {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ARORB & obj.AMask:    {Flags: gc.SizeB | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ARORL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ARORW & obj.AMask:    {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ASAHF & obj.AMask:    {Flags: gc.OK},
	x86.ASALB & obj.AMask:    {Flags: gc.SizeB | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ASALL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ASALW & obj.AMask:    {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ASARB & obj.AMask:    {Flags: gc.SizeB | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ASARL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ASARW & obj.AMask:    {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ASBBB & obj.AMask:    {Flags: gc.SizeB | gc.LeftRead | RightRdwr | gc.SetCarry | gc.UseCarry},
	x86.ASBBL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.SetCarry | gc.UseCarry},
	x86.ASBBW & obj.AMask:    {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.SetCarry | gc.UseCarry},
	x86.ASETCC & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETCS & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETEQ & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETGE & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETGT & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETHI & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETLE & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETLS & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETLT & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETMI & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETNE & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETOC & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETOS & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETPC & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETPL & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASETPS & obj.AMask:   {Flags: gc.SizeB | RightRdwr | gc.UseCarry},
	x86.ASHLB & obj.AMask:    {Flags: gc.SizeB | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ASHLL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ASHLW & obj.AMask:    {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ASHRB & obj.AMask:    {Flags: gc.SizeB | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ASHRL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ASHRW & obj.AMask:    {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.ShiftCX | gc.SetCarry},
	x86.ASTOSB & obj.AMask:   {Flags: gc.OK},
	x86.ASTOSL & obj.AMask:   {Flags: gc.OK},
	x86.ASTOSW & obj.AMask:   {Flags: gc.OK},
	obj.ADUFFZERO:            {Flags: gc.OK},
	x86.ASUBB & obj.AMask:    {Flags: gc.SizeB | gc.LeftRead | RightRdwr | gc.SetCarry},
	x86.ASUBL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.SetCarry},
	x86.ASUBW & obj.AMask:    {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.SetCarry},
	x86.ASUBSD & obj.AMask:   {Flags: gc.SizeD | gc.LeftRead | RightRdwr},
	x86.ASUBSS & obj.AMask:   {Flags: gc.SizeF | gc.LeftRead | RightRdwr},
	x86.ATESTB & obj.AMask:   {Flags: gc.SizeB | gc.LeftRead | gc.RightRead | gc.SetCarry},
	x86.ATESTL & obj.AMask:   {Flags: gc.SizeL | gc.LeftRead | gc.RightRead | gc.SetCarry},
	x86.ATESTW & obj.AMask:   {Flags: gc.SizeW | gc.LeftRead | gc.RightRead | gc.SetCarry},
	x86.AUCOMISD & obj.AMask: {Flags: gc.SizeD | gc.LeftRead | gc.RightRead},
	x86.AUCOMISS & obj.AMask: {Flags: gc.SizeF | gc.LeftRead | gc.RightRead},
	x86.AXCHGB & obj.AMask:   {Flags: gc.SizeB | LeftRdwr | RightRdwr},
	x86.AXCHGL & obj.AMask:   {Flags: gc.SizeL | LeftRdwr | RightRdwr},
	x86.AXCHGW & obj.AMask:   {Flags: gc.SizeW | LeftRdwr | RightRdwr},
	x86.AXORB & obj.AMask:    {Flags: gc.SizeB | gc.LeftRead | RightRdwr | gc.SetCarry},
	x86.AXORL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | RightRdwr | gc.SetCarry},
	x86.AXORW & obj.AMask:    {Flags: gc.SizeW | gc.LeftRead | RightRdwr | gc.SetCarry},
}

func proginfo(p *obj.Prog) gc.ProgInfo {
	info := progtable[p.As&obj.AMask]
	if info.Flags == 0 {
		gc.Fatalf("unknown instruction %v", p)
	}

	if info.Flags&gc.ImulAXDX != 0 && p.To.Type != obj.TYPE_NONE {
		info.Flags |= RightRdwr
	}

	return info
}
