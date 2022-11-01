package containers

import "fmt"

type Stack[T any] struct {
	items []T
	pos   int
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{items: make([]T, 0), pos: 0}
}

func (stack *Stack[T]) Len() int {
	return stack.pos
}

func (stack *Stack[T]) Push(value T) {
	pos := stack.pos
	if pos >= len(stack.items) {
		stack.items = append(stack.items, value)
	} else {
		stack.items[pos] = value
	}
	stack.pos = pos + 1
}

func (stack *Stack[T]) Pop() (res T, err error) {
	pos := stack.pos
	if pos == 0 {
		err = fmt.Errorf("stack is empty")
		return
	}

	res = stack.items[pos-1]
	stack.pos = pos - 1
	return
}
