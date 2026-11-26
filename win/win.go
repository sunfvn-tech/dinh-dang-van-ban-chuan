//go:build windows

// Package win: lớp mỏng gọi Win32 API (không dùng thư viện ngoài).
package win

import (
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	comctl32 = syscall.NewLazyDLL("comctl32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")

	pRegisterClassEx    = user32.NewProc("RegisterClassExW")
	pCreateWindowEx     = user32.NewProc("CreateWindowExW")
	pDefWindowProc      = user32.NewProc("DefWindowProcW")
	pGetMessage         = user32.NewProc("GetMessageW")
	pTranslateMessage   = user32.NewProc("TranslateMessage")
	pDispatchMessage    = user32.NewProc("DispatchMessageW")
	pIsDialogMessage    = user32.NewProc("IsDialogMessageW")
	pPostQuitMessage    = user32.NewProc("PostQuitMessage")
	pSendMessage        = user32.NewProc("SendMessageW")
	pPostMessage        = user32.NewProc("PostMessageW")
	pShowWindow         = user32.NewProc("ShowWindow")
	pUpdateWindow       = user32.NewProc("UpdateWindow")
	pMoveWindow         = user32.NewProc("MoveWindow")
	pSetWindowText      = user32.NewProc("SetWindowTextW")
	pGetWindowText      = user32.NewProc("GetWindowTextW")
	pGetWindowTextLen   = user32.NewProc("GetWindowTextLengthW")
	pEnableWindow       = user32.NewProc("EnableWindow")
	pMessageBox         = user32.NewProc("MessageBoxW")
	pLoadCursor         = user32.NewProc("LoadCursorW")
	pLoadImage          = user32.NewProc("LoadImageW")
	pGetClientRect      = user32.NewProc("GetClientRect")
	pGetWindowRect      = user32.NewProc("GetWindowRect")
	pGetSystemMetrics   = user32.NewProc("GetSystemMetrics")
	pGetDC              = user32.NewProc("GetDC")
	pReleaseDC          = user32.NewProc("ReleaseDC")
	pSetProcessDPIAware = user32.NewProc("SetProcessDPIAware")
	pGetSysColorBrush   = user32.NewProc("GetSysColorBrush")
	pSetFocus           = user32.NewProc("SetFocus")
	pDestroyWindow      = user32.NewProc("DestroyWindow")
	pInvalidateRect     = user32.NewProc("InvalidateRect")
	pSetForeground      = user32.NewProc("SetForegroundWindow")

	pGetModuleHandle = kernel32.NewProc("GetModuleHandleW")
	pMulDiv          = kernel32.NewProc("MulDiv")

	pCreateFont     = gdi32.NewProc("CreateFontW")
	pGetDeviceCaps  = gdi32.NewProc("GetDeviceCaps")
	pSetTextColor   = gdi32.NewProc("SetTextColor")
	pSetBkMode      = gdi32.NewProc("SetBkMode")
	pCreateSolidBr  = gdi32.NewProc("CreateSolidBrush")

	pInitCommonControlsEx = comctl32.NewProc("InitCommonControlsEx")
	pGetOpenFileName      = comdlg32.NewProc("GetOpenFileNameW")
	pSHBrowseForFolder    = shell32.NewProc("SHBrowseForFolderW")
	pSHGetPathFromIDList  = shell32.NewProc("SHGetPathFromIDListW")
	pShellExecute         = shell32.NewProc("ShellExecuteW")
	pDragAcceptFiles      = shell32.NewProc("DragAcceptFiles")
	pDragQueryFile        = shell32.NewProc("DragQueryFileW")
	pDragFinish           = shell32.NewProc("DragFinish")
	pCoInitializeEx       = ole32.NewProc("CoInitializeEx")
	pCoTaskMemFree        = ole32.NewProc("CoTaskMemFree")

	pRegCreateKeyEx  = advapi32.NewProc("RegCreateKeyExW")
	pRegSetValueEx   = advapi32.NewProc("RegSetValueExW")
	pRegCloseKey     = advapi32.NewProc("RegCloseKey")
	pRegDeleteTree   = advapi32.NewProc("RegDeleteTreeW")
	pRegOpenKeyEx    = advapi32.NewProc("RegOpenKeyExW")
	pRegQueryValueEx = advapi32.NewProc("RegQueryValueExW")
)

