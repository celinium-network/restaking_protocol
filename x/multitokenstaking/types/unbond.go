package types

import time "time"

func (ubd *MTStakingUnbonding) RemoveEntry(i int64) {
	ubd.Entries = append(ubd.Entries[:i], ubd.Entries[i+1:]...)
}

func (e MTStakingUnbondingEntry) IsMature(currentTime time.Time) bool {
	return !e.CompletionTime.After(currentTime)
}
