package utils

func Hold(a any) {

}

func Ptr[T any](a T) *T {
	return &a
}
