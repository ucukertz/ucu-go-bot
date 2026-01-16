package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/samber/lo"
	"go.mau.fi/whatsmeow/types/events"
)

func IsAdmin(msg *events.Message) bool {
	if strings.Contains(msg.Info.Sender.User, "035") {
		return true
	}
	return false
}

func AdminCmdChk(msg *events.Message, cmd string) bool {
	if !IsAdmin(msg) {
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
		query := WaMsgPrompt(msg)
		newlvl := LoggerSetLvl(query, 24*time.Hour)
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
