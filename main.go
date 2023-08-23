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
	PATH = "/sys/class/backlight/"
)

type light struct {
	Class  string
	Vendor string
	Path   string
}

func (l *light) findPath() error {
	files, err := os.ReadDir(l.Class)
	if err != nil {
		return err
	}
	for _, file := range files {
		vendor := file.Name()
		vendorPath := filepath.Join(l.Class, file.Name())
		brightPath := filepath.Join(vendorPath, "brightness")
		dat, err := os.ReadFile(brightPath)
		if err != nil {
			continue
		}
		level, err := strconv.Atoi(strings.TrimSpace(string(dat)))
		if err != nil {
			continue
		}
		if level > 0 {
			l.Vendor = vendor
			l.Path = filepath.Join(l.Class, l.Vendor, "brightness")
			return nil
		}
	}
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
	err = l.fade(brightnessp+percentage, "50ms")
	return err
}

func (l *light) decBrightness(percentage int) error {
	brightness, err := l.getBrightness()
	if err != nil {
		return err
	}
	brightnessp := levelToPercent(brightness)
	err = l.fade(brightnessp-percentage, "50ms")
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
	var fadeTime string
	var target string
	var l light
	l.Class = PATH
	err := l.findPath()
	if err != nil {
		log.Fatal(err)
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
			err = l.fade(level, fadeTime)
			return err
		},
		Commands: []*cli.Command{
			{
				Name:  "low",
				Usage: "Set brightness to low",
				Action: func(cCtx *cli.Context) error {
					err := l.fade(5, fadeTime)
					return err
				},
			},
			{
				Name:  "mid",
				Usage: "Set brightness to mid",
				Action: func(cCtx *cli.Context) error {
					err := l.fade(50, fadeTime)
					return err
				},
			},
			{
				Name:  "max",
				Usage: "Set brightness to max",
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
					err := l.decBrightness(10)
					return err
				},
			},
			{
				Name:    "inc",
				Aliases: []string{"+"},
				Usage:   "Increase screen brightness",
				Action: func(cCtx *cli.Context) error {
					err := l.incBrightness(10)

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
