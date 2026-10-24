// The oracle for regex's tests: runs every case through Go's regexp
// package (RE2, the same syntax and matching rules) and writes what it
// finds to tests/cases.json, which `tessel run tests` checks the library
// against. Run it from the library's folder: go run tools/oracle/main.go
//
// Author: Alexius Academia
// Date: 2026-10-24
package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"regexp"
	"unicode/utf8"
)

// One case: a pattern and a text, and what Go makes of them. Positions are
// in characters (code points), as Tessel counts them.
type Case struct {
	Pattern  string   `json:"pattern"`
	Text     string   `json:"text"`
	Ok       bool     `json:"ok"`
	Matches  [][]int  `json:"matches"`
	Replaced string   `json:"replaced"`
	Split    []string `json:"split"`
	Whole    bool     `json:"whole"`
}

// The template every case's replacement uses.
const template = "<${1}|$0|$$>"

func chars(s string, byteIndex int) int {
	if byteIndex < 0 {
		return -1
	}
	return utf8.RuneCountInString(s[:byteIndex])
}

func run(pattern, text string) Case {
	c := Case{Pattern: pattern, Text: text, Matches: [][]int{}, Split: []string{}}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return c
	}
	c.Ok = true
	for _, m := range re.FindAllStringSubmatchIndex(text, -1) {
		row := make([]int, len(m))
		for i, b := range m {
			row[i] = chars(text, b)
		}
		c.Matches = append(c.Matches, row)
	}
	c.Replaced = re.ReplaceAllString(text, template)
	c.Split = re.Split(text, -1)
	whole := regexp.MustCompile(`^(?:` + pattern + `)$`)
	c.Whole = whole.MatchString(text)
	return c
}

// Hand-written patterns, each tried on every text.
var patterns = []string{
	// Literals, any, classes
	`a`, `abc`, `.`, `a.c`, `(?s)a.c`, `[abc]`, `[^abc]`, `[a-c]+`, `[^a-c]+`, `[a\-z]`, `[]a]`, `[^]a]+`, `[a-]`,
	`[\d]+`, `[^\d\s]+`, `[\w.]+`, `[[:alpha:]]+`, `[[:^alpha:]]+`, `[[:digit:][:upper:]]+`, `[[:punct:]]`, `[[:xdigit:]]+`,
	`\d`, `\d+`, `\D+`, `\w+`, `\W+`, `\s+`, `\S+`, `\x41`, `\x{263a}`, `\t`, `\n`, `[\n\t]`, `\.`, `\*\+\?`, `\(\)`, `\\`,
	`é`, `[à-ÿ]+`, `☺`, `😀+`, `[😀-😂]`,
	// Anchors and boundaries
	`^a`, `a$`, `^$`, `^`, `$`, `(?m)^a`, `(?m)a$`, `(?m)^$`, `(?m)^\w+$`, `\Aa`, `a\z`, `\bfoo\b`, `\Bo\B`, `\b`, `\B`,
	// Repeats
	`a*`, `a+`, `a?`, `a*?`, `a+?`, `a??`, `a{2}`, `a{2,}`, `a{2,3}`, `a{2,3}?`, `a{0}`, `a{0,1}`, `(ab){2}`, `a{,3}`, `a{`, `a{1`, `a{x}`,
	`.*`, `.*?`, `.+`, `(a|b)*`, `(a|b)*?c`, `(a*)*`, `(a*)+`, `(a|)+`, `(|a)+`, `(a?)*`, `x*`, `(x*)*y`,
	// Alternation and groups
	`a|b`, `ab|cd`, `a|ab`, `ab|a`, `(a|ab)(c|bcd)(d*)`, `(a)|(b)`, `(a)|b`, `(?:a|b)+`, `((a)(b))`, `(a(b(c)))`, `()`, `(|)`, `a(|b)`,
	`(?P<x>a)(?P<y>b)?`, `(?<word>\w+)\s(?<num>\d+)`,
	// Flags
	`(?i)abc`, `(?i)[a-c]+`, `(?i)ÀÉ`, `(?i)straße`, `(?i)ΣΑΣ`, `(?i)привет`, `(?i:a)b`, `a(?i)b`, `(?i)a(?-i)b`, `(?im)^A$`, `(?s:.)+`,
	// Things that mean something only in context
	`a.*b`, `a.*?b`, `[^\n]+`, `\d{3}-\d{4}`, `(\d+)\.(\d+)`, `\$\d+(\.\d\d)?`, `[A-Z][a-z]+`, `(\w+)@(\w+)\.com`,
	// Errors
	`(`, `)`, `a)`, `[a`, `[z-a]`, `*a`, `a**`, `a*+`, `+`, `?`, `{1}`, `a{2}{3}`, `a{1001}`, `\1`, `(a)\1`, `(?=a)`, `(?!a)`, `(?<=a)`, `(?<!a)`,
	`\q`, `\`, `x{2,1}`, `(?z)`, `[[:foo:]]`, `(?P<1a>x)`,
}

var texts = []string{
	"", "a", "abc", "aaa", "abab", "aXbXc", "ab cd", "a\nb", "a\nb\n", "\nab\n", "foo bar foo.bar", "xoox",
	"Hello World 123", "3.14 and 2.71", "$5.00 $12", "ann@mail.com bo@site.com", "555-1234 x 9999-12",
	"ÀÉ àé", "Straße STRASSE", "σας ΣΑΣ", "Привет мир", "☺😀😁😂", "aabbaabb", "abcd", "a\tb  c", "ABC abc AbC",
	"word 42 other 7",
}

// Random patterns over a small alphabet, to stress the matching machine.
func randomPattern(r *rand.Rand, depth int) string {
	if depth <= 0 {
		atoms := []string{"a", "b", "c", ".", "[ab]", "[^a]", `\b`, "^", "$", `\w`, "x"}
		return atoms[r.Intn(len(atoms))]
	}
	switch r.Intn(9) {
	case 0, 1:
		return randomPattern(r, depth-1) + randomPattern(r, depth-1)
	case 2:
		return randomPattern(r, depth-1) + "|" + randomPattern(r, depth-1)
	case 3:
		return "(" + randomPattern(r, depth-1) + ")"
	case 4:
		return "(?:" + randomPattern(r, depth-1) + ")"
	case 5:
		reps := []string{"*", "+", "?", "*?", "+?", "??", "{2}", "{1,2}", "{0,2}?", "{2,}"}
		return "(" + randomPattern(r, depth-1) + ")" + reps[r.Intn(len(reps))]
	case 6:
		return randomPattern(r, depth-1) + "*"
	default:
		return randomPattern(r, depth-1)
	}
}

func randomText(r *rand.Rand) string {
	letters := []string{"a", "b", "c", "x", " ", "\n"}
	n := r.Intn(9)
	s := ""
	for i := 0; i < n; i++ {
		s += letters[r.Intn(len(letters))]
	}
	return s
}

func main() {
	cases := []Case{}
	for _, p := range patterns {
		for _, t := range texts {
			cases = append(cases, run(p, t))
		}
	}
	r := rand.New(rand.NewSource(2026))
	for len(cases) < len(patterns)*len(texts)+3000 {
		p := randomPattern(r, 1+r.Intn(4))
		if _, err := regexp.Compile(p); err != nil {
			continue
		}
		for k := 0; k < 3; k++ {
			cases = append(cases, run(p, randomText(r)))
		}
	}
	out, _ := json.Marshal(cases)
	os.WriteFile("tests/cases.json", out, 0o644)
}
