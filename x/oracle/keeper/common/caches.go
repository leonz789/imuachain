package common

import "github.com/imua-xyz/imuachain/x/oracle/types"

type Caches struct {
	nstStakerList map[uint64][]*types.StakerListEntry
}

func NewCaches() *Caches {
	return &Caches{
		nstStakerList: make(map[uint64][]*types.StakerListEntry),
	}
}

func (c *Caches) ensureInitialized() {
	if c.nstStakerList == nil {
		c.nstStakerList = make(map[uint64][]*types.StakerListEntry)
	}
}

func (c *Caches) GetNSTStakerList(chainID uint64) []*types.StakerListEntry {
	c.ensureInitialized()
	return c.nstStakerList[chainID]
}

func (c *Caches) ClearNSTStaker(chainID uint64, stakerIdx uint32) bool {
	c.ensureInitialized()
	sl := c.nstStakerList[chainID]
	if sl == nil || int(stakerIdx) >= len(sl) {
		return false
	}
	// Clear the staker at the given index
	sl[stakerIdx] = &types.StakerListEntry{
		StakerAddr:    sl[stakerIdx].StakerAddr,
		ActiveVersion: 0,
		LatestVersion: 0,
	}
	c.nstStakerList[chainID] = sl
	return true
}

func (c *Caches) SetNSTStakerList(chainID uint64, sl []*types.StakerListEntry) {
	c.ensureInitialized()
	c.nstStakerList[chainID] = sl
}

func (c *Caches) RemoveNSTStakerList(chainID uint64) {
	c.ensureInitialized()
	delete(c.nstStakerList, chainID)
}

func (c *Caches) UpdateNSTStakerLatestVersion(chainID uint64, stakerIdx uint32, stakerAddr string, version uint64) bool {
	c.ensureInitialized()
	sl := c.nstStakerList[chainID]
	if sl == nil || int(stakerIdx) >= len(sl) {
		return false
	}
	if sl[stakerIdx].StakerAddr != stakerAddr {
		return false
	}
	// Update the latest version if it's greater than the current one
	if sl[stakerIdx].LatestVersion < version {
		sl[stakerIdx].LatestVersion = version
		c.nstStakerList[chainID] = sl
		return true
	}
	return false
}

// AddNSTStaker adds a staker to the list for the given chainID at the specified index.
// NOTE: not concurrent safe, caller must ensure synchronization.
func (c *Caches) AddNSTStaker(chainID uint64, staker string, index uint32, activeVersion uint64) bool {
	if c.nstStakerList == nil {
		if index > 0 {
			return false
		}
		c.nstStakerList = make(map[uint64][]*types.StakerListEntry)
	}
	sl := c.nstStakerList[chainID]
	if int(index) != len(sl) {
		return false
	}
	c.nstStakerList[chainID] = append(sl, &types.StakerListEntry{
		StakerAddr:    staker,
		ActiveVersion: activeVersion,
		LatestVersion: activeVersion,
	})
	return true
}
