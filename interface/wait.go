package _interface

type wait interface {
	cusmer(c chan func(v ...interface{}))
	push(c func(v ...interface{})) bool
}
