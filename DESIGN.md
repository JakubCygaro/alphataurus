# Alphataurus Assembly

The machine will be based around a virtual value stack.
Each value is a pointer to a boxed object allocated on the heap.
Additionally there will be a set of registers.
Operations on those objects are determined at runtime.

```
AObject[<somesize>] stack;
```

The stack can be manipulated through a set of instructions.

registers:

- sp -> stack pointer, points to the top of the stack
- ip -> instruction pointer, points to the next instruction
- bp -> base pointer, points to the base of the stack (frame)
- r0-r7 -> general purpose registers
- special comparison registers

instructions: (instructions are resolved at runtime)

- add <dest>, <src> -> add dest to src and store in dest
- sub -//-
- mul -//-
- div -//-
- mod -//-
- push <VAL> -> push value
- pop [dest] -> pop to value
- call [EXTERNAL] <func> -> push ip and jump to adress
- jmp(and others) -> jump to adress
- cmp
- test
- int -> interrupt signal, a way to implement pseudo syscalls

labels:
    lab_name: -> resolved to adresses by the assembler

@entry -> marks the begining of the code

@entry
main:
    push bp
    mov bp, sp
    mov r1, "Hello, World\n"
    call EXTERNAL print
    ret

push [value]
pop
mov [src], [dst]

function main(arr_size) {
   elem = get(arr_size);
}
function test(){}

@entry
main:
   mov r1, r1
   call get
   mov [bp + 1], r0
   ret
est:
   push bp,
   mov bp, sp
   ret


1: mov r1, r1
2: jmp EXTERNAL get ; call
3: mov [bp + 1], r0
4: pop bp ; ret
5: jmp [sp]
6: push bp ; push old base pointer
7: mov bp, sp ; make current stack pointer the base pointer
8: pop bp ; restore old stack pointer
9: jmp [sp] ; jump to the value at the top of the stack that is the next instruction after the call to this function


call [procedure] ->
    push ip
    jmp [procedure]

ret ->
    pop ip

## object creation

objects are quasi collections of primary type values, that need to be allocated in a special way
```
...
    makeobj r0 ; create an object and store its pointer in r0
    push [r0] 420 ; "push" the fields of this object into it
    push [r0] 69
    push [r0] "ur mum"
    ; at this point the object could be represented as
    ; struct { int, int, string }
    mov [sp], [r0 + 0] ; get the value of the first field
    lea [r1], [r0 + 0] ; get the adress of the first field
    mov [r1], 9999 ; modify the first field
    freeobj [r0] ; deletes the object at the specified adress, makes the pointer invalid
...


struct {
    int foo;
    int bar;
}
function new_struct() {
    ret = {}
    ret.foo = 69;
    ret.bar = 420;
    return ret;
}

new_struct:
    push bp
    mov bp, sp
    makeobj [bp + 1]
    mov [[bp + 1] + 1], 69
    mov [[bp + 1] + 2], 420
    mov r0, [bp + 1]
    ret

```
function main(array_sz) {
    if(array_sz > 0)
        get(0);
}

@entry
main:
    push bp
    mov bp, sp
    mov [bp + 1], r1
    mov r1, [bp + 1]
    mov r2, 0
    cmp r1, r2
    jle else
    mov r1, 0
    call EXTERNAL get
else:
    ret

# Alphataurus bytecode

assembly instructions get translated into opcodes, 64-bit wide

example:

mov r0, r1 -> (different op code depending on the parameter types)

movrr      r0  r1
0x00010001 0x0 0x1

mov [bp + 1], r1
movbpoffr (mov to bp offset from register)
0x00010069 0x1 0x1
