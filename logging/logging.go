package logging

import (
	"io/ioutil"
	"log"
	"os"
)

var (
	VerboseFlag bool
	Info        *log.Logger
	Verbose     *log.Logger
	Error       *log.Logger
)

// setupLogs provides the following logs
// Info = stdout
// Verbose = optional verbosity
// Error = stderr
func SetupLogs() {
	Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	if VerboseFlag {
		Verbose = log.New(os.Stdout, "VERBOSE: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		Verbose = log.New(ioutil.Discard, "VERBOSE: ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	Error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
