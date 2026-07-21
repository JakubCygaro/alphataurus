- fix the bullshit that is col and line
- variables in the assembler, basically how its done in FASM
- assignment statement, that can be used in any section
```
thing = 1234

```
- special variables like in FASM: $, ., etc.
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
- the decoder codegen would use comments to figure out
 how many parameters the opcode takes
```Go
//go generate tool -type=OpcodeVal
const (
    //tool:params(2)
    OP_MOVRR OpcodeVal = iota
)
```
- future decoder idea:
```Go
// radix tree for opcode lookup
type opcodeTrie struct {  }
// array of radix trees that is indexed by the first opcode byte
// OR just a single Trie from which the lookup starts
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
 maybe it would be possible to flatten everything into a single array so that
 any and all access requires only indices into the array and minimal pointer dereference

```
arr:
[... byte, child_count, children ...]

```
 maybe you could lay out the opcodes flat in an array and then use a second
 array for lookup
```
map:
0x00000011 -> OP_MOVRR
ops:
0          1         2
[OP_MOVRR, OP_MOVIR, OP_AND, ...]

```
 *build a prefix tree, then flatten it into a hashmap*
