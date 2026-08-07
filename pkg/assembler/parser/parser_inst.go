package assembler

import (
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

//go:generate stringer -type=JmpVariant
type JmpVariant int

const (
	JMP JmpVariant = iota
	JMPE
	JMPZ
	JMPNE
	JMPNZ
	JMPG
	JMPGE
	JMPL
	JMPLE
	JMPS
	JMPNS
	JMPC
	JMPNC
	JMPO
	JMPNO
	JMPIP0R
	JMPIP1R
)

//go:generate stringer -type=ArthTy
type ArthTy int

const (
	ADD ArthTy = iota
	SUB
	DIV
	MUL
)

//go:generate stringer -type=LogTy
type LogTy int

const (
	AND LogTy = iota
	OR
	XOR
	LSH
	RSH
)

const (
	ARTH_TUNSIGNED = vm.TY_UINT
	ARTH_TSIGNED   = vm.TY_SINT
	ARTH_TFLOAT    = vm.TY_FLOAT
)

const (
	ARTH_TADD = iota
	ARTH_TSUB
	ARTH_TDIV
	ARTH_TMUL
)
const (
	LOG_TNOT = iota
	LOG_TAND
	LOG_TOR
	LOG_TXOR
	LOG_TLSH
	LOG_TRSH
)

type Instruction struct {
	// This is a Inst... struct type that contains the actual instruction specific data
	Data      any
	Line, Col int
}
type DataSize struct {
	Line, Col int
	Size      byte
}
type InstGenericMov struct {
	Dest, Src *Expr
	DataSize  *DataSize
}
type InstMovIR struct {
	Dest     lx.RegisterData
	Imm      uint64
	DataSize byte
}
type InstMovRR struct {
	Src, Dest lx.RegisterData
}
type InstMovSXRR struct {
	Mov InstMovRR
}
type InstGenericMovZX struct {
	Mov InstGenericMov
}
type InstMovZXRR struct {
	Mov InstMovRR
}
type InstMovZXIR struct {
	Mov InstMovIR
}
type InstMovZXDR struct {
	Mov InstMovDR
}
type InstMovZXDRO1 struct {
	Mov InstMovDRO1
}
type InstMovZXDRO2 struct {
	Mov InstMovDRO2
}
type InstGenericXCHG struct {
	Mov InstGenericMov
}
type InstXCHGRR struct {
	Mov InstMovRR
}
type InstXCHGDR struct {
	Mov InstMovDR
}
type InstXCHGRD struct {
	Mov InstMovRD
}
type InstXCHGDRO1 struct {
	Mov InstMovDRO1
}
type InstXCHGDRO2 struct {
	Mov InstMovDRO2
}
type InstInc struct {
	Reg lx.RegisterData
}
type InstDec struct {
	Reg lx.RegisterData
}
type InstGenericArth struct {
	Src, Dest *Expr
	Ty        int
	DataSize  byte
	ArthTy    ArthTy
}
type InstArthRR struct {
	Src, Dest lx.RegisterData
	Ty        int
	DataSize  byte
	ArthTy    ArthTy
}
type InstArthIR struct {
	Dest     lx.RegisterData
	Imm      uint64
	Ty       int
	DataSize byte
	ArthTy   ArthTy
}
type InstNotR struct {
	First lx.RegisterData
}
type InstNegR struct {
	Arg lx.RegisterData
	Float bool
}
type InstRotate struct {
	Arg lx.RegisterData
	Left bool
	RotateBy uint64
}
type InstGenericLogical struct {
	First, Second *Expr
	LogTy         LogTy
}
type InstLogicalRR struct {
	First, Second lx.RegisterData
	LogTy         LogTy
}
type InstLogicalIR struct {
	First lx.RegisterData
	Imm   uint64
	LogTy LogTy
}
type InstGenericCmp struct {
	Ty       int
	Sub, Min *Expr
}
type InstCmpRR struct {
	Ty       int
	Sub, Min lx.RegisterData
}
type InstCmpIR struct {
	Ty  int
	Min lx.RegisterData
	Imm uint64
}
type InstGenericJmp struct {
	Expr     *Expr
	Absolute bool
	Variant  JmpVariant
}
type InstJmpI struct {
	Address int64
	JmpTy   JmpVariant
}
type InstJmpIP0R struct {
	Offset int64
	OpTy   int
	JmpTy  JmpVariant
}
type InstJmpIP1R struct {
	Reg    lx.RegisterData
	Offset int64
	OpTy   int
	JmpTy  JmpVariant
}
type InstCallI struct {
	Address int64
}
type InstCallIP0R struct {
	Offset int64
	OpTy   int
}
type InstCallIP1R struct {
	Reg    lx.RegisterData
	Offset int64
	OpTy   int
}
type InstCallIPData struct {
	Reg    lx.RegisterData
	Offset int64
	OpTy   int
}
type InstMovSB struct{ Rep bool }
type InstMovSQ struct{ Rep bool }
type InstMovSH struct{ Rep bool }
type InstMovSW struct{ Rep bool }
type InstRet struct{}
type InstLab struct {
	LabelName string
}
type InstGenericPush struct {
	Expr   *Expr
	DataSz *DataSize
}
type InstPushR struct {
	Reg lx.RegisterData
}
type InstPushI struct {
	Imm    uint64
	DataSz byte
}
type InstPop struct {
	DataSz byte
}
type InstPopR struct {
	Reg    uint64
	DataSz byte
}
type InstNop struct{}

// move dereference of an arbitrary address into register
type InstMovDR struct {
	Dest    lx.RegisterData
	Address int64
}
type InstMovDRO1 struct {
	Dest   lx.RegisterData
	Offset int64
	Base   lx.RegisterData
	// Stored as Token type enum value ie TOKEN_TPLUS and so on,
	// use lexer.TokenTToOpT to get the vm compatible OP type
	DispOp int
	SF     Scale
}
type InstMovDRO2 struct {
	Dest  lx.RegisterData
	Disp  int64
	Base  lx.RegisterData
	Index lx.RegisterData
	// Stored as Token type enum value ie TOKEN_TPLUS and so on,
	// use lexer.TokenTToOpT to get the vm compatible OP type
	RegOp, DispOp int
	SF            Scale
}
type InstMovID struct {
	DataSize byte
	Imm      uint64
	Address  int64
}
type InstMovRD struct {
	Src     lx.RegisterData
	Address int64
}
type InstMovIDO1 struct {
	DataSize byte
	Imm      uint64
	Disp     int64
	Base     lx.RegisterData
	// Stored as Token type enum value ie TOKEN_TPLUS and so on,
	// use lexer.TokenTToOpT to get the vm compatible OP type
	DispOp, RegOp int
	NoOff         bool
	SF Scale
}
type InstMovRDO1 struct {
	Src  lx.RegisterData
	Disp int64
	Base lx.RegisterData
	// Stored as Token type enum value ie TOKEN_TPLUS and so on,
	// use lexer.TokenTToOpT to get the vm compatible OP type
	DispOp int
	SF     Scale
}
type InstMovIDO2 struct {
	DataSize byte
	Imm      uint64
	Disp     int64
	Base     lx.RegisterData
	Index    lx.RegisterData
	// Stored as Token type enum value ie TOKEN_TPLUS and so on,
	// use lexer.TokenTToOpT to get the vm compatible OP type
	RegOp, DispOp int
	NoOff         bool
	SF Scale
}
type InstMovRDO2 struct {
	Src   lx.RegisterData
	Disp  int64
	Base  lx.RegisterData
	Index lx.RegisterData
	// Stored as Token type enum value ie TOKEN_TPLUS and so on,
	// use lexer.TokenTToOpT to get the vm compatible OP type
	RegOp, DispOp int
	SF            Scale
}
type InstGenericCall struct {
	Expr *Expr
}
type InstExport struct {
	Name string
	Weak bool
}
type InstImport struct {
	Name string
	Weak bool
}
type InstGenericExit struct {
	Expr *Expr
}
type InstExitI struct {
	Val uint64
}
type InstExitR struct {
	Reg lx.RegisterData
}
type InstEntry struct{}
type InstClr struct{}
type InstSDF struct{}
type InstCDF struct{}
type InstSecCode struct{}
type InstGenericLea struct {
	Mov InstGenericMov
}
type InstLeaO1 struct {
	Mov InstMovDRO1
}
type InstLeaO2 struct {
	Mov InstMovDRO2
}
