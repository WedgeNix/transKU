package transku

import "regexp"

const (
	// HTMLPattern catches anything in HTML brackets.
	HTMLPattern = `<[^<>]*>`

	// WordPattern catches alphabetical words in english.
	WordPattern = `[A-Za-z'][A-Za-z'-]*[A-Za-z']{2,}|[AEIOUaeiou]|[A-Za-z']{2,}`

	// PhrasePattern catches phrases at a time.
	PhrasePattern = `(` + WordPattern + `)` + `( (` + WordPattern + `))*`
)

var (
	// HTML is the regular expression compiled.
	htmlRegex = regexp.MustCompile(HTMLPattern)

	// Phrase is the regular expression compiled.
	phraseRegex = regexp.MustCompile(PhrasePattern)
)
