package main

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

var origTermios *unix.Termios

func TcSetAttr(fd int, termios *unix.Termios) error {
	err := unix.IoctlSetTermios(fd, unix.TIOCSETA+1, termios)
	if err != nil {
		return err
	}

	return nil
}

func TcGetAttr(fd int) (*unix.Termios, error) {
	termios, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		return nil, err
	}

	return termios, nil
}

func enableRawMode() {
	STDIN := int(os.Stdin.Fd())
	origTermios, _ := TcGetAttr(STDIN)
	raw := *origTermios
	raw.Iflag &^= unix.IXON | unix.ICRNL
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.IEXTEN | unix.ISIG
	err := TcSetAttr(STDIN, &raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to enable raw mode: %s\n", err)
	}
}

func disableRawMode() {
	err := TcSetAttr(int(os.Stdin.Fd()), origTermios)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Problem disabling raw mode: %s\n", err)
	}
}

func main() {
	enableRawMode()
	defer disableRawMode()
	buffer := make([]byte, 1)
	for cc, err := os.Stdin.Read(buffer); buffer[0] != 'q' && err == nil && cc == 1; cc, err = os.Stdin.Read(buffer) {
		r := rune(buffer[0])
		if strconv.IsPrint(r) {
			fmt.Printf("%d %c\n", buffer[0], r)
		} else {
			fmt.Printf("%d\n", buffer[0])
		}
	}
	os.Exit(0)
}
