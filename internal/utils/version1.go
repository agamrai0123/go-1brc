package utils

import (
	"bufio"
	"fmt"
	"go_1brc/internal/models"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

func MeasureVersion1(inputfile string, output io.Writer) error {
	stationStats := make(map[string]models.StationStats)
	// Read the input file
	file, err := os.Open(inputfile)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		station, tempStr, hasColon := strings.Cut(line, ";")
		if !hasColon {
			continue
		}
		temp, err := strconv.ParseFloat(tempStr, 64)
		if err != nil {
			return err
		}
		s, ok := stationStats[station]
		if !ok {
			s.Min = temp
			s.Max = temp
			s.Sum = temp
			s.Count = 1
		} else {
			if temp < s.Min {
				s.Min = temp
			}
			if temp > s.Max {
				s.Max = temp
			}
			s.Sum += temp
			s.Count++
		}
		stationStats[station] = s
	}

	stations := make([]string, 0, len(stationStats))
	for station := range stationStats {
		stations = append(stations, station)
	}
	sort.Strings(stations)
	fmt.Fprint(output, "{")
	for i, station := range stations {
		if i > 0 {
			fmt.Fprint(output, ", ")
		}
		s := stationStats[station]
		mean := s.Sum / float64(s.Count)
		fmt.Fprintf(output, "%s=%.1f/%.1f/%.1f", station, s.Min, mean, s.Max)
	}
	fmt.Fprint(output, "}\n")
	return nil
}
