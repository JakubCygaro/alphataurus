package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

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

//go:generate stringer -type=InstTy
type InstTy int

const (
	INST_TMOVRR InstTy = iota
	INST_TMOVIR
	INST_TMOVDRI
	INST_TMOVDRO0
	INST_TMOVDRO1
	INST_TMOVDRO2
	INST_TMOVID
	INST_TMOVRD
	INST_TMOVRDO1
	INST_TMOVIDO1
	INST_TMOVIDO1_NO
	INST_TMOVRDO2
	INST_TMOVIDO2
	INST_TMOVIDO2_NO
	INST_TADDRR
	INST_TSUBRR
	INST_TDIVRR
	INST_TMULRR
	INST_TADDIR
	INST_TSUBIR
	INST_TINCR
	INST_TDECR
	INST_TNOT
	INST_TORIR
	INST_TORRR
	INST_TANDIR
	INST_TANDRR
	INST_TXORIR
	INST_TXORRR
	INST_TLSHRR
	INST_TLSHIR
	INST_TRSHRR
	INST_TRSHIR
	INST_TCMPRR
	INST_TCMPIR
	// INST_TJMP
	// INST_TJMPE
	// INST_TJMPZ
	// INST_TJMPNE
	// INST_TJMPNZ
	// INST_TJMPG
	// INST_TJMPGE
	// INST_TJMPL
	// INST_TJMPLE
	// INST_TJMPS
	// INST_TJMPNS
	// INST_TJMPC
	// INST_TJMPNC
	// INST_TJMPO
	// INST_TJMPNO
	// INST_TJMPIP0R
	// INST_TJMPIP1R
	INST_TCALLIP0R
	INST_TCALLIP1R
	INST_TLABEL
	INST_TPUSHR
	INST_TPUSHI
	INST_TPOP
	INST_TCLR
	INST_TNOP
	INST_TCALL
	INST_TRET
	INST_TSECCODE
	INST_TSECDATA
	INST_TIMPORT
	INST_TEXPORT
	INST_TATTRENTRY
	INST_TEXITI
	INST_TEXITR
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
	Reg RegisterData
}
type InstDec struct {
	Reg RegisterData
}
type InstArthRR struct {
	Src, Dest RegisterData
	Ty        int
	DataSize  byte
	ArthTy    InstTy
}
type InstArthIR struct {
	Dest RegisterData
	Imm       uint64
	Ty        int
	DataSize  byte
	ArthTy    InstTy
}
type InstNot struct {
	First RegisterData
}
type InstLogicalRR struct {
	First, Second RegisterData
	LogTy         InstTy
}
type InstLogicalIR struct {
	First RegisterData
	Imm           uint64
	LogTy         InstTy
}
type InstCmpRR struct {
	Ty       int
	Sub, Min RegisterData
}
type InstCmpIR struct {
	Ty       int
	Min RegisterData
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
	Reg    RegisterData
	Offset int64
	OpTy   int
	JmpTy  JmpVariant
}
type InstCallIP0R struct {
	Offset int64
	OpTy   int
}
type InstCallIP1R struct {
	Reg    RegisterData
	Offset int64
	OpTy   int
}
type InstCallIPData struct {
	Reg    RegisterData
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
	Dest   RegisterData
	Offset int64
	OpTy   int
}
type InstMovDRO1 struct {
	Dest   RegisterData
	Offset int64
	OReg1  RegisterData
	OpTy   int
}
type InstMovDRO2 struct {
	Dest   RegisterData
	Offset int64
	OReg1  RegisterData
	OReg2  RegisterData
	OpTy   int
}
type InstMovID struct {
	DataSize byte
	Imm      uint64
	Offset   int64
	OpTy     int
}
type InstMovRD struct {
	Src      RegisterData
	Offset   int64
	OpTy     int
}
type InstMovIDO1 struct {
	DataSize byte
	Imm      uint64
	Offset   int64
	OReg1    RegisterData
	OpTy     int
	NoOff    bool
}
type InstMovRDO1 struct {
	Src      RegisterData
	Offset   int64
	OReg1    RegisterData
	OpTy     int
}
type InstMovIDO2 struct {
	DataSize byte
	Imm      uint64
	Offset   int64
	OReg1    RegisterData
	OReg2    RegisterData
	OpTy     int
	NoOff    bool
}
type InstMovRDO2 struct {
	Src      RegisterData
	Offset   int64
	OReg1    RegisterData
	OReg2    RegisterData
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
	Reg RegisterData
}
type InstExitR struct {
	Val uint64
	Reg RegisterData
}
type InstEntry struct{}
type InstClr struct{}
type InstSecCode struct{}
