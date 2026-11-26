package game

import "os"

// The game's data before step C5.2 (wag/data/frozen/afd3335a23f12f41), which
// the tests of the game's mechanics play on: they were written with its
// schemes and questions, which the switch changes (docs/c5-game.md, C5.2).
// Tests of the data the game ships (TestCoreDataLoading, the golden set on the
// new data, item F) read package data.
var prevCore, prevSchemes, prevCQs, prevEdges = readPrevData()

func readPrevData() (core, schemes, cqs, edges []byte) {
	const dir = "../../data/frozen/afd3335a23f12f41/"
	read := func(p string) []byte {
		b, err := os.ReadFile(dir + p)
		if err != nil {
			panic("the frozen pre-C5.2 data: " + err.Error())
		}
		return b
	}
	return read("schemes/core.yml"), read("nascq/schemes.json"), read("nascq/critical_questions.json"), read("nascq/edges.json")
}
