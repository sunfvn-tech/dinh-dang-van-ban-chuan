//go:build windows

// ĐỊNH DẠNG VĂN BẢN CHUẨN — giao diện Windows (Win32 thuần).
package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"ddvbc/core"
	"ddvbc/win"
)

const (
	idEditIn = 100 + iota
	idBtnIn
	idRadAuto
	idRadVBHC
	idRadQPPL
	idRadDocx
	idRadPDF
	idEditOut
	idBtnOut
	idChkOpen
	idBtnRun
	idBtnOpenDir
	idBtnClear
	idLog
	idStatus
	idTitle
	idSub
	idHint
	idLblOut
	idLblLog
	idGrp1
	idGrp2
	idGrp3
	idBtnHelp
)

const (
	wmLog  = win.WM_APP + 1
	wmDone = win.WM_APP + 2
	regKey = `Software\DinhDangVanBanChuan`
	subtitle = "Định dạng lại / soạn văn bản đúng thể thức pháp luật Việt Nam (NĐ 30/2020 cho văn bản hành chính; NĐ 78/2025 sửa bởi NĐ 187/2025 cho VBQPPL)"
)

var (
	hwnd   uintptr
	ctl    = map[int]uintptr{}
	dpi    = 96
	fontUI, fontTitle, fontSmall, fontBtn uintptr
	brushBg uintptr

	mu      sync.Mutex
	pending []string
	busy    bool
	inputs  []string
	lastOut string
	toOpen  []string // file cần mở sau khi xong (mở trên luồng giao diện)
)

func sc(v int) int { return v * dpi / 96 }

func init() { runtime.LockOSThread() }

func main() {
	win.SetDPIAware()
	win.InitControls()
	dpi = win.DPI()
	fontUI = win.Font("Segoe UI", 9, false, dpi)
	fontSmall = win.Font("Segoe UI", 8, false, dpi)
	fontBtn = win.Font("Segoe UI", 10, true, dpi)
	fontTitle = win.Font("Segoe UI", 17, true, dpi)
	brushBg = win.SysColorBrush(win.COLOR_BTNFACE)

	icon := win.LoadAppIcon(32)
	win.RegisterClass("DDVBCMain", wndProc, icon)
	w, h := sc(820), sc(680)
	x := (win.SystemMetric(0) - w) / 2
	y := (win.SystemMetric(1) - h) / 2
	if y < 0 {
		y = 0
	}
	hwnd = win.CreateWindow(win.WS_EX_ACCEPTFILES, "DDVBCMain", core.AppName,
		win.WS_OVERLAPPEDWINDOW|win.WS_CLIPCHILDREN, x, y, w, h, 0, 0)
	if hwnd == 0 {
		win.MsgBox(0, "Không tạo được cửa sổ ứng dụng.", core.AppName, win.MB_ICONERROR)
		return
	}
	win.Show(hwnd, win.SW_SHOW)
	// File truyền qua dòng lệnh (kéo thả lên biểu tượng / Send To)
	var args []string
	for _, a := range os.Args[1:] {
		if st, err := os.Stat(a); err == nil && !st.IsDir() {
			args = append(args, a)
		}
	}
	if len(args) > 0 {
		setInputs(args)
	}
	win.RunLoop(hwnd)
}

func mk(ex uint32, class, text string, style uint32, id int, font uintptr) uintptr {
	h := win.CreateWindow(ex, class, text, win.WS_CHILD|win.WS_VISIBLE|style, 0, 0, 10, 10, hwnd, uintptr(id))
	win.SetFont(h, font)
	ctl[id] = h
	return h
}

