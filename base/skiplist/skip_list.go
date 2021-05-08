package skip_list

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

//https://github.com/q191201771/skiplist

var (
	ErrKeyAlreadyExist = errors.New("key already exist")
	ErrKeyNotExist     = errors.New("key not exist")
)

var kDefaultMaxHeight = 32
var kChance = 0.5

type Compare func(l, r interface{}) int

type SkipList interface {
	Insert(k, v interface{}) error /// nil | `ErrKeyAlreadyExist`.
	Erase(k interface{}) error     /// nil | `ErrKeyNotExist`,most of scenario you can ignore it.
	Find(k interface{}) (v interface{}, exist bool)
	Clear()
	Size() int
	Empty() bool
	Min() (k, v interface{}) /// if Empty(),return nil,nil
	Max() (k, v interface{}) /// if Empty(),return nil,nil

	SkipListDebug
	SkipListIterator
}

type SkipListDebug interface {
	debugPrint()
}

type SkipListIterator interface {
	Begin() Iterator /// minimum,or nil if Empty()
	End() Iterator   /// same as nil
}

type skiplist struct {
	maxHeight  int
	curHeight  int
	numOfNodes int
	head       *node
	tail       *node
	cache      []*node
	compare    Compare
}

type level struct {
	prev *node
	next *node
}

type node struct {
	levels []*level
	height int
	k, v   interface{}
}

func New(c Compare, maxHeight int) SkipList {
	return &skiplist{
		maxHeight:  maxHeight,
		curHeight:  1,
		numOfNodes: 0,
		head:       createNode(maxHeight, nil, nil),
		cache:      make([]*node, maxHeight),
		compare:    c,
	}
}

func (sl *skiplist) Insert(k, v interface{}) error {
	node := sl.head
	for i := sl.curHeight - 1; i >= 0; i-- {
		for ; node.levels[i].next != nil; node = node.levels[i].next {
			cmpRes := sl.compare(k, node.levels[i].next.k)
			if cmpRes < 0 {
				break
			} else if cmpRes == 0 {
				return ErrKeyAlreadyExist
			}
		}
		sl.cache[i] = node
	}

	height := sl.randomHeight()
	if height > sl.curHeight {
		for i := height - 1; i >= sl.curHeight; i-- {
			sl.cache[i] = sl.head
		}
		sl.curHeight = height
	}

	newNode := createNode(height, k, v)
	for i := height - 1; i >= 0; i-- {
		cacheNode := sl.cache[i]
		newNode.levels[i].next = cacheNode.levels[i].next
		newNode.levels[i].prev = sl.cache[i]
		if cacheNode.levels[i].next != nil {
			cacheNode.levels[i].next.levels[i].prev = newNode
		}
		cacheNode.levels[i].next = newNode
	}

	if newNode.levels[0].next == nil {
		sl.tail = newNode
	}

	sl.numOfNodes++
	return nil
}

func (sl *skiplist) Erase(k interface{}) error {
	node := sl.find(k)
	if node == nil {
		return ErrKeyNotExist
	}

	for i := node.height - 1; i >= 0; i-- {
		node.levels[i].prev.levels[i].next = node.levels[i].next
		if node.levels[i].next != nil {
			node.levels[i].next.levels[i].prev = node.levels[i].prev
		}
	}

	if node == sl.tail {
		if sl.head == node.levels[0].prev {
			sl.tail = nil
		} else {
			sl.tail = node.levels[0].prev
		}
	}

	if node.height == sl.curHeight && node.height != 1 {
		for i := node.height - 1; i > 0; i-- {
			if sl.head.levels[i].next == nil {
				sl.curHeight--
			} else {
				break
			}
		}
	}

	sl.numOfNodes--
	return nil
}

func (sl *skiplist) Find(k interface{}) (v interface{}, exist bool) {
	node := sl.find(k)
	if node == nil {
		return nil, false
	}
	return node.v, true
}

func (sl *skiplist) Size() int {
	return sl.numOfNodes
}

func (sl *skiplist) Clear() {
	sl.curHeight = 1
	sl.numOfNodes = 0
	sl.head = createNode(sl.maxHeight, nil, nil)
}

func (sl *skiplist) Empty() bool {
	return sl.numOfNodes == 0
}

func (sl *skiplist) Min() (k, v interface{}) {
	if sl.numOfNodes == 0 {
		return nil, nil
	}
	minNode := sl.head.levels[0].next
	return minNode.k, minNode.v
}

func (sl *skiplist) Max() (k, v interface{}) {
	if sl.tail == nil {
		return nil, nil
	}
	return sl.tail.k, sl.tail.v
}

func (sl *skiplist) Begin() Iterator {
	if sl.numOfNodes == 0 {
		return nil
	}
	return &iterator{
		sl: sl,
		n:  sl.head.levels[0].next,
	}
}

func (sl *skiplist) End() Iterator {
	return nil
}

func createNode(height int, k, v interface{}) *node {
	n := &node{
		k:      k,
		v:      v,
		height: height,
		levels: make([]*level, height),
	}
	for i := 0; i < height; i++ {
		n.levels[i] = &level{}
	}
	return n
}

func (sl *skiplist) find(k interface{}) *node {
	node := sl.head
	for i := sl.curHeight - 1; i >= 0; i-- {
		for ; node.levels[i].next != nil; node = node.levels[i].next {
			cmpRes := sl.compare(k, node.levels[i].next.k)
			if cmpRes < 0 {
				break
			} else if cmpRes == 0 {
				return node.levels[i].next
			}
		}
	}
	return nil
}

func (sl *skiplist) debugPrint() {
	fmt.Println("-------------------")
	for i := sl.curHeight - 1; i >= 0; i-- {
		node := sl.head.levels[i].next
		for ; node != nil; node = node.levels[i].next {
			fmt.Printf("%v ", node.k)
		}
		fmt.Printf("\n")
	}
	fmt.Println("-------------------")
}

func (sl *skiplist) randomHeight() (n int) {
	//return 2
	for n = 1; n < sl.maxHeight && rand.Float64() < kChance; n++ {
	}
	return
}

func init() {
	rand.Seed(time.Now().Unix())
}
