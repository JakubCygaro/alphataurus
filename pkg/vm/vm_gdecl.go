package vm

const (
	ADDRESSDEADZONE_SIZE = 0x1000
)
const (
	R0_IDX = 0
	R1_IDX = iota
	R2_IDX
	R3_IDX
	R4_IDX
	R5_IDX
	R6_IDX
	R7_IDX
	BP_IDX
	SP_IDX
	IP_IDX
)
const (
	GP_REG_MAX = R7_IDX
	MAX_REG_IDX = IP_IDX
	MAX_SZ = SZ_64
)

const (
	TY_SINT  = iota
	TY_UINT  = iota
	TY_FLOAT = iota
)
const (
	SZ_8  = iota
	SZ_16 = iota
	SZ_32
	SZ_64
)