func createControls() {
	mk(0, "STATIC", core.AppName, win.SS_CENTER, idTitle, fontTitle)
	mk(0, "STATIC", "Chọn văn bản → chọn loại văn bản → chọn kiểu xuất → bấm ĐỊNH DẠNG VĂN BẢN", win.SS_CENTER, idSub, fontUI)

	mk(0, "BUTTON", " 1. Văn bản đầu vào ", win.BS_GROUPBOX, idGrp1, fontUI)
	e := mk(win.WS_EX_CLIENTEDGE, "EDIT", "", win.ES_AUTOHSCROLL|win.ES_READONLY|win.WS_TABSTOP, idEditIn, fontUI)
	_ = e
	mk(0, "BUTTON", "Chọn file...", win.BS_PUSHBUTTON|win.WS_TABSTOP, idBtnIn, fontUI)
	mk(0, "STATIC", "Nhận file .docx, .doc, .txt, .pdf (có thể chọn nhiều file hoặc kéo thả file vào cửa sổ). File gốc giữ nguyên, kết quả lưu thành <tên>_chuanhoa.docx", win.SS_LEFT, idHint, fontSmall)

	mk(0, "BUTTON", " 2. Loại văn bản (căn cứ áp dụng) ", win.BS_GROUPBOX, idGrp2, fontUI)
	mk(0, "BUTTON", "Tự nhận diện (số ký hiệu có năm như …/2026/QĐ-UBND, Luật, Nghị định, Thông tư → VBQPPL; còn lại → văn bản hành chính)", win.BS_AUTORADIOBUTTON|win.WS_GROUP|win.WS_TABSTOP, idRadAuto, fontUI)
	mk(0, "BUTTON", "Văn bản hành chính — Nghị định 30/2020/NĐ-CP (công văn, quyết định, tờ trình, báo cáo, kế hoạch, thông báo...)", win.BS_AUTORADIOBUTTON, idRadVBHC, fontUI)
	mk(0, "BUTTON", "Văn bản quy phạm pháp luật — Nghị định 78/2025/NĐ-CP, sửa đổi bởi Nghị định 187/2025/NĐ-CP", win.BS_AUTORADIOBUTTON, idRadQPPL, fontUI)

	mk(0, "BUTTON", " 3. Kết quả xuất ra ", win.BS_GROUPBOX, idGrp3, fontUI)
	mk(0, "BUTTON", "Chỉ xuất file Word (.docx)", win.BS_AUTORADIOBUTTON|win.WS_GROUP|win.WS_TABSTOP, idRadDocx, fontUI)
	mk(0, "BUTTON", "Xuất file Word (.docx) và file PDF (.pdf)", win.BS_AUTORADIOBUTTON, idRadPDF, fontUI)
	mk(0, "STATIC", "Lưu vào thư mục:", win.SS_LEFT, idLblOut, fontUI)
	mk(win.WS_EX_CLIENTEDGE, "EDIT", "", win.ES_AUTOHSCROLL|win.WS_TABSTOP|win.WS_GROUP, idEditOut, fontUI)
	mk(0, "BUTTON", "Chọn thư mục...", win.BS_PUSHBUTTON|win.WS_TABSTOP, idBtnOut, fontUI)
	mk(0, "BUTTON", "Mở file kết quả sau khi định dạng xong", win.BS_AUTOCHECKBOX|win.WS_TABSTOP, idChkOpen, fontUI)

	mk(0, "BUTTON", "ĐỊNH DẠNG VĂN BẢN", win.BS_DEFPUSHBUTTON|win.WS_TABSTOP|win.WS_GROUP, idBtnRun, fontBtn)
	mk(0, "BUTTON", "Mở thư mục kết quả", win.BS_PUSHBUTTON|win.WS_TABSTOP, idBtnOpenDir, fontUI)
	mk(0, "BUTTON", "Xóa nhật ký", win.BS_PUSHBUTTON|win.WS_TABSTOP, idBtnClear, fontUI)
	mk(0, "BUTTON", "Hướng dẫn", win.BS_PUSHBUTTON|win.WS_TABSTOP, idBtnHelp, fontUI)
	mk(0, "STATIC", "Nhật ký xử lý (các thành phần nhận diện được, các điểm đã sửa, chỗ còn \"...\" cần điền):", win.SS_LEFT, idLblLog, fontUI)
	lg := mk(win.WS_EX_CLIENTEDGE, "EDIT", "", win.ES_MULTILINE|win.ES_READONLY|win.ES_AUTOVSCROLL|win.WS_VSCROLL|win.WS_TABSTOP, idLog, win.Font("Consolas", 9, false, dpi))
	win.Send(lg, win.EM_SETLIMITTEXT, 0, 0)

	st := win.CreateWindow(0, "msctls_statusbar32", "", win.WS_CHILD|win.WS_VISIBLE|win.SBARS_SIZEGRIP, 0, 0, 0, 0, hwnd, idStatus)
	win.SetFont(st, fontSmall)
	ctl[idStatus] = st

	// Nạp lựa chọn lần trước
	prof, _ := win.RegGetString(win.HKCU, regKey, "Profile")
	switch prof {
	case core.ProfileVBHC:
		win.SetCheck(ctl[idRadVBHC], true)
	case core.ProfileVBQPPL:
		win.SetCheck(ctl[idRadQPPL], true)
	default:
		win.SetCheck(ctl[idRadAuto], true)
	}
	if v, _ := win.RegGetString(win.HKCU, regKey, "PDF"); v == "1" {
		win.SetCheck(ctl[idRadPDF], true)
	} else {
		win.SetCheck(ctl[idRadDocx], true)
	}
	if v, ok := win.RegGetString(win.HKCU, regKey, "OpenAfter"); !ok || v == "1" {
		win.SetCheck(ctl[idChkOpen], true)
	}
	if v, _ := win.RegGetString(win.HKCU, regKey, "OutDir"); v != "" {
		win.SetText(ctl[idEditOut], v)
	}
	win.AppendText(ctl[idLog], fmt.Sprintf("%s — phiên bản %s\r\nSẵn sàng. Hãy chọn hoặc kéo thả văn bản cần định dạng vào cửa sổ.\r\nĐể trống ô \"Lưu vào thư mục\" thì kết quả lưu cùng thư mục với file gốc.\r\n", core.AppName, core.Version))
	win.Focus(ctl[idBtnIn])
}

