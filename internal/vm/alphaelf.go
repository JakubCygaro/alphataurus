package vm

const (
	AELF_FILE_MAG         = "AELF"
	AELF_FILE_HEADER_SIZE = 128
)

type AlphaELFFile struct {
	Version                         uint32
	CodeStart, CodeSize             uint64
	StaticDataStart, StaticDataSize uint64
	Data                            []byte
}
