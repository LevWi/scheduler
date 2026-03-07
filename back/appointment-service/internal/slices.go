package common

func MapE[T any, R any](in []T, f func(T) (R, error)) ([]R, error) {
	out := make([]R, len(in))

	for i, v := range in {
		r, err := f(v)
		if err != nil {
			return nil, err
		}
		out[i] = r
	}

	return out, nil
}
