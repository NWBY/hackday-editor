package main

import (
	"log"
	"os"

	"golang.org/x/sys/unix"
)

/*** data ***/
var origTermios *unix.Termios

/*** terminal ***/
func die(err error) {
	disableRawMode()
	log.Fatal(err)
}

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
		log.Fatalf("Failed to get terminal attributes: %s\n", err)
		return nil, err
	}

	return termios, nil
}

func enableRawMode() {
	STDIN := int(os.Stdin.Fd())
	origTermios, _ := TcGetAttr(STDIN)
	raw := *origTermios
	raw.Iflag &^= unix.BRKINT | unix.ICRNL | unix.INPCK | unix.ISTRIP | unix.IXON
	raw.Oflag &^= unix.OPOST
	raw.Cflag |= unix.CS8
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.IEXTEN | unix.ISIG
	raw.Cc[unix.VMIN+1] = 0
	raw.Cc[unix.VTIME+1] = 1
	err := TcSetAttr(STDIN, &raw)
	if err != nil {
		log.Fatalf("Failed to enable raw mode: %s\n", err)
	}
}

func disableRawMode() {
	err := TcSetAttr(int(os.Stdin.Fd()), origTermios)
	if err != nil {
		log.Fatalf("Failed to disable raw mode: %s\n", err)
	}
}

func editorReadKey() byte {
	var buffer [1]byte
	var cc int
	var err error

	for cc, err = os.Stdin.Read(buffer[:]); cc != 1; cc, err = os.Stdin.Read(buffer[:]) {
	}
	if err != nil {
		die(err)
	}
	return buffer[0]

}

/*** input ***/
func editorProcessKeypress() {
	c := editorReadKey()
	switch c {
	case ('q' & 0x1f):
		disableRawMode()
		os.Exit(0)
	}
}

/*** init ***/
func main() {
	enableRawMode()
	defer disableRawMode()

	for {
		editorProcessKeypress()
	}
}
