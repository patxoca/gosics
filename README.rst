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

Each *value* take 2 bytes of memory. The value ``01234`` is stored
in memory as::

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

The assembler supports both literal and symbolic adresses, labels:

.. code-block:: go
   :linenos:

   as := assembler.New()
   ...
   as.SBNZ(Label("DATA"), ...)
   ...
   as.Label("DATA")
   ...

In line 5 we define the label ``DATA`` pointing to some address. Here
``Label`` is a method of the assembler object, it stores the label and
the address it points to.

In line 3 we reference the memory address pointed to by ``DATA``. Here
``Label`` is a type implementing the ``labeler`` interface (a type
that is able to compute an address).

To avoid collisions, labels starting by a double underscore are
reserved for the assembler.


Directives
----------

The assembler defines the directives ``DB`` and ``DD`` to insert data
into the program.

.. code-block:: go
   :linenos:

   as := assembler.New()
   ...
   as.Label("DATA")
   as.DD(1, 2, ...)

``DB`` inserts a sequence of bytes while ``DD`` inserts a sequence of
two bytes (*doubles*).


Macro instructions
------------------

The assembler define some macro instructions that generate code for
some usual opcodes in terms of the SBNZ instructions:

- ``MOV(src, dst)``: move content of ``src`` to ``dst``

  .. code-block:: asm

     SBNZ src, __ZERO, dst, __next
     __next:


- ``JMP(dst)``: jump inconditionally to ``dst``

  .. code-block:: asm

     SBNZ __ONE, __ZERO, __JUNK, dst


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

The assembler inserts the following prologue in each program:

.. code-block:: asm

   SBNZ __ONE, __ZERO, __JUNK, __start
   __ONE:
   DD 0x0001
   __ZERO:
   DD 0x0000
   __JUNK:
   DD 0x0000
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
    ass.DD(0x01)
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
