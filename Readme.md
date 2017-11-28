# The Santa Claus Problem implementation using my reworked Go STM

## Author: Sidharth Mishra

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

My implementation is a clone of S.P. Jone's Haskell implementation, but, using my STM
library. There will be modifications and I'll note those modifications in the source code
and here.

## References

[[1]](https://www.schoolofhaskell.com/school/advanced-haskell/beautiful-concurrency/4-the-santa-claus-problem)
S. P. Jones, "Beautiful concurrency", ch. 4, [Online] Available.
https://www.schoolofhaskell.com/school/advanced-haskell/beautiful-concurrency/4-the-santa-claus-problem
