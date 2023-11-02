package _interface

type wait interface {
	Cusmer(c chan func(v ...interface{}))
	Push(c func(v ...interface{})) bool
}
