package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Options là tùy chọn xử lý một file.
type Options struct {
	Input   string
	OutDir  string // rỗng = cùng thư mục file gốc
	Profile string // auto | vbhc | vbqppl
	PDF     bool   // xuất thêm PDF
}

// Result là kết quả xử lý.
type Result struct {
	Docx, PDF string
}

// Process: đọc → nhận diện → chuẩn hóa → ghi .docx (và .pdf nếu chọn).
func Process(o Options, logf func(string)) (*Result, error) {
	ext := strings.ToLower(filepath.Ext(o.Input))
	switch ext {
	case ".docx", ".doc", ".txt", ".text", ".pdf", ".rtf", ".odt", ".docm":
	default:
		return nil, fmt.Errorf("không hỗ trợ định dạng %s (chỉ nhận .docx, .doc, .txt, .pdf)", ext)
	}
	if _, err := os.Stat(o.Input); err != nil {
		return nil, fmt.Errorf("không tìm thấy file: %s", o.Input)
	}
	outDir := o.OutDir
	if outDir == "" {
		outDir = filepath.Dir(o.Input)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, err
	}
	base := strings.TrimSuffix(filepath.Base(o.Input), filepath.Ext(o.Input))
	src := o.Input
	tmpDir, _ := os.MkdirTemp("", "ddvbc")
	defer os.RemoveAll(tmpDir)

	if ext == ".doc" || ext == ".pdf" || ext == ".rtf" || ext == ".odt" {
		logf(fmt.Sprintf("Chuyển %s sang .docx (Microsoft Word, hoặc LibreOffice nếu không có Word)...", ext))
		conv := filepath.Join(tmpDir, "input.docx")
		via, err := ConvertToDocx(o.Input, conv)
		if err != nil {
			return nil, fmt.Errorf("không chuyển được %s sang .docx: %v", ext, err)
		}
		logf("  → đã chuyển bằng " + via)
		src = conv
	}
	logf("Đọc nội dung văn bản...")
	items, err := ReadInput(src)
	if err != nil {
		return nil, err
	}
	logf(fmt.Sprintf("  → %d dòng/đoạn, %d bảng", countText(items), len(items)-countText(items)))
	d, logs := Parse(items, o.Profile)
	for _, l := range logs {
		logf(l)
	}
	out := uniquePath(filepath.Join(outDir, base+"_chuanhoa.docx"))
	if err := BuildDocx(d, out); err != nil {
		return nil, fmt.Errorf("lỗi ghi file Word: %v", err)
	}
	res := &Result{Docx: out}
	logf("Đã tạo file Word: " + out)
	if o.PDF {
		pdf := strings.TrimSuffix(out, ".docx") + ".pdf"
		logf("Xuất PDF...")
		via, err := ConvertToPDF(out, pdf)
		if err != nil {
			logf("  ✗ Không xuất được PDF: " + err.Error())
		} else {
			res.PDF = pdf
			logf("Đã tạo file PDF (bằng " + via + "): " + pdf)
		}
	}
	return res, nil
}

func countText(items []Item) int {
	n := 0
	for _, it := range items {
		if it.Table == nil {
			n++
		}
	}
	return n
}

// uniquePath tránh ghi đè: thêm (2), (3)... nếu file đã có và đang bị khóa/không ghi được.
func uniquePath(p string) string {
	if _, err := os.Stat(p); err != nil {
		return p
	}
	if f, err := os.OpenFile(p, os.O_WRONLY, 0); err == nil {
		f.Close()
		return p // ghi đè bản chuẩn hóa cũ (không phải bản gốc)
	}
	ext := filepath.Ext(p)
	b := strings.TrimSuffix(p, ext)
	for i := 2; ; i++ {
		q := fmt.Sprintf("%s (%d)%s", b, i, ext)
		if _, err := os.Stat(q); err != nil {
			return q
		}
	}
}
