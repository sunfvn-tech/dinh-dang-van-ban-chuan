//go:build windows

package core

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
)

// Script PowerShell điều khiển Microsoft Word (đường dẫn truyền qua biến môi trường để tránh lỗi ký tự).
const wordScript = `
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
try { [Console]::OutputEncoding = [System.Text.Encoding]::UTF8 } catch {}
try {
  foreach ($v in '16.0','15.0') {
    $k = "HKCU:\Software\Microsoft\Office\$v\Word\Options"
    if (Test-Path $k) { New-ItemProperty -Path $k -Name DisableConvertPdfWarning -Value 1 -PropertyType DWord -Force | Out-Null }
  }
} catch {}
# Word đã mở sẵn? (Office 365 thường gắn vào Word đang chạy) → không ẩn, không đóng Word của người dùng
$pre = @(Get-Process WINWORD -ErrorAction SilentlyContinue).Count -gt 0
$w = $null; $d = $null; $err = ''
try {
  $w = New-Object -ComObject Word.Application
  if (-not $pre) { $w.Visible = $false }
  $w.DisplayAlerts = 0
  try { $w.AutomationSecurity = 3 } catch {}
  $d = $w.Documents.Open($env:DDVB_IN, $false, $true, $false)
  if ($env:DDVB_FMT -eq '17') { $d.ExportAsFixedFormat($env:DDVB_OUT, 17) } else { $d.SaveAs2($env:DDVB_OUT, 16) }
} catch { $err = $_.Exception.Message }
# Chỉ đóng tài liệu do chương trình mở; chỉ thoát Word nếu Word do chương trình khởi động
if ($d) { try { $d.Close(0) } catch {} }
if ($w -and -not $pre) { try { $w.Quit(0) } catch {} }
if ($w) { try { [void][System.Runtime.InteropServices.Marshal]::ReleaseComObject($w) } catch {} }
if ((Test-Path -LiteralPath $env:DDVB_OUT) -and ((Get-Item -LiteralPath $env:DDVB_OUT).Length -gt 0)) { exit 0 }
if ($err -eq '') { $err = 'Word không tạo được file kết quả' }
[Console]::Error.WriteLine($err)
exit 3
`

var reCLIXML = regexp.MustCompile(`<S S="Error">([^<]*)</S>`)

// cleanPSError: rút gọn thông báo lỗi PowerShell dạng CLIXML thành chữ đọc được.
func cleanPSError(s string) string {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "CLIXML") {
		var parts []string
		for _, m := range reCLIXML.FindAllStringSubmatch(s, -1) {
			parts = append(parts, m[1])
		}
		s = strings.Join(parts, "")
		s = strings.ReplaceAll(s, "_x000D__x000A_", " ")
		s = strings.ReplaceAll(s, "&quot;", "\"")
	}
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > 300 {
		s = string(r[:300]) + "..."
	}
	return s
}

func encodePS(s string) string {
	u := utf16.Encode([]rune(s))
	b := make([]byte, len(u)*2)
	for i, c := range u {
		b[2*i] = byte(c)
		b[2*i+1] = byte(c >> 8)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func hidden(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}

func viaWord(in, out string, fmtCode int) error {
	err := viaWordOnce(in, out, fmtCode)
	if err != nil {
		os.Remove(out)
		if err2 := viaWordOnce(in, out, fmtCode); err2 == nil {
			return nil
		}
	}
	return err
}

func viaWordOnce(in, out string, fmtCode int) error {
	start := time.Now()
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-EncodedCommand", encodePS(wordScript))
	abs, _ := filepath.Abs(in)
	absOut, _ := filepath.Abs(out)
	cmd.Env = append(os.Environ(), "DDVB_IN="+abs, "DDVB_OUT="+absOut, fmt.Sprintf("DDVB_FMT=%d", fmtCode))
	hidden(cmd)
	b, err := cmd.CombinedOutput()
	if st, e := os.Stat(out); e == nil && st.Size() > 0 && !st.ModTime().Before(start.Add(-2*time.Second)) {
		return nil // file đã tạo xong, bỏ qua lỗi khi đóng Word
	}
	if err != nil {
		msg := cleanPSError(string(b))
		if msg == "" {
			msg = err.Error()
		}
		return errors.New(msg)
	}
	if st, e := os.Stat(out); e != nil || st.Size() == 0 {
		return errors.New("Word không tạo được file kết quả")
	}
	return nil
}

func findSoffice() string {
	if p, err := exec.LookPath("soffice.exe"); err == nil {
		return p
	}
	for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)", "ProgramW6432"} {
		if d := os.Getenv(env); d != "" {
			p := filepath.Join(d, "LibreOffice", "program", "soffice.exe")
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}

func viaSoffice(in, out, format, ext string) error {
	so := findSoffice()
	if so == "" {
		return errors.New("không có LibreOffice")
	}
	dir, _ := os.MkdirTemp("", "so")
	defer os.RemoveAll(dir)
	args := []string{"--headless", "--convert-to", format, "--outdir", dir, in}
	if strings.EqualFold(filepath.Ext(in), ".pdf") {
		args = append([]string{"--infilter=writer_pdf_import"}, args...)
	}
	cmd := exec.Command(so, args...)
	hidden(cmd)
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%v %s", err, b)
	}
	gen := filepath.Join(dir, strings.TrimSuffix(filepath.Base(in), filepath.Ext(in))+ext)
	b, err := os.ReadFile(gen)
	if err != nil {
		return errors.New("LibreOffice không tạo được file")
	}
	return os.WriteFile(out, b, 0644)
}

// ConvertToDocx chuyển .doc/.pdf/.rtf/.odt → .docx (ưu tiên Microsoft Word).
func ConvertToDocx(in, out string) (string, error) {
	e1 := viaWord(in, out, 16)
	if e1 == nil {
		return "Microsoft Word", nil
	}
	e2 := viaSoffice(in, out, "docx:MS Word 2007 XML", ".docx")
	if e2 == nil {
		return "LibreOffice", nil
	}
	return "", fmt.Errorf("Word: %v; LibreOffice: %v", e1, e2)
}

// ConvertToPDF chuyển .docx → .pdf (ưu tiên Microsoft Word).
func ConvertToPDF(in, out string) (string, error) {
	e1 := viaWord(in, out, 17)
	if e1 == nil {
		return "Microsoft Word", nil
	}
	e2 := viaSoffice(in, out, "pdf", ".pdf")
	if e2 == nil {
		return "LibreOffice", nil
	}
	return "", fmt.Errorf("cần Microsoft Word hoặc LibreOffice. Word: %v; LibreOffice: %v", e1, e2)
}
