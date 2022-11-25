package main

import (
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

func getWindowSize() (*unix.Winsize, error) {
	winSize, err := unix.IoctlGetWinsize(int(os.Stdin.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return nil, err
	}

	return winSize, nil
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
	win, err := getWindowSize()
	if err != nil {
		log.Fatalf("Unable to get window size: %s\n", err)
	}

	E.screenCols = int(win.Col)
	E.screenRows = int(win.Row)
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
