package tree

import "testing"

func TestInfixToPostfix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		infix  string
		result string
	}{
		{name: "and of ands with paren", infix: "(1&2)& (3 &4)", result: "1 2 & 3 4 & &"},
		{name: "or of ands with paren", infix: "(1&2)|(3&4)", result: "1 2 & 3 4 & |"},
		{name: "and no paren", infix: "1&2", result: "1 2 &"},
		{name: "no operator", infix: "1", result: "1"},
		{name: "no paren lot of &", infix: "1&2&3&4", result: "1 2 & 3 & 4 &"},
	}
	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			if postfix := infixToPostfix(c.infix); postfix != c.result {
				tt.Errorf("expected %s but got %s", c.result, postfix)
			}
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		visited    []bool
		expression string
		eval       bool
	}{
		{name: "and with all visited", visited: []bool{true, true, true, true}, expression: "(0|1)&(2|3)", eval: true},
		{name: "or with enough visited", visited: []bool{true, false}, expression: "0|1", eval: true},
		{name: "and with not all visited", visited: []bool{true, false}, expression: "0&1", eval: false},
		{name: "or with none visited", visited: []bool{false, false}, expression: "0|1", eval: false},
		{name: "and of lot expressions no paren", visited: []bool{true, true, true, true}, expression: "0&1&2&3", eval: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			t := Parse(c.expression)
			if eval := t.Eval(c.visited); eval != c.eval {
				tt.Fatalf("expected %s to eval to %v but got %v", c.expression, c.eval, eval)
			}
		})
	}
}
