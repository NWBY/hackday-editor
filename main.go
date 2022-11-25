package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/sys/unix"
)

/*** data ***/
var origTermios *unix.Termios

/*** terminal ***/
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

/*** init ***/
func main() {
	enableRawMode()
	defer disableRawMode()
	buffer := make([]byte, 1)
	var cc int
	var err error

	for cc, err = os.Stdin.Read(buffer); buffer[0] != 'q' && cc >= 0; cc, err = os.Stdin.Read(buffer) {
		if buffer[0] > 20 && buffer[0] < 0x7f {
			fmt.Printf("%3d %d %c\r\n", buffer[0], buffer[0], cc)
		} else {
			fmt.Printf("%3d %d\r\n", buffer[0], cc)
		}
		buffer[0] = 0
	}
	if err != nil {
		disableRawMode()
		log.Fatal(err)
	}
}
