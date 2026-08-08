package opt

type Either[Left any, Right any] struct {
	l       Left
	r       Right
	hasLeft bool
}

func MakeLeft[Left any, Right any](l Left) Either[Left, Right] {
	return Either[Left, Right]{
		l:       l,
		hasLeft: true,
	}
}
func MakeRight[Left any, Right any](r Right) Either[Left, Right] {
	return Either[Left, Right]{
		r:       r,
		hasLeft: false,
	}
}
func (eth *Either[Left, Right]) HasLeft() bool {
	return eth.hasLeft
}
func (eth *Either[Left, Right]) HasRight() bool {
	return !eth.hasLeft
}
func (eth *Either[Left, Right]) GetLeft() Left {
	if !eth.hasLeft {
		panic("GetLeft called on an Either with Right vairant")
	}
	return eth.l
}
func (eth *Either[Left, Right]) GetRight() Right {
	if eth.hasLeft {
		panic("GetRight called on an Either with Left vairant")
	}
	return eth.r
}
