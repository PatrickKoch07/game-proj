package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var LOG zerolog.Logger

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("|%-6s|", i))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("**%s**", i)
	}
	LOG = zerolog.New(output).With().Timestamp().Caller().Logger()
}
