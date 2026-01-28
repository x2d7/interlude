package types

type Stream[T any] interface {
    Next() bool   // advance; returns false on EOF or error
    Current() T   // the current element; valid only if Last Next() returned true
    Err() error   // non-nil if the stream ended because of an error
    Close() error // release resources
}
