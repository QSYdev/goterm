package tree

// Stack is a LIFO data structure.
type Stack struct {
	top  *Element
	size int
}

// Element is an element in the stack.
type Element struct {
	value interface{}
	next  *Element
}

// Empty returns true if no elements are in the stack.
func (s *Stack) Empty() bool {
	return s.size == 0
}

// Top returns the top value of the stack without poping.
func (s *Stack) Top() interface{} {
	return s.top.value
}

// Push pushes value into the stack.
func (s *Stack) Push(value interface{}) {
	s.size++
	s.top = &Element{value, s.top}
}

// Pop popes the top value.
func (s *Stack) Pop() (value interface{}) {
	if s.size == 0 {
		return nil
	}
	s.size--
	value, s.top = s.top.value, s.top.next
	return value
}
