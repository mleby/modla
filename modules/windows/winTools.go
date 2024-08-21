package main

import (
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

type RECT struct {
	Left, Top, Right, Bottom int32
}

var vda = syscall.NewLazyDLL("VirtualDesktopAccessor.dll")
var (
	user32                  = windows.NewLazyDLL("user32.dll")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGetWindowText       = user32.NewProc("GetWindowTextW")
	procGetWindowTextLength = user32.NewProc("GetWindowTextLengthW")
)

func IsFullScreen(hwnd win.HWND) bool {
	a := GetWindowDimensions(hwnd)
	b := GetWindowDimensions(win.GetDesktopWindow())
	result := (a.Left == b.Left) &&
		(a.Top == b.Top) &&
		(a.Right == b.Right) &&
		(a.Bottom == b.Bottom)
	return result
}

func WinActivate(hwnd win.HWND) {
	if win.IsIconic(hwnd) {
		win.ShowWindow(hwnd, win.SW_RESTORE)
	}
	win.SetForegroundWindow(hwnd)
	win.SetActiveWindow(hwnd)
}

func GetWindowThreadProcessId(hwnd uintptr) uintptr {
	var prcsId uintptr = 0
	us32 := syscall.MustLoadDLL("user32.dll")
	prc := us32.MustFindProc("GetWindowThreadProcessId")
	_, _, _ = prc.Call(hwnd, uintptr(unsafe.Pointer(&prcsId)))
	//log.Println("ProcessId: ", prcsId, "ret", ret, " Message: ", err)
	return prcsId
}

func GetWindowDimensions(hwnd win.HWND) *RECT {
	var rect RECT

	us32 := syscall.MustLoadDLL("user32.dll")
	prc := us32.MustFindProc("GetWindowRect")

	_, _, _ = prc.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rect)), 0)

	return &rect
}

//func GetWindowProcessPath(hWindow win.HWND) string {
//
//}

func IsPinnedWindow(hwnd win.HWND) int {
	var IsPinnedWindow = vda.NewProc("IsPinnedWindow") // : Integer;
	r1, _, _ := IsPinnedWindow.Call(uintptr(hwnd))
	isPinnedWindow := int(r1)
	return isPinnedWindow
}

func GetWindowDesktopNumber(hwnd win.HWND) int {
	var GetWindowDesktopNumber = vda.NewProc("GetWindowDesktopNumber") // : Integer;
	r1, _, _ := GetWindowDesktopNumber.Call(uintptr(hwnd))
	windowDesktopNumber := int(r1)
	return windowDesktopNumber
}

func GetCurrentDesktopNumber() int {
	var GetCurrentDesktopNumber = vda.NewProc("GetCurrentDesktopNumber") // : Integer;
	r1, _, _ := GetCurrentDesktopNumber.Call()
	currentDesktopNumber := int(r1)
	return currentDesktopNumber
}

func GetWindowTextLength(hwnd win.HWND) int {
	ret, _, _ := procGetWindowTextLength.Call(
		uintptr(hwnd))

	return int(ret)
}

func GetWindowText(hwnd win.HWND) string {
	textLen := GetWindowTextLength(hwnd) + 1

	buf := make([]uint16, textLen)
	procGetWindowText.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(textLen))

	return syscall.UTF16ToString(buf)
}

//func getWindow(funcName string) uintptr {
//    proc := mod.NewProc(funcName)
//    hwnd, _, _ := proc.Call()
//    return hwnd
//}
