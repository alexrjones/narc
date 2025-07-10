package idle

import "C"
import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

var (
	procDefWindowProc = user32.NewProc("DefWindowProcW")
)

const (
	WM_POWERBROADCAST      = 0x0218
	PBT_APMSUSPEND         = 0x0004
	PBT_APMRESUMEAUTOMATIC = 0x0012
)

type syscallCallback = func(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr

func makeSleepwatcher(onSleepStateChange func(awake bool)) syscallCallback {
	return func(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
		switch msg {
		case WM_POWERBROADCAST:
			switch wParam {
			case PBT_APMSUSPEND:
				onSleepStateChange(false)
			case PBT_APMRESUMEAUTOMATIC:
				onSleepStateChange(true)
			}
		}
		ret, _, _ := procDefWindowProc.Call(
			uintptr(hwnd), uintptr(msg), wParam, lParam)
		return ret
	}
}

func wndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_POWERBROADCAST:
		switch wParam {
		case PBT_APMSUSPEND:
			fmt.Println("[Windows] System is going to sleep", time.Now())
		case PBT_APMRESUMEAUTOMATIC:
			fmt.Println("[Windows] System just woke up", time.Now())
		}
	}
	ret, _, _ := procDefWindowProc.Call(
		uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
}

func runMessageLoop(cb any) {
	// This creates a minimal window to receive messages
	modUser32 := syscall.NewLazyDLL("user32.dll")
	modKernel32 := syscall.NewLazyDLL("kernel32.dll")
	procCreateWindowEx := modUser32.NewProc("CreateWindowExW")
	procRegisterClassEx := modUser32.NewProc("RegisterClassExW")
	procGetModuleHandle := modKernel32.NewProc("GetModuleHandleW")
	procGetMessage := modUser32.NewProc("GetMessageW")
	procTranslateMessage := modUser32.NewProc("TranslateMessage")
	procDispatchMessage := modUser32.NewProc("DispatchMessageW")

	type WNDCLASSEX struct {
		cbSize        uint32
		style         uint32
		lpfnWndProc   uintptr
		cbClsExtra    int32
		cbWndExtra    int32
		hInstance     syscall.Handle
		hIcon         syscall.Handle
		hCursor       syscall.Handle
		hbrBackground syscall.Handle
		lpszMenuName  *uint16
		lpszClassName *uint16
		hIconSm       syscall.Handle
	}

	className, _ := syscall.UTF16PtrFromString("MySleepWatcher")
	hInstance, _, _ := procGetModuleHandle.Call(0)

	var wc WNDCLASSEX
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	wc.lpfnWndProc = syscall.NewCallback(cb)
	wc.hInstance = syscall.Handle(hInstance)
	wc.lpszClassName = className

	procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))

	// Create invisible message-only window
	procCreateWindowEx.Call(
		0, uintptr(unsafe.Pointer(className)), 0, 0,
		0, 0, 0, 0, 0, 0, uintptr(hInstance), 0,
	)

	var msg [4]uintptr
	for {
		r, _, _ := procGetMessage.Call(
			uintptr(unsafe.Pointer(&msg[0])), 0, 0, 0)
		if int32(r) == -1 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg[0])))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg[0])))
	}
}

func StartSleepWatcher(onSleepStateChange func(bool)) {
	runMessageLoop(makeSleepwatcher(onSleepStateChange))
}
