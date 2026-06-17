package to

func Val[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func Ptr[T any](v T) *T {
	return &v
}
