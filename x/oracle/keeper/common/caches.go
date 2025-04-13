package common

import (
	"fmt"
	"slices"
)

type Caches struct {
	nstStakerList map[uint64][]string
}

func NewCaches() *Caches {
	return &Caches{
		nstStakerList: make(map[uint64][]string),
	}
}

func (c *Caches) GetNSTStakerList(chainID uint64) []string {
	if c.nstStakerList == nil {
		c.nstStakerList = make(map[uint64][]string)
	}
	return c.nstStakerList[chainID]
}

func (c *Caches) SetNSTStakerList(chainID uint64, sl []string) {
	if c.nstStakerList == nil {
		c.nstStakerList = make(map[uint64][]string)
	}
	c.nstStakerList[chainID] = sl
}

func (c *Caches) RemoveNSTStakerList(chainID uint64) {
	if c.nstStakerList == nil {
		c.nstStakerList = make(map[uint64][]string)
	}
	delete(c.nstStakerList, chainID)
}

func (c *Caches) UpdateNSTStakerList(chainID uint64, from, to int, stakerFrom, stakerTo string) bool {
	if c.nstStakerList == nil {
		c.nstStakerList = make(map[uint64][]string)
	}

	sl := c.nstStakerList[chainID]
	// append
	if to == len(sl) {
		sl = append(sl, stakerTo)
		c.nstStakerList[chainID] = sl
		return true
	}

	if from != len(sl)-1 {
		return false
	}

	// remove
	if to == -1 {
		if from < 0 {
			// we have verified from == len(sl) - 1, so this just means a useless operation
			return true
		}

		if from == 0 {
			delete(c.nstStakerList, chainID)
			return true
		}

		sl = sl[:from]
		c.nstStakerList[chainID] = sl
		return true
	}

	// rotate
	if len(sl) <= 1 {
		return false
	}

	if to < 0 || to > len(sl) ||
		stakerFrom != sl[from] ||
		stakerTo != sl[to] {
		return false
	}
	sl[to] = sl[from]
	sl = sl[:from]
	c.nstStakerList[chainID] = sl
	return true
}

func (c *Caches) RotateStakerList(chainID uint64, indexes []uint32) (map[uint32]string, error) {
	if c.nstStakerList == nil {
		c.nstStakerList = make(map[uint64][]string)
	}
	l := len(indexes)
	if l == 0 {
		return nil, nil
	}
	sl := c.nstStakerList[chainID]
	l2 := len(sl)
	if len(sl) == 0 {
		return nil, fmt.Errorf("remove more stakers than exists, existing:%d, remove:%d", l2, l)
	}
	slices.Sort(indexes)
	// remove duplicates
	for i := 0; i < l-1; i++ {
		if indexes[i] == indexes[i+1] {
			indexes = slices.Delete(indexes, i, i+1)
			l--
			i--
		}
	}
	// make sure all indexes are valid
	if int(indexes[l-1]) >= l2 {
		return nil, fmt.Errorf("remove index exceeds exisintg max index, max:%d, remove:%d", l2, indexes[l-1])
	}
	ret := make(map[uint32]string)
	removeMap := make(map[uint32]bool)
	for _, i := range indexes {
		removeMap[i] = true
	}

	i := 0
	j := 1
	for ; j <= l2 && j <= l; j++ {
		//#nosec G115
		if int(indexes[i]) < l2-j && removeMap[uint32(l2-j)] {
			continue
		}
		if int(indexes[i]) == l2-j {
			j++
			break
		}
		if int(indexes[i]) > l2-j {
			break
		}
		ret[indexes[i]] = sl[l2-j]
		sl[indexes[i]] = sl[l2-j]
		i++
	}

	if j > l2 {
		j = l2
		//#nosec G115
	} else if j > 0 {
		j--
	}
	sl = sl[:l2-j]
	if len(sl) == 0 {
		delete(c.nstStakerList, chainID)
	} else {
		c.nstStakerList[chainID] = sl
	}
	if len(ret) == 0 {
		ret = nil
	}
	return ret, nil
}
