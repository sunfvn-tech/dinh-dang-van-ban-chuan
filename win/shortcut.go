//go:build windows

package win

import (
	"fmt"
	"syscall"
	"unsafe"
)

// GUID là định danh COM.
type GUID struct {
	D1 uint32
	D2 uint16
	D3 uint16
	D4 [8]byte
}

var (
	clsidShellLink  = GUID{0x00021401, 0, 0, [8]byte{0xC0, 0, 0, 0, 0, 0, 0, 0x46}}
	iidIShellLinkW  = GUID{0x000214F9, 0, 0, [8]byte{0xC0, 0, 0, 0, 0, 0, 0, 0x46}}
	iidIPersistFile = GUID{0x0000010b, 0, 0, [8]byte{0xC0, 0, 0, 0, 0, 0, 0, 0x46}}

	// Thư mục đặc biệt của người dùng hiện tại
	FolderPrograms = GUID{0xA77F5D77, 0x2E2B, 0x44C3, [8]byte{0xA6, 0xA2, 0xAB, 0xA6, 0x01, 0x05, 0x4A, 0x51}}
	FolderSendTo   = GUID{0x8983036C, 0x27C0, 0x404B, [8]byte{0x8F, 0x08, 0x10, 0x2D, 0x10, 0xDC, 0xFD, 0x74}}
	FolderDesktop  = GUID{0xB4BFCC3A, 0xDB2C, 0x424C, [8]byte{0xB0, 0x29, 0x7F, 0xE9, 0x9A, 0x87, 0xC6, 0x41}}

	pCoCreateInstance      = ole32.NewProc("CoCreateInstance")
	pSHGetKnownFolderPath  = shell32.NewProc("SHGetKnownFolderPath")
)

// KnownFolder trả đường dẫn thư mục đặc biệt (Unicode, theo cả chuyển hướng OneDrive).
func KnownFolder(id GUID) (string, error) {
	var p *uint16
	r, _, _ := pSHGetKnownFolderPath.Call(uintptr(unsafe.Pointer(&id)), 0, 0, uintptr(unsafe.Pointer(&p)))
	if r != 0 {
		return "", fmt.Errorf("SHGetKnownFolderPath 0x%x", r)
	}
	defer pCoTaskMemFree.Call(uintptr(unsafe.Pointer(p)))
	return utf16PtrToString(p), nil
}

func utf16PtrToString(p *uint16) string {
	if p == nil {
		return ""
	}
	var s []uint16
	for ptr := unsafe.Pointer(p); ; ptr = unsafe.Add(ptr, 2) {
		c := *(*uint16)(ptr)
		if c == 0 {
			break
		}
		s = append(s, c)
	}
	return syscall.UTF16ToString(s)
}

func vcall(obj uintptr, idx int, args ...uintptr) uintptr {
	vtbl := *(*uintptr)(unsafe.Pointer(obj))
	fn := *(*uintptr)(unsafe.Pointer(vtbl + uintptr(idx)*unsafe.Sizeof(uintptr(0))))
	r, _, _ := syscall.SyscallN(fn, append([]uintptr{obj}, args...)...)
	return r
}

// CreateShortcut tạo file .lnk bằng IShellLinkW (hỗ trợ tên tiếng Việt có dấu).
func CreateShortcut(lnkPath, target, workDir, desc, icon string, iconIdx int) error {
	pCoInitializeEx.Call(0, 2)
	var sl uintptr
	r, _, _ := pCoCreateInstance.Call(uintptr(unsafe.Pointer(&clsidShellLink)), 0, 1 /*CLSCTX_INPROC_SERVER*/, uintptr(unsafe.Pointer(&iidIShellLinkW)), uintptr(unsafe.Pointer(&sl)))
	if r != 0 || sl == 0 {
		return fmt.Errorf("không khởi tạo được ShellLink (0x%x)", r)
	}
	defer vcall(sl, 2) // Release
	vcall(sl, 20, U(target))           // SetPath
	vcall(sl, 9, U(workDir))           // SetWorkingDirectory
	vcall(sl, 7, U(desc))              // SetDescription
	vcall(sl, 17, U(icon), uintptr(iconIdx)) // SetIconLocation
	var pf uintptr
	if r := vcall(sl, 0, uintptr(unsafe.Pointer(&iidIPersistFile)), uintptr(unsafe.Pointer(&pf))); r != 0 || pf == 0 {
		return fmt.Errorf("không lấy được IPersistFile (0x%x)", r)
	}
	defer vcall(pf, 2)
	if r := vcall(pf, 6, U(lnkPath), 1); r != 0 { // Save
		return fmt.Errorf("không lưu được lối tắt %s (0x%x)", lnkPath, r)
	}
	return nil
}
