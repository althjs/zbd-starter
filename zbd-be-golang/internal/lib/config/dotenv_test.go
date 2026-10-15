package config

import "testing"

func TestDotEnvValue(t *testing.T) {
	for in, want := range map[string]string{
		`"abc def" # 설명`:   "abc def",
		`'x#y'`:            "x#y",
		`plain value here`: "plain value here",
		`plain # 설명`:       "plain",
		`a#b`:              "a#b",
		` "q" `:            "q",
	} {
		if got := dotEnvValue(in); got != want {
			t.Errorf("%q → %q, want %q", in, got, want)
		}
	}
}
