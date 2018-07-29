package tree

import (
	"strconv"
	"strings"
)

const (
	openParen  = '('
	closeParen = ')'
	andOp      = '&'
	orOp       = '|'
)

// Node has an eval method that returns true depending
// on the visited elements.
type Node interface {
	Eval(visited []bool) bool
}

// And is a node that implements the and binary expression.
type And struct {
	Left, Right Node
}

// Eval implements the Node interface for and.
func (and And) Eval(visited []bool) bool {
	return and.Left.Eval(visited) && and.Right.Eval(visited)

}

// Or is a node that implements the or binary expression.
type Or struct {
	Left, Right Node
}

// Eval implements the Node interface for or.
func (or Or) Eval(visited []bool) bool {
	return or.Left.Eval(visited) || or.Right.Eval(visited)
}

// Leaf represents a leaf in the tree.
type Leaf struct {
	Value int
}

// Eval implements the Node interface for Leaft.
func (l Leaf) Eval(visited []bool) bool {
	return visited[l.Value]
}

// Parse parses the expression and returns the tree.
func Parse(expression string) Node {
	var stack Stack
	postfix := infixToPostfix(expression)
	for _, c := range postfix {
		switch c {
		case ' ':
			break
		case andOp:
			a := And{}
			a.Left, _ = stack.Pop().(Node)
			a.Right, _ = stack.Pop().(Node)
			stack.Push(a)
		case orOp:
			o := Or{}
			o.Left, _ = stack.Pop().(Node)
			o.Right, _ = stack.Pop().(Node)
			stack.Push(o)
		default:
			v, err := strconv.Atoi(string(c))
			if err != nil {
				return Leaf{}
			}
			stack.Push(Leaf{Value: v})
		}
	}
	t, _ := stack.Pop().(Node)
	return t
}

func infixToPostfix(infix string) string {
	var stack Stack
	postfix := ""
	var j int
	for i, c := range infix {
		switch c {
		case ' ':
		case andOp, orOp:
			for !stack.Empty() {
				top, _ := stack.Top().(rune)
				if top == openParen {
					break
				}
				postfix += " " + string(top)
				stack.Pop()
			}
			stack.Push(c)
		case openParen:
			stack.Push(c)
		case closeParen:
			for !stack.Empty() {
				str, _ := stack.Top().(rune)
				if str == openParen {
					break
				}
				postfix += " " + string(str)
				stack.Pop()
			}
			stack.Pop()
		default:
			if i < j {
				break
			}
			j = i
			number := ""
			for ; j < len(infix) && (infix[j] >= '0' && infix[j] <= '9'); j++ {
				number = number + string(infix[j])
			}
			postfix += " " + number
		}
	}
	for !stack.Empty() {
		str, _ := stack.Pop().(rune)
		postfix += " " + string(str)
	}
	return strings.TrimSpace(postfix)
}
