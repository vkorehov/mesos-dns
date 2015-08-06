// Package patterns contains types to compile and interpolate domain patterns
// with a given context.
package patterns

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"

	"github.com/mesosphere/mesos-dns/logging"
	"github.com/mesosphere/mesos-dns/records/labels"
	"github.com/mesosphere/mesos-dns/records/state"
)

// PatternContext is the namespace to resolve DomainPattern variables in
type PatternContext map[string]string

// DomainPattern holds a text/template pattern for a domain name to be interpolated
// with a PatternContext.
type DomainPattern string

type token interface {
	interpolate(context PatternContext) (string, error)
	isSeparator() bool
}

type separatorToken struct{}
type stringToken string
type variableToken string

// CompiledDomainPattern is a compiled DomainPattern that can be executed efficiently
type CompiledDomainPattern struct {
	tokens  []token
	pattern DomainPattern
}

// defaultDomainPatterns is an non-exported default for the domain patterns. It
// is not made public to avoid being mutable from outside.
var defaultDomainPatterns = []DomainPattern{"{name}.{framework}"}

// DefaultDomainPatterns returns a clone of the default domain patterns
func DefaultDomainPatterns() []DomainPattern {
	clone := make([]DomainPattern, len(defaultDomainPatterns))
	copy(clone, defaultDomainPatterns)
	return clone
}

// addNonVariableTokens splits the given string s into partial labels and adds
// tokens for them. It accepts labels with "_" depending on whether s is in the
// left or right most position in the pattern.
func addNonVariableTokens(tokens []token, s string, leftMost, rightMost bool, spec labels.Func) ([]token, error) {
	if s == "" {
		return tokens, nil
	}

	if s == "." {
		tokens = append(tokens, separatorToken{})
		return tokens, nil
	}

	labels := strings.Split(s, ".")
	for i, label := range labels {
		if i != 0 {
			tokens = append(tokens, separatorToken{})
		}

		// "" and first or last => . at the left or right of s, skip empty string
		if label == "" && (i == 0 || i == len(labels)-1) {
			continue
		}

		// "" and not the first or last => consecutive separators
		if label == "" && i > 0 && i < len(labels)-1 {
			return nil, fmt.Errorf("invalid consecutive separators")
		}

		// special case for valid underscores: trim them for spec comparison below
		labelWithoutValidUnderscore := label
		if (label[0] == '_' && label != "_") || (label == "_" && !rightMost) {
			labelWithoutValidUnderscore = strings.TrimLeft(label, "_")
		}

		// prepend or append some character to check for RFC compatibility for non-inner labels.
		//
		// But don't do this if there is no variableToken on the left or the right, i.e. these
		// tokens are the most left or the most right ones in the pattern.
		pre := ""
		post := ""
		if i == 0 && !leftMost {
			pre = "a"
		}
		if i == len(labels)-1 && !rightMost {
			post = "a"
		}

		if escapedLabel := spec(pre + labelWithoutValidUnderscore + post); pre+labelWithoutValidUnderscore+post != escapedLabel {
			return nil, fmt.Errorf("pattern substring %v is no valid label", label)
		}
		tokens = append(tokens, stringToken(label))
	}
	return tokens, nil
}

// Compile compiles domainPatterns. This code only runs for a handful of patterns
// per InsertState run, i.e. is not time critical.
func (dp DomainPattern) Compile(spec labels.Func) (*CompiledDomainPattern, error) {
	tokens := []token{}

	if string(dp) == "" {
		return nil, fmt.Errorf("invalid empty domain pattern")
	}

	// split pattern into tokens: strings, separators and {variables}
	varRE, err := regexp.Compile(`{[\s\w-:]*}`)
	if err != nil {
		logging.Error.Fatalf("invalid regular expression for variables in domain pattern: %v", err)
	}

	// find variable references and work through the index list
	varMatches := varRE.FindAllStringIndex(string(dp), -1)
	oldRight := 0
	for i, m := range varMatches {
		// extract variable identifier
		left := m[0]
		right := m[1]
		identifier := strings.Trim(string(dp)[left+1:right-1], " \t")
		if identifier == "" {
			return nil, fmt.Errorf("empty variable reference found in domain pattern %v", dp)
		}

		// create token for everything in front of the variable
		leftMost := i == 0
		tokens, err = addNonVariableTokens(tokens, string(dp)[oldRight:left], leftMost, false, spec)
		if err != nil {
			return nil, fmt.Errorf("invalid domain pattern %v: %v", dp, err)
		}

		// add the actual variable token
		tokens = append(tokens, variableToken(identifier))

		// prepare for next round
		oldRight = right
	}

	// add pending tokens behind the last variable
	leftMost := len(varMatches) == 0
	rightMost := true
	tokens, err = addNonVariableTokens(tokens, string(dp)[oldRight:len(dp)], leftMost, rightMost, spec)
	if err != nil {
		return nil, fmt.Errorf("invalid domain pattern %v: %v", dp, err)
	}

	// check that the first and the last token is not a separator
	if tokens[0].isSeparator() {
		return nil, fmt.Errorf("domain pattern cannot start with a dot")
	}
	if tokens[len(tokens)-1].isSeparator() {
		return nil, fmt.Errorf("domain pattern cannot end with a dot")
	}

	return &CompiledDomainPattern{
		tokens:  tokens,
		pattern: dp,
	}, nil
}