// Hằng số Win32 thường dùng.
const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_CHILD            = 0x40000000
	WS_VISIBLE          = 0x10000000
	WS_TABSTOP          = 0x00010000
	WS_GROUP            = 0x00020000
	WS_VSCROLL          = 0x00200000
	WS_HSCROLL          = 0x00100000
	WS_BORDER           = 0x00800000
	WS_CLIPCHILDREN     = 0x02000000
	WS_CAPTION          = 0x00C00000
	WS_SYSMENU          = 0x00080000
	WS_MINIMIZEBOX      = 0x00020000
	WS_EX_CLIENTEDGE    = 0x00000200
	WS_EX_ACCEPTFILES   = 0x00000010
	CW_USEDEFAULT       = 0x80000000

	BS_PUSHBUTTON      = 0
	BS_DEFPUSHBUTTON   = 1
	BS_AUTOCHECKBOX    = 3
	BS_AUTORADIOBUTTON = 9
	BS_GROUPBOX        = 7
	ES_AUTOHSCROLL     = 0x80
	ES_MULTILINE       = 0x4
	ES_READONLY        = 0x800
	ES_AUTOVSCROLL     = 0x40
	SS_LEFT            = 0
	SS_CENTER          = 1
	SBARS_SIZEGRIP     = 0x100

	WM_CREATE         = 0x0001
	WM_DESTROY        = 0x0002
	WM_SIZE           = 0x0005
	WM_CLOSE          = 0x0010
	WM_SETFONT        = 0x0030
	WM_GETMINMAXINFO  = 0x0024
	WM_COMMAND        = 0x0111
	WM_CTLCOLORSTATIC = 0x0138
	WM_DROPFILES      = 0x0233
	WM_USER           = 0x0400
	WM_APP            = 0x8000
	BM_GETCHECK       = 0x00F0
	BM_SETCHECK       = 0x00F1
	EM_SETSEL         = 0x00B1
	EM_REPLACESEL     = 0x00C2
	EM_SETLIMITTEXT   = 0x00C5
	SB_SETPARTS       = WM_USER + 4
	SB_SETTEXTW       = WM_USER + 11
	BN_CLICKED        = 0

	SW_SHOW        = 5
	SW_SHOWNORMAL  = 1
	MB_OK          = 0
	MB_YESNO       = 4
	MB_ICONINFO    = 0x40
	MB_ICONWARNING = 0x30
	MB_ICONERROR   = 0x10
	MB_ICONQUEST   = 0x20
	IDYES          = 6
	COLOR_BTNFACE  = 15
	COLOR_WINDOW   = 5
	IMAGE_ICON     = 1
	LR_DEFAULTSIZE = 0x40
	LR_SHARED      = 0x8000
)

// MSG là cấu trúc thông điệp.
type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
	_       uint32
}

// RECT là hình chữ nhật.
type RECT struct{ Left, Top, Right, Bottom int32 }

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}

