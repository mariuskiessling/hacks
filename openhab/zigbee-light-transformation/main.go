package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mariuskiessling/hacks/openhab"
)

type Event struct {
	State      string `json:"state"`
	Brightness int    `json:"brightness"`
}

func main() {
	apiBaseURL := flag.String("url", "http://127.0.0.1:8080/rest", "The base URL of the OpenHab API.")
	flag.Parse()

	c := &openhab.Client{
		HTTPClient: http.Client{
			Timeout: 5 * time.Second,
		},
		APIBaseURL: *apiBaseURL,
	}
	if len(flag.Args()) < 1 {
		help("Missing subcommand", 1)
	}

	switch flag.Args()[0] {
	case "power":
		if len(flag.Args()) < 2 {
			help("Missing argument. Possible value 'out' for JSON output or 'in' for JSON ingestion.", 1)
		}

		switch flag.Args()[1] {
		case "in":
			if len(flag.Args()) < 3 {
				help(`Missing argument. Possible value is a JSON payload of this structure: { "state": "ON/OFF", "brightness": 0-255 }`, 1)
			}
			PowerIn(strings.Join(flag.Args()[2:], " "))

		case "out":
			if len(flag.Args()) != 3 {
				help("Missing argument. Possible values ON/1 or OFF/0.", 1)
			}
			PowerOut(flag.Args()[2], c)

		default:
			help("Invalid argument. Possible value 'out' for JSON output or 'in' for JSON ingestion.", 1)
		}

	case "brightness":
		if len(flag.Args()) < 2 {
			help("Missing argument. Possible value 'out' for JSON output or 'in' for JSON ingestion.", 1)
		}
		switch flag.Args()[1] {
		case "in":
			if len(flag.Args()) < 3 {
				help(`Missing argument. Possible value is a JSON payload of this structure: { "state": "ON/OFF", "brightness": 0-255 }`, 1)
			}
			BrightnessIn(strings.Join(flag.Args()[2:], " "))

		case "out":
			if len(flag.Args()) != 3 {
				help("Missing argument. Possible values are any integer between 1-255.", 1)
			}
			level, err := strconv.Atoi(flag.Args()[2])
			if err != nil {
				fmt.Println(err)
				help("Invalid argument. Brightness level not a valid number.", 1)
			}
			BrightnessOut(int(level))

		default:
			help("Invalid argument. Possible value 'out' for JSON output or 'in' for JSON ingestion.", 1)
		}

	default:
		help(fmt.Sprintf("Unknown command '%v'", flag.Args()[0]), 1)
	}
}

func PowerOut(cmd string, openHabClient *openhab.Client) {
	i, err := openHabClient.GetItem("Innr_RF264_1_Brightness")
	if err != nil {
		panic(err)
	}

	// Overwrite any NULL (unknown by OpenHAB) item state with 0 (off) to satisfy
	// the int casting.
	if i.State == "NULL" {
		i.State = "0"
	}

	rel, err := strconv.Atoi(i.State)
	if err != nil {
		panic(err)
	}

	abs := BrightnessRelativeToAbsolute(rel)

	if cmd == "ON" || cmd == "1" {
		fmt.Printf(`{ "state": "ON", "brightness": %v }`, abs)
	} else {
		fmt.Printf(`{ "state": "OFF", "brightness": %v }`, abs)
	}
}

func BrightnessIn(payload string) {
	e := decodeEvent(payload)

	level := BrightnessAbsoluteToRelative(e.Brightness)
	fmt.Println(level)
}

func PowerIn(payload string) {
	e := decodeEvent(payload)
	fmt.Println(e.State)
}

func decodeEvent(payload string) Event {
	e := &Event{}

	payload = strings.ReplaceAll(payload, "'", "")

	err := json.Unmarshal([]byte(payload), e)
	if err != nil {
		fmt.Println(payload)
		help("Invalid argument. JSON payload has no valid structure.", 1)
	}

	return *e
}

func BrightnessOut(level int) {
	level = BrightnessRelativeToAbsolute(level)

	if level <= 5 {
		fmt.Printf(`{ "state": "OFF", "brightness": %v }`, level)
	} else {
		fmt.Printf(`{ "state": "ON", "brightness": %v }`, level)
	}
}

func BrightnessRelativeToAbsolute(rel int) int {
	return int(float64(rel) / 100.0 * 255.0)
}

func BrightnessAbsoluteToRelative(abs int) int {
	return int(float64(abs) / 255.0 * 100.0)
}

func help(err string, exitCode int) {
	fmt.Printf(`Error: %v 

Possible commands:
  power
  brightness
`, err)

	os.Exit(exitCode)
}
