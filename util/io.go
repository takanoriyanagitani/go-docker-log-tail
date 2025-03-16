package util

import (
	"context"
)

type IO[T any] func(context.Context) (T, error)

func (i IO[T]) Or(alt T) IO[T] {
	return func(ctx context.Context) (T, error) {
		t, e := i(ctx)
		switch e {
		case nil:
			return t, nil
		default:
			return alt, nil
		}
	}
}

func Of[T any](t T) IO[T] {
	return func(_ context.Context) (T, error) { return t, nil }
}

func Lift[T, U any](pure func(T) (U, error)) func(T) IO[U] {
	return func(t T) IO[U] {
		return func(_ context.Context) (U, error) {
			return pure(t)
		}
	}
}

func Bind[T, U any](
	i IO[T],
	f func(T) IO[U],
) IO[U] {
	return func(ctx context.Context) (u U, e error) {
		t, e := i(ctx)
		switch e {
		case nil:
			return f(t)(ctx)
		default:
			return u, e
		}
	}
}

func All[T any](ios ...IO[T]) IO[[]T] {
	return func(ctx context.Context) ([]T, error) {
		var ret []T = make([]T, 0, len(ios))
		for _, io := range ios {
			t, e := io(ctx)
			if nil != e {
				return nil, e
			}
			ret = append(ret, t)
		}
		return ret, nil
	}
}

func AllMap[T any](ios map[string]IO[T]) IO[map[string]T] {
	return func(ctx context.Context) (map[string]T, error) {
		ret := map[string]T{}
		for key, val := range ios {
			t, e := val(ctx)
			if nil != e {
				return ret, e
			}

			ret[key] = t
		}
		return ret, nil
	}
}

type Void struct{}

var Empty Void = Void{}
