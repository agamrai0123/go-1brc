# The One Billion Row Challenge in Go
I am following the article by [Ben Hoyt](https://benhoyt.com/writings/go-one-billion-rows/) for the **The One Billion Row Challenge** in Go.

## 🚀 Quick Points to Remember
- The `measurements.txt` file contains **1 Billion rows** (~13 GB), so **do not add it to GitHub**.  
  Add it to `.gitignore`
- use the command `./go-1brc -cpuprofile cpu.prof` to create a pprof CPU profile.
- use the command `go tool pprof -http=:8080 cpu.prof` to run the pprof profile.
- [Graphviz](https://graphviz.org/download/) is required for generating call graphs in pprof.
- In pprof, if a function takes X seconds (Y%) of A seconds (B%) then:
    - **Self-time**: The *X seconds* represents time spent inside function itself (excluding time in functions it calls).
    - **Cumulative time**: The *A seconds* includes *X seconds* self-time + time spent in any other functions that the function calls.

- ***How to optimise a function***:
    - If **self-time** is high, optimizing this function directly will have a big impact.
    - If **cumulative-time** is high, most of the cost is in its children, and optimizing those might be more effective.

- ***Max Goroutines recommendations**:

| Workload Type 							     | Suggested maxGoroutines  |
|------------------------------------------------|--------------------------|
| `CPU-Intensive (e.g., hash computation)`       | **runtime.NumCPU()**  	|
| `I/O-Bound (e.g., DB queries, HTTP requests)`  | **100-1000**             |
| `Mixed Workload`  							 | **NumCPU() * 10**        |
| `Memory-Constrained` 							 | **Test & adjust**        |

## 🚨 Why is this hard?
- It is **1 Billion Rows!!!!!!!!**
- You must:
    1. Read 1 billion lines efficiently.
    2. Store and process station temperature data.
    3. Write results to an output file.
### ❓ Key Questions to Answer
- **File Handling**: How would you read the file? As the stations need to be read line by line, how would you optimise 1 billion reads?
- **Data Processing**: As after reading you have to process mean of the station temperatures, which data structure would you use? 
- **Running Mean vs. Sum/Count**: Would you use running mean? If not why not?
- **Parallelization**: How would you parallelize the workload? As read and writes are blocking and sequential.

## 📊 Baseline Testing
- Just counting the words in the file:
```
$ time wc measurements.txt 
 1000000000  1179195629 13795508817 measurements.txt

real    1m25.350s
user    1m20.717s
sys     0m3.936s
```

- using `gawk` to Process the data:
```
time gawk -b -f 1brc.awk measurements.txt >measurements.out

real    12m45.927s
user    12m35.029s
sys     0m7.453s
```
## Solutions

### Solution 1: Simple Go
- [Solution](https://github.com/agamrai0123/go-1brc/blob/main/internal/utils/version1.go)
- Time taken: 126.89 seconds
- [pprof](https://github.com/agamrai0123/go-1brc/blob/main/pprof_graphs/version1.png)

##### Observations:
- Map operations taking almost 27% of the runtime. 16% for access/read and 11% for assign/write.
- bufio scan taking almost 20% of the runtime.
- bufio text, for reading line by line is taking 17.5% of the runtime.
- 20% of the runtime is taken up for parsing float from string.

##### Next Steps:
- As suggested by Ben Hoyt, Map read and writes can be reduced if we use `map[string]*stats` (pointer values) instead of a `map[string]stats`.
    - As we for each line, we’re hashing the string twice: once when we try to fetch the value from the map, and once when we update the map.
    - if we use pointer instead of the actual structure, we can reduce the hashing to only once.
    - Avoids struct copy; updates are made directly in memory.

- Use scanner.Bytes() to read bytes instead of scanner.Text()
    - returns a []byte that directly references the scanner's buffer, avoiding extra allocations.
    - reuses the internal buffer, the garbage collector has fewer temporary string objects to clean up.

| Method             | Time Taken (1M lines) | Memory Allocations         |
|--------------------|-----------------------|----------------------------|
| `scanner.Text()`   | **1.23s**             | **High (many allocations)**|
| `scanner.Bytes()`  | **0.85s**             | **Low (buffer reuse)**     |


- Avoid using strconv.ParseFloat()

```
	negative := false                   // Flag to check if the number is negative
	index := 0
	if tempBytes[index] == '-' {        // if the first byte read is negative
		index++                         // Move to next index   
		negative = true
	}
	temp := float64(tempBytes[index] - '0')     // This converts the current digit ('0' to '9') to an integer by subtracting ASCII '0'
	index++
	if tempBytes[index] != '.' {
		temp = temp*10 + float64(tempBytes[index]-'0')
		index++
	}
	index++ // skip '.'                         // skipping the decimal(.) and then dividing by 10 as valuse are upto single decimal place
	temp += float64(tempBytes[index]-'0') / 10
	if negative {                               // Assigning the negative sign to the number
		temp = -temp
	}
```
### Solution 2:
- [Solution](https://github.com/agamrai0123/go-1brc/blob/main/internal/utils/version2.go)
- Time taken: 49.62 seconds
- [pprof](https://github.com/agamrai0123/go-1brc/blob/main/pprof_graphs/version2.png)

##### Observations:
- Map Access takes up 22.60% of the runtime.
- `bytes.Cut` takes up 11.76% of the runtime.
- bufio Scan takes up 36.16% of the runtime.

##### Next Steps:
- Avoid bufio.Scanner
	- Read bytes one by one to locate the newline.
	- Process the same bytes again to read and process the line.
	- This results in double processing of bytes, which is inefficient.
- Using Fixed Point Integers instead of float
	- Using integers to represent the temperature (in tenths) is typically better than using floats because it’s more efficient and avoids potential pitfalls with floating-point precision.
- Avoid bytes.Cut
	- the *line* is in the format `<station>;<temperature>`
	- `<station>` is a variable-length string representing a station identifier.
	- `<temperature>` is a floating-point value formatted as either:
		- N.N
		- NN.N
		- -N.N
		- -NN.N
```
end := len(line)
tenths := int32(line[end-1] - '0')
ones := int32(line[end-3] - '0') // line[end-2] is '.'
var temp int32
var semicolon int
if line[end-4] == ';' {          // positive N.N temperature
    temp = ones*10 + tenths
    semicolon = end - 4
} else if line[end-4] == '-' {   // negative -N.N temperature
    temp = -(ones*10 + tenths)
    semicolon = end - 5
} else {
    tens := int32(line[end-4] - '0')
    if line[end-5] == ';' {      // positive NN.N temperature
        temp = tens*100 + ones*10 + tenths
        semicolon = end - 5
    } else {                     // negative -NN.N temperature
        temp = -(tens*100 + ones*10 + tenths)
        semicolon = end - 6
    }
}
station := line[:semicolon]
```

###### How buffer Size depends on read time:
- **Buffer Size & Read Time**
	- **Small Buffer (< 4KB)**
		- Too many system calls → slow performance.

	- **Medium Buffer (4KB - 64KB)**
		- Generally optimal for performance. OS likes block sizes of 4KB-16KB.

	- **Large Buffer (> 1MB)**
		- Can reduce system calls but may consume too much memory without a major speed gain.

#### Lets try some Benchmarking:
- [Benchmarking](https://github.com/agamrai0123/go-1brc/blob/main/internal/tests/buffer_test.go) File Reading using `bufio.Reader` and `file.Read()` for different buffer sizes
	- Keep in mind that this is a large file, so it might be better to try with 1 Million Rows for the tests

- [Test Results](https://github.com/agamrai0123/go-1brc/blob/main/internal/tests/buffer_test.xlsx)

#### Inferences
- **Small Buffer Sizes (1KB–8KB)**: Performance varies. Direct is faster at 1KB (24.616s vs 25.154s), Bufio is faster at 2KB (18.827s vs 20.911s), and Direct outperforms Bufio significantly at 8KB (19.648s vs 24.692s).
- **Medium to Large Sizes (16KB–64MB)**: Times decrease as buffer size increases, reaching a minimum at 4MB (Direct: 10.918s, Bufio: 10.810s), then slightly increase or stabilize. Bufio is marginally faster at 4MB.
- **Overall Trend**: Larger buffers (e.g., 4MB) reduce read time significantly compared to 1KB (by over 50%), but beyond 4MB, gains diminish, and times rise slightly, possibly due to memory overhead.


### Solution 3:
- [Solution](https://github.com/agamrai0123/go-1brc/blob/main/internal/utils/version3.go)
- Time taken: 28.33 seconds
- [pprof](https://github.com/agamrai0123/go-1brc/blob/main/pprof_graphs/version3.png)

##### Observations:
- Map Access dominates the runtime and takes up 59.91% of the runtime.

##### Next Steps:
- Custom Map Access
- Try to parallelise