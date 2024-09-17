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
			rmodel := relaymodel{Number: tint}
			switch wordsii[1] {
			case "ON":
				rmodel.Status = true
			default:
				continue
			}
			s.Relay = append(s.Relay, rmodel)

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
	Relay     string     `json:"rele"`
	Owi_temp [][]string `json:"owi_temp"`
}

func getinfolau5(resBody []byte, s *sensorconfig) {
	res := stlaurent{}
	json.Unmarshal(resBody, &res)
	for _, temp_info := range res.Owi_temp {
		if len(temp_info) != 4 || temp_info[3] == "" {
			continue
		}
		tname := string(charset.Cp1251BytesToRunes([]byte(temp_info[2])))
		tvalue, err := strconv.ParseFloat(temp_info[3], 32)
		if err != nil {
			continue
		}
		tmodel := tempmodel{Name: tname, Mac: temp_info[1], Value: float32(tvalue)}
		s.Temperature = append(s.Temperature, tmodel)
	}
	for i, r := range res.Relay {
		if i >= s.Relaylimit {
			break
		}
		rmodel := relaymodel{Number: i}
		if r != 48 {
			rmodel.Status = true
		}
		s.Relay = append(s.Relay, rmodel)
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
