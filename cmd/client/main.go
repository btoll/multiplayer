package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/term"
)

var logFile, _ = os.Create("debug.log")

func main() {
	conn, err := net.Dial("tcp", ":1111")
	if err != nil {
		panic(err)
	}

	for {
		buf := make([]byte, 128)
		n, _ := conn.Read(buf)
		fmt.Printf("%s", string(buf[:n]))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		fmt.Fprintln(conn, scanner.Text())
		_, _ = conn.Read(buf)
		break
	}

	oldState, _ := term.MakeRaw(int(os.Stdin.Fd()))
	defer func() {
		fmt.Fprintln(logFile, time.Now(), "restoring state")
		term.Restore(int(os.Stdin.Fd()), oldState)
	}()

	go func() {
		buf := make([]byte, 3)
		seq := make([]byte, 0, 3)
		for {
			n, _ := os.Stdin.Read(buf)
			seq = append(seq, buf[:n]...)
			fmt.Fprintf(logFile, "seq=%q\n", string(seq))
			fmt.Fprintf(conn, "%s\n", seq)
			seq = seq[:0] // Zero-out the length so we can append above.
		}
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	fmt.Fprintln(logFile, time.Now(), "server closed the connection")
	err = conn.Close()
	if err != nil {
		fmt.Fprintln(logFile, err)
	}
}
