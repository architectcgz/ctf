package queries

func nonNilSlice[T any](items []T) []T {
	if len(items) == 0 {
		return []T{}
	}
	return items
}
