package assembler

import (
	"github.com/JakubCygaro/alphataurus/pkg/vm"
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
	INST_TJMP
	INST_TJMPE
	INST_TJMPZ
	INST_TJMPNE
	INST_TJMPNZ
	INST_TJMPG
	INST_TJMPGE
	INST_TJMPL
	INST_TJMPLE
	INST_TJMPS
	INST_TJMPNS
	INST_TJMPC
	INST_TJMPNC
	INST_TJMPO
	INST_TJMPNO
	INST_TJMPIP0R
	INST_TJMPIP1R
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
	Ty        InstTy
	Data      any
	Line, Col int
}
type InstMovData struct {
	Src, Dest int
	Imm       uint64
	DataSize  byte
}
type InstIncDecData struct {
	Reg RegisterData
}
type InstArthData struct {
	Src, Dest RegisterData
	Imm       uint64
	Ty        int
	DataSize  byte
}
type InstLogicalData struct {
	First, Second RegisterData
	Imm           uint64
}
type InstCmpData struct {
	Ty       int
	Sub, Min RegisterData
	Imm      uint64
}
type InstJmpData struct {
	Address  any
	Absolute bool
}
type InstJmpIPData struct {
	Reg    RegisterData
	Offset int64
	JmpTy  InstTy
	OpTy   int
}
type InstCallIPData struct {
	Reg    RegisterData
	Offset int64
	OpTy   int
}
type InstLabData struct {
	Label      string
	DeclaredAt string
}
type InstPushPopData struct {
	Reg    uint64
	Imm    uint64
	DataSz byte
}
type InstDerefMovData struct {
	Dest   RegisterData
	Offset int64
	OReg1  RegisterData
	OReg2  RegisterData
	Label  string
	OpTy   int
}
type InstMovDerefData struct {
	DataSize byte
	Src      RegisterData
	Imm      uint64
	Offset   int64
	OReg1    RegisterData
	OReg2    RegisterData
	Label    string
	OpTy     int
}
type InstCallData struct {
	Addr  uint64
	Ident string
	Expr  *Expr
}
type InstImportExportData struct {
	Name string
	Weak bool
}
type InstExitData struct {
	Val uint64
	Reg RegisterData
}
