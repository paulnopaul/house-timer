package regularity

import (
	"strconv"
	"strings"
	"time"

	"github.com/agnivade/levenshtein"
)

func day(count int) time.Duration {
	return time.Duration(count) * time.Hour * 24
}

func week(count int) time.Duration {
	return day(7)
}

func month(count int) time.Duration {
	return day(30)
}

type durFunc func(int) time.Duration

var regularitites = []struct {
	word string
	dur  durFunc
}{
	{"день", day},
	{"неделя", week},
	{"месяц", month},
}

func nearestDuration(word string) durFunc {
	min := levenshtein.ComputeDistance(word, regularitites[0].word)
	minI := 0
	for i := 1; i < len(regularitites); i++ {
		if dist := levenshtein.ComputeDistance(word, regularitites[i].word); dist < min {
			min = dist
			minI = i
		}
	}
	return regularitites[minI].dur
}

func ExtractRegularity(regularityString string) (time.Duration, error) {
	splitted := strings.Split(regularityString, " ")
	if len(splitted) != 2 {
		return time.Duration(0), ErrArgCount
	}
	count, err := strconv.Atoi(splitted[0])
	if err != nil {
		return time.Duration(0), ErrFirstArg
	}
	if count == 0 {
		return time.Duration(0), ErrZero
	}
	dur := nearestDuration(splitted[1])
	return dur(count), nil
}
