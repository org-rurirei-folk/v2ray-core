package log

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/v2fly/v2ray-core/v4/common/platform"
	"github.com/v2fly/v2ray-core/v4/common/signal/done"
	"github.com/v2fly/v2ray-core/v4/common/signal/semaphore"
)

// Printer is the interface for logger.Print()
type Printer interface {
	Print(v ...interface{})
}

// UseAlternativeLogPrinter provides a alternative Printer
var UseAlternativeLogPrinter = func() Printer {
	return nil
}

// Writer is the interface for writing logs.
type Writer interface {
	Write(string) error
	io.Closer
}

// WriterCreator is a function to create LogWriters.
type WriterCreator func() Writer

type generalLogger struct {
	creator WriterCreator
	buffer  chan Message
	access  *semaphore.Instance
	done    *done.Instance
}

// NewLogger returns a generic log handler that can handle all type of messages.
func NewLogger(logWriterCreator WriterCreator) Handler {
	return &generalLogger{
		creator: logWriterCreator,
		buffer:  make(chan Message, 16),
		access:  semaphore.New(1),
		done:    done.New(),
	}
}

func (l *generalLogger) run() {
	defer l.access.Signal()

	dataWritten := false
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	logger := l.creator()
	if logger == nil {
		return
	}
	defer logger.Close()

	for {
		select {
		case <-l.done.Wait():
			return
		case msg := <-l.buffer:
			logger.Write(msg.String() + platform.LineSeparator())
			dataWritten = true
		case <-ticker.C:
			if !dataWritten {
				return
			}
			dataWritten = false
		}
	}
}

func (l *generalLogger) Handle(msg Message) {
	select {
	case l.buffer <- msg:
	default:
	}

	select {
	case <-l.access.Wait():
		go l.run()
	default:
	}
}

func (l *generalLogger) Close() error {
	return l.done.Close()
}

type consoleLogWriter struct {
	Printer
}

func (w *consoleLogWriter) Write(s string) error {
	w.Print(s)
	return nil
}

func (w *consoleLogWriter) Close() error {
	return nil
}

type fileLogWriter struct {
	file *os.File
	Printer
}

func (w *fileLogWriter) Write(s string) error {
	w.Print(s)
	return nil
}

func (w *fileLogWriter) Close() error {
	return w.file.Close()
}

// CreateStdoutLogWriter returns a LogWriterCreator that creates LogWriter for stdout.
func CreateStdoutLogWriter() WriterCreator {
	return func() Writer {
		printer := Printer(log.New(os.Stdout, "", log.Ldate|log.Ltime))
		if prt := UseAlternativeLogPrinter(); prt != nil {
			printer = prt
		}
		return &consoleLogWriter{
			Printer: printer,
		}
	}
}

// CreateStderrLogWriter returns a LogWriterCreator that creates LogWriter for stderr.
func CreateStderrLogWriter() WriterCreator {
	return func() Writer {
		printer := Printer(log.New(os.Stderr, "", log.Ldate|log.Ltime))
		if prt := UseAlternativeLogPrinter(); prt != nil {
			printer = prt
		}
		return &consoleLogWriter{
			Printer: printer,
		}
	}
}

// CreateFileLogWriter returns a LogWriterCreator that creates LogWriter for the given file.
func CreateFileLogWriter(path string) WriterCreator {
	return func() Writer {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
		if err != nil {
			return nil
		}
		printer := Printer(log.New(file, "", log.Ldate|log.Ltime))
		if prt := UseAlternativeLogPrinter(); prt != nil {
			printer = prt
		}
		return &fileLogWriter{
			file:    file,
			Printer: printer,
		}
	}
}

func init() {
	RegisterHandler(NewLogger(CreateStdoutLogWriter()))
}
