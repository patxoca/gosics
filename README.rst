.. -*- ispell-local-dictionary: "british" -*-

This is a toy project to get into the go language. It implements a
SBNZ based single instruction set computer.

According to the entry in the
`wikipedia <https://en.wikipedia.org/wiki/One_instruction_set_computer>`_,
a *one instruction set computer* (OISC), sometimes called an *ultimate
reduced instruction set computer* (URISC), is an abstract machine that
uses only one instruction – obviating the need for a machine language
opcode. With a judicious choice for the single instruction and given
infinite resources, an OISC is capable of being a universal computer
in the same manner as traditional computers that have multiple
instructions. OISCs have been recommended as aids in teaching computer
architecture and have been used as computational models in structural
computing research.

The instruction chosen for that simulator is ``SBNZ``, *Subtract and
Branch if Not equal to Zero*: the ``SBNZ a, b, c, d`` instruction
subtracts the contents at address ``b`` from the contents at address
``a``, stores the result at address ``c``, and then, if the result is
not 0, transfers control to address ``d`` (if the result is equal
zero, execution proceeds to the next instruction in sequence).


Architecture
============

The ``Computer`` has 64Kb of memory shared by both program an data.
The memory is big endian.

Each instruction takes 8 bytes of memory (4 pointers * 2
bytes/pointer). The instruction ``SBNZ 0x0123, 0x4567, 0x89AB,
0xCDEF`` is stored in memory as::

  i    i+1  i+2  i+3  i+4  i+5  i+6  i+7
  0x01 0x23 0x45 0x67 0x89 0xAB 0xCD 0xEF

Each *value* take 2 bytes of memory. The value ``0x1234`` is stored in
memory as::

  i    i+1
  0x12 0x32

The address 65535 (0xFFFF) is special, jumping to that address halts
the computer.


The assembler
=============

At this point there's an *in memory* assembler that helps writing
programs. In a future a real assembler, that parses the program from a
file, may be implemented.

Labels
------

The assembler supports both literal and symbolic adresses, labels::

   SBNZ DATA, ...
   ; mode code here

   DATA:
   ; data here

To avoid collisions, labels starting with a double underscore are
reserved for the assembler.


Directives
----------

The assembler defines the directives ``DB`` and ``DD`` to insert
literal data into the program::

   DATA:
   DD 1 2

``DB`` inserts a sequence of bytes while ``DD`` inserts a sequence of
doubles (two bytes).


Macro instructions
------------------

The assembler define some macro instructions that generate code, in
terms of the SBNZ instructions, for some usual opcodes:

- ``MOV(src, dst)``: move content of ``src`` to ``dst``

  .. code-block:: asm

     SBNZ src, __ZERO, dst, __next
     __next:


- ``JMP(dst)``: jump inconditionally to ``dst``

  .. code-block:: asm

     SBNZ __ONE, __ZERO, __JUNK, dst

Run ``go doc assembler.Assembler`` to get a listing of all
opcodes.

The ``PUSH`` and ``POP`` macro instructions are more interesting,
those instructions modify the program in order to simulate the stack.
The implementation relies on some suport code in the program's
*preamble*.

Here's the suport code for the ``PUSH`` opcode:

.. code-block:: asm

   __push_operand:
     DD 0
   __SP:
     DD 0xFFFE
   __push:
     ; copy the content of __SP in the C operand of the next instruction
     SBNZ __SP, __ZERO, <12>, <8>
     ; copy the value to the top of the stack
     SBNZ __push_operand, __ZERO, 0xFFFE, <8>
     ; decrease __SP twice
     SBNZ __SP, __ONE, __SP, <8>
     SBNZ __SP, __ONE, __SP, <8>
     ; insert a SBNZ instruction that will jump inconditionally
     ; the client code must overwrite the contents of __push_ret
     ; with the "return" address
     DD __ONE __ZERO __JUNK
   __push_ret:
     DD 0xFFFF

The ``<>`` represent offsets relative to the IP of the current
instruction. That's not supported by the assembler, is just for
illustrative purposes. So ``<12>`` points to the C operand and ``<8>``
to the begining of the next instruction.

The ``PUSH`` opcode is something like:

.. code-block:: asm

   ; store to operand in __push_operand
     SBNZ SRC, __ZERO, __push_operand, <8>
   ; overwrite the "return"" address
     SBNZ data, __ZERO, __push_ret, <8>
   ; jump to __push
     SBNZ __ONE, __ZERO, __JUNK, __push
   ; jump over the data. The return address points here
     SBNZ __ONE, __ZERO, __JUNK, exit
   data:
     DD <-8>
   exit:


