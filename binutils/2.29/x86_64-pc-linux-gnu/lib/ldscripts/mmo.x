/* Default linker script, for normal executables */
/* Copyright (C) 2014-2017 Free Software Foundation, Inc.
   Copying and distribution of this script, with or without modification,
   are permitted in any medium without royalty provided the copyright
   notice and this notice are preserved.  */
OUTPUT_FORMAT("mmo")
OUTPUT_ARCH(mmix)
ENTRY(Main)
SECTIONS
{
  .text  DEFINED (__.MMIX.start..text) ? __.MMIX.start..text : 0:
  {
    *(.text)
    *(.text.*)
    *(.gnu.linkonce.t*)
    *(.rodata)
    *(.rodata.*)
    *(.gnu.linkonce.r*)
    /* FIXME: Move .init, .fini, .ctors and .dtors to their own sections.  */
     PROVIDE (_init_start = .);
     PROVIDE (_init = .);
     KEEP (*(SORT_NONE(.init)))
     PROVIDE (_init_end = .);
     PROVIDE (_fini_start = .);
     PROVIDE (_fini = .);
     KEEP (*(SORT_NONE(.fini)))
     PROVIDE (_fini_end = .);
    /* FIXME: Align ctors, dtors, ehframe.  */
     PROVIDE (_ctors_start = .);
     PROVIDE (__ctors_start = .);
     PROVIDE (_ctors = .);
     PROVIDE (__ctors = .);
     KEEP (*crtbegin.o(.ctors))
     KEEP (*crtbegin?.o(.ctors))
     KEEP (*(EXCLUDE_FILE (*crtend.o *crtend?.o) .ctors))
     KEEP (*(SORT(.ctors.*)))
     KEEP (*(.ctors))
     PROVIDE (_ctors_end = .);
     PROVIDE (__ctors_end = .);
     PROVIDE (_dtors_start = .);
     PROVIDE (__dtors_start = .);
     PROVIDE (_dtors = .);
     PROVIDE (__dtors = .);
     KEEP (*crtbegin.o(.dtors))
     KEEP (*crtbegin?.o(.dtors))
     KEEP (*(EXCLUDE_FILE (*crtend.o *crtend?.o) .dtors))
     KEEP (*(SORT(.dtors.*)))
     KEEP (*(.dtors))
     PROVIDE (_dtors_end = .);
     PROVIDE (__dtors_end = .);
    KEEP (*(.jcr))
    KEEP (*(.eh_frame))
    *(.gcc_except_table)
    Main = DEFINED (Main) ? Main : (DEFINED (_start) ? _start : ADDR (.text));
  }
  /* The following NOP assignment and those after .data and .bss, are
     necessary to get orphan sections adopted by the .text inserted before
     the following end-section symbols.  An output section would also serve
     this purpose, but we can't do that.  */
  . = .;
   PROVIDE(etext = .);
   PROVIDE(_etext = .);
   PROVIDE(__etext = .);
  .data  DEFINED (__.MMIX.start..data) ? __.MMIX.start..data : 0x2000000000000000:
  {
     PROVIDE(__Sdata = .);
    *(.data);
    *(.data.*)
    *(.gnu.linkonce.d*)
  }
  . = .;
   PROVIDE(__Edata = .);
  /* Deprecated, use __Edata.  */
   PROVIDE(edata = .);
   PROVIDE(_edata = .);
   PROVIDE(__edata = .);
  /* At the moment, although perhaps we should, we can't map sections
     without contents to sections *with* contents due to FIXME: a BFD bug.
     Anyway, the mmo back-end ignores sections without contents when
     writing out sections, so this works fine.   */
  .bss :
  {
     PROVIDE(__Sbss = .);
     PROVIDE(__bss_start = .);
     *(.sbss);
     *(.bss);
    *(.bss.*)
     *(COMMON);
  }
  . = .;
   PROVIDE(__Ebss = .);
  /* Deprecated, use __Ebss or __Eall as appropriate.  */
   PROVIDE(end = .);
   PROVIDE(_end = .);
   PROVIDE(__end = .);
   PROVIDE(__Eall = .);
  .stab 0 : { *(.stab) }
  .stabstr 0 : { *(.stabstr) }
  .stab.excl 0 : { *(.stab.excl) }
  .stab.exclstr 0 : { *(.stab.exclstr) }
  .stab.index 0 : { *(.stab.index) }
  .stab.indexstr 0 : { *(.stab.indexstr) }
  /* DWARF debug sections.
     Symbols in the DWARF debugging sections are relative to the beginning
     of the section so we begin them at 0.  */
  /* DWARF 1 */
  .debug          0 : { *(.debug) }
  .line           0 : { *(.line) }
  /* GNU DWARF 1 extensions */
  .debug_srcinfo  0 : { *(.debug_srcinfo) }
  .debug_sfnames  0 : { *(.debug_sfnames) }
  /* DWARF 1.1 and DWARF 2 */
  .debug_aranges  0 : { *(.debug_aranges) }
  .debug_pubnames 0 : { *(.debug_pubnames) }
  /* DWARF 2 */
  .debug_info     0 : { *(.debug_info .gnu.linkonce.wi.*) }
  .debug_abbrev   0 : { *(.debug_abbrev) }
  .debug_line     0 : { *(.debug_line .debug_line.* .debug_line_end ) }
  .debug_frame    0 : { *(.debug_frame) }
  .debug_str      0 : { *(.debug_str) }
  .debug_loc      0 : { *(.debug_loc) }
  .debug_macinfo  0 : { *(.debug_macinfo) }
  /* SGI/MIPS DWARF 2 extensions */
  .debug_weaknames 0 : { *(.debug_weaknames) }
  .debug_funcnames 0 : { *(.debug_funcnames) }
  .debug_typenames 0 : { *(.debug_typenames) }
  .debug_varnames  0 : { *(.debug_varnames) }
  /* DWARF 3 */
  .debug_pubtypes 0 : { *(.debug_pubtypes) }
  .debug_ranges   0 : { *(.debug_ranges) }
  /* DWARF Extension.  */
  .debug_macro    0 : { *(.debug_macro) }
  .debug_addr     0 : { *(.debug_addr) }
  .MMIX.reg_contents :
  {
    /* Note that this section always has a fixed VMA - that of its
       first register * 8.  */
    *(.MMIX.reg_contents.linker_allocated);
    *(.MMIX.reg_contents);
  }
  /* By default, put the high end of the stack where the register stack
     begins.  They grow in opposite directions.  */
  PROVIDE (__Stack_start = 0x6000000000000000);
  /* Unfortunately, stabs are not mappable from ELF to MMO.
     It can probably be fixed with some amount of work.  */
  /DISCARD/ :
  {  *(.gnu.warning.*); }
  .gnu.attributes 0 : { KEEP (*(.gnu.attributes)) }
}
