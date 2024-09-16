package main

import (
	"fmt"
	"github.com/gelsrc/go-charset"
	"github.com/goccy/go-json"
	"strconv"
	"strings"
)

func getinfors04(resBody []byte, s *sensorconfig) {
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
			s.Temperature = []tempmodel{{Value: float32(tint) / 10}}
		default:
			continue
		}
		fmt.Printf("Word %d is: %s\n", i, wordi)
	}
}

type stlaurent struct {
	Owi_temp [][]string `json:"owi_temp"`
}

func getinfolau5(resBody []byte, s *sensorconfig) {
	res := stlaurent{}
	json.Unmarshal(resBody, &res)
	for _, wordi := range res.Owi_temp {
		if len(wordi) != 4 || wordi[3] == "" {
			continue
		}
		tname := string(charset.Cp1251BytesToRunes([]byte(wordi[2])))
		tvalue, err := strconv.ParseFloat(wordi[3], 32)
		if err != nil {
			continue
		}
		tmodel := tempmodel{Name: tname, Mac: wordi[1], Value: float32(tvalue)}
		s.Temperature = append(s.Temperature, tmodel)
	}
}

func getinfofromres(resBody []byte, s *sensorconfig) {
	switch s.Model {
	case "rs-04", "rs-38", "rs-044":
		getinfors04(resBody, s)
	case "laurent-5":
		getinfolau5(resBody, s)
	default:
		return
	}

}
