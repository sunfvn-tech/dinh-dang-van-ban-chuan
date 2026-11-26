//go:build !windows

package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func soffice(in, out, format string) error {
	dir, _ := os.MkdirTemp("", "so")
	defer os.RemoveAll(dir)
	cmd := exec.Command("soffice", "--headless", "--convert-to", format, "--outdir", dir, in)
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%v: %s", err, b)
	}
	ext := "." + format
	if format == "docx:MS Word 2007 XML" {
		ext = ".docx"
	}
	gen := filepath.Join(dir, trimExt(filepath.Base(in))+ext)
	b, err := os.ReadFile(gen)
	if err != nil {
		return err
	}
	return os.WriteFile(out, b, 0644)
}

func trimExt(s string) string { return s[:len(s)-len(filepath.Ext(s))] }

// ConvertToDocx chuyển .doc/.pdf → .docx.
func ConvertToDocx(in, out string) (string, error) {
	if filepath.Ext(in) == ".pdf" {
		return "", fmt.Errorf("trên Linux chỉ hỗ trợ PDF qua Word (Windows)")
	}
	return "LibreOffice", soffice(in, out, "docx:MS Word 2007 XML")
}

// ConvertToPDF chuyển .docx → .pdf.
func ConvertToPDF(in, out string) (string, error) {
	return "LibreOffice", soffice(in, out, "pdf")
}
