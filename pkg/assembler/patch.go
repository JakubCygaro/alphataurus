package assembler

import (
	"encoding/binary"

	"github.com/JakubCygaro/alphataurus/pkg/vm"
	decls "github.com/JakubCygaro/alphataurus/pkg/vm/decls"
	aobj "github.com/JakubCygaro/alphataurus/pkg/vm/obj"
)

func (a *Assembler) patchCall(pos int, address uint64) error {
	opcode := a.opCodes.GetBytes(vm.OP_CALL)
	binary.BigEndian.PutUint32(a.bytecode[pos:], uint32(opcode))
	binary.BigEndian.PutUint64(a.bytecode[pos+decls.OPCODE_SIZE:], uint64(address))
	return nil
}
func (a *Assembler) patchJmp(data PatchJmp, pos int, address uint64) error {
	opcode := a.jmpInstToOpCode(data.Variant)
	binary.BigEndian.PutUint32(a.bytecode[pos:], uint32(opcode))
	binary.BigEndian.PutUint64(a.bytecode[pos+decls.OPCODE_SIZE:], address)
	return nil
}
func (a *Assembler) patchCallIP(sym *aobj.SymbolData, pos int) error {
	opcode := a.opCodes.GetBytes(vm.OP_CALLIP)
	var reg byte
	reg = vm.OP_TADD
	reg <<= 4
	reg |= 0x0f
	posAsInstAddr := uint64(pos + vm.ADDRESSDEADZONE_SIZE)
	diff := int64(sym.Loc) - int64(posAsInstAddr)

	binary.BigEndian.PutUint32(a.bytecode[pos:], uint32(opcode))
	a.bytecode[pos] = reg
	binary.BigEndian.PutUint64(a.bytecode[pos+decls.OPCODE_SIZE:], uint64(diff))
	return nil
}
func (a *Assembler) patchJmpIP(data PatchJmp, sym *aobj.SymbolData, pos int) error {
	opcode := a.absoluteJmpToIPJmp(data.Variant)
	var reg byte
	reg = vm.OP_TADD
	reg <<= 4
	reg |= 0x0f
	posAsInstAddr := uint64(pos + vm.ADDRESSDEADZONE_SIZE)
	diff := int64(sym.Loc) - int64(posAsInstAddr)

	binary.BigEndian.PutUint32(a.bytecode[pos:], uint32(opcode))
	a.bytecode[pos] = reg
	binary.BigEndian.PutUint64(a.bytecode[pos+decls.OPCODE_SIZE:], uint64(diff))
	return nil
}
