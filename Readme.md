### Allocation

Any function that constructs a value, is the owner of that value.
If the value needs to exist after the owning function
returns, then the value must be constructed in the heap.

If the function can be inlined, the ownership of the construction moves up to the calling function.

```
func fn() {
  // Before inlining
  input := bytes.NewReader(data) // <- Original Call
   
  // After inlining
  input := &bytes.Reader{ buf: data } // <- After Inlining Optimization
}
```

### GC mark assist

GC activates the GC ```mark assist``` to speed up the marking process.
GC mark assist in Go helps to maintain the performance and responsiveness of programs by dynamically adjusting the
resources allocated to the garbage collector during the mark phase to keep up with the pace of object allocation

Could be the situation when all the Goroutines in "mark assist" to help slowdown allocations and get the initial GC
finished.

### Performance notes

- If you notice a lot of time spent in the ```runtime.mallocgc``` function, it suggests that the program may be making
  too
  many small memory allocations

- If you're spending a significant amount of time managing channel operations, ```sync.Mutex``` code, or other
  synchronization
  elements in your program, it's likely facing contention issues. To improve performance, think about restructuring the
  program to reduce the frequent access of shared resources. Common techniques for this include sharding/partitioning,
  local buffering/batching and copy-on-write technique

- If your program spends a significant amount of time in ```syscall.Read/Write```, it might be doing too many small
  reads and writes. Using bufio wrappers around os.File or net.Conn can be helpful in this situation

- If your program is spending a lot of time in the ```GC (Garbage Collection)``` component, it could be because it's
  either creating too many temporary objects or because the heap size is too small, leading to frequent garbage
  collections
    - Large objects impact memory usage and GC pacing, whereas numerous small allocations affect marking speed.

    - Combine values into larger ones to reduce memory allocations and alleviate pressure on the garbage collector,
      resulting in faster garbage collections

    - Values without pointers aren't scanned by the garbage collector. Eliminating pointers from actively used values
      can enhance garbage collection efficiency

### Run Bench

```
generate pprof and binary
// go test -bench . -benchmem -memprofile p.out -gcflags -m=2
// go test -bench . -benchtime 3s -benchmem -memprofile p.out
// go test -bench . -benchtime 3s -benchmem -cpuprofile p.out -gcflags -m=2

run tools
// go tool pprof -noinlines p.out
// go tool pprof --noinlines  memcpu.test p.out


Profiling commands
// go tool pprof -http :8080 stream.test p.out
// press ```o```
// list <func_name>
//weblist <func_name>

```

### Inlining

Inlining is an optimization technique that eliminates function calls by replacing
them with a duplicate copy of the code contained within the function being called.






