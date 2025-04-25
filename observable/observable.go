package observable

type Observable[T any] interface {
	Current() T
	Changes() <-chan T
}
