// Package tmpl contains types to compile and interpolate templates
// with a given context.
package tmpl

import (
	"errors"
	"strings"
)

type (
	// Context holds state to be interpolated in a Template.
	Context map[string]string

	// Template is a compiled Template that can be executed efficiently.
	Template struct {
		template string
		tokens   []token
	}

	// token is a parsed template token
	token struct {
		value string
		ident bool
	}
)

var (
	// ErrUnbalanced is returned when a template doesn't have balanced braces.
	ErrUnbalanced = errors.New("tmpl: unbalanced braces")
)

// New returns a compiled Template and an error in case of compilation failure.
func New(template string) (Template, error) {
	t := Template{template: template}

	next, prev, last := '{', '}', 0
	for i, r := range template {
		switch r {
		case next:
			if value := strings.TrimSpace(template[last:i]); value != "" {
				t.tokens = append(t.tokens, token{value, r == '}'})
			}
			next, prev, last = prev, next, i+1
		case prev:
			return t, ErrUnbalanced
		}
	}

	if last < len(template) {
		t.tokens = append(t.tokens, token{template[last:], false})
	}

	if next != '{' {
		return t, ErrUnbalanced
	}

	return t, nil
}

// Must is the same as New but panics instead of returning an error.
func Must(template string) Template {
	t, err := New(template)
	if err != nil {
		panic(err)
	}
	return t
}

// Execute replaces matching Context variables in a Template and returns the
// result.
func (t Template) Execute(c Context) string {
	out := make([]byte, 0, len(t.template)*2) // optimistic allocation
	for _, token := range t.tokens {
		value := token.value
		if token.ident && c != nil {
			if value = c[value]; value == "" {
				value = token.value
			}
		}
		out = append(out, value...)
	}
	return string(out)
}

// String implements the fmt.Stringer interface.
func (t Template) String() string { return t.template }
