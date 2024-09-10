package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type defaultmodel struct {
	Name       string
	Login      string
	Password   string
	Relaylimit int
	Infolink   string
}

var defaultmodele defaultmodel

type sensorconfig struct {
	Address     string
	Model       string
	Login       string
	Password    string
	Relaylimit  int
	Relay       [10]bool
	Temperature int
	Infolink    string
	IgnoreEmpty bool
	Err         int
}

func (s *sensorconfig) update() {
	linfo := s.Address + s.Infolink
	linfo = strings.ReplaceAll(linfo, "_login_", s.Login)
	linfo = strings.ReplaceAll(linfo, "_password_", s.Password)
	req, err := http.NewRequest(http.MethodGet, linfo, nil)
	if err != nil {
		log.Printf("client: could not create request: %s\n", err)
		s.Err++
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("client: error making http request: %s\n", err)
		s.Err++
	}
	defer res.Body.Close()

	rescode := res.StatusCode
	if rescode != 200 {
		log.Print("client: HTTP response status codes not 200\n")
		s.Err++
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("client: could not read response body: %s\n", err)
		s.Err++
	}
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
			tint, _ := strconv.Atoi(strings.ReplaceAll(wordsii[1], ".", ""))
			if err != nil || tint == 999 {
				continue
			}
			s.Temperature = tint
		default:
			continue
		}
		fmt.Printf("Word %d is: %s\n", i, wordi)
	}
}

func main() {
	defaultmodels := make(map[string]defaultmodel)
	var err error
	//Read config for sensor models
	err = filepath.Walk("./sensormodel/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			jsonFile, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer jsonFile.Close()

			byteValue, _ := io.ReadAll(jsonFile)
			json.Unmarshal(byteValue, &defaultmodels)

		}
		fmt.Printf("dir: %v: name: %s\n", info.IsDir(), path)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	sensorconfigs := make(map[string]sensorconfig)
	//Read info for used sensors
	err = filepath.Walk("./sensorconfig/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			jsonFile, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer jsonFile.Close()

			byteValue, _ := io.ReadAll(jsonFile)
			json.Unmarshal(byteValue, &sensorconfigs)

		}
		fmt.Printf("dir: %v: name: %s\n", info.IsDir(), path)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	//Fill missed info
	for key, sensor := range sensorconfigs {
		//check link for geting info
		if sensor.Infolink == "" {
			if sensor.Model == "" || defaultmodels[sensor.Model] == defaultmodele || defaultmodels[sensor.Model].Infolink == "" {
				delete(sensorconfigs, key)
				continue
			}
		}
		//Check - if can skip or no default config
		if sensor.IgnoreEmpty || sensor.Model == "" || defaultmodels[sensor.Model] == defaultmodele {
			continue
		}
		Defaultc := reflect.ValueOf(defaultmodels[sensor.Model])
		v := reflect.ValueOf(&sensor).Elem()
		typeOfS := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			value := Defaultc.FieldByName(typeOfS.Field(i).Name)
			//Continue if no default field
			if !field.IsZero() || !field.CanSet() || !value.IsValid() || value.IsZero() {
				continue
			}
			field.Set(value)
			fmt.Printf("[%v]: Setting -%v- for -%s-\n", key, field, typeOfS.Field(i).Name)

		}
		sensorconfigs[key] = sensor
	}

	time.Sleep(1 * time.Second)

	for key, sensor := range sensorconfigs {
		sensor.update()
		sensorconfigs[key] = sensor
	}
}
