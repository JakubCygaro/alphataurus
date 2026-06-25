package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/vm"
	lx "github.com/JakubCygaro/alphataurus/pkg/assembler/lexer"
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
type InstMovIR struct {
	Dest     int
	Imm      uint64
	DataSize byte
}
type InstMovRR struct {
	Src, Dest int
	DataSize  byte
}
type InstInc struct {
	Reg lx.RegisterData
}
type InstDec struct {
	Reg lx.RegisterData
}
type InstArthRR struct {
	Src, Dest lx.RegisterData
	Ty        int
	DataSize  byte
	ArthTy    ArthTy
}
type InstArthIR struct {
	Dest lx.RegisterData
	Imm       uint64
	Ty        int
	DataSize  byte
	ArthTy    ArthTy
}
type InstNot struct {
	First lx.RegisterData
}
type InstLogicalRR struct {
	First, Second lx.RegisterData
	LogTy         LogTy
}
type InstLogicalIR struct {
	First lx.RegisterData
	Imm           uint64
	LogTy         LogTy
}
type InstCmpRR struct {
	Ty       int
	Sub, Min lx.RegisterData
}
type InstCmpIR struct {
	Ty       int
	Min lx.RegisterData
	Imm      uint64
}
type InstJmp struct {
	Address  any
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
type InstPushR struct {
	Reg    uint64
	Imm    uint64
	DataSz byte
}
type InstPushI struct {
	Reg    uint64
	Imm    uint64
	DataSz byte
}
type InstPop struct {
	Reg    uint64
	Imm    uint64
	DataSz byte
}
type InstNop struct{}
type InstMovDRI struct {
	Dest   lx.RegisterData
	Offset int64
	OpTy   int
}
type InstMovDRO1 struct {
	Dest   lx.RegisterData
	Offset int64
	OReg1  lx.RegisterData
	OpTy   int
}
type InstMovDRO2 struct {
	Dest   lx.RegisterData
	Offset int64
	OReg1  lx.RegisterData
	OReg2  lx.RegisterData
	OpTy   int
}
type InstMovID struct {
	DataSize byte
	Imm      uint64
	Offset   int64
	OpTy     int
}
type InstMovRD struct {
	Src      lx.RegisterData
	Offset   int64
	OpTy     int
}
type InstMovIDO1 struct {
	DataSize byte
	Imm      uint64
	Offset   int64
	OReg1    lx.RegisterData
	OpTy     int
	NoOff    bool
}
type InstMovRDO1 struct {
	Src      lx.RegisterData
	Offset   int64
	OReg1    lx.RegisterData
	OpTy     int
}
type InstMovIDO2 struct {
	DataSize byte
	Imm      uint64
	Offset   int64
	OReg1    lx.RegisterData
	OReg2    lx.RegisterData
	OpTy     int
	NoOff    bool
}
type InstMovRDO2 struct {
	Src      lx.RegisterData
	Offset   int64
	OReg1    lx.RegisterData
	OReg2    lx.RegisterData
	Label    string
	OpTy     int
}
type InstCall struct {
	Addr  uint64
	Ident string
	Expr  *Expr
}
type InstExport struct {
	Name string
	Weak bool
}
type InstImport struct {
	Name string
	Weak bool
}
type InstExitI struct {
	Val uint64
	Reg lx.RegisterData
}
type InstExitR struct {
	Val uint64
	Reg lx.RegisterData
}
type InstEntry struct{}
type InstClr struct{}
type InstSecCode struct{}
