// study this very well later
package main

// import (
// 	"bufio"
// 	"bytes"
// 	"context"
// 	"encoding/hex"
// 	"encoding/json"
// 	"flag"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net"
// 	"os"
// 	"os/signal"
// 	"runtime"
// 	"strings"
// 	"sync"
// 	"sync/atomic"
// 	"syscall"
// 	"time"
// 	"unicode/utf8"
// )

// // ANSI colors for terminal per-connection
// var colors = []string{
// 	"\x1b[31m", // red
// 	"\x1b[32m", // green
// 	"\x1b[33m", // yellow
// 	"\x1b[34m", // blue
// 	"\x1b[35m", // magenta
// 	"\x1b[36m", // cyan
// 	"\x1b[37m", // white
// }

// const resetColor = "\x1b[0m"

// // Message is a logger that also implements io.Writer so it can be used in TeeReaders/MultiWriters.
// type Message struct {
// 	*log.Logger
// 	colorPrefix string // ANSI color sequence for terminal
// 	toFileOnly  bool   // if true, don't colorize output (useful for file logger)
// }

// // Write implements io.Writer so Message can be used in io.TeeReader, io.MultiWriter, etc.
// func (m *Message) Write(p []byte) (int, error) {
// 	// Keep writes atomic-ish by printing as a single log entry
// 	msg := string(bytes.TrimRight(p, "\n")) // avoid double newlines
// 	if m.toFileOnly {
// 		_ = m.Output(3, msg)
// 		return len(p), nil
// 	}
// 	// Colorize and include caller info
// 	colored := fmt.Sprintf("%s%s%s", m.colorPrefix, msg, resetColor)
// 	_ = m.Output(3, colored)
// 	return len(p), nil
// }

// // createGlobalLogger creates a file-backed logger and returns an io.Writer for file logging.
// func createGlobalLogger(logFile string) (fileWriter io.Writer, cleanup func(), err error) {
// 	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	// global file logger (no color)
// 	fileLogger := log.New(f, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
// 	// wrap into Message if you need writer interface, but here we return file writer.
// 	cleanup = func() { _ = f.Close() }
// 	return fileWriterWithLogger{f: f, logger: fileLogger}, cleanup, nil
// }

// // small wrapper so we can use file logger with Output (preserve file caller lines)
// type fileWriterWithLogger struct {
// 	f      *os.File
// 	logger *log.Logger
// }

// func (fw fileWriterWithLogger) Write(p []byte) (int, error) {
// 	// Use logger.Output so file gets timestamp + file:line info
// 	s := string(bytes.TrimRight(p, "\n"))
// 	_ = fw.logger.Output(3, s)
// 	return len(p), nil
// }

// var connCounter uint64
// var bufPool = sync.Pool{
// 	New: func() any { b := make([]byte, 4096); return &b },
// }

// func RunLogger() {
// 	var addr = flag.String("listen", "127.0.0.1:9090", "listen address")
// 	var logfile = flag.String("log", "server.log", "log file path")
// 	flag.Parse()

// 	fileWriter, cleanup, err := createGlobalLogger(*logfile)
// 	if err != nil {
// 		log.Fatalf("cannot create log file: %v", err)
// 	}
// 	defer cleanup()

// 	// consoleLogger: base logger that writes to stdout (colored per conn via Message)
// 	consoleBase := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)

// 	// Start listener
// 	ln, err := net.Listen("tcp", *addr)
// 	if err != nil {
// 		consoleBase.Fatalf("listen error: %v", err)
// 	}
// 	consoleBase.Printf("listening on %s (pid=%d, go=%s)", ln.Addr(), os.Getpid(), runtime.Version())

// 	// context + signal handling for graceful shutdown
// 	ctx, cancel := context.WithCancel(context.Background())
// 	go handleSignals(cancel, consoleBase)

// 	var wg sync.WaitGroup
// 	acceptLoop := func() {
// 		for {
// 			conn, err := ln.Accept()
// 			if err != nil {
// 				select {
// 				case <-ctx.Done():
// 					return
// 				default:
// 					consoleBase.Printf("accept error: %v", err)
// 					continue
// 				}
// 			}
// 			wg.Add(1)
// 			go func(c net.Conn) {
// 				defer wg.Done()
// 				handleConn(ctx, c, consoleBase, fileWriter)
// 			}(conn)
// 		}
// 	}

// 	go acceptLoop()

// 	// Wait until canceled
// 	<-ctx.Done()
// 	consoleBase.Println("shutdown requested — closing listener")
// 	_ = ln.Close()
// 	// wait active connections
// 	wg.Wait()
// 	consoleBase.Println("shutdown complete")
// }

// // handleSignals listens for SIGINT/SIGTERM and cancels context
// func handleSignals(cancel context.CancelFunc, l *log.Logger) {
// 	ch := make(chan os.Signal, 1)
// 	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
// 	sig := <-ch
// 	l.Printf("signal received: %v", sig)
// 	cancel()
// }

