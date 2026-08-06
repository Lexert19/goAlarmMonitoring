package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"
)

func ManualInputLoop(ctx context.Context, bus *EventBus) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("  m - Motion (INFO)")
	fmt.Println("  d - Door (WARNING)")
	fmt.Println("  s - Smoke (CRITICAL)")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if !scanner.Scan() {
				return
			}
			text := scanner.Text()
			if len(text) == 0 {
				continue
			}
			ch := text[0]
			var typ EventType
			var level Level
			switch ch {
			case 'm':
				typ, level = Motion, INFO
			case 'd':
				typ, level = Door, WARNING
			case 's':
				typ, level = Smoke, CRITICAL
			default:
				fmt.Println("Unknown command. Available: m, d, s")
				continue
			}
			event := Event{Type: typ, Time: time.Now(), Level: level}
			bus.Publish(event)
		}
	}
}
