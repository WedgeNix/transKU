package regex

import "regexp"

const (
	// HTMLPattern catches anything in HTML brackets.
	HTMLPattern = `<[^<>]*>`

	// WordPattern catches alphabetical words in english.
	WordPattern = `[A-Za-z'-]+`

	// PhrasePattern catches phrases at a time.
	PhrasePattern = `(` + WordPattern + `)` + `( ` + WordPattern + `)*`
)

var (
	// HTML is the regular expression compiled.
	HTML = regexp.MustCompile(HTMLPattern)

	// Phrase is the regular expression compiled.
	Phrase = regexp.MustCompile(PhrasePattern)
)
