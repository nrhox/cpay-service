package utils

import (
	"regexp"
	"strconv"
	"time"
)

var durationRegex = regexp.MustCompile(`(\d+)(y|mo|d|h|m|s)`)

func ParseDuration(input string) time.Duration {
	matches := durationRegex.FindAllStringSubmatch(input, -1)

	var total time.Duration

	for _, match := range matches {
		valueInt, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			continue
		}

		value := time.Duration(valueInt)
		unit := match[2]

		switch unit {
		case "y":
			total += value * 24 * 365 * time.Hour
		case "mo":
			total += value * 24 * 30 * time.Hour
		case "d":
			total += value * 24 * time.Hour
		case "h":
			total += value * time.Hour
		case "m":
			total += value * time.Minute
		case "s":
			total += value * time.Second
		}
	}

	return total
}