func layout() {
	r := win.ClientRect(hwnd)
	W, H := int(r.Right), int(r.Bottom)
	st := ctl[idStatus]
	win.Send(st, win.WM_SIZE, 0, 0)
	stH := int(win.WindowRect(st).Bottom - win.WindowRect(st).Top)
	verW := sc(130)
	parts := [2]int32{int32(W - verW), -1}
	win.Send(st, win.SB_SETPARTS, 2, uintptrOf(&parts))
	win.Send(st, win.SB_SETTEXTW, 0, win.U(subtitle))
	win.Send(st, win.SB_SETTEXTW, 1, win.U("Phiên bản "+core.Version))

	m := sc(12)
	cw := W - 2*m
	y := sc(8)
	win.Move(ctl[idTitle], m, y, cw, sc(34))
	y += sc(36)
	win.Move(ctl[idSub], m, y, cw, sc(18))
	y += sc(24)

	// Nhóm 1
	g1 := sc(82)
	win.Move(ctl[idGrp1], m, y, cw, g1)
	bw := sc(120)
	win.Move(ctl[idEditIn], m+sc(12), y+sc(22), cw-bw-sc(32), sc(24))
	win.Move(ctl[idBtnIn], m+cw-bw-sc(12), y+sc(21), bw, sc(26))
	win.Move(ctl[idHint], m+sc(12), y+sc(52), cw-sc(24), sc(26))
	y += g1 + sc(6)

	// Nhóm 2
	g2 := sc(90)
	win.Move(ctl[idGrp2], m, y, cw, g2)
	win.Move(ctl[idRadAuto], m+sc(12), y+sc(20), cw-sc(24), sc(20))
	win.Move(ctl[idRadVBHC], m+sc(12), y+sc(42), cw-sc(24), sc(20))
	win.Move(ctl[idRadQPPL], m+sc(12), y+sc(64), cw-sc(24), sc(20))
	y += g2 + sc(6)

	// Nhóm 3
	g3 := sc(108)
	win.Move(ctl[idGrp3], m, y, cw, g3)
	half := (cw - sc(24)) / 2
	win.Move(ctl[idRadDocx], m+sc(12), y+sc(20), half, sc(20))
	win.Move(ctl[idRadPDF], m+sc(12)+half, y+sc(20), half, sc(20))
	lw := sc(110)
	win.Move(ctl[idLblOut], m+sc(12), y+sc(50), lw, sc(20))
	win.Move(ctl[idEditOut], m+sc(12)+lw, y+sc(47), cw-lw-bw-sc(32), sc(24))
	win.Move(ctl[idBtnOut], m+cw-bw-sc(12), y+sc(46), bw, sc(26))
	win.Move(ctl[idChkOpen], m+sc(12), y+sc(80), cw-sc(24), sc(20))
	y += g3 + sc(10)

	// Nút
	win.Move(ctl[idBtnRun], m, y, sc(230), sc(36))
	win.Move(ctl[idBtnOpenDir], m+sc(240), y+sc(4), sc(150), sc(28))
	win.Move(ctl[idBtnClear], m+cw-sc(110), y+sc(4), sc(110), sc(28))
	win.Move(ctl[idBtnHelp], m+cw-sc(230), y+sc(4), sc(110), sc(28))
	y += sc(44)

	win.Move(ctl[idLblLog], m, y, cw, sc(18))
	y += sc(20)
	lh := H - stH - y - sc(8)
	if lh < sc(40) {
		lh = sc(40)
	}
	win.Move(ctl[idLog], m, y, cw, lh)
}