// // handleConn manages a single connection lifecycle
// func handleConn(ctx context.Context, conn net.Conn, consoleBase *log.Logger, fileWriter io.Writer) {
// 	id := atomic.AddUint64(&connCounter, 1)
// 	color := colors[id%uint64(len(colors))]
// 	prefix := fmt.Sprintf("[conn:%d %s->%s]", id, conn.RemoteAddr(), conn.LocalAddr())

// 	// per-connection logger that writes colored output to stdout and also to file via MultiWriter
// 	stdoutMsg := &Message{
// 		Logger:      log.New(os.Stdout, prefix+" ", log.LstdFlags|log.Lmicroseconds),
// 		colorPrefix: color,
// 	}
// 	fileMsg := &Message{
// 		Logger:     log.New(io.Discard, prefix+" ", 0), // we use fileWriter below
// 		toFileOnly: true,
// 	}
// 	// combine: stdout colored + file
// 	multi := io.MultiWriter(stdoutMsg, fileWriter)

// 	consoleBase.Printf("accepted %s (id=%d)", conn.RemoteAddr(), id)

// 	// Use buffered reader to avoid small syscalls
// 	r := bufio.NewReader(conn)

// 	defer func() {
// 		_ = conn.Close()
// 		consoleBase.Printf("closed conn id=%d", id)
// 	}()

// 	// read loop with deadlines and buffer pool
// 	for {
// 		// Respect context cancellation
// 		select {
// 		case <-ctx.Done():
// 			return
// 		default:
// 		}

// 		// Set a read deadline so dead connections don't hang forever
// 		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))

// 		bp := bufPool.Get().(*[]byte)
// 		buf := *bp
// 		n, err := r.Read(buf)
// 		if err != nil {
// 			if ne, ok := err.(net.Error); ok && ne.Timeout() {
// 				// continue to next read
// 				bufPool.Put(bp)
// 				continue
// 			}
// 			if err == io.EOF {
// 				bufPool.Put(bp)
// 				return
// 			}
// 			// log and close
// 			fmt.Fprintf(os.Stderr, "read error id=%d: %v\n", id, err)
// 			bufPool.Put(bp)
// 			return
// 		}

// 		payload := make([]byte, n)
// 		copy(payload, buf[:n])
// 		bufPool.Put(bp)

// 		// Detect and format payload for logs
// 		payloadType, formatted := detectAndFormat(payload)

// 		// Compose structured line
// 		meta := fmt.Sprintf("%s dir=%s bytes=%d type=%s", prefix, "recv", n, payloadType)
// 		// Write metadata and the content via multiwriter so both console & file get the same stream
// 		_, _ = multi.Write([]byte(meta + "\n"))
// 		_, _ = multi.Write([]byte(formatted + "\n"))

// 		// Echo back with safeguard (set write deadline)
// 		_ = conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
// 		_, werr := conn.Write(payload)
// 		if werr != nil {
// 			// log write error and return
// 			_, _ = multi.Write([]byte(fmt.Sprintf("%s dir=send error=%v\n", prefix, werr)))
// 			return
// 		}
// 		// log successful send
// 		_, _ = multi.Write([]byte(fmt.Sprintf("%s dir=sent bytes=%d\n", prefix, n)))
// 	}
// }

// // detectAndFormat evaluates payload and returns a short human representation
// func detectAndFormat(b []byte) (typ string, out string) {
// 	if len(b) == 0 {
// 		return "empty", "<empty>"
// 	}
// 	// Check if UTF-8 valid
// 	if utf8.Valid(b) {
// 		// trim and replace newlines for single-line log readability
// 		s := string(b)
// 		// try JSON pretty print if JSON
// 		var js json.RawMessage
// 		if json.Unmarshal(b, &js) == nil {
// 			// pretty print but cap size
// 			var pretty bytes.Buffer
// 			_ = json.Indent(&pretty, b, "", "  ")
// 			txt := compactOrTruncate(pretty.String(), 1000)
// 			return "json", txt
// 		}
// 		// text - sanitize control chars
// 		s = strings.ReplaceAll(s, "\r\n", "\\r\\n")
// 		s = strings.ReplaceAll(s, "\n", "\\n")
// 		s = strings.ReplaceAll(s, "\t", "\\t")
// 		return "text", compactOrTruncate(s, 1000)
// 	}
// 	// Binary - hex dump first N bytes
// 	limit := 256
// 	if len(b) < limit {
// 		limit = len(b)
// 	}
// 	dump := hex.Dump(b[:limit])
// 	if len(b) > limit {
// 		dump += fmt.Sprintf("\n... (total %d bytes, truncated dump)", len(b))
// 	}
// 	return "binary", dump
// }

// // compactOrTruncate returns a one-line compact form or truncated multiline up to maxLen
// func compactOrTruncate(s string, maxLen int) string {
// 	if len(s) > maxLen {
// 		return s[:maxLen] + fmt.Sprintf("... (truncated, total %d bytes)", len(s))
// 	}
// 	return s
// }
