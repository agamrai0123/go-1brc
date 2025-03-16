package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"go_1brc/internal/models"
	"io"
	"log"
	"os"
	"sort"
)

func MeasureVersion2(inputfile string, output io.Writer) error {
	stationStats := make(map[string]*models.StationStats)
	// Read the input file
	file, err := os.Open(inputfile)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		station, tempBytes, hasColon := bytes.Cut(line, []byte(";"))
		if !hasColon {
			continue
		}

		negative := false
		index := 0
		if tempBytes[index] == '-' {
			index++
			negative = true
		}
		temp := float64(tempBytes[index] - '0')
		index++
		if tempBytes[index] != '.' {
			temp = temp*10 + float64(tempBytes[index]-'0')
			index++
		}
		index++ // skip '.'
		temp += float64(tempBytes[index]-'0') / 10
		if negative {
			temp = -temp
		}

		s := stationStats[string(station)]
		if s == nil {
			stationStats[string(station)] = &models.StationStats{
				Min:   temp,
				Max:   temp,
				Sum:   temp,
				Count: 1,
			}
		} else {
			s.Min = min(s.Min, temp)
			s.Max = max(s.Max, temp)
			s.Sum += temp
			s.Count++
		}
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
	log.Println("Stations count:", len(stations))
	return nil
}