Memory layout
-------------

For the sake of convenience the assembler pre-allocates 6 bytes of
memory for 3 operands and defines 3 labels pointing to them:

- ``__ONE``: contains a 1

- ``__ZERO``: contains a 0. That's not strictly required since we can
  get a 0 substracting 1 from 1, buts it's convenient.

- ``__JUNK``: temporary storage, use with care.

When writing a program we can use the constants ``assembler.ONE``,
``assembler.ZERO`` and ``assembler.JUNK`` to reference those
addresses.

The assembler inserts the following preamble in each program:

.. code-block:: asm

   SBNZ __ONE, __ZERO, __JUNK, __start
   __ONE:
   DD 0x0001
   __ZERO:
   DD 0x0000
   __JUNK:
   DD 0x0000
   ;; runtime
   __start:

the first instruction jumps over the data block and the program code
starts at address ``__start``.


Example
-------

First we need to create an assembler and *write* the program. In this
example we'll multiply the numbers in adresses 0 and 1, by repeated
sums, and store the result in address 3. The address 4 is used for a
counter. For the sake of simplicity we assume that both operands are
possitive.

.. code-block:: go

    // pre define labels for readability
    OP1 := assembler.Label("OP1")
    OP2 := assembler.Label("OP2")
    DST := assembler.Label("DST")
    CNT := assembler.Label("CNT")
    LOO := assembler.Label("loop")
    ELO := assembler.Label("exit_loop")

    ass := assembler.New()

    ass.MOV(OP1, CNT)
    ass.MOV(assembler.ZERO, DST)
    ass.Label(LOO)
    ass.BEQ(CNT, assembler.ZERO, ELO)
    ass.ADD(OP2, DST, DST)
    ass.DEC(CNT)
    ass.JMP(LOO)
    ass.Label(ELO)
    ass.HLT()

    ass.Label(OP1)
    ass.DD(0x03)
    ass.Label(OP2)
    ass.DD(0x02)
    ass.Label(DST)
    ass.DD(0x00)
    ass.Label(CNT)
    ass.DD(0x00)

Then we create the computer and load it's memory:

.. code-block:: go

    computer := new(Computer)
    computer.LoadMemory(ass.Assemble())

And finally we can run the program:

.. code-block:: go

    c.Print(N)
    for !c.Halted() {
        c.Step()
        c.Print(N)
    }

And we'll get the result at address 0x5a, 2 * 3 = 6, great!!


The memory contents, conveniently annotated for readability, after
loading the previous program are::

  00: 0008 000a 000c 006a

      ; __ONE:
  08: 0001
      ; __ZERO:
  0A: 0000
      ; __JUNK:
  0C: 0000

      ; __push_operand:
  0E: 0faba

      ; __SP:
  10: fffe

      ; __push:
  12: 0010 000a 001e 001a
  1A: 000e 000a fffe 0022
  22: 0010 0008 0010 002a
  2A: 0010 0008 0010 0032
  32: 0008 000a 000c ffff ; __push_ret points to 38

      ;; __pop:
  3A: 000a 0008 000c 0042
  42: 0010 000c 0010 004a
  4A: 0010 000c 0010 0052
  52: 0010 000a 005a 005a
  5A: fffe 000a 000e 0062
  62: 0008 000a 000c ffff ; __pop_ret points to ffff

      ; end of preamble

      ; __start:

      ; MOV OP1, CNT
  6A: 00b2 000a 00b8 0072

      ; MOV __ZERO, DST
  72: 000a 000a 00b6 007a

      ; LOO:
      ; BEQ CNT, __ZERO, ELO
  7A: 00b8 000a 000c 008a
  82: 0008 000a 000c 00aa

      ; ADD OP2, DST, DST
  8A: 000a 00b6 000c 0092
  92: 00b4 000c 00b6 009a

      ; DEC CNT
  9A: 00b8 0008 00b8 00a2

      ; JMP LOO
  A2: 0008 000a 000c 007a

      ; ELO:
      ; HLT
  AA: 0008 000a 000c ffff

      ; OP1
  B2: 0003

      ; OP2
  B4: 0002

      ; DST
  B6: 0000

      ; CNT
  B8: 0000
