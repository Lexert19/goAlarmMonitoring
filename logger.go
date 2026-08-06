package main

import "fmt"

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Log(event Event) {
	fmt.Printf("[%s] %s: %s\n", event.Level, event.Time.Format("15:04:05"), event.Type)
}
