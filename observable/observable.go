package observable

type Observable[T any] interface {
	Current() T
	Changes() <-chan T
}

type Mutable[T any] interface {
	Current() T
	Changes() <-chan T
	Mutate(state T)
}
