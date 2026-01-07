package readline

import (
	"syscall"
	"unsafe"
)

func getTermios() (*syscall.Termios, error) {
	fd := uintptr(syscall.Stdin)
	termios := new(syscall.Termios)
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCGETA, uintptr(unsafe.Pointer(termios)), 0, 0, 0)
	if err != 0 {
		return nil, err
	}
	return termios, nil
}

func setTermios(termios *syscall.Termios) error {
	fd := uintptr(syscall.Stdin)
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TIOCSETA, uintptr(unsafe.Pointer(termios)), 0, 0, 0)
	if err != 0 {
		return err
	}
	return nil
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
