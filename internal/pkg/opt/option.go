package opt

type Opt[T any] struct {
	val   T
	isSet bool
}

func MakeSome[T any](v T) Opt[T] {
	return Opt[T]{
		val:   v,
		isSet: true,
	}
}
func MakeNone[T any]() Opt[T] {
	return Opt[T]{
		isSet: false,
	}
}
func (o *Opt[T]) HasVal() bool {
	return o.isSet
}
func (o *Opt[T]) Get() T {
	if !o.isSet {
		panic("Called GetVal on None value")
	}
	return o.val
}
func (o *Opt[T]) TryGet() (T, bool) {
	return o.val, o.isSet
}
func (o *Opt[T]) Take() T {
	if !o.isSet {
		panic("Called GetVal on None value")
	}
	var d T
	d, o.val, o.isSet = o.val, d, false
	return d
}
func (o *Opt[T]) Set(v T) {
	o.val = v
	o.isSet = true
}
