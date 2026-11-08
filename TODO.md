- mul should output a result of double the size of its operands,
 multiplying 2 halfwords yields a word
- TESTING FOR: rolr, rorr, test, lea
- static data in object files with proper linking and relocation
- fix the bullshit that is col and line
- variables in the assembler, basically how its done in FASM
- assignment statement, that can be used in any section
```
thing = 1234

```
- special variables like in FASM: $, ., etc.
- syscall
- add linker option for selecting entry point from a specific file:
ald foo.ao bar.ao -e foo.ao -o baz.aelf
 could also support picking a specific label from a file:
ald ... -e foo.ao:_entry -o baz.aelf
