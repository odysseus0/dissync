package discord

import (
	"strconv"
	"time"
)

const discordEpochMs = 1420070400000

func TimestampFromID(id string) time.Time {
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return time.Time{}
	}
	ms := (n >> 22) + discordEpochMs
	return time.UnixMilli(ms)
}

func CompareIDs(a, b string) int {
	ai, _ := strconv.ParseInt(a, 10, 64)
	bi, _ := strconv.ParseInt(b, 10, 64)
	if ai < bi {
		return -1
	}
	if ai > bi {
		return 1
	}
	return 0
}

func MaxID(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	best := ids[0]
	for _, id := range ids[1:] {
		if CompareIDs(id, best) > 0 {
			best = id
		}
	}
	return best
}

func MinID(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	best := ids[0]
	for _, id := range ids[1:] {
		if CompareIDs(id, best) < 0 {
			best = id
		}
	}
	return best
}
