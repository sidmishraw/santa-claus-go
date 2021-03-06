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
	"bytes"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"time"

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

// Elves `MeetInStudy` and Reindeers `DeliverToys`

// loggerMutex will used for taking lock on the logger

func main() {
	// for i := 0; i < 100; i++ {
	// 	MySTM.Log("Iteration #", i)
	// TestAssemble1()
	// TestAssemble2()
	// 	MySTM.Log()
	// }
	SantaRun()
	// Test1()
}

// Test1 tests the basic workflow
func Test1() {
	elfInGate := NewGate(3)
	elfOutGate := NewGate(3)
	reindeerInGate := NewGate(9)
	reindeerOutGate := NewGate(9)

	MySTM.Log("elfInGate = ", *elfInGate)
	MySTM.Log("elfOutGate = ", *elfOutGate)
	MySTM.Log("reindeerInGate = ", *reindeerInGate)
	MySTM.Log("reindeerOutGate = ", *reindeerOutGate)

	gateCells := make([]*stm.MemoryCell, 0)
	gateCells = append(gateCells, elfInGate, elfOutGate, reindeerInGate, reindeerOutGate)

	//# initialize the gate's remaining so that they can pass through
	for _, gateCell := range gateCells {
		OperateGate(gateCell)
	}
	//# initialize the gate's remaining so that they can pass through

	tLog := MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			MySTM.Log("gatecells = ", gateCells)
			for _, gcell := range gateCells {
				gate := t.ReadT(gcell).(*Gate)
				MySTM.Log(*gcell, " = ", gate)
			}
			return true
		}).
		Done()

	MySTM.Log("initial state - ")
	MySTM.Exec(tLog)

	//# operate
	PassGate(elfInGate)
	PassGate(elfInGate)
	PassGate(reindeerOutGate)
	//# operate

	MySTM.Log("final state - ")
	MySTM.Exec(tLog)
}

// TestAssemble1 tests the assembly and validity of the functions for elves
func TestAssemble1() {
	elfGroup := NewGroup(1) // group of capacity 1
	go func() {
		inGateCell, outGateCell := AwaitGroup(elfGroup) // AwaitGroup awaits for the group to be filled atomically - transactionally
		OperateGate(inGateCell)
		OperateGate(outGateCell)
	}()
	MySTM.Exec(NewElf(12, elfGroup))
}

// TestAssemble2 tests the assembly and validity of the functions for reindeers
func TestAssemble2() {
	reindeerGroup := NewGroup(1)                     // group of capacity 1
	MySTM.ForkAndExec(NewReindeer(1, reindeerGroup)) // fork and execute the transaction in another thread
	inGateCell, outGateCell := AwaitGroup(reindeerGroup)
	OperateGate(inGateCell)
	OperateGate(outGateCell)
}

// SantaRun runs the main scenario
func SantaRun() {
	elfGroup := NewGroup(3)      // Santa can be woken up by 3 elves
	reindeerGroup := NewGroup(9) // Santa can be woken up by 9 reindeers
	for i := 1; i <= 10; i++ {
		// creates all the elves
		go Forever(NewElf(i, elfGroup)) // spawn - fork and go on forever
	}
	for i := 1; i <= 9; i++ {
		// creates all the reindeers
		go Forever(NewReindeer(i, reindeerGroup)) // spawn - fork and go on forever
	}
	log.Println("Beginning Santa's workplace simulation ~~~")
	workReindeers, reindeerChannel, workElves, elfChannel := Santa(elfGroup, reindeerGroup) // just to reuse the transactions
	for i := 0; i < 3; i++ {
		//# RUN!
		MySTM.ForkAndExec(workReindeers, workElves)
		//# RUN!
		//# Selection logic
		// The selection logic, as per the problem description
		// reindeers are given higher precedence, hence I wait for them to write
		// to the reindeer channel before proceeding forward
		// then I slip in a thread sleep to give the elves some time to reassemble.
		if rGates := <-reindeerChannel; len(rGates) == 2 {
			MySTM.Log("Ho! Ho! Hoo! Let's go deliver some toys! :D")
			MySTM.Log("-------------------------------------------")
			for _, gatecell := range rGates {
				OperateGate(gatecell) // operate the gates
			}
			time.Sleep(5 * time.Second) // for pretty printing
		}
		if eGates := <-elfChannel; len(eGates) == 2 {
			MySTM.Log("Ho! Ho! Hoo! Let's hold the meeting in the study! :D")
			MySTM.Log("----------------------------------------------------")
			for _, gatecell := range eGates {
				OperateGate(gatecell) // operate the gates
			}
			time.Sleep(5 * time.Second) // just for pretty printing
		}
		//# Selection logic
	}
	log.Println("Santa is done for the day. Bye Santa ~~~")
}

//# Santa's task

