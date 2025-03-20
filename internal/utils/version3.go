package utils

import (
	"bytes"
	"fmt"
	"go_1brc/internal/models"
	"io"
	"log"
	"os"
	"sort"
)

func MeasureVersion3(inputfile string, output io.Writer) error {
	stationStats := make(map[string]*models.StationStatsv3)
	// Read the input file
	file, err := os.Open(inputfile)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, 4*1024*1024)
	readStart := 0

	for {
		n, err := file.Read(buf[readStart:])
		if err != nil && err != io.EOF {
			return err
		}
		if readStart+n == 0 {
			break
		}
		chunk := buf[:readStart+n]

		newline := bytes.LastIndexByte(chunk, '\n')
		if newline < 0 {
			break
		}
		remaining := chunk[newline+1:]
		chunk = chunk[:newline+1]
		for {
			newline := bytes.IndexByte(chunk, '\n')
			if newline < 0 {
				break
			}
			// line includes the '\n'
			line := chunk[:newline+1]
			chunk = chunk[newline+1:]

			end := len(line)
			tenths := int32(line[end-1] - '0')
			ones := int32(line[end-3] - '0') // line[end-2] is '.'
			var temp int32
			var semicolon int
			if line[end-4] == ';' {
				temp = ones*10 + tenths
				semicolon = end - 4
			} else if line[end-4] == '-' {
				temp = -(ones*10 + tenths)
				semicolon = end - 5
			} else {
				tens := int32(line[end-4] - '0')
				if line[end-5] == ';' {
					temp = tens*100 + ones*10 + tenths
					semicolon = end - 5
				} else { // '-'
					temp = -(tens*100 + ones*10 + tenths)
					semicolon = end - 6
				}
			}
			station := line[:semicolon]
			s := stationStats[string(station)]
			if s == nil {
				stationStats[string(station)] = &models.StationStatsv3{
					Min:   temp,
					Max:   temp,
					Sum:   int64(temp),
					Count: 1,
				}
			} else {
				s.Min = min(s.Min, temp)
				s.Max = max(s.Max, temp)
				s.Sum += int64(temp)
				s.Count++
			}
		}
		readStart = copy(buf, remaining)
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
		mean := float64(s.Sum) / float64(s.Count) / 10
		fmt.Fprintf(output, "%s=%.1f/%.1f/%.1f", station, float64(s.Min)/10, mean, float64(s.Max)/10)
	}
	fmt.Fprint(output, "}\n")
	log.Println("Stations count:", len(stations))
	return nil
}
