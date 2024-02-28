### RSS parser

Program that counts the frequency a topic is found in a collection of
RSS news feed documents.

### Single thread application

Program uses a single threaded algorithm
named freq that iterates over the collection, processing each document one at a time,
and returns the number of times the target is found.

### Run the trace server

- ```go build main.go```
- ```time ./main > t.out```
- ```go tool trace t.out```

### Check race detector

- ```go build --race main.go```
- ```./main > t.out```

### Performance:

| Solution      | Runtime   | Top heap | GC Occurrences | GC Avg Wall Duration | GC Wall Duration |  
|---------------|-----------|----------|----------------|----------------------|------------------|
| Single thread | 1326ms    | 6.6MB    | 387            | 299ns                | 116ms            | 
| Fan-Out       | 325.738ms | 59.8MB   | 63             | 2.5ms                | 157ms            |
| Pool          | 474.902ms | 8.6MB    | 396            | 795ns                | 314ms            |
| Pool 80Mb GC  | 228.026ms | 74.76MB  | 22             | 904ns                | 19.89ms          |

### Pool solution (performance vs memory usage)

- First, that the GC should start when 8.6MB of memory is in-use on the heap.
- Second, that the next GC should start when 100% more memory is allocated on the heap, based on the result of the
  marked live value from the previous GC

I can increase the GC memory usages and not start the first collection until the memory in-use reaches 40MB.
To do this, I need to change the GOGC value by an order of magnitude, which is 1000.

- ```time GOGC=1000 ./main > t.out``` or ```debug.SetGCPercent(1000)```

### GC mark assist

GC activates the GC mark assist to speed up the marking process.
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

















