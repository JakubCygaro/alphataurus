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
type MovSize struct {
	Line, Col int
	Size      byte
}
type InstGenericMov struct {
	Dest, Src *Expr
	DataSize  *MovSize
}
type InstMovIR struct {
	Dest     lx.RegisterData
	Imm      uint64
	DataSize byte
}
type InstMovRR struct {
	Src, Dest lx.RegisterData
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
type InstNot struct {
	First lx.RegisterData
}
type InstLogical struct {
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
type InstCmp struct {
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
type InstJmp struct {
	Address  *Expr
	Absolute bool
	Variant  JmpVariant
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
type InstRet struct{}
type InstLab struct {
	Label      string
	DeclaredAt string
}
type InstPush struct {
	Val    *Expr
	DataSz byte
}
type InstPushR struct {
	Reg    uint64
	DataSz byte
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
	OReg1  lx.RegisterData
	OffOp  int
}
type InstMovDRO2 struct {
	Dest         lx.RegisterData
	Offset       int64
	OReg1        lx.RegisterData
	OReg2        lx.RegisterData
	OffOp, RegOp int
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
	Offset   int64
	OReg1    lx.RegisterData
	OffsetOp int
	NoOff    bool
}
type InstMovRDO1 struct {
	Src      lx.RegisterData
	Offset   int64
	OReg1    lx.RegisterData
	OffsetOp int
}
type InstMovIDO2 struct {
	DataSize        byte
	Imm             uint64
	Offset          int64
	OReg1           lx.RegisterData
	OReg2           lx.RegisterData
	OffsetOp, RegOp int
	NoOff           bool
}
type InstMovRDO2 struct {
	Src             lx.RegisterData
	Offset          int64
	OReg1           lx.RegisterData
	OReg2           lx.RegisterData
	RegOp, OffsetOp int
}
type InstCall struct {
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
type InstExit struct {
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
type InstSecCode struct{}
