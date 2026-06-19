package assembler

// start with a known value e.g 3030
// then do operations on it and recursively build an expression tree
//     add    mul
// 3000-- 1000
//     \- 2000-- 2
//            \- 1000
// after this tree is built it can be used to emit an expression string
// and then the const evaluated expression can be checked for corectness
// also the expression tree can be evaluated
