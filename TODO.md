- TryEvalConstExprWithLabels -> takes a map of label-value pairs that it uses while trying to evaluate expressions
- better commutative expressions comp-time evalutaion wit fat trees,
 basically the Add and Mul Expression nodes keep a list of children expressions
 instead of being binary.

    +
   /|\
  / | \
  1 2 -
      |\
      | \
      2 3
- fix the bullshit that is col and line
- lea
- syscall
- variables in the assembler, basically how its done in FASM
- label deref [start]
- add more tests
- add movzz


r0  -> 64-bits
r0h -> 32-bits (h for half word)
r0q -> 16-bits (q for quarter word)
r0o -> 8-bits (o for octave)

mov BYTE [bp + 8], r0o
