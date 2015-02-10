package proxy

import (
	"errors"
	"fmt"
	"sync"
)

type Container interface {
	Get() (interface{}, error)
	Put(item interface{})
	Del(item interface{}) error
}

type RRContainer struct {
	lk        sync.Mutex
	count     int
	container []interface{}
}

func NewRRContainer() *RRContainer {
	return &RRContainer{}
}

func (rr *RRContainer) Get() (interface{}, error) {
	rr.lk.Lock()
	defer rr.lk.Unlock()

	size := len(rr.container)
	if size == 0 {
		return nil, errors.New("RRContainer is Empty")
	}
	count := rr.count
	rr.count = (rr.count + 1) % size
	return rr.container[count], nil
}

func (rr *RRContainer) Put(item interface{}) {
	rr.lk.Lock()
	defer rr.lk.Unlock()

	rr.container = append(rr.container, item)
	fmt.Println(rr.container)
}

func (rr *RRContainer) Del(item interface{}) (err error) {
	rr.lk.Lock()
	defer rr.lk.Unlock()

	size := len(rr.container)
	for i, v := range rr.container {
		if v == item {
			if i == 0 {
				rr.container = rr.container[1:]
				return
			}
			if i == size-1 {
				rr.container = rr.container[0 : i-1]
				return
			}
			var c []interface{}
			for x := 0; x < i; x = x + 1 {
				c = append(c, rr.container[x])
			}
			for x := i + 1; x < size; x = x + 1 {
				c = append(c, rr.container[x])
			}
			rr.container = c
			return
		}
	}
	return errors.New(fmt.Sprintf("not found item %v", item))
}
