package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var dlvl = AdminDevDiff(zerolog.DebugLevel, zerolog.InfoLevel) // Default level
var log = zerolog.New(zerolog.ConsoleWriter{TimeFormat: "2006/01/02 - 15:04:05.000", Out: os.Stdout, NoColor: true}).With().Timestamp().Logger()

func LoggerInit() {
	zerolog.SetGlobalLevel(dlvl)
	zerolog.TimeFieldFormat = time.RFC3339Nano
}

func LoggerStr2Lvl(s string) zerolog.Level {
	notfound := zerolog.TraceLevel
	switch s[0] {
	case 'T', 't':
		return zerolog.TraceLevel
	case 'D', 'd':
		return zerolog.DebugLevel
	case 'I', 'i':
		return zerolog.InfoLevel
	case 'W', 'w':
		return zerolog.WarnLevel
	case 'E', 'e':
		return zerolog.ErrorLevel
	case 'F', 'f':
		return zerolog.FatalLevel
	case 'P', 'p':
		return zerolog.PanicLevel
	case 'N', 'n':
		return zerolog.NoLevel
	default:
		log.Warn().Msgf("Can't identify desired level, using %s instead", notfound.String())
		return notfound
	}
}

var logResetIntr = make(chan struct{})
var logResetterExist bool = false

func logResetter(dr time.Duration) {
	if logResetterExist {
		logResetIntr <- struct{}{}
	}
	logResetterExist = true
	select {
	case <-time.After(dr):
		log.Info().Msgf("Returning log level to %s", dlvl.String())
		zerolog.SetGlobalLevel(dlvl)
		break
	case <-logResetIntr:
		break
	}
	logResetterExist = false
}

func LoggerSetLvl(newlvl string, duration time.Duration) string {
	nlvl := LoggerStr2Lvl(newlvl)
	log.Info().Msgf("Setting log level to %s", nlvl.String())
	zerolog.SetGlobalLevel(nlvl)
	go logResetter(duration)
	return nlvl.String()
}
