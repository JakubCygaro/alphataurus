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