// Santa represents what Santa has to do. He waits until there is either a group of reindeers,
// or a group of elves. Once he has made his choice of which group to attend to, he must
// take them through their task.
// According to the problem description, a group of reindeers has higher priority than a group of elves.
func Santa(elfGroupCell *stm.MemoryCell, reindeerGroupCell *stm.MemoryCell) (workReindeers *stm.Transaction, reindeerChannel chan [2]*stm.MemoryCell, workElves *stm.Transaction, elfChannel chan [2]*stm.MemoryCell) {
	reindeerChannel = make(chan [2]*stm.MemoryCell) // channel holding the reindeer gates
	elfChannel = make(chan [2]*stm.MemoryCell)      // channel holding the elven gates
	//# work those reindeers Santa!
	workReindeers = MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			inGateCell, outGateCell := AwaitGroup(reindeerGroupCell)
			reindeerChannel <- [2]*stm.MemoryCell{inGateCell, outGateCell} // dump into the reindeer channel
			return true
		}).
		Done()
	//# work those reindeers Santa!
	//# work those elves Santa!
	workElves = MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			inGateCell, outGateCell := AwaitGroup(elfGroupCell)
			elfChannel <- [2]*stm.MemoryCell{inGateCell, outGateCell} // dump into the elf channel
			return true
		}).
		Done()
	//# work those elves Santa!
	return workReindeers, reindeerChannel, workElves, elfChannel
}

//# Santa's task

// # Helper's tasks

// NewElf simulates the actions an elf needs to do.
// An elf joins the group, then passes through the inGate, meets with Santa,
// then passes through the outGate and out of Santa's study.
// The ID represents the ID of the elf, and the groupCell is the MemoryCell containing the
// Group the elf has to join. NewElfMeet returns the stm.Transaction pointer that can be
// used to execute the transaction at a later time(when ever you need)
func NewElf(ID int, groupCell *stm.MemoryCell) (elf *stm.Transaction) {
	elf = MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			inGate, outGate := JoinGroup(groupCell) // join the group
			PassGate(inGate)                        // elf passes through the inGate
			MeetInStudy(ID)                         // Meets with Santa in the Study
			PassGate(outGate)                       // elf leaves the study and passes out of the outGate
			return true
		}).
		Done()
	return elf
}

// NewReindeer simulates the action a reindeer needs to do.
// A reindeer joins the group, then passes through the inGate, delivers the toys with Santa,
// passes through the outGate and is done.
func NewReindeer(ID int, groupCell *stm.MemoryCell) (reindeer *stm.Transaction) {
	reindeer = MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			inGate, outGate := JoinGroup(groupCell) // join the group
			PassGate(inGate)                        // reindeer passes through the inGate
			DeliverToys(ID)                         // meets with Santa and delivers toys
			PassGate(outGate)                       // reindeer leaves the study and passes out of the outGate
			return true
		}).
		Done()
	return reindeer
}

// MeetInStudy is the operation carried out by Elves. It takes in the Elf's ID and prints
// "Elf #ID meeting in the study".
func MeetInStudy(elfID int) {
	MySTM.Log("Elf " + fmt.Sprint(elfID) + " meeting in the study")
}

// DeliverToys is the operation carried out by Reindeers. It takes in the Reindeer's ID and prints
// "Reindeer #ID delivering toys".
func DeliverToys(reinID int) {
	MySTM.Log("Reindeer " + fmt.Sprint(reinID) + " delivering toys")
}

// # Helper's tasks

// MySTM represents the single shared data store
var MySTM = stm.NewSTM()

//# Group

// Group represents the group of helpers. It is created empty with a specified capacity.
// A helper (Elf of reindeer) may join a group by calling the `JoinGroup` function.
// The call to `JoinGroup` blocks if the group is full. Santa calls `AwaitGroup` function to
// wait for the group to be full, when it is full he gets the Group's gates and the group is
// immediately re-initialized with fresh gates so that another group of eager elves can start
// assembling.
type Group struct {
	capacity   int
	spacesLeft int
	inGate     *stm.MemoryCell `name:"inGate" type:"*Gate"`
	outGate    *stm.MemoryCell `name:"outGate" type:"*Gate"`
}

// NewGroup makes a new group of the desired capacity in the STM and returns the stm.MemoryCell
// that holds the new group.
func NewGroup(capacity int) *stm.MemoryCell {
	group := new(Group)
	group.capacity = capacity   // the groups capacity
	group.spacesLeft = capacity // the group is initially empty,hence the spacesLeft = capacity, it gets decrement with each addition
	group.inGate = NewGate(capacity)
	group.outGate = NewGate(capacity)
	groupData := stm.Data(group)
	groupCell := MySTM.MakeMemCell(&groupData)
	return groupCell
}

