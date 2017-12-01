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

// Elves `MeetInStudy` and Reindeers `DeliverToys`

func main() {
	// for i := 0; i < 100; i++ {
	// 	fmt.Println("Iteration #", i)
	TestAssemble1()
	// 	TestAssemble2()
	// 	fmt.Println()
	// }
	// SantaRun()
	// Test1()
}

// Test1 tests the basic workflow
func Test1() {
	elfInGate := NewGate(3)
	elfOutGate := NewGate(3)
	reindeerInGate := NewGate(9)
	reindeerOutGate := NewGate(9)

	fmt.Println("elfInGate = ", *elfInGate)
	fmt.Println("elfOutGate = ", *elfOutGate)
	fmt.Println("reindeerInGate = ", *reindeerInGate)
	fmt.Println("reindeerOutGate = ", *reindeerOutGate)

	gateCells := make([]*stm.MemoryCell, 0)
	gateCells = append(gateCells, elfInGate, elfOutGate, reindeerInGate, reindeerOutGate)

	//# initialize the gate's remaining so that they can pass through
	for _, gateCell := range gateCells {
		OperateGate(gateCell)
	}
	//# initialize the gate's remaining so that they can pass through

	tLog := MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			fmt.Println("gatecells = ", gateCells)
			for _, gcell := range gateCells {
				gate := t.ReadT(gcell).(*Gate)
				fmt.Println(*gcell, " = ", gate)
			}
			return true
		}).
		Done()

	fmt.Println("initial state - ")
	MySTM.Exec(tLog)

	//# operate
	PassGate(elfInGate)
	PassGate(elfInGate)
	PassGate(reindeerOutGate)
	//# operate

	fmt.Println("final state - ")
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
	elves := make([]*stm.Transaction, 0)
	for i := 1; i <= 10; i++ {
		// creates all the elves
		elves = append(elves, NewElf(i, elfGroup))
	}
	reindeers := make([]*stm.Transaction, 0)
	for i := 1; i <= 9; i++ {
		// creates all the reindeers
		reindeers = append(reindeers, NewReindeer(i, reindeerGroup))
	}
	workReindeers, workElves := Santa(elfGroup, reindeerGroup)
	for {
		MySTM.ForkAndExec(reindeers...)
		MySTM.ForkAndExec(elves...)
		MySTM.Exec(workReindeers)
		MySTM.Exec(workElves)
	}
}

//# Santa's task

// Santa represents what Santa has to do. He waits until there is either a group of reindeers,
// or a group of elves. Once he has made his choice of which group to attend to, he must
// take them through their task.
// According to the problem description, a group of reindeers has higher priority than a group of elves.
// It returns the two atomic transactions for reindeers and elves ready to be executed
func Santa(elfGroupCell *stm.MemoryCell, reindeerGroupCell *stm.MemoryCell) (workReindeers, workElves *stm.Transaction) {
	//# work those reindeers Santa!
	workReindeers = MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			inGateCell, outGateCell := AwaitGroup(reindeerGroupCell)
			fmt.Println("Ho! Ho! Hoo! Let's go deliver some toys! :D")
			OperateGate(inGateCell)
			OperateGate(outGateCell)
			return true
		}).
		Done()
	//# work those reindeers Santa!
	//# work those elves Santa!
	workElves = MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			inGateCell, outGateCell := AwaitGroup(elfGroupCell)
			fmt.Println("Ho! Ho! Hoo! Let's hold the meeting in the study! :D")
			OperateGate(inGateCell)
			OperateGate(outGateCell)
			return true
		}).
		Done()
	//# work those elves Santa!
	return workReindeers, workElves
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
			inGate, _ := JoinGroup(groupCell) // join the group
			PassGate(inGate)                  // elf passes through the inGate
			// wg := new(sync.WaitGroup)
			// wg.Add(1)
			// go func() {
			// 	MeetInStudy(ID)
			// 	wg.Done()
			// }() // meets with Santa in his study
			// wg.Wait()
			// PassGate(outGate)                       // elf leaves the study and passes out of the outGate
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
			inGate, _ := JoinGroup(groupCell) // join the group
			PassGate(inGate)                  // reindeer passes through the inGate
			DeliverToys(ID)                   // meets with Santa and delivers toys
			// PassGate(outGate)                       // reindeer leaves the study and passes out of the outGate
			return true
		}).
		Done()
	return reindeer
}

