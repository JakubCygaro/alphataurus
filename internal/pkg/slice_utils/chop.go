package sliceutils

// Reslices a slice so that:
//
// chopped = slice[:count]
//
// remaining = slice[count:]
//
// chopped will be set to nil if the lenght of the input slice is less than count
func Chop(slice []byte, count int) (chopped, remaining []byte) {
	if len(slice) < count {
		return nil, slice
	}
	return slice[:count], slice[count:]
}
