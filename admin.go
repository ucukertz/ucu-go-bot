package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types/events"
)

func AdminCmdChk(msg *events.Message, cmd string) bool {
	if !strings.Contains(msg.Info.Sender.User, "3405") {
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
		WaText(msg, stats)
		return true
	case "!a.log":
		query := WaMsgQry(msg)
		newlvl := LoggerSetLvl(query, 24*time.Hour)
		WaText(msg, fmt.Sprintf("Log level set to %s for 24 hours", newlvl))
		return true
	}

	return false
}
