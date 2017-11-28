/**
* main.go
* @author Sidharth Mishra
* @description The Santa Claus Problem implementation

Problem Statement:
Santa repeatedly sleeps until wakened by either all of his nine reindeer, back from their holidays, or by a group of three of his ten elves. If awakened by the reindeer, he harnesses each of them to his sleigh, delivers toys with them and finally unharnesses them (allowing them to go off on holiday). If awakened by a group of elves, he shows each of the group into his study, consults with them on toy R&D and finally shows them each out (allowing them to go back to work). Santa should give priority to the reindeer in the case that there is both a group of elves and a group of reindeer waiting.

[1] S. P. Jones, "Beautiful concurrency", ch. 4, [Online] Available. https://www.schoolofhaskell.com/school/advanced-haskell/beautiful-concurrency/4-the-santa-claus-problem

* @created Mon Nov 27 2017 21:44:28 GMT-0800 (PST)
* @copyright 2017 Sidharth Mishra
* @last-modified Mon Nov 27 2017 21:44:28 GMT-0800 (PST)
*/

package main

import (
	"fmt"

	"github.com/sidmishraw/stm-reworked/stm"
)

// According to S.P. Jones, Santa makes 2 `Group`s: Elves and Reindeers. Each Elf and Reindeer tries to
// join their respective group. Upon successful joining, two Gates are returned. `EntryGate` and `ExitGate`.
// The `EntryGate` allows Santa to control when the elf can enter the study, and also lets Santa know when
// they are all inside. The `ExitGate` controls the elves leaving the study.
//
// Santa waits for either of his two groups to be ready, and then uses that Group's Gate's to marshal
// his helpers(elves or reindeers) through their task.
//
// The helpers spend their time in an infinite loop: try to join a group, move through the gates under Santa's
// control, and then delay for a random interval before trying to join a group again.

// My interpretation:
// From the last description, it can be inferred that the helpers are threads that do `join a group`,
// `move through gates`, and `delay for some time` actions in sequence infinitely.
// From my STM model, I can represent each of these sequences as a transaction?

// Elves `meetInStudy` and Reindeers `deliverToys`

func main() {

}

// meetInStudy is the operation carried out by Elves. It takes in the Elf's ID and prints
// "Elf #ID meeting in the study".
func meetInStudy(elfID int) {
	fmt.Println("Elf ", elfID, " meeting in the study")
}

// deliverToys is the operation carried out by Reindeers. It takes in the Reindeer's ID and prints
// "Reindeer #ID delivering toys".
func deliverToys(reinID int) {
	fmt.Println("Reindeer ", reinID, " delivering toys")
}

// MySTM represents the single shared data store
var MySTM = stm.NewSTM()

// Group represents the group of helpers
type Group struct {
}

// Gate represents the gate held by Santa. A Gate has a fixed `capacity` which we need to specify
// while making a new Gate, and a mutable `remaining` capacity. The `remaining` capacity is
// decremented whenever a helper calls `passGate` to go through the gate.
// If the capacity is 0, passGate blocks.
// A Gate is created with zero remaining capacity, so that no helpers can pass through it.
// Santa opens the gate with `operateGate`, which sets its remaining capacity back to n.
type Gate struct {
	capacity  int
	remaining int
}

// NewGate makes a new Gate of the given capacity and 0 remaining capacity.
// The new gate is made inside the STM, so the NewGate returns a new stm.MemoryCell.
// Now, the value of this structure can only be accessed inside a stm.TransactionContext.
func NewGate(capacity int) *stm.MemoryCell {
	gate := new(Gate)
	gate.capacity = capacity
	gate.remaining = 0
	gateData := stm.Data(gate)
	return MySTM.MakeMemCell(&gateData)
}

// PassGate allows the helper to pass through the gate. When a helper passes through
// the gate, the remaining capacity is decremented. If the remaining capacity of
// the gate is 0, then, the call blocks.
func (gate *Gate) PassGate() {
	if gate.remaining > 0 {
		gate.remaining--
	} else {

	}
}

// OperateGate is used by Santa to reset the remaining capacity of the gate back to n or capacity.
func (gate *Gate) OperateGate() {
	gate.remaining = gate.capacity
}
