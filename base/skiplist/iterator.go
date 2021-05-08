package skip_list

/// NOTICE
/// like c++ stl
/// after mod skiplist, an available iterator will trans to unavailable.

type Iterator interface {
	Next() Iterator /// return's key is bigger than caller's,or nil if not exist
	Key() interface{}
	Value() interface{}
}

type iterator struct {
	sl SkipList
	n  *node
}

func (iter *iterator) Next() Iterator {
	if iter.n.levels[0].next == nil {
		return nil
	}
	iter.n = iter.n.levels[0].next
	return iter
}

func (iter *iterator) Key() interface{} {
	if iter.n == nil {
		return nil
	}
	return iter.n.k
}

func (iter *iterator) Value() interface{} {
	if iter.n == nil {
		return nil
	}
	return iter.n.v
}
