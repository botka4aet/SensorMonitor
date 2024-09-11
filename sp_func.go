package main

import (
	"fmt"
	"strconv"
	"strings"
)

func getinfors04(resBody string, s *sensorconfig) {
	wordsi := strings.Split(string(resBody[:]), ";")

	for i, wordi := range wordsi {
		wordsii := strings.Split(wordi, ":")
		//skip, if not parameter : value
		if len(wordsii) != 2 {
			continue
		}
		wordii := wordsii[0]
		if len([]rune(wordii)) > 3 {
			wordii = string(wordii)[0:4]
		}
		switch wordii {
		case "gpio":
			tstring := strings.TrimPrefix(wordsii[0], "gpio")
			tint, err := strconv.Atoi(tstring)
			if err != nil {
				continue
			}
			switch wordsii[1] {
			case "ON":
				s.Relay[tint-1] = true
			case "OFF":
				s.Relay[tint-1] = false
			default:
				continue
			}

		case "dws":
			tint, err := strconv.Atoi(strings.ReplaceAll(wordsii[1], ".", ""))
			if err != nil || tint == 999 {
				continue
			}
			s.Temperature = []tempmodel{tempmodel{Value: float32(tint) / 10}}
		default:
			continue
		}
		fmt.Printf("Word %d is: %s\n", i, wordi)
	}
}

func getinfofromres(resBody string, s *sensorconfig) {
	switch s.Model {
	case "rs-04":
		getinfors04(resBody, s)
	case "rs-38":
		getinfors04(resBody, s)
	case "rs-044":
		getinfors04(resBody, s)
	case "laurent-5":
	default:
		return
	}

}
