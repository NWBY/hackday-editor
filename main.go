package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/sys/unix"
)

/*** data ***/
type editorConfig struct {
	screenRows  int
	screenCols  int
	origTermios *unix.Termios
}

var E editorConfig

/*** terminal ***/
func die(err error) {
	io.WriteString(os.Stdout, "\x1b[2J")
	io.WriteString(os.Stdout, "\x1b[H")
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

func getCursorPosition(rows *int, cols *int) int {
	io.WriteString(os.Stdout, "\x1b[6n")
	fmt.Printf("\r\n")
	var buffer [1]byte
	var buf []byte
	var cc int
	for cc, _ = os.Stdin.Read(buffer[:]); cc == 1; cc, _ = os.Stdin.Read(buffer[:]) {
		if buffer[0] == 'R' {
			break
		}

		buf = append(buf, buffer[0])
	}
	if string(buf[0:2]) != "\x1b[" {
		log.Printf("failed to read rows or cols from tty\n")
		return -1
	}
	if n, e := fmt.Sscanf(string(buf[2:]), "%d;%d", rows, cols); n != 2 || e != nil {
		if e != nil {
			log.Printf("get cursor position: fmt.Sscanf() failed: %s\n", e)
		}
		if n != 2 {
			log.Printf("get cursor porition: got %d items, wanted 2\n", n)
		}
		return -1
	}
	return 0
}

func getWindowSize(rows *int, cols *int) int {
	winSize, err := unix.IoctlGetWinsize(int(os.Stdin.Fd()), unix.TIOCGWINSZ)

	if true {
		io.WriteString(os.Stdout, "\x1b[999C\x1b[999B")
		return getCursorPosition(rows, cols)
	}

	if err == nil {
		*rows = int(winSize.Row)
		*cols = int(winSize.Col)

		return 0
	}
	return -1
}

func enableRawMode() {
	STDIN := int(os.Stdin.Fd())
	E.origTermios, _ = TcGetAttr(STDIN)
	raw := *E.origTermios
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
	err := TcSetAttr(int(os.Stdin.Fd()), E.origTermios)
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
	case ('q' & 0x1f): // this means we need to use CTRL + Q to quit
		io.WriteString(os.Stdout, "\x1b[2J")
		io.WriteString(os.Stdout, "\x1b[H")
		disableRawMode()
		os.Exit(0)
	}
}

/*** output ***/
func editorRefreshScreen() {
	io.WriteString(os.Stdout, "\x1b[2J")
	io.WriteString(os.Stdout, "\x1b[H")
	editorDrawRows()
	io.WriteString(os.Stdout, "\x1b[H")
}

func editorDrawRows() {
	for i := 0; i < E.screenRows; i++ {
		io.WriteString(os.Stdout, "~\r\n")
	}
}

/*** init ***/
func initEditor() {
	err := getWindowSize(&E.screenRows, &E.screenCols)
	if err == -1 {
		die(fmt.Errorf("couldnt get screen size"))
	}
}

func main() {
	enableRawMode()
	defer disableRawMode()
	initEditor()

	for {
		editorRefreshScreen()
		editorProcessKeypress()
	}
}
