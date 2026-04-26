package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/samber/lo"
	"go.mau.fi/whatsmeow/types/events"
)

func AdminChk(msg *events.Message) bool {
	if strings.Contains(msg.Info.Sender.User, "234") {
		return true
	}
	return false
}

func AdminCmdChk(msg *events.Message, cmd string) bool {
	if !AdminChk(msg) {
		return false
	}

	switch cmd {
	case "!a.stat":
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		mem := fmt.Sprintf("Memory Usage: %.2f MB", float64(m.Alloc)/1024/1024)
		uptime := time.Since(synctime).Round(time.Second).String()
		up := fmt.Sprintf("Uptime: %s", uptime)
		stats := strings.Join([]string{mem, up}, "\n")
		WaReplyText(msg, stats)
		return true
	case "!a.log":
		prompt := WaMsgPrompt(msg)
		newlvl := LoggerLvlSet(prompt, 24*time.Hour)
		WaReplyText(msg, fmt.Sprintf("Log level set to %s for 24 hours", newlvl))
		return true
	case "!a.tune":
		go SdTune(msg)
		return true
	case "!a.bake":
		go SdBake(msg)
		return true
	case "!a.take":
		go SdTake(msg)
		return true
	}

	return false
}

// Returns the development version of a variable if in dev mode, otherwise returns the production version.
func AdminDevDiff[T any](dev T, prod T) T {
	return lo.If(ENV_DEV_MODE == "1", dev).Else(prod)
}

func AdminBackoff(attempt int) {
	if attempt <= 0 {
		return
	}
	sec := 5 + (attempt-1)*2
	time.Sleep(time.Duration(sec) * time.Second)
}

var fireReactLastTime time.Time

func AdminFireReact(msg *events.Message, text string) {
	if !strings.Contains(text, "233222") {
		return
	}

	now := time.Now().UTC()
	var lastRecharge time.Time
	y, m, d := now.Date()

	t1 := time.Date(y, m, d, 2, 0, 0, 0, time.UTC)
	t2 := time.Date(y, m, d, 7, 0, 0, 0, time.UTC)
	t3 := time.Date(y, m, d, 12, 0, 0, 0, time.UTC)

	if now.After(t3) || now.Equal(t3) {
		lastRecharge = t3
	} else if now.After(t2) || now.Equal(t2) {
		lastRecharge = t2
	} else if now.After(t1) || now.Equal(t1) {
		lastRecharge = t1
	} else {
		lastRecharge = time.Date(y, m, d-1, 12, 0, 0, 0, time.UTC)
	}

	if fireReactLastTime.Before(lastRecharge) {
		fireReactLastTime = now
		WaReact(msg, "🔥")
	}
}
