package reporting

import (
	"log"
	"sync"
)

// safe
var hadError = false
var hadRuntimeError = false
var mu sync.Mutex

func ReportError(line int, msg string) {
	Report(line, "", msg)
}

func Report(line int, where string, msg string) {
	log.Printf("[line %d] Error %s: %s", line, where, msg)
	SetError()
}

func ReportRuntimeError(err error) {
	log.Println(err.Error())
	SetRuntimeError()
}

func SetError() {
	mu.Lock()
	hadError = true
	mu.Unlock()
}

func SetRuntimeError() {
	mu.Lock()
	hadRuntimeError = true
	mu.Unlock()
}

func UnsetError() {
	mu.Lock()
	hadError = false
	mu.Unlock()
}

func HadError() bool {
	mu.Lock()
	f := hadError
	mu.Unlock()
	return f
}

func HadRuntimeError() bool {
	mu.Lock()
	f := hadRuntimeError
	mu.Unlock()
	return f
}