// String returns the domain pattern string
func (dp DomainPattern) String() string {
	return string(dp)
}

// Execute applies a patternContext to a pre-compiled DomainPattern by interpolating
// the context values using the text.Template syntax.
// This function is called for basically every record. Make sure it's fast.
func (cdp *CompiledDomainPattern) Execute(context PatternContext, domain string) (string, error) {
	labels := make([]string, 0, len(cdp.tokens)+3)
	for _, t := range cdp.tokens {
		label, err := t.interpolate(context)
		if err != nil {
			return "", err
		}
		labels = append(labels, label)
	}
	labels = append(labels, ".", domain, ".")
	return strings.Join(labels, ""), nil
}

// Pattern returns the original uncompiled pattern
func (cdp *CompiledDomainPattern) Pattern() DomainPattern {
	return cdp.pattern
}

func (t separatorToken) interpolate(context PatternContext) (string, error) {
	return ".", nil
}

func (t stringToken) interpolate(context PatternContext) (string, error) {
	return string(t), nil
}

func (t variableToken) interpolate(context PatternContext) (string, error) {
	value := context[string(t)]
	if value == "" {
		return "", fmt.Errorf("%q is not defined in the pattern context %v", t, context)
	}
	return value, nil
}

func (t separatorToken) isSeparator() bool { return true }
func (t stringToken) isSeparator() bool    { return false }
func (t variableToken) isSeparator() bool  { return false }

// NewPatternContext creates a patternContext for a given task and a label spec
func NewPatternContext(task *state.Task, framework string, spec labels.Func) PatternContext {
	context := PatternContext{
		"framework":      framework,
		"slave-id-short": slaveIDTail(task.SlaveID),
		"slave-id":       task.SlaveID,
		"task-id":        task.ID,
		"task-id-hash":   hashString(task.ID),
		"name":           specEachLabel(task.Name, spec),
	}

	if task.Discovery != nil {
		possiblySet := func(key string, value *string) {
			if value != nil && *value != "" {
				context[key] = specEachLabel(*value, spec)
			}
		}
		possiblySet("version", task.Discovery.Version)
		possiblySet("location", task.Discovery.Location)
		possiblySet("environment", task.Discovery.Environment)

		for _, label := range task.Discovery.Labels.Labels {
			context["label:"+label.Key] = spec(label.Value)
		}

		// use discovery name of task name if defined
		if task.Discovery.Name != nil && *task.Discovery.Name != "" {
			context["name"] = spec(*task.Discovery.Name)
		}
	}
	return context
}

// specEachLabel
func specEachLabel(s string, spec labels.Func) string {
	labels := strings.Split(s, ".")
	specLabels := make([]string, 0, len(labels))
	for _, l := range labels {
		specLabels = append(specLabels, spec(l))
	}
	return strings.Join(specLabels, ".")
}

// return the slave number from a Mesos slave id
func slaveIDTail(slaveID string) string {
	fields := strings.Split(slaveID, "-")
	return strings.ToLower(fields[len(fields)-1])
}

// BUG: The probability of hashing collisions is too high with only 17 bits.
// NOTE: Using a numerical base as high as valid characters in DNS names would
// reduce the resulting length without risking more collisions.
func hashString(s string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	sum := h.Sum32()
	lower, upper := uint16(sum), uint16(sum>>16)
	return strconv.FormatUint(uint64(lower+upper), 10)
}
