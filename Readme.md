# The Santa Claus Problem implementation using my reworked Go STM

> Author: Sidharth Mishra

## Description

The problem is stated as follows:

```text
Santa repeatedly sleeps until wakened by either all of his nine reindeer, back from their
holidays, or by a group of three of his ten elves. If awakened by the reindeer, he
harnesses each of them to his sleigh, delivers toys with them and finally unharnesses them
(allowing them to go off on holiday). If awakened by a group of elves, he shows each of
the group into his study, consults with them on toy R&D and finally shows them each out
(allowing them to go back to work). Santa should give priority to the reindeer in the case
that there is both a group of elves and a group of reindeer waiting.
```

My implementation is a clone of S. P. Jones' Haskell implementation, but, using my STM
library. There will be modifications and I'll note those modifications in the source code
and here.

## Known deviations

* I used Go's channels for selection. In the Haskell code, S. P. Jones, uses `choose` to
  choose reindeer and elven groups.

```haskell
santa :: Group -> Group -> IO ()
santa elf_gp rein_gp = do
    putStr "----------\n"
    choose [(awaitGroup rein_gp, run "deliver toys"),
            (awaitGroup elf_gp, run "meet in my study")]
  where
    run :: String -> (Gate,Gate) -> IO ()
    run task (in_gate,out_gate) = do
        putStr ("Ho! Ho! Ho! let’s " ++ task ++ "\n")
        operateGate in_gate
        operateGate out_gate
```

I emulate the same by using Go's channels:

```Go
// My Shitty Go code :/

// for making the transactions for reindeers and elves
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

// For running the transactions
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
```

* I used a failing transaction pattern to emulate S. P. Jones' `check` function. This
  function blocks until the condition is satisfied:

```Go
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
```

### Caveats

* Verbosity

* Unable to emulate the IO Monadic behavior of Haskell

* Random parks of the goroutines and freeze ups when Santa is run more than 2-3 times.

* Very sad with GoLang's `goroutines` - unable to readily get metadata for debugging like
  in case of Java is a big problem for me.

### Conclusion

I was able to generate the output like S. P. Jones' Haskell implementation, but, my code
is ugly and inelegant :(

[Output](./root.log):

```text
2017/12/02 15:15:52 Beginning Santa's workplace simulation ~~~
2017/12/02 15:15:52 Ho! Ho! Hoo! Let's go deliver some toys! :D
2017/12/02 15:15:52 -------------------------------------------
2017/12/02 15:15:52 Reindeer 6 delivering toys
2017/12/02 15:15:52 Reindeer 8 delivering toys
2017/12/02 15:15:52 Reindeer 7 delivering toys
2017/12/02 15:15:52 Reindeer 4 delivering toys
2017/12/02 15:15:52 Reindeer 1 delivering toys
2017/12/02 15:15:52 Reindeer 2 delivering toys
2017/12/02 15:15:52 Reindeer 3 delivering toys
2017/12/02 15:15:52 Reindeer 5 delivering toys
2017/12/02 15:15:52 Reindeer 9 delivering toys
2017/12/02 15:15:57 Ho! Ho! Hoo! Let's hold the meeting in the study! :D
2017/12/02 15:15:57 ----------------------------------------------------
2017/12/02 15:15:57 Elf 3 meeting in the study
2017/12/02 15:15:57 Elf 1 meeting in the study
2017/12/02 15:15:57 Elf 2 meeting in the study
2017/12/02 15:16:02 Ho! Ho! Hoo! Let's go deliver some toys! :D
2017/12/02 15:16:02 -------------------------------------------
2017/12/02 15:16:02 Reindeer 8 delivering toys
2017/12/02 15:16:02 Reindeer 4 delivering toys
2017/12/02 15:16:02 Reindeer 3 delivering toys
2017/12/02 15:16:02 Reindeer 2 delivering toys
2017/12/02 15:16:02 Reindeer 6 delivering toys
2017/12/02 15:16:02 Reindeer 1 delivering toys
2017/12/02 15:16:02 Reindeer 9 delivering toys
2017/12/02 15:16:02 Reindeer 5 delivering toys
2017/12/02 15:16:02 Reindeer 7 delivering toys
2017/12/02 15:16:07 Ho! Ho! Hoo! Let's hold the meeting in the study! :D
2017/12/02 15:16:07 ----------------------------------------------------
2017/12/02 15:16:07 Elf 4 meeting in the study
2017/12/02 15:16:07 Elf 6 meeting in the study
2017/12/02 15:16:07 Elf 5 meeting in the study
2017/12/02 15:16:12 Ho! Ho! Hoo! Let's go deliver some toys! :D
2017/12/02 15:16:12 -------------------------------------------
2017/12/02 15:16:12 Reindeer 3 delivering toys
2017/12/02 15:16:12 Reindeer 8 delivering toys
2017/12/02 15:16:12 Reindeer 9 delivering toys
2017/12/02 15:16:12 Reindeer 6 delivering toys
2017/12/02 15:16:12 Reindeer 4 delivering toys
2017/12/02 15:16:12 Reindeer 5 delivering toys
2017/12/02 15:16:12 Reindeer 7 delivering toys
2017/12/02 15:16:12 Reindeer 1 delivering toys
2017/12/02 15:16:12 Reindeer 2 delivering toys
2017/12/02 15:16:17 Ho! Ho! Hoo! Let's hold the meeting in the study! :D
2017/12/02 15:16:17 ----------------------------------------------------
2017/12/02 15:16:17 Elf 3 meeting in the study
2017/12/02 15:16:17 Elf 1 meeting in the study
2017/12/02 15:16:17 Elf 9 meeting in the study
2017/12/02 15:16:22 Santa is done for the day. Bye Santa ~~~
```

### Changelog v0.0.2

* Switched versions of the underlying STM

### References

[[1]](https://www.schoolofhaskell.com/school/advanced-haskell/beautiful-concurrency/4-the-santa-claus-problem)
S. P. Jones, "Beautiful concurrency", ch. 4, [Online] Available.
https://www.schoolofhaskell.com/school/advanced-haskell/beautiful-concurrency/4-the-santa-claus-problem
