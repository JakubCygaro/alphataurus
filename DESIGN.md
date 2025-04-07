# Alphataurus Assembly

The machine will be based around a virtual value stack.
Each value is a pointer to a boxed object allocated on the heap.
Additionally there will be a set of registers.
Operations on those objects are determined at runtime.

```
AObject[<somesize>] stack;
```

The stack can be manipulated through a set of instructions.

push [value]
pop
mov [src], [dst]

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


```
