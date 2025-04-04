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