// S chuyển chuỗi sang UTF-16.
func S(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

// U trả uintptr của chuỗi UTF-16.
func U(s string) uintptr { return uintptr(unsafe.Pointer(S(s))) }

// Instance trả handle module.
func Instance() uintptr { h, _, _ := pGetModuleHandle.Call(0); return h }

// SetDPIAware bật nhận biết DPI hệ thống.
func SetDPIAware() { pSetProcessDPIAware.Call() }

// DPI trả DPI màn hình.
func DPI() int {
	dc, _, _ := pGetDC.Call(0)
	v, _, _ := pGetDeviceCaps.Call(dc, 88) // LOGPIXELSX
	pReleaseDC.Call(0, dc)
	if v == 0 {
		return 96
	}
	return int(v)
}

// InitControls khởi tạo common controls và COM.
func InitControls() {
	icc := struct{ Size, ICC uint32 }{8, 0x4 | 0x4000} // ICC_BAR_CLASSES | ICC_STANDARD_CLASSES
	pInitCommonControlsEx.Call(uintptr(unsafe.Pointer(&icc)))
	pCoInitializeEx.Call(0, 2) // COINIT_APARTMENTTHREADED
}

// LoadAppIcon nạp icon tài nguyên số 1.
func LoadAppIcon(size int) uintptr {
	h, _, _ := pLoadImage.Call(Instance(), 1, IMAGE_ICON, uintptr(size), uintptr(size), LR_SHARED)
	return h
}

// RegisterClass đăng ký lớp cửa sổ.
func RegisterClass(name string, proc func(h, m, w, l uintptr) uintptr, icon uintptr) {
	cur, _, _ := pLoadCursor.Call(0, 32512)
	bg, _, _ := pGetSysColorBrush.Call(COLOR_BTNFACE)
	wc := wndClassEx{
		WndProc:    syscall.NewCallback(proc),
		Instance:   Instance(),
		Icon:       icon,
		IconSm:     icon,
		Cursor:     cur,
		Background: bg,
		ClassName:  S(name),
	}
	wc.Size = uint32(unsafe.Sizeof(wc))
	pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
}

// CreateWindow tạo cửa sổ/điều khiển.
func CreateWindow(exStyle uint32, class, title string, style uint32, x, y, w, h int, parent, id uintptr) uintptr {
	r, _, _ := pCreateWindowEx.Call(uintptr(exStyle), U(class), U(title), uintptr(style),
		uintptr(int32(x)), uintptr(int32(y)), uintptr(int32(w)), uintptr(int32(h)), parent, id, Instance(), 0)
	return r
}

// DefWindowProc gọi thủ tục mặc định.
func DefWindowProc(h, m, w, l uintptr) uintptr { r, _, _ := pDefWindowProc.Call(h, m, w, l); return r }

// RunLoop chạy vòng lặp thông điệp.
func RunLoop(main uintptr) {
	var msg MSG
	for {
		r, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			return
		}
		if ok, _, _ := pIsDialogMessage.Call(main, uintptr(unsafe.Pointer(&msg))); ok != 0 {
			continue
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		pDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

// Các hàm tiện ích.
func PostQuit()                                   { pPostQuitMessage.Call(0) }
func Send(h uintptr, m uint32, w, l uintptr) uintptr { r, _, _ := pSendMessage.Call(h, uintptr(m), w, l); return r }
func Post(h uintptr, m uint32, w, l uintptr)       { pPostMessage.Call(h, uintptr(m), w, l) }
func Show(h uintptr, cmd int)                      { pShowWindow.Call(h, uintptr(cmd)); pUpdateWindow.Call(h) }
func Move(h uintptr, x, y, w, hh int)              { pMoveWindow.Call(h, uintptr(x), uintptr(y), uintptr(w), uintptr(hh), 1) }
func SetText(h uintptr, s string)                  { pSetWindowText.Call(h, U(s)) }
func Enable(h uintptr, on bool)                    { v := 0; if on { v = 1 }; pEnableWindow.Call(h, uintptr(v)) }
func Focus(h uintptr)                              { pSetFocus.Call(h) }
func Destroy(h uintptr)                            { pDestroyWindow.Call(h) }
func Foreground(h uintptr)                         { pSetForeground.Call(h) }
func Redraw(h uintptr)                             { pInvalidateRect.Call(h, 0, 1) }
func SetFont(h, font uintptr)                      { Send(h, WM_SETFONT, font, 1) }
func Checked(h uintptr) bool                       { return Send(h, BM_GETCHECK, 0, 0) == 1 }
func SetCheck(h uintptr, on bool)                  { v := uintptr(0); if on { v = 1 }; Send(h, BM_SETCHECK, v, 0) }
func SysColorBrush(i int) uintptr                  { r, _, _ := pGetSysColorBrush.Call(uintptr(i)); return r }
func SetTextColor(dc uintptr, rgb uint32)          { pSetTextColor.Call(dc, uintptr(rgb)) }
func SetBkTransparent(dc uintptr)                  { pSetBkMode.Call(dc, 1) }
func SolidBrush(rgb uint32) uintptr                { r, _, _ := pCreateSolidBr.Call(uintptr(rgb)); return r }
func SystemMetric(i int) int                       { r, _, _ := pGetSystemMetrics.Call(uintptr(i)); return int(r) }

// GetText đọc chữ của điều khiển.
func GetText(h uintptr) string {
	n, _, _ := pGetWindowTextLen.Call(h)
	buf := make([]uint16, n+1)
	pGetWindowText.Call(h, uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}

// ClientRect trả kích thước vùng khách.
func ClientRect(h uintptr) RECT {
	var r RECT
	pGetClientRect.Call(h, uintptr(unsafe.Pointer(&r)))
	return r
}

// WindowRect trả kích thước cửa sổ.
func WindowRect(h uintptr) RECT {
	var r RECT
	pGetWindowRect.Call(h, uintptr(unsafe.Pointer(&r)))
	return r
}

// Font tạo phông chữ theo điểm và DPI.
func Font(face string, pt int, bold bool, dpi int) uintptr {
	h, _, _ := pMulDiv.Call(uintptr(pt), uintptr(dpi), 72)
	weight := 400
	if bold {
		weight = 700
	}
	f, _, _ := pCreateFont.Call(uintptr(-int32(h)), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, U(face))
	return f
}

// MsgBox hiển thị hộp thoại.
func MsgBox(owner uintptr, text, title string, flags uint32) int {
	r, _, _ := pMessageBox.Call(owner, U(text), U(title), uintptr(flags))
	return int(r)
}

// AppendText nối chữ vào ô nhiều dòng.
func AppendText(h uintptr, s string) {
	Send(h, EM_SETSEL, 0x7fffffff, 0x7fffffff)
	Send(h, EM_REPLACESEL, 0, U(s))
}

// ShellOpen mở file/thư mục bằng chương trình mặc định.
func ShellOpen(path string) {
	pShellExecute.Call(0, U("open"), U(path), 0, 0, SW_SHOWNORMAL)
}

// ShellSelect mở Explorer và chọn file.
func ShellSelect(path string) {
	pShellExecute.Call(0, U("open"), U("explorer.exe"), U("/select,\""+path+"\""), 0, SW_SHOWNORMAL)
}

type openFileName struct {
	StructSize    uint32
	Owner         uintptr
	Instance      uintptr
	Filter        *uint16
	CustomFilter  *uint16
	MaxCustFilter uint32
	FilterIndex   uint32
	File          *uint16
	MaxFile       uint32
	FileTitle     *uint16
	MaxFileTitle  uint32
	InitialDir    *uint16
	Title         *uint16
	Flags         uint32
	FileOffset    uint16
	FileExtension uint16
	DefExt        *uint16
	CustData      uintptr
	FnHook        uintptr
	TemplateName  *uint16
	PvReserved    uintptr
	DwReserved    uint32
	FlagsEx       uint32
}

// OpenFileDialog mở hộp chọn file. filter dạng "Mô tả|*.a;*.b|Tất cả|*.*".
func OpenFileDialog(owner uintptr, title, filter string) (string, bool) {
	f := []uint16{}
	for _, part := range splitBar(filter) {
		u, _ := syscall.UTF16FromString(part)
		f = append(f, u...)
	}
	f = append(f, 0)
	buf := make([]uint16, 4096)
	ofn := openFileName{
		Owner:  owner,
		Filter: &f[0],
		File:   &buf[0],
		MaxFile: uint32(len(buf)),
		Title:  S(title),
		Flags:  0x00080000 | 0x00001000 | 0x00000800 | 0x00000008, // EXPLORER|FILEMUSTEXIST|PATHMUSTEXIST|NOCHANGEDIR
		FilterIndex: 1,
	}
	ofn.StructSize = uint32(unsafe.Sizeof(ofn))
	r, _, _ := pGetOpenFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return "", false
	}
	return syscall.UTF16ToString(buf), true
}

func splitBar(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == '|' {
			out = append(out, cur)
			cur = ""
		} else {
			cur += string(r)
		}
	}
	return append(out, cur)
}

type browseInfo struct {
	Owner       uintptr
	Root        uintptr
	DisplayName *uint16
	Title       *uint16
	Flags       uint32
	Fn          uintptr
	LParam      uintptr
	Image       int32
}

// FolderDialog mở hộp chọn thư mục.
func FolderDialog(owner uintptr, title string) (string, bool) {
	name := make([]uint16, 260)
	bi := browseInfo{Owner: owner, DisplayName: &name[0], Title: S(title), Flags: 0x1 | 0x40} // RETURNONLYFSDIRS|NEWDIALOGSTYLE
	pidl, _, _ := pSHBrowseForFolder.Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 {
		return "", false
	}
	defer pCoTaskMemFree.Call(pidl)
	buf := make([]uint16, 1024)
	pSHGetPathFromIDList.Call(pidl, uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf), true
}

// AcceptFiles bật kéo thả file.
func AcceptFiles(h uintptr) { pDragAcceptFiles.Call(h, 1) }

// DroppedFiles đọc danh sách file kéo thả.
func DroppedFiles(hdrop uintptr) []string {
	n, _, _ := pDragQueryFile.Call(hdrop, 0xFFFFFFFF, 0, 0)
	var out []string
	for i := uintptr(0); i < n; i++ {
		buf := make([]uint16, 1024)
		pDragQueryFile.Call(hdrop, i, uintptr(unsafe.Pointer(&buf[0])), 1024)
		out = append(out, syscall.UTF16ToString(buf))
	}
	pDragFinish.Call(hdrop)
	return out
}

// ---- Registry (HKCU) ----

// HKCU là gốc HKEY_CURRENT_USER.
const HKCU = 0x80000001

// RegSetStrings ghi các giá trị chuỗi (và DWORD nếu tên bắt đầu bằng "#") vào khóa.
func RegSetStrings(root uintptr, path string, vals map[string]string, dwords map[string]uint32) error {
	var key uintptr
	r, _, _ := pRegCreateKeyEx.Call(root, U(path), 0, 0, 0, 0x20006, 0, uintptr(unsafe.Pointer(&key)), 0)
	if r != 0 {
		return syscall.Errno(r)
	}
	defer pRegCloseKey.Call(key)
	for k, v := range vals {
		u, _ := syscall.UTF16FromString(v)
		pRegSetValueEx.Call(key, U(k), 0, 1, uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)*2))
	}
	for k, v := range dwords {
		vv := v
		pRegSetValueEx.Call(key, U(k), 0, 4, uintptr(unsafe.Pointer(&vv)), 4)
	}
	return nil
}

// RegGetString đọc một giá trị chuỗi.
func RegGetString(root uintptr, path, name string) (string, bool) {
	var key uintptr
	if r, _, _ := pRegOpenKeyEx.Call(root, U(path), 0, 0x20019, uintptr(unsafe.Pointer(&key))); r != 0 {
		return "", false
	}
	defer pRegCloseKey.Call(key)
	buf := make([]uint16, 2048)
	size := uint32(len(buf) * 2)
	if r, _, _ := pRegQueryValueEx.Call(key, U(name), 0, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size))); r != 0 {
		return "", false
	}
	return syscall.UTF16ToString(buf), true
}

// RegDeleteTree xóa khóa và toàn bộ khóa con.
func RegDeleteTree(root uintptr, path string) { pRegDeleteTree.Call(root, U(path)) }
