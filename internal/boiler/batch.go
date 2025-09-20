package boiler

func batchMapKeys[K comparable, V any](m map[K]V, size int) [][]K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return batchSlice(keys, size)
}

func batchSlice[T any](slice []T, size int) [][]T {
	out := make([][]T, 0)

	if size <= 1 {
		panic("slice batch size must be greater than 1")
	}

	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		out = append(out, slice[i:end])
	}

	return out
}
