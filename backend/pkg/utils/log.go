package utils

import (
	"fmt"
	"log"
	"os"
)

var logFile *os.File
// setupLogFile ensures the logs directory exists and returns a file for logging.
func SetupLogFile() (*os.File, error) {
	// Ensure the logs directory exists
	if _, statErr := os.Stat("logs"); os.IsNotExist(statErr) {
		mkdirErr := os.Mkdir("logs", 0755)
		if mkdirErr != nil {
			return nil,fmt.Errorf("error creating logs directory: %w", mkdirErr)
		}
	}

	file, fileErr := os.OpenFile("logs/application.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if fileErr != nil {
		return  nil,fmt.Errorf("error opening log file: %w", fileErr)
	}
	logFile=file

	defer logFile.Close()

	return logFile,nil
}

// LogInfo logs an informational message to the log file.
func LogInfo(message string) {
	

	logger := log.New(logFile, "[INFO]: ", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println(message)
}

// LogWarning logs a warning message to the log file.
func LogWarning(message string) {
	

	logger := log.New(logFile, "[WARNING]: ", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println(message)
}

// LogError logs an error message to the log file and prints it to the terminal.
func LogError(err error) {
	
	fileLogger := log.New(logFile, "[ERROR]: ", log.Ldate|log.Ltime|log.Lshortfile)
	fileLogger.Println(err.Error())

	// Also print to terminal
	terminalLogger := log.New(os.Stderr, "[ERROR]: ", log.Ldate|log.Ltime|log.Lshortfile)
	terminalLogger.Println(err.Error())
}
