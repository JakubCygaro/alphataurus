package assembler

import (
	"bufio"

	pr "github.com/JakubCygaro/alphataurus/pkg/assembler/parser"
	"github.com/JakubCygaro/alphataurus/pkg/vm"
	aobj "github.com/JakubCygaro/alphataurus/pkg/vm/obj"
)

type void struct{}

var Void void = void{}

type PatchCall void
type PatchJmp struct {
	Variant pr.JmpVariant
}

type unresolvedJump struct {
	Ident string
	// what instruction is gonna get patched, type of Patch... struct
	PatchTy  any
	Absolute bool
	Expr     *pr.Expr
}
type unresolvedJumpMap map[int]unresolvedJump
type unevaledInstMap map[int]pr.Instruction

type AssemblerWarningData struct {
	Line, Col int
	Message   string
}

type Assembler struct {
	parser          *pr.Parser
	opCodes         vm.OpCodeMap
	unresolvedJumps unresolvedJumpMap
	// instructions that depend on expressions that have yet to be evaluated
	unevalInsts unevaledInstMap
	//relating to the current instruction
	line, col int
	// labels          labelMap
	lastInst pr.Instruction
	bytecode []byte
	// position in the emited bytecode
	pos         int
	instCount   int
	symbols     aobj.SymbolTable
	relocations aobj.RelocationTable
	hasEntry    bool
	entry       uint64
	// pointer to a function that recieves warnings emitted by the assembler
	WarningSink func(AssemblerWarningData)
	ev          ExpressionEvaluator
	prov        *assemblerVarProvider
}

type assemblerVarProvider struct {
	syms             *aobj.SymbolTable
	LastFailedAccess *string
}

func (avp *assemblerVarProvider) Get(name string) any {
	if sd, _, ok := avp.syms.GetByName(name); ok {
		return sd.Loc
	}
	avp.LastFailedAccess = &name
	return nil
}

func (a *Assembler) InstructionCount() int {
	return a.instCount
}

func aInitialState() Assembler {
	a := Assembler{
		opCodes:         vm.GenerateOpcodeMap(),
		unresolvedJumps: make(unresolvedJumpMap),
		unevalInsts:     make(unevaledInstMap),
		bytecode:        make([]byte, 0, 64),
		symbols:         aobj.NewSymbolTable(),
		relocations:     make(aobj.RelocationTable, 0),
		hasEntry:        false,
		prov:            &assemblerVarProvider{},
	}
	return a
}
func (a *Assembler) clearState() {
	clear(a.unresolvedJumps)
	clear(a.bytecode)
	clear(a.bytecode)
	clear(a.unevalInsts)
	a.symbols.Clear()
	clear(a.relocations)
	a.hasEntry = false
	a.instCount = 0
}
func NewAssembler(reader *bufio.Reader) *Assembler {
	a := aInitialState()
	a.parser = pr.NewParser(reader)
	a.parser.WarningSink = a.parserWarningHandler
	a.prov.syms = &a.symbols
	a.ev = ExpressionEvaluator{
		Ctx: EvaluationContext{
			VarProvider: a.prov,
		},
	}
	return &a
}

// reset the state of the assembler and load new reader input
func (a *Assembler) LoadNew(reader *bufio.Reader) {
	p := a.parser
	p.LoadNew(reader)
	ws := a.WarningSink
	a.clearState()
	a.parser = p
	a.WarningSink = ws
}

func (a *Assembler) codePos(at int) (byte uint64, address uint64) {
	position := uint64(at)
	posAsInstAddr := uint64(position + vm.ADDRESSDEADZONE_SIZE)
	return position, posAsInstAddr
}
func (a *Assembler) currentCodePos() (byte uint64, address uint64) {
	return a.codePos(a.pos)
}
