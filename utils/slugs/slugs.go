package slugs

import (
	"math/rand"
	"regexp"
	"strings"
	"time"
)

var (
	nonAlnum = regexp.MustCompile("[^a-z0-9]+")
	dashes   = regexp.MustCompile("-+")
)

func Slugify(text string) string {
	s := strings.ToLower(strings.TrimSpace(text))
	s = strings.ReplaceAll(s, "&", " and ")
	s = nonAlnum.ReplaceAllString(s, "-")
	s = dashes.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func Unique(base string, exists func(string) (bool, error)) (string, error) {
	base = strings.Trim(base, "-")
	if base == "" {
		base = "community"
	}
	if exists == nil {
		return base, nil
	}

	taken, err := exists(base)
	if err != nil {
		return "", err
	}
	if !taken {
		return base, nil
	}

	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	for i := 0; i < 12; i++ {
		candidate := base + "-" + randomSuffix(r)
		taken, err := exists(candidate)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
	}
	return base + "-" + randomSuffix(r), nil
}

func randomSuffix(r *rand.Rand) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
