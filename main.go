package main

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func TcSetAttr(fd int, termios *unix.Termios) error {
	err := unix.IoctlSetTermios(fd, unix.TIOCSETA, termios)
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
	raw, _ := TcGetAttr(STDIN)
	raw.Lflag &^= syscall.ECHO
	_ = TcSetAttr(STDIN, raw)
}

func main() {
	enableRawMode()
	buffer := make([]byte, 1)
	for cc, err := os.Stdin.Read(buffer); buffer[0] != 'q' && err == nil && cc == 1; cc, err = os.Stdin.Read(buffer) {
		// ignore
	}
	os.Exit(0)
}
