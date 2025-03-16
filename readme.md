# The One Billion Row Challenge in Go
I am following the article by [Ben Hoyt](https://benhoyt.com/writings/go-one-billion-rows/) for the **The One Billion Row Challenge** in Go.

## 🚀 Quick Points to Remember
- The `measurements.txt` file contains **1 Billion rows** (~13 GB), so **do not add it to GitHub**.  
  Add it to `.gitignore
- use the command `./go-1brc -cpuprofile cpu.prof` to create a pprof CPU profile.
- use the command `go tool pprof -http=:8080 cpu.prof` to run the pprof profile.
- [Graphviz](https://graphviz.org/download/) is required for generating call graphs in pprof.

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
- [pprof]()

##### Observation:
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
