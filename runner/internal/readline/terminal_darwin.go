package readline

import (
	"syscall"
	"unsafe"
)

type Termios syscall.Termios

func getTermios() (*Termios, error) {
	fd := uintptr(syscall.Stdin)
	termios := new(Termios)
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCGETA, uintptr(unsafe.Pointer(termios)), 0, 0, 0)
	if err != 0 {
		return nil, err
	}
	return termios, nil
}

func setTermios(termios *Termios) error {
	fd := uintptr(syscall.Stdin)
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TIOCSETA, uintptr(unsafe.Pointer(termios)), 0, 0, 0)
	if err != 0 {
		return err
	}
	return nil
}

func applyRawMode(termios *Termios) {
	termios.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	termios.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	termios.Cflag &^= syscall.CSIZE | syscall.PARENB
	termios.Cflag |= syscall.CS8
	termios.Cc[syscall.VMIN] = 1
	termios.Cc[syscall.VTIME] = 0
}

func (t *Terminal) GetWidth() (int, error) {
	ws := &struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}{}

	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(syscall.Stdout), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(ws)))
	if err != 0 {
		return 0, err
	}

	return int(ws.Col), nil
}
