package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var kernel, rootfs string

func main() {
	flag.StringVar(&kernel, "kernel", "linux", "absolute path to the UML kernel")
	flag.StringVar(&rootfs, "rootfs", "rootfs.raw", "absolute path to the root filesystem file")
	flag.Parse()

	binpath := flag.Arg(0)
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	bindir := filepath.Dir(binpath)
	bin := filepath.Base(binpath)
	hostname := strings.TrimSuffix(bin, filepath.Ext(binpath))

	// create pipes to attach to interact with the console
	ustdin, hstdin, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	hstdout, ustdout, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	// an additional pipe to transmit the suite's exit code
	hexit, uexit, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	// ensure all reads of Linux's stdout are copied to our stdout
	tstdout := io.TeeReader(hstdout, os.Stdout)

	umlAttrs := &os.ProcAttr{
		Files: []*os.File{ustdin, ustdout, os.Stderr, uexit},
	}

	uml, err := os.StartProcess(
		kernel,
		[]string{
			"linux",

			// To allow multiple instances of UML simultaneously, configure a
			// rootfs with a copy-on-write file in the bindir
			fmt.Sprintf("ubd0=%s.cow,%s", binpath, rootfs),
			"rw",

			// Disable automatic console bindings.
			"con=null",

			// Bind the output of console 0 to stderr, to capture kernel messages.
			"con0=null,fd:2",

			// Bind input and output of console 1 to our pipes.
			"con1=fd:0,fd:1",

			// Bind serial port 1 to capture the process exit code.
			"ssl1=fd:3",

			fmt.Sprintf("systemd.hostname=%s", hostname),
			"systemd.tty.term.tty1=dumb",

			// mount bindir and working directory
			fmt.Sprintf("systemd.mount-extra=none:/mnt/tmp:hostfs:%s", bindir),
			fmt.Sprintf("systemd.mount-extra=none:/mnt/work:hostfs:%s", pwd),
		},
		umlAttrs,
	)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(tstdout)
	scanner.Split(SplitPrompt)

	exitCh := make(chan int)

	go func() {
		// log the UML kernel information
		scanner.Scan()
		io.WriteString(hstdin, "uname -a\n")

		// change to working directory
		scanner.Scan()
		io.WriteString(hstdin, "cd /mnt/work\n")

		// run the process!
		scanner.Scan()
		fmt.Fprintf(hstdin, "/mnt/tmp/%s %s\n", bin, strings.Join(flag.Args()[2:], " "))

		// write the exit code to the serial port
		scanner.Scan()
		io.WriteString(hstdin, "echo $? > /dev/ttyS1\n")

		scanner.Scan()
		io.WriteString(hstdin, "poweroff\n")

		// read the rest of stdout so it gets tee'd
		io.ReadAll(tstdout)
	}()

	go func() {
		b := make([]byte, 1)
		hexit.Read(b)
		c, _ := strconv.Atoi(string(b))
		exitCh <- c
	}()

	if _, err := uml.Wait(); err != nil {
		panic(err)
	}

	os.Exit(<-exitCh)
}

// SplitPrompt is a split function for a Scanner that returns a token if a "dumb" prompt is detected.
// A dumb prompt would match the RegExp `\n\$` (a newline followed by a dollar sign), which is commonly
// used when TERM is "dumb".
func SplitPrompt(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.LastIndexByte(data, '\n'); i >= 0 {
		if len(data) > i+1 {
			if data[i+1] == '$' {
				return i + 2, []byte{0}, nil
			}
			return i + 1, nil, nil
		}
	}

	if atEOF {
		return len(data), nil, nil
	}

	return 0, nil, nil
}
