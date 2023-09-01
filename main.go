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
	// Set the linux device class path for the backlight
	PATH = "/sys/class/backlight/"
	// Set the lowest accepted backlight level for when decreasing incrementally
	BOTTOM_THRESHOLD = 1 // percent
)

// this struct contains the path to the backlight device
type light struct {
	Path string
}

func (l *light) findPath() error {
	// findPath() will find the backlight device in the linux backlight device PATH
	//     and ensure that it contains the file 'brightness'
	files, err := os.ReadDir(PATH)
	if err != nil {
		return err
	}

	// iterate over each file in PATH
	for _, file := range files {
		brightnessPath := filepath.Join(PATH, file.Name(), "brightness")
		dat, err := os.ReadFile(brightnessPath)
		if err != nil {
			continue // if the file 'brightness' cannot be read, go to next iteration
		}
		level, err := strconv.Atoi(strings.TrimSpace(string(dat)))
		if err != nil {
			continue // if the contents of file 'brightness' cannot be parsed to int, go to next iteration
		}
		if level > 0 { // if the contents of file 'brightness' is a number > 0, then set the l.Path and return
			l.Path = brightnessPath
			return nil
		}
	}
	// If file 'brightness' with an integer value was not found, return error
	return fmt.Errorf("could not find path")
}

func (l *light) getBrightness() (int, error) {
	dat, err := os.ReadFile(l.Path)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(dat))
	bright, err := strconv.Atoi(s)
	return bright, err
}

func percentToLevel(percentage int) int {
	level := percentage * 120000 / 100
	return level
}

func levelToPercent(level int) int {
	return level * 100 / 120000
}

func (l *light) setBrightLevel(level int) error {
	levelS := fmt.Sprintf("%d", level)
	err := os.WriteFile(l.Path, []byte(levelS), 0644)
	return err
}

func (l *light) incBrightness(percentage int) error {
	brightness, err := l.getBrightness()
	if err != nil {
		return err
	}
	brightnessp := levelToPercent(brightness)
	target := brightnessp + percentage
	err = l.fade(target, "125ms")
	return err
}

func (l *light) decBrightness(percentage int) error {
	brightness, err := l.getBrightness()
	if err != nil {
		return err
	}
	brightnessp := levelToPercent(brightness)
	target := brightnessp - percentage
	if target < BOTTOM_THRESHOLD {
		target = BOTTOM_THRESHOLD
	}
	err = l.fade(target, "125ms")
	return err
}

func (l *light) fade(target int, duration string) error {
	// target will be accepted in percent
	fadeTime, err := time.ParseDuration(duration)
	if err != nil {
		return err
	}
	// We'll assume 60 fps
	targetLevel := percentToLevel(target)
	currentLevel, err := l.getBrightness()
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
		l.setBrightLevel(currentLevel)
		time.Sleep(stepinterval)
	}
	return nil
}

func (l *light) pulse(amp int) error {
	currentLevel, err := l.getBrightness()
	if err != nil {
		return err
	}
	currentPercent := levelToPercent(currentLevel)
	for {
		l.fade(currentPercent+amp, "75ms")
		l.fade(currentPercent-amp, "75ms")
		time.Sleep(time.Millisecond * 200)
	}
}

func main() {
	var fadeTime, targetBrightness string // declare the variables used for flags
	var l light                           // declare an instance of light struct
	err := l.findPath()                   // find the backlight device path under /sys/class/backlight
	if err != nil {
		log.Fatal(err) // if we cannot find the backlight device, exit the program
	}

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
				Destination: &targetBrightness,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if targetBrightness == "" {
				fmt.Fprintf(os.Stdout, "Please set a valid target brightness.\n")
				return nil
			}
			level, err := strconv.Atoi(targetBrightness)
			if err != nil {
				err = fmt.Errorf("invalid target: %w", err)
				return err
			}
			err = l.fade(level, fadeTime)
			return err
		},
		Commands: []*cli.Command{
			{
				Name:    "low",
				Aliases: []string{"lo"},
				Usage:   "Set brightness to low",
				Action: func(cCtx *cli.Context) error {
					err := l.fade(5, fadeTime)
					return err
				},
			},
			{
				Name:    "mid",
				Aliases: []string{"medium"},
				Usage:   "Set brightness to mid",
				Action: func(cCtx *cli.Context) error {
					err := l.fade(50, fadeTime)
					return err
				},
			},
			{
				Name:    "high",
				Aliases: []string{"hi", "max"},
				Usage:   "Set brightness to max",
				Action: func(cCtx *cli.Context) error {
					err := l.fade(100, fadeTime)
					return err
				},
			},
			{
				Name:    "dec",
				Aliases: []string{"-"},
				Usage:   "Decrease screen brightness",
				Action: func(cCtx *cli.Context) error {
					err := l.decBrightness(5)
					return err
				},
			},
			{
				Name:    "inc",
				Aliases: []string{"+"},
				Usage:   "Increase screen brightness",
				Action: func(cCtx *cli.Context) error {
					err := l.incBrightness(5)

					return err
				},
			},
			{
				Name:  "pulse",
				Usage: "Pulse Effect",
				Action: func(cCtx *cli.Context) error {
					err := l.pulse(25)
					return err
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
