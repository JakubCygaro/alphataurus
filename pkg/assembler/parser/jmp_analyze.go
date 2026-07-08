package assembler

import (
	"fmt"

	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func GetConcreteJmpInst(genericCmp InstGenericJmp) (any, error) {
	switch addr := genericCmp.Expr.Val.(type) {
	case ConstExpr:
		if i, ok := addr.Val.(ConstExprILit); !ok {
			return nil, fmt.
				Errorf("TODO: Invalid expression as jump address %v %v",
					genericCmp.Expr.Line, genericCmp.Expr.Col)
		} else {
			return InstJmpI{
				Address: i.Signed(),
				JmpTy:   genericCmp.Variant,
			}, nil
		}
	case DerefExpr:
		switch deref := addr.Inner.Val.(type) {
		case RegExpr:
			if deref.Reg.Reg != vm.IP_IDX {
				return nil, fmt.Errorf("TODO: expected IP register in jump")
			}
			return InstJmpIP0R{
				Offset: 0,
				OpTy: vm.OP_TADD,
				JmpTy:   genericCmp.Variant,
			}, nil
		case OneRegOffsetExpr:
			if deref.Reg.Reg != vm.IP_IDX {
				return nil, fmt.Errorf("TODO: expected IP register in jump")
			}
			if off, ok := getAsOffset(deref.Offset); !ok {
				return nil, fmt.Errorf("TODO: bad offset")
			} else {
				return InstJmpIP0R{
					Offset: off,
					OpTy: vm.OP_TADD,
					JmpTy:   genericCmp.Variant,
				}, nil
			}
		case TwoRegOffsetExpr:
			//TODO: more complex checks
			if deref.Reg1.Reg != vm.IP_IDX {
				return nil, fmt.Errorf("TODO: expected IP register in jump")
			}
			if off, ok := getAsOffset(deref.Offset); !ok {
				return nil, fmt.Errorf("TODO: bad offset")
			} else {
				return InstJmpIP1R{
					Reg: deref.Reg2,
					Offset: off,
					OpTy: deref.RegOp,
					JmpTy:   genericCmp.Variant,
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("TODO: invalid jump instruction expression")
}
func GetConcreteCallInst(genericCmp InstGenericCall) (any, error) {
	switch addr := genericCmp.Expr.Val.(type) {
	case ConstExpr:
		if i, ok := addr.Val.(ConstExprILit); !ok {
			return nil, fmt.
				Errorf("TODO: Invalid expression as call address %v %v",
					genericCmp.Expr.Line, genericCmp.Expr.Col)
		} else {
			return InstCallI{
				Address: i.Signed(),
			}, nil
		}
	case DerefExpr:
		switch deref := addr.Inner.Val.(type) {
		case RegExpr:
			if deref.Reg.Reg != vm.IP_IDX {
				return nil, fmt.Errorf("TODO: expected IP register in call")
			}
			return InstCallIP0R{
				Offset: 0,
				OpTy: vm.OP_TADD,
			}, nil
		case OneRegOffsetExpr:
			if deref.Reg.Reg != vm.IP_IDX {
				return nil, fmt.Errorf("TODO: expected IP register in call")
			}
			if off, ok := getAsOffset(deref.Offset); !ok {
				return nil, fmt.Errorf("TODO: bad offset")
			} else {
				return InstCallIP0R{
					Offset: off,
					OpTy: vm.OP_TADD,
				}, nil
			}
		case TwoRegOffsetExpr:
			//TODO: more complex checks
			if deref.Reg1.Reg != vm.IP_IDX {
				return nil, fmt.Errorf("TODO: expected IP register in call")
			}
			if off, ok := getAsOffset(deref.Offset); !ok {
				return nil, fmt.Errorf("TODO: bad offset")
			} else {
				return InstCallIP1R{
					Reg: deref.Reg2,
					Offset: off,
					OpTy: deref.RegOp,
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("TODO: invalid call instruction expression")
}
