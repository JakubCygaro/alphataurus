package vm

const (
	AEXE_FILE_MAG         = "AELF"
)

type AlphaEXEFile struct {
	Version                         uint32
	HeaderSize                      uint16
	HasEntry                        bool
	Entry                           uint64
	CodeStart, CodeSize             uint64
	StaticDataStart, StaticDataSize uint64
	SymbolsStart, SymbolsSize       uint64
	RelocsStart, RelocsSize         uint64
	Data                            []byte
}
