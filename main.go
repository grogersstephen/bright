package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

const (
	PATH = "/sys/class/backlight/intel_backlight/"
)

func getIntFromFile(path string) (int, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(dat))
	i, err := strconv.Atoi(s)
	return i, err
}

func getBrightness() (int, error) {
	bright, err := getIntFromFile(filepath.Join(PATH, "brightness"))
	return bright, err
}

func percentToLevel(percentage int) int {
	level := percentage * 120000 / 100
	return level
}

func levelToPercent(level int) int {
	return level * 100 / 120000
}

func setBrightness(percentage int) error {
	err := fade(percentage, "500ms")
	return err
}

func setBrightLevel(level int) error {
	path := filepath.Join(PATH, "brightness")
	levelS := fmt.Sprintf("%d", level)
	err := os.WriteFile(path, []byte(levelS), 0644)
	return err
}

func incBrightness(percentage int) error {
	brightness, err := getBrightness()
	if err != nil {
		return err
	}
	brightnessp := levelToPercent(brightness)
	err = setBrightness(brightnessp + percentage)
	return err
}

func decBrightness(percentage int) error {
	brightness, err := getBrightness()
	if err != nil {
		return err
	}
	brightnessp := levelToPercent(brightness)
	err = setBrightness(brightnessp - percentage)
	return err
}

func fade(target int, duration string) error {
	// target will be accepted in percent
	fadeTime, err := time.ParseDuration(duration)
	if err != nil {
		return err
	}
	// We'll assume 60 fps
	targetLevel := percentToLevel(target)
	currentLevel, err := getBrightness()
	if err != nil {
		return err
	}
	difference := currentLevel - targetLevel
	stepinterval := time.Millisecond * 17 // This will yield ~ 60hz
	stepcount := int(fadeTime / stepinterval)
	step := difference / stepcount
	for difference > 100 || difference < -100 {
		currentLevel -= step
		difference = currentLevel - targetLevel
		setBrightLevel(currentLevel)
		time.Sleep(stepinterval)
	}
	return nil
}

func pulse(amp int) error {
	currentLevel, err := getBrightness()
	if err != nil {
		return err
	}
	currentPercent := levelToPercent(currentLevel)
	for {
		fade(currentPercent+amp, "75ms")
		fade(currentPercent-amp, "75ms")
		time.Sleep(time.Millisecond * 200)
	}
}

func main() {
	var fadeTime string
	var target string

	app := &cli.App{
		Name:  "bright",
		Usage: "Set the screen brightness",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "duration",
				Aliases:     []string{"d"},
				Usage:       "Set a fade duration",
				Value:       "500ms",
				Destination: &fadeTime,
			},
			&cli.StringFlag{
				Name:        "target",
				Aliases:     []string{"t"},
				Usage:       "Set target brightness level in percent",
				Value:       "50",
				Destination: &target,
			},
		},
		Action: func(cCtx *cli.Context) error {
			level, err := strconv.Atoi(target)
			if err != nil {
				err = fmt.Errorf("invalid target: %w", err)
				return err
			}
			err = fade(level, fadeTime)
			return err
		},
		Commands: []*cli.Command{
			{
				Name:  "low",
				Usage: "Set brightness to low",
				Action: func(cCtx *cli.Context) error {
					err := fade(5, fadeTime)
					return err
				},
			},
			{
				Name:  "mid",
				Usage: "Set brightness to mid",
				Action: func(cCtx *cli.Context) error {
					err := fade(50, fadeTime)
					return err
				},
			},
			{
				Name:  "max",
				Usage: "Set brightness to max",
				Action: func(cCtx *cli.Context) error {
					err := fade(100, fadeTime)
					return err
				},
			},
			{
				Name:    "dec",
				Aliases: []string{"-"},
				Usage:   "Decrease screen brightness",
				Action: func(cCtx *cli.Context) error {
					err := decBrightness(5)
					return err
				},
			},
			{
				Name:    "inc",
				Aliases: []string{"+"},
				Usage:   "Increase screen brightness",
				Action: func(cCtx *cli.Context) error {
					err := incBrightness(5)

					return err
				},
			},
			{
				Name:  "pulse",
				Usage: "Pulse Effect",
				Action: func(cCtx *cli.Context) error {
					err := pulse(25)
					return err
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
