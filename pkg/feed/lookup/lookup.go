// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

/*
Package lookup defines feed lookup algorithms and provides tools to place updates,
so they can be found
*/
package lookup

import (
	"context"
	"time"
)

const maxuint64 = ^uint64(0)

// LowestLevel establishes the frequency resolution of the lookup algorithm as a power of 2.
const LowestLevel uint8 = 0 // default is 0 (1 second)

// HighestLevel sets the lowest frequency the algorithm will operate at, as a power of 2.
// 31 -> 2^31 equals to roughly 38 years.
const HighestLevel = 31

// DefaultLevel sets what level will be chosen to search when there is no hint
const DefaultLevel = HighestLevel

// Algorithm is the function signature of a lookup algorithm
type Algorithm func(ctx context.Context, now uint64, hint Epoch, read ReadFunc) (value interface{}, err error)

// Lookup finds the update with the highest timestamp that is smaller or equal than 'now'
// It takes a hint which should be the epoch where the last known update was
// If you don't know in what epoch the last update happened, simply submit lookup.NoClue
// read() will be called on each lookup attempt
// Returns an error only if read() returns an error
// Returns nil if an update was not found
var Lookup Algorithm = LongEarthAlgorithm

// TimeAfter must point to a function that returns a timer
// This is here so that tests can replace it with
// a mock-up timer factory to simulate time deterministically
var TimeAfter = time.After

// ReadFunc is a handler called by Lookup each time it attempts to find a value
// It should return <nil> if a value is not found
// It should return <nil> if a value is found, but its timestamp is higher than "now"
// It should only return an error in case the handler wants to stop the
// lookup process entirely.
// If the context is canceled, it must return context.Canceled

// ReadFunc
type ReadFunc func(ctx context.Context, epoch Epoch, now uint64) (interface{}, error)

// NoClue is a hint that can be provided when the Lookup caller does not have
// a clue about where the last update may be
var NoClue = Epoch{}

// getBaseTime returns the epoch base time of the given
// time and level
func getBaseTime(t uint64, level uint8) uint64 {
	return t & (maxuint64 << level)
}

// Hint creates a hint based only on the last known update time
// skipcq: TCV-001
func Hint(last uint64) Epoch {
	return Epoch{
		Time:  last,
		Level: DefaultLevel,
	}
}

// GetNextLevel returns the frequency level a next update should be placed at, provided where
// the last update was and what time it is now.
// This is the first nonzero bit of the XOR of 'last' and 'now', counting from the highest significant bit
// but limited to not return a level that is smaller than the last-1
func GetNextLevel(last Epoch, now uint64) uint8 {
	// First XOR the last epoch base time with the current clock.
	// This will set all the FairOS most significant bits to zero.
	mix := last.Base() ^ now

	// Then, make sure we stop the below loop before one level below the current, by setting
	// that level's bit to 1.
	// If the next level is lower than the current one, it must be exactly level-1 and not lower.
	mix |= 1 << (last.Level - 1)

	// if the last update was more than 2^highestLevel seconds ago, choose the highest level
	if mix > (maxuint64 >> (64 - HighestLevel - 1)) {
		return HighestLevel
	}

	// set up a mask to scan for nonzero bits, starting at the highest level
	mask := uint64(1 << (HighestLevel))

	for i := uint8(HighestLevel); i > LowestLevel; i-- {
		if mix&mask != 0 { // if we find a nonzero bit, this is the level the next update should be at.
			return i
		}
		mask = mask >> 1 // move our bit one position to the right
	}
	return 0
}

// GetNextEpoch returns the epoch where the next update should be located
// according to where the previous update was
// and what time it is now.
func GetNextEpoch(last Epoch, now uint64) Epoch {
	if last == NoClue {
		return GetFirstEpoch(now)
	}
	level := GetNextLevel(last, now)
	return Epoch{
		Level: level,
		Time:  now,
	}
}

// GetFirstEpoch returns the epoch where the first update should be located
// based on what time it is now.
func GetFirstEpoch(now uint64) Epoch {
	return Epoch{Level: HighestLevel, Time: now}
}

var worstHint = Epoch{Time: 0, Level: 63}

var trace = func(id int32, formatString string, a ...interface{}) {
	// fmt.Printf("Step ID #%d "+formatString+"\n", append([]interface{}{id}, a...)...)
}
