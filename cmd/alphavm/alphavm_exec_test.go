package alphavm

import (
	// "encoding/binary"
	"fmt"
	"math/rand"

	// "regexp"
	// "strings"
	"testing"

	"github.com/JakubCygaro/alphataurus/pkg/vm"
	// "github.com/JakubCygaro/alphataurus/pkg/vm"
)

func TestExitI1(t *testing.T) {
	exitV := uint64(rand.Int())
	lines := fmt.Sprintf(`
		section '.code'
		@entry
			exit %v
	`, exitV)
	if s, err := assembleAndExecute(lines); err != nil {
		t.Error(compilationOfErr(lines))
		t.Error(err)
	} else if s.GetExitCode() != exitV {
		t.Error(compilationOfErr(lines))
		t.Errorf("Expected exit code %v, got %v", exitV, s.GetExitCode())
	}
}
func TestExitR1(t *testing.T) {
	exitV := uint64(rand.Int())
	r, rsz := randomGpRegisterWithSize()
	lines := fmt.Sprintf(`
		section '.code'
		@entry
			mov %s, %v
			exit %s
	`, regStr(r, rsz), exitV, 
		regStr(r, rsz))
	switch rsz {
	case vm.SZ_8:
		exitV = uint64(byte(exitV))
	case vm.SZ_16:
		exitV = uint64(uint16(exitV))
	case vm.SZ_32:
		exitV = uint64(uint32(exitV))
	}
	if s, err := assembleAndExecute(lines); err != nil {
		t.Error(compilationOfErr(lines))
		t.Error(err)
	} else if s.GetExitCode() != exitV {
		t.Error(compilationOfErr(lines))
		t.Errorf("Expected exit code %v, got %v", exitV, s.GetExitCode())
	}
}
