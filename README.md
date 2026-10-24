# regex

Regular expressions for [Tessel](https://tessel-lang.syncster.dev)
programs, with the syntax of RE2 (the one Go and Rust use). Matching
always takes time in proportion to the text, whatever the pattern, so no
pattern can hang a program, even one a user typed in.

```tessel
import regex

fn main() {
    let date = regex.must("(?P<year>\\d\{4})-(\\d\{2})-(\\d\{2})")
    if let m = date.find("Due 2026-10-31, paid 2026-11-02.") {
        print("{m.text} at {m.start}, year {m.named("year") ?? ""}")
    }
    print(date.replaceAll("Due 2026-10-31.", with: "$3/$2/$\{year}"))
}
```

```text
2026-10-31 at 4, year 2026
Due 31/10/2026.
```

## Using it

Add it to a project (Tessel 0.10 or later):

```sh
tessel add regex
```

then `import regex`.

### Backslashes and braces in Tessel strings

A pattern is a Tessel string, where a backslash is written `\\` and a
brace `\{`. So the pattern `\d{3}\.\w+` is written
`"\\d\{3}\\.\\w+"`. A closing brace `}` needs nothing.

## Making a pattern

| Function | |
|---|---|
| `regex.compile(pattern)` | A `Result<Regex>`: the compiled pattern, or a failure saying what's wrong and where, like `missing ) for this ( at index 0 of "(\d+"`. For patterns from users or files. |
| `regex.must(pattern)` | The `Regex`, for patterns written in the code: a mistake in the pattern stops the program with the reason. |
| `regex.escape(text)` | The text with a backslash before every character that means something in a pattern, to match it exactly. |

Compile a pattern once and use it many times: compiling does all the work
of reading the pattern.

## Using a pattern

| Method of `Regex` | Gives |
|---|---|
| `matches(text)` | Whether it matches anywhere in the text. |
| `matchesWhole(text)` | Whether it matches the whole text, first character to last. |
| `find(text, from:)` | The first `Match` (starting at character `from` or later, 0 unless you say), or `nil`. |
| `findAll(text)` | Every match, left to right, none overlapping. |
| `count(text)` | How many matches there are. |
| `replace(text, with:)` | The text with the first match replaced. |
| `replaceAll(text, with:)` | The text with every match replaced. In `with`, `$1` is group 1's text, `$\{name}` or `$<name>` a named group's, `$0` the whole match and `$$` a dollar sign. |
| `replaceEach(text) { m in … }` | The text with every match replaced by what the block gives for it. |
| `split(text, limit:)` | The parts between the matches; with a `limit` above 0, at most that many parts, the last holding the rest. |
| `groupCount`, `groupNames`, `groupIndex(name)` | The pattern's capturing groups. |

A `Match` has `start` and `end` (character indexes: its text is
`text.substring(from: start, to: end)`), `text`, `group(i)` (group 0 is the
whole match; `nil` for a group that took no part), `named(name)`, and
`groupStart(i)` and `groupEnd(i)`.

```tessel
let r = regex.must("\\b[a-z]")
print(r.replaceEach("hello big world") { m in m.text.uppercase() })
print(regex.must("\\s*,\\s*").split("a , b,c ,d"))
print(regex.must("[A-Z]\{3}-\\d\{4}").matchesWhole("ABC-1234"))
for m in regex.must("\\d+").findAll("3 apples, 12 pears") {
    print("{m.text} at {m.start}")
}
```

```text
Hello Big World
["a", "b", "c", "d"]
true
3 at 0
12 at 10
```

## The syntax

| | |
|---|---|
| `x` | The character itself, unless it's one of `\ . + * ? ( ) \| [ ] { } ^ $` |
| `\.` `\*` `\\` … | A punctuation mark taken literally |
| `.` | Any character but a line break (with `(?s)`, any at all) |
| `[abc]` `[a-z]` `[^a-z]` | One of a set, a range, or anything not in it |
| `[[:alpha:]]` `[[:^digit:]]` | POSIX classes: `alnum alpha ascii blank cntrl digit graph lower print punct space upper word xdigit` |
| `\d` `\w` `\s` | A digit `[0-9]`, a word character `[0-9A-Za-z_]`, a space `[\t\n\v\f\r ]` |
| `\D` `\W` `\S` | Anything but those |
| `\n` `\t` `\r` `\f` `\v` `\a` `\0` | Line feed, tab, return, form feed, vertical tab, bell, zero |
| `\x41` `\x{1F600}` `\u00e9` `\u{1F600}` | A character by its code |
| `^` `$` | The start and end of the text (with `(?m)`, of any line) |
| `\A` `\z` | The start and end of the text, always |
| `\b` `\B` | Between a word character and another, or not |
| `x*` `x+` `x?` | 0 or more, 1 or more, 0 or 1, as many as can be |
| `x{3}` `x{2,}` `x{2,5}` | Exactly 3, 2 or more, 2 to 5 (up to 1000) |
| `x*?` `x+?` `x??` `x{2,5}?` | The same, as few as can be |
| `x\|y` | Either |
| `(x)` | A group: what it matched is kept, numbered from 1 |
| `(?P<name>x)` `(?<name>x)` | A named group |
| `(?:x)` | A group that isn't kept |
| `(?i)` `(?m)` `(?s)` `(?i-s)` | Flags, from there to the end of the group: `i` ignore case, `m` `^` and `$` at lines, `s` `.` matches line breaks |
| `(?i:x)` | Flags for `x` only |

Which match is found: the leftmost, and from there the one a
backtracking engine would find, trying alternatives in order and repeats
greedily (or lazily). That's what Go, Rust, Python and JavaScript agree on.

### What isn't there, and why

**Backreferences** (`\1`) and **lookarounds** (`(?=…)`, `(?!…)`,
`(?<=…)`, `(?<!…)`) can't be matched in linear time, so the library
refuses them with a message saying so, as RE2, Go and Rust do. **Unicode
classes** like `\p{L}` aren't there yet.

### Unicode

Positions count characters as Tessel does, by code point (an emoji is
one character), so they work with `substring` directly. `\w`, `\d`, `\s`
and `\b` are ASCII, as in RE2. `(?i)` knows the cases of ASCII, Latin-1,
Latin Extended-A, Greek and Cyrillic letters, one to one: `(?i)straße`
doesn't match `STRASSE`.

## Speed

Matching runs every way the pattern could match at once (a Pike VM), so
the time is the text's length times at most the pattern's size. On an
Apple M5 Pro, a log of 850 000 characters takes about 0.15 to 0.25 s to
search with a pattern like `(\w+)@(\w+)\.com`, and a pattern that starts
with a plain character (`ERROR`) skips ahead to it and takes a few
milliseconds. Patterns that take a backtracking engine longer than the
age of the universe, like `(a+)+$` on a long run of a's, take the same
time as any other.

## Tests

```sh
tessel run tests
```

checks the library against Go's `regexp` package on 6 915 cases, every
feature on 25 texts plus 3 000 random patterns: which patterns are
accepted, every match and the positions of its groups, a replacement, a
split, and `matchesWhole`. Then it tests the API, the error messages,
Unicode positions, and that the time grows in proportion to the text.
`go run tools/oracle/main.go` makes the cases again (`tests/cases.json`)
from Go.

## What's where

| File | |
|---|---|
| `regex.tsl` | `compile`, `must`, `escape`, `Regex` and `Match`. |
| `parse.tsl` | Reading a pattern into a tree, and its error messages. |
| `compile.tsl` | Turning the tree into a program for the machine. |
| `machine.tsl` | The matching machine (Pike VM). |
| `classes.tsl` | Character classes: ranges, \d \w \s, POSIX classes, case folding. |
| `text.tsl` | Text as code points, and letter cases. |
| `tests/` | The tests and the cases from Go. |
| `tools/oracle/main.go` | Makes the cases from Go. |

## Licence

MIT: see [LICENSE](LICENSE).