// JoinGroup lets the helpers join the group. This is a transactional operation. It updates the
// group in the STM. JoinGroup first checks if the Group is full. If the group is full, the call
// blocks. Otherwise, it updates its member count and the member is added. It returns the MemoryCells
// of the inGate and outGate of the Group.
//
// @transactional
func JoinGroup(groupCell *stm.MemoryCell) (ingateCell, outGateCell *stm.MemoryCell) {
	// MySTM.Log("Trying to join group = ", *groupCell)
	t := MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			group := t.ReadT(groupCell).(*Group) // read transactionally
			checkExpression := func() bool {
				if group.spacesLeft <= 0 {
					return false // block, since no spaces remaining in the group, the new members cannot be added
				}
				return true // pass through
			}
			Check(checkExpression)                      // blocks till checkExpression evaluates to true
			group.spacesLeft--                          // update the spacesLeft
			ingateCell = group.inGate                   // ingate cell
			outGateCell = group.outGate                 // out gate cell
			return t.WriteT(groupCell, stm.Data(group)) // write transactionally to the STM
		}).
		Done()
	MySTM.Exec(t)
	// MySTM.Log("Joined group = ", *groupCell)
	return ingateCell, outGateCell
}

// AwaitGroup makes new Gates when it re-initializes the Group. This ensures that
// a new group can assemble while the old one is still talking to Santa in the study,
// with no danger of an elf from the new group overtaking a sleepy elf from the old one.
// It returns the stm.MemoryCells containing the newly created Gates.
//
// @transactional
func AwaitGroup(groupCell *stm.MemoryCell) (inGateCell, outGateCell *stm.MemoryCell) {
	// MySTM.Log("Awaiting group = ", *groupCell)
	t := MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			group := t.ReadT(groupCell).(*Group) // read transactionally
			inGateCell = group.inGate            // to be returned
			outGateCell = group.outGate          // to be returned
			checkExpression := func() bool {
				if group.spacesLeft <= 0 {
					return true // go through
				}
				return false // block
			}
			Check(checkExpression) // blocks until checkExpression evaluates to true
			// the group is full and Santa can start processing, meet with elves
			// or go deliver gifts with the reindeers.
			group.spacesLeft = group.capacity           // restore spaces_left
			group.inGate = NewGate(group.capacity)      // new ingate
			group.outGate = NewGate(group.capacity)     // new outgate
			return t.WriteT(groupCell, stm.Data(group)) // update the group transactionally
		}).
		Done()
	MySTM.Exec(t)
	// MySTM.Log("Done Awaiting group = ", *groupCell)
	return inGateCell, outGateCell
}

//# Group

//# Gate

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
// The new gate is made inside the STM, so the NewGate returns a new `stm.MemoryCell`.
// Now, the value of this structure can only be accessed inside a `stm.TransactionContext`.
func NewGate(capacity int) *stm.MemoryCell {
	gate := new(Gate)
	gate.capacity = capacity
	gate.remaining = 0
	gateData := stm.Data(gate) // convert into `stm.Data` to store in a `stm.MemoryCell`
	return MySTM.MakeMemCell(&gateData)
}

// PassGate allows the helper to pass through the gate. When a helper passes through
// the gate, the remaining capacity is decremented. If the remaining capacity of
// the gate is 0, then, the call blocks. It is a transactional operation hence,
// it takes in the MemoryCell containing the Gate instance.
//
// To simulate blocking, I've failed the transaction and forced to retry it till it
// succeeds. Similar to blocking?
// @transactional
func PassGate(gateCell *stm.MemoryCell) {
	// MySTM.Log(getGID(), " Passing through gate = ", *gateCell)
	t := MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			gate := t.ReadT(gateCell).(*Gate) // read from MemoryCell transactionally
			checkingExpression := func() bool {
				if gate.remaining <= 0 {
					return false // block
				}
				return true // go through
			}
			Check(checkingExpression)                 // blocks till checkingExpression is false
			gate.remaining--                          // update the remaining count
			return t.WriteT(gateCell, stm.Data(gate)) // write the updated gate's content
		}).
		Done()
	MySTM.Exec(t) // execute the transaction, this will block the main thread?
	// MySTM.Log(getGID(), " Passed through gate = ", *gateCell)
}

// OperateGate is used by Santa to reset the remaining capacity of the gate back to n or capacity.
// It takes in a stm.MemoryCell that contains the Gate inside it.
func OperateGate(gateCell *stm.MemoryCell) {
	// MySTM.Log("Operating gate = ", *gateCell)
	t := MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			gate := t.ReadT(gateCell).(*Gate)         // read transactionally
			gate.remaining = gate.capacity            // restore to full capacity
			return t.WriteT(gateCell, stm.Data(gate)) // write transactionally
		}).
		Done()
	MySTM.Exec(t)
	// MySTM.Log("Operated gate = ", *gateCell)
}

//# Gate

//# Transactional checks

// Check evaluates the expression and verifies if the operation was successful or not.
// If the expression was not successful, the Check blocks.
func Check(expression func() bool) {
	// MySTM.Log("Goroutine#", getGID())
	checkingT := MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			return expression() // the transaction succeeds if expression evaluates to true else it blocks
		}).
		Done()
	MySTM.Exec(checkingT)
	// MySTM.Log("Goroutine#", getGID())
}

//# Transactional checks

// Forever runs the transaction forever.
func Forever(t *stm.Transaction) {
	for {
		MySTM.Exec(t)
	}
}

// getGID fetches the goroutine's ID from the stacktrace
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
