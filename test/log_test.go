package test

import (
	"log"
	"os"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	log.SetPrefix("[index]")
	lo := log.New(os.Stdout, "[pool]"+time.Now().Format(time.DateTime)+" ", log.Llongfile)
	lo.Println(123)
	log.Println(111)
}
