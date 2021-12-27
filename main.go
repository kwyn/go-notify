package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/gregdel/pushover"
)

const (
	APIKeyEnvVariable           = "PUSHOVER_API_TOKEN"
	UserTokenEnvVariable        = "PUSHOVER_USER_TOKEN"
	GoNotifyDebugEnvVariable    = "GO_NOTIFY_DEBUG"
	GoNotifySkipSendEnvVariable = "GO_NOTIFY_SKIP_SEND"
)

// Logger interface for basic logs
type Logger interface {
	Log(string)
	Logf(string, ...interface{})
	Enable()
}

type logger struct {
	enabled bool
	out     io.Writer
}

// NewLogger creates a new logger that satisfies the Logger interface.
// Defaults to print to os.Stderr
func NewLogger() Logger {
	return &logger{
		enabled: false,
		out:     os.Stderr,
	}
}

func (l *logger) Enable() {
	l.enabled = true
}

func (l *logger) Log(s string) {
	if l.enabled {
		fmt.Fprint(l.out, s+"\n")
	}
}

func (l *logger) Logf(s string, args ...interface{}) {
	if l.enabled {
		fmt.Fprintf(l.out, s+"\n", args)
	}
}

func main() {
	token := os.Getenv(APIKeyEnvVariable)
	user := os.Getenv(UserTokenEnvVariable)
	if token == "" || user == "" {
		panic(fmt.Sprintf("no api or user tokens found ensure that environment %s and %s variables are set", APIKeyEnvVariable, UserTokenEnvVariable))
	}

	log := NewLogger()

	debug := os.Getenv(GoNotifyDebugEnvVariable)
	if debug == "true" {
		log.Enable()
	}
	cmd := os.Args[1:]

	if len(cmd) < 1 {
		panic(fmt.Sprintf("No cmd provided to %s", os.Args[0]))
	}
	c := exec.Command(cmd[0], cmd[1:]...)

	var stdout bytes.Buffer
	var combinedOut bytes.Buffer
	c.Stderr = io.MultiWriter(&combinedOut, os.Stderr)
	c.Stdout = io.MultiWriter(&combinedOut, &stdout, os.Stdout)

	app := pushover.New(token)
	recipient := pushover.NewRecipient(user)

	message := &pushover.Message{}

	_ = c.Run()
	exitCode := c.ProcessState.ExitCode()
	time := c.ProcessState.UserTime()
	prefix := fmt.Sprintf("✅ (%s): ", exitCode, time)
	cmdStr := c.String()
	if exitCode != 0 {
		prefix = fmt.Sprintf("❌ exit %v (%s): ", exitCode, time)
		s := combinedOut.String()
		if len(s) > pushover.MessageMaxLength {
			message.AddAttachment(&combinedOut)
		}
		// TODO: handle too large of an attachment
		message.Message = combinedOut.String()
	} else {
		s := stdout.String()
		if len(s) > pushover.MessageMaxLength {
			message.AddAttachment(&stdout)
		}
		// TODO: handle too large of an attachment
		message.Message = stdout.String()
	}

	if len(cmdStr)+len(prefix) > pushover.MessageTitleMaxLength {
		truncationString := " (sic...)"
		trimIndex := pushover.MessageTitleMaxLength - len(prefix) - len(truncationString)
		truncation := prefix + cmdStr[:trimIndex] + truncationString
		message.Title = truncation
	} else {
		message.Title = prefix + cmdStr
	}

	if os.Getenv(GoNotifySkipSendEnvVariable) == "true" {
		fmt.Fprintf(os.Stderr, "%s=true skipping send step\n", GoNotifySkipSendEnvVariable)
		return
	}

	resp, err := app.SendMessage(message, recipient)
	if err != nil {
		panic(fmt.Errorf("Failed to send message: %w", err))
	}

	log.Logf("%s", resp)
}
