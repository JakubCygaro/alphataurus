- TryEvalConstExprWithLabels -> takes a map of label-value pairs that it uses while trying to evaluate expressions
- fix the bullshit that is col and line
- move labes into symbols, so there is only one place where the assembler stores symbols and looks for them
- @entry and AELF file support
- lea
- syscall
- variables in the assembler, basically how its done in FASM
- label deref [start]
- add more tests

