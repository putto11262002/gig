package util

// CollectionIter[T] defines an iterator over a collection of type T.
type CollectionIter[T any] struct {
	collection []T
	c          int // Index of the current element in the collection
}

func NewCollectionIter[T any](collection []T) *CollectionIter[T] {
	return &CollectionIter[T]{
		collection: collection,
		c:          0,
	}
}

// Next returns the next item in the iteration and a boolean indicating whether there are more items.
//
// If there are no more items in the collection, it returns the zero value of T and false.
// Otherwise, it returns the next item in the collection and true.
func (iter *CollectionIter[T]) Next() (item T, ok bool) {
	if iter.c >= len(iter.collection) {
		return item, false
	}
	p := iter.collection[iter.c]
	iter.c++
	return p, true
}
