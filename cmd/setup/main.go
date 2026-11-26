//go:build windows

// Bộ cài đặt ĐỊNH DẠNG VĂN BẢN CHUẨN.
//
//	setup.exe              cài đặt có giao diện
//	setup.exe /S [/D=dir]  cài im lặng (triển khai hàng loạt)
//	uninstall.exe /uninstall [/S]  gỡ cài đặt
package main

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"ddvbc/core"
	"ddvbc/win"
)

//go:embed payload/DinhDangVanBanChuan.exe.gz
var appExeGz []byte

// appExe giải nén chương trình nhúng trong bộ cài.
func appExe() []byte {
	zr, err := gzip.NewReader(bytes.NewReader(appExeGz))
	if err != nil {
		return nil
	}
	b, _ := io.ReadAll(zr)
	return b
}

//go:embed payload/HuongDanSuDung.txt
var guide []byte

const (
	exeName   = "DinhDangVanBanChuan.exe"
	uninstKey = `Software\Microsoft\Windows\CurrentVersion\Uninstall\DinhDangVanBanChuan`
	publisher = "Phòng Kỹ thuật - VTTB - CNTT, Bệnh viện Hữu nghị đa khoa Nghệ An - Giai đoạn 2"
)

var (
	lastWarn string // cảnh báo không chặn cài đặt
	hwnd uintptr
	ctl  = map[int]uintptr{}
	dpi  = 96
	fUI, fTitle uintptr
)

const (
	idDir = 200 + iota
	idBrowse
	idDesk
	idRunAfter
	idInstall
	idExit
	idInfo
	idTitle
	idLbl
	idStat
)

func init() { runtime.LockOSThread() }

func sc(v int) int { return v * dpi / 96 }

func defaultDir() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}
	return filepath.Join(base, "Programs", "DinhDangVanBanChuan")
}

func main() {
	silent, uninstall := false, false
	dir := ""
	for _, a := range os.Args[1:] {
		switch {
		case strings.EqualFold(a, "/S"):
			silent = true
		case strings.EqualFold(a, "/uninstall"):
			uninstall = true
		case strings.HasPrefix(strings.ToUpper(a), "/D="):
			dir = a[3:]
		}
	}
	if strings.EqualFold(filepath.Base(os.Args[0]), "uninstall.exe") {
		uninstall = true
	}
	win.SetDPIAware()
	win.InitControls()
	if uninstall {
		doUninstall(silent)
		return
	}
	if dir == "" {
		if d, ok := win.RegGetString(win.HKCU, uninstKey, "InstallLocation"); ok && d != "" {
			dir = d
		} else {
			dir = defaultDir()
		}
	}
	if silent {
		if err := install(dir, true); err != nil {
			os.Exit(1)
		}
		return
	}
	win.InitControls()
	dpi = win.DPI()
	fUI = win.Font("Segoe UI", 9, false, dpi)
	fTitle = win.Font("Segoe UI", 14, true, dpi)
	win.RegisterClass("DDVBCSetup", wndProc, win.LoadAppIcon(32))
	w, h := sc(560), sc(380)
	hwnd = win.CreateWindow(0, "DDVBCSetup", "Cài đặt "+core.AppName+" "+core.Version,
		win.WS_CAPTION|win.WS_SYSMENU|win.WS_MINIMIZEBOX|win.WS_CLIPCHILDREN,
		(win.SystemMetric(0)-w)/2, (win.SystemMetric(1)-h)/2, w, h, 0, 0)
	win.SetText(ctl[idDir], dir)
	win.Show(hwnd, win.SW_SHOW)
	win.RunLoop(hwnd)
}

func mk(ex uint32, class, text string, style uint32, id int, f uintptr, x, y, w, h int) uintptr {
	c := win.CreateWindow(ex, class, text, win.WS_CHILD|win.WS_VISIBLE|style, sc(x), sc(y), sc(w), sc(h), hwnd, uintptr(id))
	win.SetFont(c, f)
	ctl[id] = c
	return c
}

