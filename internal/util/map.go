package shared

// Pair that contains two generic elements
type Pair[T, R any] struct {
	First  T
	Second R
}

// Returns a slice key, value pair
func Entries[K comparable, V any](hashMap map[K]V) []*Pair[K, V] {
	var entries []*Pair[K, V]
	for key, value := range hashMap {
		entries = append(entries, &Pair[K, V]{key, value})
	}
	return entries
}

// Returns the slice of all keys
func Keys[K comparable, V any](hashMap map[K]V) []K {
	var keys []K
	for key := range hashMap {
		keys = append(keys, key)
	}
	return keys
}

// Returns the slice of all values
func Values[K comparable, V any](hashMap map[K]V) []V {
	var values []V
	for _, value := range hashMap {
		values = append(values, value)
	}
	return values
}

// Returns whether the map contains the key
func ContainsKey[K comparable, V any](hashMap map[K]V, target K) bool {
	_, found := hashMap[target]
	return found
}

// Get the value corresponding to the key, returns default value if the key is not there.
func GetOrDefault[K comparable, V any](hashMap map[K]V, key K, defaultVal V) V {
	if value, found := hashMap[key]; found {
		return value
	}
	return defaultVal
}

// Filter keys based on the given predicate
func FilterKeys[K comparable, V any](hashMap map[K]V, predicate func(K) bool) map[K]V {
	filtered := make(map[K]V)
	ForEach(Filter(Keys(hashMap), predicate), func(key K) {
		filtered[key] = hashMap[key]
	})
	return filtered
}

// Maps the map into a slice based on the given transform
func FlatMap[K comparable, V, R any](hashMap map[K]V, transform func(K, V) R) []R {
	var slice []R
	if hashMap == nil || transform == nil {
		return slice
	}
	for key, value := range hashMap {
		slice = append(slice, transform(key, value))
	}
	return slice
}

// Runs operation on each entry of the map
func ForEachEntry[K comparable, V any](hashMap map[K]V, operation func(K, V)) {
	if hashMap == nil || operation == nil {
		return
	}
	for key, value := range hashMap {
		operation(key, value)
	}
}
