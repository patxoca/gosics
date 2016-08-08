.. -*- ispell-local-dictionary: "british" -*-

This is a toy project to get into the go language. It implements a
SBNZ based single instruction set computer.

According to the entry in the
`wikipedia<https://en.wikipedia.org/wiki/One_instruction_set_computer>_`,
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
subtracts the contents at address ``a`` from the contents at address
``b``, stores the result at address ``c``, and then, if the result is
not 0, transfers control to address ``d`` (if the result is equal
zero, execution proceeds to the next instruction in sequence).


Architecture
============

In order to keep things simple the simulator has two independent
address spaces for program code and data. Each address block is 256
words long, so the pointers are 8 bits long (``uint8``). For the
program block the words are 32 bits long (4 bytes == 4 pointers) and
for the data block the words are 8 bits long (a numerical value in the
range -128..127).

For the program block the address 255 is reserved and can't contain an
opcode. Jumping to that address halts the simulation. The
``MaxProgramAddress`` constant points to that address.

Although not strictly required, for the data block there are three
reserved addresses:

- 253: contains a 1
- 254: contains a 0
- 255: used for intermediate results

The constants ``ONE``, ``ZERO`` and ``JUNK`` point to those addresses.

By the way, if we want to stop the simulation we can do::

  SBNZ ONE, ZERO, JUNK, MaxProgramAddress

Wich will substract ZERO from ONE, store the result in JUNK, an since
the result is 1 (!= 0), it will jump to the halt address.


Current status
==============

At this point the ``Computer`` class is funcional, and there's a
simple *in memory* ``Assembler``class which emulates complex
instructions.


Example
=======

First we need to create an assembler and *write* the program. In this
example we'll multiply the numbers in adresses 0 and 1, by repeated
sums, and store the result in address 3. The address 4 is used for a
counter. For the sake of simplicity we assume that both operands are
possitive.

.. code-block:: go

    const COUNTER DataAddress = 3
    const RESULT DataAddress = 2

    ass := NewAssembler()

    ass.MOV(0, COUNTER)
    ass.MOV(ZERO, RESULT)
    ass.label("loop")
    ass.BEQ(COUNTER, ZERO, Label("exit_loop"))
    ass.ADD(1, RESULT, RESULT)
    ass.DEC(COUNTER)
    ass.JMP(Label("loop"))
    ass.label("exit_loop")
    ass.HLT()

We need the initial contents for the data memory:

.. code-block:: go

    data := []DataCell{
        2, 3,
    }

Then we create the computer and load both the program and the data:

.. code-block:: go

    computer := new(Computer)
    computer.load_program(ass.assemble())
    computer.load_data(data)

And finally we can run the program:

.. code-block:: go

    computer.PrintProgramMemory()
    computer.PrintDataMemory(4)
    for !computer.Halted() {
        computer.Step()
    }
    computer.PrintDataMemory(4)

And we should get in the screen the result: the program dump in terms
of SBNZ instructions and the memory dumps before and after the
execution::

  Program memory dump
    0   0   z   3   1
    1   z   z   2   2
    2   3   z   j   4
    3   o   z   j   8
    4   z   2   j   5
    5   1   j   2   6
    6   3   o   3   7
    7   o   z   j   2
    8   o   z   j   h
      ...
  IP= 0
  02 03 00 00 ... 01 00 00
  IP= 255
  02 03 06 00 ... 01 00 01

So 2 * 3 = 6, great!!
