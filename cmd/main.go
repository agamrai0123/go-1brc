package main

import (
	"flag"
	"fmt"
	"go_1brc/internal/utils"
	"os"
	"runtime/pprof"
)

func main() {
	var cpuProfile = flag.String("cpuprofile", "", "write CPU profile to file")
	flag.Parse()
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	inputfile := "../measurements.txt"
	outputfile := "stations.txt"
	output, err := os.Create(outputfile)
	if err != nil {
		panic(err)
	}
	defer output.Close()
	err = utils.MeasureVersion2(inputfile, output)
	if err != nil {
		panic(err)
	}
	fmt.Println("Measurements processed and written to", outputfile)
}
