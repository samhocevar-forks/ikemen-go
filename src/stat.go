package main

import (
)

type StatName int

const (
	StatRAM         StatName = iota
	StatVRAM

	StatSoundCount
	StatSoundMemory
	StatTextureCount
	StatTextureMemory
)

var _globalStats = make(map[StatName]float64)

func init() {
}

func GetStat(name StatName) float64 {
	return _globalStats[name]
}

func UpdateStat(name StatName, delta float64) {
	_globalStats[name] += delta

	switch name {
	case StatSoundMemory:
		_globalStats[StatRAM] += delta
	case StatTextureMemory:
		_globalStats[StatVRAM] += delta
	}
}