func wndProc(h, msg, wp, lp uintptr) uintptr {
	switch uint32(msg) {
	case win.WM_CREATE:
		hwnd = h
		createControls()
		return 0
	case win.WM_SIZE:
		if len(ctl) > 0 {
			layout()
		}
		return 0
	case win.WM_GETMINMAXINFO:
		mmi := (*[10]int32)(ptrOf(lp))
		mmi[6], mmi[7] = int32(sc(700)), int32(sc(620)) // ptMinTrackSize
		return 0
	case win.WM_CTLCOLORSTATIC:
		if lp == ctl[idTitle] {
			win.SetTextColor(wp, 0x00803300) // xanh đậm (BGR)
			win.SetBkTransparent(wp)
			return brushBg
		}
	case win.WM_DROPFILES:
		files := win.DroppedFiles(wp)
		if len(files) > 0 && !busy {
			setInputs(files)
		}
		return 0
	case win.WM_COMMAND:
		id := int(wp & 0xFFFF)
		code := int(wp >> 16 & 0xFFFF)
		if code == win.BN_CLICKED {
			onCommand(id)
		}
		return 0
	case wmLog:
		flushLog()
		return 0
	case wmDone:
		flushLog()
		busy = false
		mu.Lock()
		open := toOpen
		toOpen = nil
		mu.Unlock()
		for _, f := range open {
			if !strings.HasPrefix(f, "#") {
				win.ShellOpen(f)
			} else {
				lastOut = strings.TrimPrefix(f, "#")
			}
		}
		win.Enable(ctl[idBtnRun], true)
		win.Enable(ctl[idBtnIn], true)
		win.SetText(ctl[idBtnRun], "ĐỊNH DẠNG VĂN BẢN")
		return 0
	case win.WM_CLOSE:
		if busy && win.MsgBox(h, "Đang xử lý văn bản. Vẫn thoát?", core.AppName, win.MB_YESNO|win.MB_ICONQUEST) != win.IDYES {
			return 0
		}
		win.Destroy(h)
		return 0
	case win.WM_DESTROY:
		saveSettings()
		win.PostQuit()
		return 0
	}
	return win.DefWindowProc(h, msg, wp, lp)
}

func setInputs(files []string) {
	var ok []string
	for _, f := range files {
		switch strings.ToLower(filepath.Ext(f)) {
		case ".docx", ".doc", ".txt", ".pdf", ".rtf", ".odt", ".docm", ".text":
			ok = append(ok, f)
		}
	}
	if len(ok) == 0 {
		win.MsgBox(hwnd, "Chỉ nhận file .docx, .doc, .txt, .pdf.", core.AppName, win.MB_ICONWARNING)
		return
	}
	inputs = ok
	if len(ok) == 1 {
		win.SetText(ctl[idEditIn], ok[0])
	} else {
		names := []string{}
		for _, f := range ok {
			names = append(names, filepath.Base(f))
		}
		win.SetText(ctl[idEditIn], fmt.Sprintf("%d file: %s", len(ok), strings.Join(names, "; ")))
	}
	win.AppendText(ctl[idLog], fmt.Sprintf("\r\nĐã chọn %d file.\r\n", len(ok)))
}

func profileChoice() string {
	switch {
	case win.Checked(ctl[idRadVBHC]):
		return core.ProfileVBHC
	case win.Checked(ctl[idRadQPPL]):
		return core.ProfileVBQPPL
	}
	return core.ProfileAuto
}

func saveSettings() {
	pdf := "0"
	if win.Checked(ctl[idRadPDF]) {
		pdf = "1"
	}
	op := "0"
	if win.Checked(ctl[idChkOpen]) {
		op = "1"
	}
	win.RegSetStrings(win.HKCU, regKey, map[string]string{
		"Profile": profileChoice(), "PDF": pdf, "OpenAfter": op, "OutDir": strings.TrimSpace(win.GetText(ctl[idEditOut])),
	}, nil)
}