func create() {
	mk(0, "STATIC", core.AppName, 0, idTitle, fTitle, 20, 14, 510, 30)
	mk(0, "STATIC", "Phiên bản "+core.Version+"\r\nĐịnh dạng lại / soạn văn bản đúng thể thức pháp luật Việt Nam (NĐ 30/2020 cho văn bản hành chính; NĐ 78/2025 sửa bởi NĐ 187/2025 cho VBQPPL).\r\nCài cho người dùng hiện tại, không cần quyền quản trị (Admin).", 0, idInfo, fUI, 20, 48, 510, 70)
	mk(0, "STATIC", "Thư mục cài đặt:", 0, idLbl, fUI, 20, 128, 200, 18)
	mk(win.WS_EX_CLIENTEDGE, "EDIT", "", win.ES_AUTOHSCROLL|win.WS_TABSTOP, idDir, fUI, 20, 148, 400, 24)
	mk(0, "BUTTON", "Chọn...", win.WS_TABSTOP, idBrowse, fUI, 430, 147, 100, 26)
	d := mk(0, "BUTTON", "Tạo biểu tượng ngoài màn hình Desktop", win.BS_AUTOCHECKBOX|win.WS_TABSTOP, idDesk, fUI, 20, 186, 400, 20)
	win.SetCheck(d, true)
	r := mk(0, "BUTTON", "Mở chương trình sau khi cài đặt", win.BS_AUTOCHECKBOX|win.WS_TABSTOP, idRunAfter, fUI, 20, 210, 400, 20)
	win.SetCheck(r, true)
	mk(0, "STATIC", "Bộ cài tạo lối tắt \"DocStd\" trong Start Menu, mục \"Gửi tới (Send to)\" khi bấm chuột phải vào file, và mục gỡ cài đặt trong Settings › Apps.", 0, idStat, fUI, 20, 238, 510, 40)
	mk(0, "BUTTON", "Cài đặt", win.BS_DEFPUSHBUTTON|win.WS_TABSTOP, idInstall, fUI, 300, 292, 110, 30)
	mk(0, "BUTTON", "Thoát", win.WS_TABSTOP, idExit, fUI, 420, 292, 110, 30)
}

func wndProc(h, msg, wp, lp uintptr) uintptr {
	switch uint32(msg) {
	case win.WM_CREATE:
		hwnd = h
		create()
		return 0
	case win.WM_COMMAND:
		switch int(wp & 0xFFFF) {
		case idBrowse:
			if d, ok := win.FolderDialog(h, "Chọn thư mục cài đặt"); ok {
				win.SetText(ctl[idDir], filepath.Join(d, "DinhDangVanBanChuan"))
			}
		case idExit:
			win.Destroy(h)
		case idInstall:
			dir := strings.TrimSpace(win.GetText(ctl[idDir]))
			if dir == "" {
				return 0
			}
			win.Enable(ctl[idInstall], false)
			win.SetText(ctl[idStat], "Đang cài đặt...")
			if err := install(dir, win.Checked(ctl[idDesk])); err != nil {
				win.SetText(ctl[idStat], "Cài đặt không thành công.")
				win.MsgBox(h, "Cài đặt không thành công:\n"+err.Error(), core.AppName, win.MB_ICONERROR)
				win.Enable(ctl[idInstall], true)
				return 0
			}
			win.SetText(ctl[idStat], "Đã cài đặt xong vào: "+dir)
			if win.Checked(ctl[idRunAfter]) {
				startDetached(filepath.Join(dir, exeName))
			}
			if lastWarn != "" {
				win.MsgBox(h, lastWarn, core.AppName, win.MB_ICONWARNING)
			}
			win.MsgBox(h, "Đã cài đặt xong "+core.AppName+" phiên bản "+core.Version+".\n\nMở chương trình từ biểu tượng \""+core.ShortName+"\" ngoài Desktop hoặc Start Menu.", core.AppName, win.MB_ICONINFO)
			win.Destroy(h)
		}
		return 0
	case win.WM_DESTROY:
		win.PostQuit()
		return 0
	}
	return win.DefWindowProc(h, msg, wp, lp)
}

func hide(c *exec.Cmd) { c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} }

func startDetached(path string) {
	c := exec.Command(path)
	c.Dir = filepath.Dir(path)
	c.Start()
}

// shortcutPaths trả các vị trí lối tắt: Start Menu, Send to, (Desktop) theo tên name.
func shortcutPaths(desktop bool, name string) []string {
	ids := []win.GUID{win.FolderPrograms, win.FolderSendTo}
	if desktop {
		ids = append(ids, win.FolderDesktop)
	}
	var out []string
	for _, id := range ids {
		if d, err := win.KnownFolder(id); err == nil && d != "" {
			out = append(out, filepath.Join(d, name+".lnk"))
		}
	}
	return out
}

