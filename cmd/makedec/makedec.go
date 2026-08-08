package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/JakubCygaro/alphataurus/internal/pkg/opt"
	toml "github.com/pelletier/go-toml/v2"

	arg "github.com/alexflint/go-arg"
)

var args struct {
	InputFile   string `arg:"positional,required"`
	// PackageName string `arg:"-p,--package,required"`
	// OutputFile  string `arg:"-o,--output, required"`
}

func exitWithErr(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

type Spec struct {
	// Instruction description
	Desc string
	// What bytes from the opcode portion of the instruction
	// are reserved as parameters to the instruction
	Reserve struct {
		// Byte count
		Size int
		// Bit documentation, the system is big endian so it goes from
		// right to left
		Doc []struct {
			Bits int
			Desc string
		}
	}
	// What bits of the argument portion of the instruction mean
	Param []struct {
		// Bit documentation, the system is big endian so it goes from
		// right to left
		Bits int
		Desc string
	}
}

type OpSpec struct {
	// Opcodes defined
	Op map[string]Spec
}

func main() {
	if err := arg.Parse(&args); err != nil {
		exitWithErr("%s\n", err.Error())
	}
	var read []byte
	if f, err := os.ReadFile(args.InputFile); err != nil {
		exitWithErr("%s\n", err.Error())
	} else {
		read = f
	}
	var spec OpSpec
	if err := toml.Unmarshal(read, &spec); err != nil {
		exitWithErr("%s\n", err.Error())
	}
	built, err := buildSource(spec)
	if err != nil {
		exitWithErr("%s\n", err.Error())
	}
	println(built)
}

func verifyOpSpec(opcode string, spec *Spec) error {
	if strings.ContainsFunc(opcode, func(r rune) bool {
		return unicode.IsSpace(r) ||
			unicode.IsControl(r) ||
			unicode.IsPunct(r)
	}) {
		return fmt.Errorf(
			"Spec for opcode `%s` contains disallowed characters",
			opcode,
		)
	}
	if spec.Reserve.Size < 0 || spec.Reserve.Size > 3 {
		return fmt.Errorf(
			"Spec for opcode `%s` reserves a disallowed amount of bytes (%d)",
			opcode,
			spec.Reserve.Size,
		)
	}
	reserved := spec.Reserve.Size * 8
	for i, bitDoc := range spec.Reserve.Doc {
		reserved -= bitDoc.Bits
		if reserved < 0 {
			return fmt.Errorf(
				"Bit doc %d `%s` for opcode `%s` uses "+
					"more bits than are reserved in the spec (%d). "+
					"%d bits over the reserved %d",
				i,
				bitDoc.Desc,
				opcode,
				spec.Reserve.Size,
				-reserved,
				spec.Reserve.Size*8,
			)
		}
	}
	if reserved > 0 {
		return fmt.Errorf(
			"Spec for opcode `%s` does not document all reserved bits."+
				" %d undocumented bits remain",
			opcode,
			reserved,
		)
	}
	paramReserved := 64
	for i, paramDoc := range spec.Param {
		paramReserved -= paramDoc.Bits
		if paramReserved < 0 {
			return fmt.Errorf(
				"Param doc %d `%s` for opcode `%s` uses "+
					"more bits than are allowed for an instruction parameter (64). "+
					"%d bits over the maximum",
				i,
				paramDoc.Desc,
				opcode,
				-paramReserved,
			)
		}
	}
	if paramReserved > 0 {
		return fmt.Errorf(
			"Spec for opcode `%s` does not document all parameter bits."+
				" %d undocumented bits remain",
			opcode,
			paramReserved,
		)
	}
	return nil
}

type OpcodeVal struct {
	Name string
	Spec Spec
}
type LayerEntry opt.Either[*OpcodeMap, OpcodeVal]

type OpcodeMap struct {
	Prefix       uint8
	Prev         *OpcodeMap
	Layer        [256]opt.Either[*OpcodeMap, OpcodeVal]
	next         int
	depth        int
	LastNewLayer *OpcodeMap
}

func (om *OpcodeMap) HasSpace() bool {
	return om.next < 256
}

func (om *OpcodeMap) PushLayer(layer *OpcodeMap) error {
	if om.depth < 0 {
		return fmt.
			Errorf("Maximum map depth reached")
	}
	if om.next >= 256 {
		return fmt.
			Errorf("Maximum layer size reached")
	}
	om.Layer[om.next] =
		opt.MakeLeft[*OpcodeMap, OpcodeVal](layer)
	om.next++
	om.LastNewLayer = layer
	layer.depth = om.depth - 1
	layer.Prefix = uint8(om.next - 1)
	layer.Prev = om
	return nil
}
func (om *OpcodeMap) PushSpec(name string, s Spec) error {
	if om.next >= 256 {
		return fmt.
			Errorf("Maximum layer size reached")
	}
	om.Layer[om.next] =
		opt.MakeRight[*OpcodeMap](OpcodeVal{name, s})
	om.next++
	return nil
}

var m = OpcodeMap{
	depth: 3,
}

type LayerList []*OpcodeMap

var layers = []LayerList{
	make(LayerList, 0),
	make(LayerList, 0),
	make(LayerList, 0),
	make(LayerList, 0),
}

func insertIntoMap(m *OpcodeMap, name string, spec Spec) error {
	if m.depth > spec.Reserve.Size {
		if m.LastNewLayer != nil &&
			m.LastNewLayer.HasSpace() {
			return insertIntoMap(m.LastNewLayer, name, spec)
		}
		// backtrack case
		if !m.HasSpace() {
			m.Prev.LastNewLayer = nil
			return insertIntoMap(m.Prev, name, spec)
		}
		lay := &OpcodeMap{}
		if err := m.PushLayer(lay); err != nil {
			return err
		} else {
			layers[lay.depth] = append(layers[lay.depth], lay)
			return insertIntoMap(m.LastNewLayer, name, spec)
		}
	} else if m.depth == spec.Reserve.Size {
		return m.PushSpec(name, spec)
	}
	return nil
}

type Assigned struct {
	Opcode uint32
	Spec   Spec
}

var assigned = map[uint32]OpcodeVal{}

func recurseUp(layer *OpcodeMap, bytes []byte) {
	if layer.Prev != nil {
		bytes[layer.depth] = layer.Prefix
		recurseUp(layer.Prev, bytes)
	}
}

func assignLayer(ll LayerList) {
	for _, layer := range ll {
		for i := range layer.next {
			if layer.Layer[i].HasLeft() {
				continue
			}
			val := layer.Layer[i].GetRight()
			b := [4]byte{}
			b[layer.depth] = byte(i)
			recurseUp(layer, b[:])
			assigned[binary.BigEndian.Uint32(b[:])] = val
		}
	}
}

func buildSource(spec OpSpec) (string, error) {
	builder := strings.Builder{}
	layers[3] = append(layers[3], &m)
	for opcode, spec := range spec.Op {
		if err := verifyOpSpec(opcode, &spec); err != nil {
			return "", err
		}
		if err := insertIntoMap(&m, opcode, spec); err != nil {
			return "", nil
		}
	}
	for i, ll := range layers {
		println(i, ll)
		assignLayer(ll)
	}
	builder.WriteString(
		"const (\n",
	)
	for code, spec := range assigned {
		builder.WriteString(
			fmt.Sprintf(
				"\t OP_%s = 0x%08x\n",
				spec.Name,
				code,
			),
		)
	}
	builder.WriteString(
		")\n",
	)
	return builder.String(), nil

}