// MeetInStudy is the operation carried out by Elves. It takes in the Elf's ID and prints
// "Elf #ID meeting in the study".
func MeetInStudy(elfID int) {
	fmt.Println("Elf ", elfID, " meeting in the study")
}

// DeliverToys is the operation carried out by Reindeers. It takes in the Reindeer's ID and prints
// "Reindeer #ID delivering toys".
func DeliverToys(reinID int) {
	fmt.Println("Reindeer ", reinID, " delivering toys")
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
	fmt.Println("Trying to join group = ", *groupCell)
	t := MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			group := t.ReadT(groupCell).(*Group) // read transactionally
			blockStatus := false
			if group.spacesLeft <= 0 {
				// if the number of spaces left is 0 or less, then block member addition to the group
				blockStatus = true // group is full, block
			} else {
				// members can be added, add the member - decrement the spacesLeft
				group.spacesLeft--                                  // update the spacesLeft
				writeStatus := t.WriteT(groupCell, stm.Data(group)) // write transactionally to the STM
				if writeStatus {
					updatedGroup := t.ReadT(groupCell).(*Group)
					ingateCell = updatedGroup.inGate
					outGateCell = updatedGroup.outGate
				}
				blockStatus = false || !writeStatus
			}
			return !blockStatus
		}).
		Done()
	MySTM.Exec(t)
	fmt.Println("Joined group = ", *groupCell)
	return ingateCell, outGateCell
}

// AwaitGroup makes new Gates when it re-initializes the Group. This ensures that
// a new group can assemble while the old one is still talking to Santa in the study,
// with no danger of an elf from the new group overtaking a sleepy elf from the old one.
// It returns the stm.MemoryCells containing the newly created Gates.
//
// @transactional
func AwaitGroup(groupCell *stm.MemoryCell) (inGateCell, outGateCell *stm.MemoryCell) {
	fmt.Println("Awaiting group = ", *groupCell)
	t := MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			group := t.ReadT(groupCell).(*Group) // read transactionally
			inGateCell = group.inGate            // to be returned
			outGateCell = group.outGate          // to be returned
			blockStatus := false                 // signifies the blocking status, true means the call should block, otherwise moves through
			if group.spacesLeft <= 0 {
				// the group is full and Santa can start processing, meet with elves
				// or go deliver gifts with the reindeers.
				group.spacesLeft = group.capacity                   // restore spaces_left
				group.inGate = NewGate(group.capacity)              // new ingate
				group.outGate = NewGate(group.capacity)             // new outgate
				writeStatus := t.WriteT(groupCell, stm.Data(group)) // update the group transactionally
				blockStatus = false || !writeStatus
			} else {
				// group is not full, so Santa needs to wait
				// by failing the transaction, it is going to force a retry
				// leading to a blocking call.
				blockStatus = true
			}
			return !blockStatus
		}).
		Done()
	MySTM.Exec(t)
	fmt.Println("Done Awaiting group = ", *groupCell)
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
	fmt.Println("Passing through gate = ", *gateCell)
	t := MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			gate := t.ReadT(gateCell).(*Gate) // read from MemoryCell transactionally
			blockStatus := false              // a positive blockStatus signifies that the call must block
			if gate.remaining <= 0 {
				// since the gate.remaining is 0 or less(?), the call should block
				// best case would be fail this transaction and retry
				blockStatus = true
			} else {
				gate.remaining--                                  // update the remaining count
				writeStatus := t.WriteT(gateCell, stm.Data(gate)) // write the updated gate's content
				blockStatus = false || !writeStatus               // only successful if writeStatus is true
			}
			return !blockStatus
		}).
		Done()
	MySTM.Exec(t) // execute the transaction, this will block the main thread?
	fmt.Println("Passed through gate = ", *gateCell)
}

// OperateGate is used by Santa to reset the remaining capacity of the gate back to n or capacity.
// It takes in a stm.MemoryCell that contains the Gate inside it.
func OperateGate(gateCell *stm.MemoryCell) {
	fmt.Println("Operating gate = ", *gateCell)
	t := MySTM.NewT().
		Do(func(t *stm.Transaction) bool {
			gate := t.ReadT(gateCell).(*Gate)         // read transactionally
			gate.remaining = gate.capacity            // restore to full capacity
			return t.WriteT(gateCell, stm.Data(gate)) // write transactionally
		}).
		Done()
	MySTM.Exec(t)
	fmt.Println("Operated gate = ", *gateCell)
}

//# Gate
