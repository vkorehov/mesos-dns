// Package patterns contains types to compile and interpolate domain patterns
// with a given context.
package patterns

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"text/template"

	"github.com/mesosphere/mesos-dns/records/labels"
	"github.com/mesosphere/mesos-dns/records/state"
)

// PatternContext is the namespace to resolve DomainPattern variables in
type PatternContext struct {
	SlaveID   string
	TaskID    string
	TaskName  string
	Discovery struct {
		Version     *string
		Name        *string
		Location    *string
		Environment *string
		Labels      map[string]string
	}
}

// DomainPattern holds a text/template pattern for a domain name to be interpolated
// with a PatternContext.
type DomainPattern string

// CompiledDomainPattern is a compiled DomainPattern that can be executed efficiently
type CompiledDomainPattern template.Template

// defaultDomainPatterns is an non-exported default for the domain patterns. It
// is not made public to avoid being mutable from outside.
var defaultDomainPatterns = []DomainPattern{"{{.TaskName}}"}

// DefaultDomainPatterns returns a clone of the default domain patterns
func DefaultDomainPatterns() []DomainPattern {
	clone := make([]DomainPattern, len(defaultDomainPatterns))
	copy(clone, defaultDomainPatterns)
	return clone
}

// Compile compiles domainPatterns into text/templates.
// This code only runs for a handful of patterns per InsertState run,
// i.e. is not time critical.
func (dp DomainPattern) Compile() (*CompiledDomainPattern, error) {
	tmpl, err := template.New(string(dp)).Parse(string(dp))
	if err != nil {
		return nil, fmt.Errorf("invalid domain pattern %q: %v", dp, err)
	}

	return (*CompiledDomainPattern)(tmpl), nil
}

// String returns the domain pattern string
func (dp DomainPattern) String() string {
	return string(dp)
}

// Execute applies a patternContext to a pre-compiled DomainPattern by interpolating
// the context values using the text.Template syntax.
func (cdp *CompiledDomainPattern) Execute(context *PatternContext) (string, error) {
	var buf bytes.Buffer
	err := (*template.Template)(cdp).Execute(&buf, context)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Pattern returns the original uncompiled pattern
func (cdp *CompiledDomainPattern) Pattern() DomainPattern {
	name := (*template.Template)(cdp).Name()
	return DomainPattern(name)
}

// NewPatternContext creates a patternContext for a given task and a label spec
func NewPatternContext(task *state.Task, spec labels.Func) *PatternContext {
	context := PatternContext{
		SlaveID:  slaveIDTail(task.SlaveID),
		TaskID:   hashString(task.ID),
		TaskName: spec(task.Name),
	}
	context.Discovery.Labels = make(map[string]string)
	if task.Discovery != nil {
		context.Discovery.Version = task.Discovery.Version
		context.Discovery.Name = task.Discovery.Name
		context.Discovery.Location = task.Discovery.Location
		context.Discovery.Environment = task.Discovery.Environment
		for _, label := range task.Discovery.Labels.Labels {
			context.Discovery.Labels[label.Key] = label.Value
		}
	}
	return &context
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