func onCommand(id int) {
	switch id {
	case idBtnIn:
		if f, ok := win.OpenFileDialog(hwnd, "Chọn văn bản cần định dạng", "Văn bản (*.docx;*.doc;*.txt;*.pdf)|*.docx;*.doc;*.txt;*.pdf;*.rtf;*.odt|Tất cả các file (*.*)|*.*"); ok {
			setInputs([]string{f})
		}
	case idBtnOut:
		if d, ok := win.FolderDialog(hwnd, "Chọn thư mục lưu văn bản đã định dạng"); ok {
			win.SetText(ctl[idEditOut], d)
		}
	case idBtnHelp:
		openGuide()
	case idBtnClear:
		win.SetText(ctl[idLog], "")
	case idBtnOpenDir:
		switch {
		case lastOut != "":
			win.ShellSelect(lastOut)
		case strings.TrimSpace(win.GetText(ctl[idEditOut])) != "":
			win.ShellOpen(strings.TrimSpace(win.GetText(ctl[idEditOut])))
		case len(inputs) > 0:
			win.ShellOpen(filepath.Dir(inputs[0]))
		}
	case idBtnRun:
		run()
	}
}

func logLine(s string) {
	mu.Lock()
	pending = append(pending, s)
	mu.Unlock()
	win.Post(hwnd, wmLog, 0, 0)
}

func flushLog() {
	mu.Lock()
	p := pending
	pending = nil
	mu.Unlock()
	if len(p) > 0 {
		win.AppendText(ctl[idLog], strings.Join(p, "\r\n")+"\r\n")
	}
}

func run() {
	if busy {
		return
	}
	if len(inputs) == 0 {
		win.MsgBox(hwnd, "Anh/chị chưa chọn văn bản đầu vào.\nBấm \"Chọn file...\" hoặc kéo thả file vào cửa sổ.", core.AppName, win.MB_ICONWARNING)
		return
	}
	outDir := strings.TrimSpace(win.GetText(ctl[idEditOut]))
	if outDir != "" {
		if st, err := os.Stat(outDir); err != nil || !st.IsDir() {
			if err := os.MkdirAll(outDir, 0755); err != nil {
				win.MsgBox(hwnd, "Thư mục lưu không hợp lệ: "+outDir, core.AppName, win.MB_ICONERROR)
				return
			}
		}
	}
	opt := core.Options{OutDir: outDir, Profile: profileChoice(), PDF: win.Checked(ctl[idRadPDF])}
	openAfter := win.Checked(ctl[idChkOpen])
	files := append([]string(nil), inputs...)
	saveSettings()
	busy = true
	win.Enable(ctl[idBtnRun], false)
	win.Enable(ctl[idBtnIn], false)
	win.SetText(ctl[idBtnRun], "Đang xử lý...")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logLine(fmt.Sprintf("✗ Lỗi không mong muốn: %v", r))
			}
			win.Post(hwnd, wmDone, 0, 0)
		}()
		okN := 0
		for i, f := range files {
			logLine(fmt.Sprintf("\r\n══════ [%d/%d] %s ══════", i+1, len(files), filepath.Base(f)))
			o := opt
			o.Input = f
			res, err := core.Process(o, logLine)
			if err != nil {
				logLine("✗ LỖI: " + err.Error())
				continue
			}
			okN++
			mu.Lock()
			toOpen = append(toOpen, "#"+res.Docx) // ghi nhớ file kết quả cuối
			if openAfter && len(files) == 1 {
				if res.PDF != "" {
					toOpen = append(toOpen, res.PDF)
				}
				toOpen = append(toOpen, res.Docx)
			}
			mu.Unlock()
		}
		logLine(fmt.Sprintf("\r\nHoàn tất: %d/%d file. Anh/chị kiểm tra lại các chỗ \"...\" cần điền và nội dung trước khi trình ký.", okN, len(files)))
	}()
}

//go:embed HuongDanSuDung.txt
var guideText []byte

// openGuide mở file hướng dẫn (cạnh file chạy; nếu không có thì ghi ra thư mục tạm).
func openGuide() {
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), "HuongDanSuDung.txt")
		if st, err := os.Stat(p); err == nil && st.Size() == int64(len(guideText)) {
			win.ShellOpen(p)
			return
		}
	}
	p := filepath.Join(os.TempDir(), "DinhDangVanBanChuan_HuongDan_"+core.Version+".txt")
	if err := os.WriteFile(p, guideText, 0644); err == nil {
		win.ShellOpen(p)
	}
}
