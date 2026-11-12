package sdata

// Comparer compares two rows on a tag.
type Comparer[T any] interface {
	Compare(o T, tag Tag) int
}

// QuickSort sorts items[lo..hi] on key1 then key2 with shStudio's
// quicksort (Collection.hpp), so that equal rows end in the same order as
// in the original. It returns the number of swaps.
func QuickSort[T Comparer[T]](items []T, lo, hi int, key1, key2 Tag, ascending bool) int {
	changes := 0
	compare := func(a, b T) int {
		c := a.Compare(b, key1)
		if c == 0 && key2 != "" {
			c = a.Compare(b, key2)
		}
		if !ascending {
			c = -c
		}
		return c
	}
	j := hi
	if hi <= lo {
		return 0
	}
	for {
		i := lo
		pivot := items[(lo+j)/2]
		for {
			for i < hi && compare(items[i], pivot) < 0 {
				i++
			}
			for ; j > 0; j-- {
				if compare(items[j], pivot) <= 0 {
					break
				}
			}
			if j < i {
				break
			}
			if j != i {
				items[i], items[j] = items[j], items[i]
				changes++
			}
			j--
			i++
			if i > j {
				break
			}
		}
		if lo < j {
			changes += QuickSort(items, lo, j, key1, key2, ascending)
		}
		j = hi
		if i >= hi {
			break
		}
		lo = i
	}
	return changes
}

// SortList sorts a whole list (sortList) and reports whether it changed.
func SortList[T Comparer[T]](items []T, key1, key2 Tag, ascending bool) bool {
	if len(items) < 2 {
		return false
	}
	return QuickSort(items, 0, len(items)-1, key1, key2, ascending) > 0
}
