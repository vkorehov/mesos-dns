package tmpl

import (
	"reflect"
	"testing"
	"testing/quick"
)

func TestNew(t *testing.T) {
	for _, tt := range []struct {
		Template
		err error
	}{
		{Template{"", nil}, nil},
		{Template{"a", []token{{"a", false}}}, nil},
		{Template{"{a}", []token{{"a", true}}}, nil},
		{Template{"{a}b{c}", []token{{"a", true}, {"b", false}, {"c", true}}}, nil},
		{Template{"a{b}c", []token{{"a", false}, {"b", true}, {"c", false}}}, nil},
		{Template{"{a}b{c", []token{{"a", true}, {"b", false}, {"c", false}}}, ErrUnbalanced},
		{Template{"}{a}b{c", nil}, ErrUnbalanced},
		{Template{"a}b{c}", nil}, ErrUnbalanced},
		{Template{"{a}b{{c}}", []token{{"a", true}, {"b", false}}}, ErrUnbalanced},
	} {
		if got, err := New(tt.template); err != tt.err {
			t.Errorf("New(%q) got err: %v, want: %v", tt.template, err, tt.err)
		} else if want := tt.Template; !reflect.DeepEqual(got, want) {
			t.Errorf("New(%q)\ngot:  %#v\nwant: %#v", tt.template, got, want)
		}
	}
}

func TestTemplate_Execute(t *testing.T) {
	for _, tt := range []struct {
		template string
		Context
		want string
	}{
		{"foo", nil, "foo"},
		{"{foo}", Context{"foo": "bar"}, "bar"},
		{"{ foo\t}", Context{"foo": "bar"}, "bar"},
		{"{   \tfoo\t \t}", Context{"foo": "bar"}, "bar"},
		{"foo.{foo}", Context{"foo": "bar"}, "foo.bar"},
		{"foo.{bar}", nil, "foo.bar"},
		{"{bar}.{foo}", Context{"foo": "bar", "bar": "foo"}, "foo.bar"},
	} {
		if template, err := New(tt.template); err != nil {
			t.Fatal("New(%q): %v", tt.template, err)
		} else if got, want := template.Execute(tt.Context), tt.want; got != want {
			t.Errorf("Execute(%v, %v)\ngot:  %#v\nwant: %#v", tt.template, tt.Context, got, want)
		}

	}
}

func TestHashString(t *testing.T) {
	t.Skip("TODO: Increase entropy, fix the bug!")
	fn := func(a, b string) bool { return hashString(a) != hashString(b) }
	if err := quick.Check(fn, &quick.Config{MaxCount: 1e9}); err != nil {
		t.Fatal(err)
	}
}
