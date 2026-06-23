- fix the bullshit that is col and line
- add remaining jump instructions
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
- variables in the assembler, basically how its done in FASM
- TryEvalConstExprWithLabels -> takes a map of variable-value pairs that it uses while trying to evaluate expressions
- lea
- syscall
- add movzz
- add test instruction
- add linker option for selecting entry point from a specific file:
ald foo.ao bar.ao -e foo.ao -o baz.aelf
 could also support picking a specific label from a file:
ald ... -e foo.ao:_entry -o baz.aelf
- possible reimplementation of the opcode decoder by writing a custom codegen tool
 that inspects the codebase for OP_CODE declarations and writes a fast decoder
- future decoder idea:
```Go
// radix tree for opcode lookup
type opcodeTrie struct {  }
// array of radix trees that is indexed by the first opcode byte
var prefixArray [256]opcodeTrie

func Lookup(opcode [4]byte)  {
    pref := prefixArray[opcode[3]]
    if pref.none {
        return nil
    }
    rem := opcode[:4]
    return pref.find(rem)
}


```
