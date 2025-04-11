package common

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
