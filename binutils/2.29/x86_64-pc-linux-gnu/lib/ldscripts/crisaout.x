/* Default linker script, for normal executables */
/* Copyright (C) 2014-2017 Free Software Foundation, Inc.
   Copying and distribution of this script, with or without modification,
   are permitted in any medium without royalty provided the copyright
   notice and this notice are preserved.  */
OUTPUT_FORMAT("a.out-cris")
OUTPUT_ARCH(cris)
ENTRY (__start)
SECTIONS
{
  .text  0:
  {
   CREATE_OBJECT_SYMBOLS;
     __Stext = .;
    *(.startup)
    *(.text)
    __start = DEFINED(__start) ? __start :
		   DEFINED(_start) ? _start :
		     DEFINED(start) ? start :
		        DEFINED(.startup) ? .startup + 2 : 2;
    *(.text.*)
    *(.gnu.linkonce.t*)
    *(.rodata)
    *(.rodata.*)
    *(.gnu.linkonce.r*)
    /* Do not "provide" init-start and fini-start symbols; they might be
       referred to weakly, so the linker would not override the zero
       default.
       FIXME: It's somewhat unexpected to have code emitted by the linker
       script.  Some other mechanism could probably do better.  */
     . = ALIGN (2);
      ___init__start = .;
     PROVIDE (___do_global_ctors = .);
     SHORT (0xe1fc); /* push srp */
     SHORT (0xbe7e);
     *(.init)
     SHORT (0x0d3e); /* jump [sp+] */
     PROVIDE (__init__end = .);
     PROVIDE (___init__end = .);
     . = ALIGN (2);
      ___fini__start = .;
     PROVIDE (___do_global_dtors = .);
     SHORT (0xe1fc); /* push srp */
     SHORT (0xbe7e);
     *(.fini)
     SHORT (0x0d3e); /* jump [sp+] */
     PROVIDE (__fini__end = .);
      ___fini__end = .;
    /* Cater to linking from ELF.  */
     PROVIDE(___ctors = .);
     ___elf_ctors_dtors_begin = .;
     KEEP (*crtbegin.o(.ctors))
     KEEP (*crtbegin?.o(.ctors))
     KEEP (*(EXCLUDE_FILE (*crtend.o *crtend?.o) .ctors))
     KEEP (*(SORT(.ctors.*)))
     KEEP (*(.ctors))
     PROVIDE(___ctors_end = .);
     PROVIDE(___dtors = .);
     KEEP (*crtbegin.o(.dtors))
     KEEP (*crtbegin?.o(.dtors))
     KEEP (*(EXCLUDE_FILE (*crtend.o *crtend?.o) .dtors))
     KEEP (*(SORT(.dtors.*)))
     KEEP (*(.dtors))
     PROVIDE(___dtors_end = .);
     ___elf_ctors_dtors_end = .;
    /* We include objects that force alignment of the data segment.
       Unfortunately that sometimes causes a gap between .text and .data,
       which is not detectable since .data does not have a start address
       of itself in the a.out header.  This should only matter for
       testing; for production use, .data is at a "known" location.
       We assume .data does not get an alignment larger than 32 bytes.  */
    . = ALIGN (32);
     __Etext = .;
    /* Deprecated, use __Etext.  */
     PROVIDE(_etext = .);
  }
  /* Any dot-relative start-expression (such as "ALIGN(2)", also including
     the "default" .data alignment expression) will use the initial, raw
     size of .text and will be incorrect if the alignment used is less
     than the alignment for .text (which might depend on input and obj
     format).  FIXME: Seems like a bug in ld.  Seems hard to fix.  Seems
     unimportant.  */
  .data :
  {
     __Sdata = .;
    *(.data);
    *(.data.*)
    *(.gnu.linkonce.d*)
    *(.eh_frame) /* FIXME: Make .text */
    *(.gcc_except_table)
    /* See comment at ALIGN before __Etext.  */
    . = ALIGN (32);
     __Edata = .;
    /* Deprecated, use __Edata.  */
     PROVIDE(_edata = .);
  }
  .bss :
  {
    /* Deprecated, use __Sbss.  */
     PROVIDE(_bss_start = .);
     __Sbss = .;
    *(.bss)
    *(.bss.*)
    *(COMMON)
     __Ebss = .;
    /* Deprecated, use __Ebss or __Eall as appropriate.  */
     PROVIDE(_end = .);
     PROVIDE(__end = .);
  }
   __Eall = .;
  /* Unfortunately, stabs are not mappable from ELF to a.out.
     It can probably be fixed with some amount of work.  */
  /DISCARD/ :
  { *(.stab) *(.stab*) *(.debug) *(.debug*) *(.comment) *(.gnu.warning.*) }
  /* For the rsim and xsim simulators.  */
   PROVIDE(__Endmem = 0x10000000);
  /* For elinux.  */
   PROVIDE(__Stacksize = 0);
}
