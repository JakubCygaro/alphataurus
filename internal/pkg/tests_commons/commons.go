package tests_commons

import (
	"cmp"
	"fmt"
	"math/rand"

	"github.com/JakubCygaro/alphataurus/pkg/vm"
)

func LabelGenerator(name string) func() string {
	start := 0
	return func() string {
		ret := start + 1
		return fmt.Sprintf("_genlab_%s_%d", name, ret)
	}
}
func RandomRegisterWord() byte {
	return byte(rand.Int()%vm.MAX_REG_IDX + 1)
}
func RandomGpRegisterWord() byte {
	return byte(rand.Int()%vm.GP_REG_MAX + 1)
}
func RandomGpRegisterWithSize() (byte, byte) {
	return RandomGpRegisterWord(), byte(rand.Int()%vm.SZ_64 + 1)
}
func RegStr(reg, sz byte) string {
	switch reg {
	case vm.BP_IDX:
		return "bp"
	case vm.IP_IDX:
		return "ip"
	case vm.SP_IDX:
		return "sp"
	default:
		var suf string = ""
		switch sz {
		case vm.SZ_8:
			suf = "b"
		case vm.SZ_16:
			suf = "q"
		case vm.SZ_32:
			suf = "h"
		}
		return fmt.Sprintf("r%v%s", reg, suf)
	}
}
func NextRandomGpRegister(reg byte) byte {
	return byte((reg+1)%vm.GP_REG_MAX + 1)
}
func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	} else {
		return b
	}
}
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	} else {
		return b
	}
}
