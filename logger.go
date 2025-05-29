package ngebut

import (
	"os"

	"github.com/ryanbekhen/ngebut/log"
)

var (
	// logger is the global logger instance
	logger *log.Logger
)

// initLogger initializes the logger with the given log level
func initLogger(level log.Level) {
	// Set up pretty logging for development
	console := log.DefaultConsoleWriter()
	console.Out = os.Stdout

	// Create a new logger with the console writer
	logger = log.New(console, log.InfoLevel)

	// Set the log level
	switch level {
	case log.DebugLevel:
		logger.SetLevel(log.DebugLevel)
	case log.InfoLevel:
		logger.SetLevel(log.InfoLevel)
	case log.WarnLevel:
		logger.SetLevel(log.WarnLevel)
	case log.ErrorLevel:
		logger.SetLevel(log.ErrorLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}

	// Set the default logger
	log.SetOutput(console)
	log.SetLevel(logger.GetLevel())
}

// displayStartupMessage displays a startup message with server information
func displayStartupMessage(addr string) {
	logger.Info().Msg("  _   _            _           _")
	logger.Info().Msg(" | \\ | | __ _  ___| |__  _   _| |_ ")
	logger.Info().Msg(" |  \\| |/ _` |/ _ \\ '_ \\| | | | __|")
	logger.Info().Msg(" | |\\  | (_| |  __/ |_) | |_| | |_ ")
	logger.Info().Msg(" |_| \\_|\\__, |\\___|_.__/ \\__,_|\\__|")
	logger.Info().Msg("        |___/")
	logger.Info().Msg(" ")
	logger.Info().Msgf("Server is running on %s", addr)
	logger.Info().Msg("Press Ctrl+C to stop the server")
	logger.Info().Msg(" ")
}