func install(dir string, desktop bool) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	exe := filepath.Join(dir, exeName)
	data := appExe()
	if len(data) == 0 {
		return fmt.Errorf("bộ cài bị hỏng (không giải nén được chương trình)")
	}
	if err := os.WriteFile(exe, data, 0755); err != nil {
		return fmt.Errorf("không ghi được %s (chương trình có đang mở không? hãy đóng lại rồi cài lại): %v", exe, err)
	}
	os.WriteFile(filepath.Join(dir, "HuongDanSuDung.txt"), guide, 0644)
	if self, err := os.Executable(); err == nil {
		un := filepath.Join(dir, "uninstall.exe")
		if !strings.EqualFold(self, un) {
			if src, err := os.Open(self); err == nil {
				if dst, err := os.Create(un); err == nil {
					io.Copy(dst, src)
					dst.Close()
				}
				src.Close()
			}
		}
	}
	var lnkErrs []string
	// xóa lối tắt tên cũ (bản 1.0.0–1.0.8) để không bị trùng biểu tượng
	for _, l := range shortcutPaths(true, core.AppName) {
		os.Remove(l)
	}
	for _, l := range shortcutPaths(desktop, core.ShortName) {
		if err := win.CreateShortcut(l, exe, dir, "Định dạng văn bản đúng thể thức NĐ 30/2020, NĐ 78/2025 (sửa bởi NĐ 187/2025)", exe, 0); err != nil {
			lnkErrs = append(lnkErrs, err.Error())
		}
	}
	un := filepath.Join(dir, "uninstall.exe")
	win.RegSetStrings(win.HKCU, uninstKey, map[string]string{
		"DisplayName":          core.ShortName,
		"Comments":             core.AppName + " - định dạng văn bản đúng thể thức NĐ 30/2020, NĐ 78/2025 (sửa bởi NĐ 187/2025)",
		"DisplayVersion":       core.Version,
		"Publisher":            publisher,
		"InstallLocation":      dir,
		"DisplayIcon":          exe + ",0",
		"UninstallString":      `"` + un + `" /uninstall`,
		"QuietUninstallString": `"` + un + `" /uninstall /S`,
	}, map[string]uint32{"NoModify": 1, "NoRepair": 1, "EstimatedSize": uint32(len(data)/1024 + 2048)})
	if len(lnkErrs) > 0 {
		lastWarn = "Chương trình đã cài vào " + dir + " nhưng có lối tắt chưa tạo được:\n" + strings.Join(lnkErrs, "\n")
	}
	return nil
}

func doUninstall(silent bool) {
	dir, ok := win.RegGetString(win.HKCU, uninstKey, "InstallLocation")
	if !ok || dir == "" {
		if self, err := os.Executable(); err == nil {
			dir = filepath.Dir(self)
		}
	}
	if !silent && win.MsgBox(0, "Gỡ bỏ "+core.ShortName+" ("+core.AppName+") khỏi máy tính?\n\nCác văn bản anh/chị đã định dạng không bị xóa.", "Gỡ cài đặt", win.MB_YESNO|win.MB_ICONQUEST) != win.IDYES {
		return
	}
	for _, name := range []string{core.ShortName, core.AppName} {
		for _, l := range shortcutPaths(true, name) {
			os.Remove(l)
		}
	}
	win.RegDeleteTree(win.HKCU, uninstKey)
	win.RegDeleteTree(win.HKCU, `Software\DinhDangVanBanChuan`)
	os.Remove(filepath.Join(dir, exeName))
	os.Remove(filepath.Join(dir, "HuongDanSuDung.txt"))
	// uninstall.exe đang chạy: xóa thư mục sau khi thoát
	c := exec.Command("cmd.exe", "/c", "ping 127.0.0.1 -n 3 >nul & del /f /q \""+filepath.Join(dir, "uninstall.exe")+"\" & rd \""+dir+"\"")
	hide(c)
	c.Start()
	if !silent {
		win.MsgBox(0, "Đã gỡ bỏ "+core.ShortName+".", "Gỡ cài đặt", win.MB_ICONINFO)
	}
}
