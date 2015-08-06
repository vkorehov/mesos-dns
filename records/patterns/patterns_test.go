package patterns

import (
	"testing"
	"testing/quick"

	"github.com/mesosphere/mesos-dns/records/labels"
)

func TestHashString(t *testing.T) {
	t.Skip("TODO: Increase entropy, fix the bug!")
	fn := func(a, b string) bool { return hashString(a) != hashString(b) }
	if err := quick.Check(fn, &quick.Config{MaxCount: 1e9}); err != nil {
		t.Fatal(err)
	}
}

func TestCompile(t *testing.T) {
	for _, ts := range []struct {
		pattern string
		rfc     labels.Func
		err     bool
	}{
		{"abc", labels.RFC952, false},

		{"", labels.RFC952, true},
		{".", labels.RFC952, true},
		{"abc.", labels.RFC952, true},
		{".abc", labels.RFC952, true},
		{".abc.", labels.RFC952, true},
		{".a.b.c.", labels.RFC952, true},
		{"a..bc", labels.RFC952, true},
		{"a...bc", labels.RFC952, true},
		{"1", labels.RFC952, true},
		{"1.2", labels.RFC952, true},
		{"-", labels.RFC952, true},
		{"a-", labels.RFC952, true},
		{"-a", labels.RFC952, true},
		{"a.-.b", labels.RFC952, true},
		{"a:b", labels.RFC952, true},

		{"_abc", labels.RFC952, false},
		{"_{abc}", labels.RFC952, false},
		{"_{abc}._tcp.mesos", labels.RFC952, false},

		{"_", labels.RFC952, true},
		{"a_b", labels.RFC952, true},
		{"abc_", labels.RFC952, true},
		{"_{abc}._", labels.RFC952, true},

		{"abc.def.ghi", labels.RFC952, false},
		{"abc.def123.ghi", labels.RFC952, false},
	} {
		_, err := DomainPattern(ts.pattern).Compile(ts.rfc)
		if err != nil && !ts.err {
			t.Errorf("cannot compile pattern %q: %v", ts.pattern, err)
			continue
		} else if err == nil && ts.err {
			t.Errorf("expected error compiling pattern %q", ts.pattern)
			continue
		}
	}
}

func TestExecute(t *testing.T) {
	for _, ts := range []struct {
		pattern string
		rfc     labels.Func
		context PatternContext
		answer  string
		err     bool
	}{
		{"abc", labels.RFC952, PatternContext{}, "abc", false},
		{"abc.def", labels.RFC952, PatternContext{}, "abc.def", false},
		{"abc.def123.ghi.j-k-l", labels.RFC952, PatternContext{}, "abc.def123.ghi.j-k-l", false},

		{"{framework}", labels.RFC952, PatternContext{"framework": "marathon"}, "marathon", false},
		{"{ framework\t}", labels.RFC952, PatternContext{"framework": "marathon"}, "marathon", false},
		{"{   \tframework\t \t}", labels.RFC952, PatternContext{"framework": "marathon"}, "marathon", false},
		{"{framework}.mesos", labels.RFC952, PatternContext{"framework": "marathon"}, "marathon.mesos", false},
		{"{name}.{framework}.mesos", labels.RFC952, PatternContext{"framework": "marathon", "name": "nginx"}, "nginx.marathon.mesos", false},
		{"{name}-{framework}.mesos", labels.RFC952, PatternContext{"framework": "marathon", "name": "nginx"}, "nginx-marathon.mesos", false},
	} {
		compiled, err := DomainPattern(ts.pattern).Compile(ts.rfc)
		if err != nil {
			t.Errorf("cannot compile pattern %q: %v", ts.pattern, err)
			continue
		}

		got, err := compiled.Execute(ts.context, "mesos")
		if err != nil && !ts.err {
			t.Errorf("unexpected execution error for pattern %v in context %v: %v", ts.pattern, ts.context, err)
			continue
		} else if err == nil && ts.err {
			t.Errorf("expected execution error for pattern %v in context %v: got %v", ts.pattern, ts.context, got)
			continue
		}

		if got != ts.answer {
			t.Errorf("invalid answer for pattern %v in context %v: got %q, want %q", ts.pattern, ts.context, got, ts.answer)
			continue
		}
	}
}
