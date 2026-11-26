package core

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// Item là một đơn vị nội dung đọc từ văn bản gốc: một dòng/đoạn văn hoặc một bảng dữ liệu.
type Item struct {
	Text   string
	Bold   bool
	Italic bool
	Table  [][]string // khác nil nếu là bảng dữ liệu
	Spans  [][]int    // số cột gộp (gridSpan) của từng ô
}

type dxCell struct {
	items []Item
	span  int
}

// ReadInput đọc .docx hoặc .txt (các định dạng khác phải chuyển sang .docx trước).
func ReadInput(path string) ([]Item, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".docx", ".docm", ".dotx":
		return readDocx(path)
	case ".txt", ".text":
		return readTxt(path)
	}
	return nil, fmt.Errorf("định dạng %s cần được chuyển sang .docx trước", filepath.Ext(path))
}

// ---------------- TXT ----------------

func readTxt(path string) ([]Item, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := decodeText(b)
	var out []Item
	for _, ln := range strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n") {
		ln = strings.TrimRight(ln, "\r")
		if strings.TrimSpace(ln) == "" {
			continue
		}
		out = append(out, Item{Text: ln})
	}
	if len(out) == 0 {
		return nil, errors.New("file văn bản trống")
	}
	return out, nil
}

func decodeText(b []byte) string {
	switch {
	case bytes.HasPrefix(b, []byte{0xEF, 0xBB, 0xBF}):
		return string(b[3:])
	case bytes.HasPrefix(b, []byte{0xFF, 0xFE}):
		return utf16Decode(b[2:], false)
	case bytes.HasPrefix(b, []byte{0xFE, 0xFF}):
		return utf16Decode(b[2:], true)
	}
	if utf8.Valid(b) {
		return string(b)
	}
	// Dự phòng: coi như Latin-1/Windows-1252
	r := make([]rune, len(b))
	for i, c := range b {
		r[i] = rune(c)
	}
	return string(r)
}

func utf16Decode(b []byte, be bool) string {
	u := make([]uint16, len(b)/2)
	for i := range u {
		if be {
			u[i] = uint16(b[2*i])<<8 | uint16(b[2*i+1])
		} else {
			u[i] = uint16(b[2*i+1])<<8 | uint16(b[2*i])
		}
	}
	return string(utf16.Decode(u))
}

// ---------------- DOCX ----------------

type dxPara struct {
	text      strings.Builder
	boldChars int
	italChars int
	allChars  int
}

type dxRun struct{ bold, italic bool }

func readDocx(path string) ([]Item, error) {
	_ = fmt.Sprint
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("không mở được file Word: %v", err)
	}
	defer zr.Close()
	var f *zip.File
	for _, z := range zr.File {
		if z.Name == "word/document.xml" {
			f = z
		}
	}
	if f == nil {
		return nil, errors.New("file không có word/document.xml")
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	dec := xml.NewDecoder(rc)
	var items []Item
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "p":
			items = append(items, parsePara(dec)...)
		case "tbl":
			rows, nested := parseTable(dec)
			items = append(items, nested...)
			items = append(items, tableToItems(rows)...)
		}
	}
	return items, nil
}

func onOff(se xml.StartElement) bool {
	for _, a := range se.Attr {
		if a.Name.Local == "val" {
			v := strings.ToLower(a.Value)
			return !(v == "0" || v == "false" || v == "off" || v == "none")
		}
	}
	return true
}

// parsePara đọc tới </w:p>; trả về các dòng (tách theo ngắt dòng mềm) và đoạn lồng trong textbox.
func parsePara(dec *xml.Decoder) []Item {
	var p dxPara
	var nested []Item
	var run dxRun
	var paraBold, paraItal bool
	inRPr, inPPr := false, false
	depth := 1
	add := func(s string) {
		p.text.WriteString(s)
		n := utf8.RuneCountInString(strings.TrimSpace(s))
		p.allChars += n
		if run.bold || paraBold {
			p.boldChars += n
		}
		if run.italic || paraItal {
			p.italChars += n
		}
	}
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p":
				nested = append(nested, parsePara(dec)...)
				continue
			case "tbl":
				rows, nn := parseTable(dec)
				nested = append(nested, nn...)
				nested = append(nested, tableToItems(rows)...)
				continue
			case "pPr":
				inPPr = true
			case "rPr":
				inRPr = true
			case "r":
				run = dxRun{}
			case "b":
				if inRPr && !inPPr {
					run.bold = onOff(t)
				}
			case "i":
				if inRPr && !inPPr {
					run.italic = onOff(t)
				}
			case "t":
				var s string
				dec.DecodeElement(&s, &t)
				add(s)
				continue
			case "tab":
				if !inPPr {
					add("\t")
				}
			case "br", "cr":
				if !inPPr {
					p.text.WriteString("\n")
				}
			case "noBreakHyphen":
				add("-")
			case "delText", "instrText":
				dec.Skip()
				continue
			}
			depth++
		case xml.EndElement:
			depth--
			switch t.Name.Local {
			case "pPr":
				inPPr = false
			case "rPr":
				inRPr = false
			}
		}
	}
	var out []Item
	bold := p.allChars > 0 && p.boldChars*10 >= p.allChars*8
	ital := p.allChars > 0 && p.italChars*10 >= p.allChars*8
	for _, ln := range strings.Split(p.text.String(), "\n") {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		out = append(out, Item{Text: ln, Bold: bold, Italic: ital})
	}
	return append(nested, out...)
}

// parseTable đọc tới </w:tbl>; trả về các hàng, mỗi ô gồm các Item và số cột gộp.
func parseTable(dec *xml.Decoder) ([][]dxCell, []Item) {
	var rows [][]dxCell
	var cur []dxCell
	var cell dxCell
	inCell, inTcPr := false, false
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "tr":
				cur = nil
			case "tc":
				cell = dxCell{span: 1}
				inCell = true
			case "tcPr":
				inTcPr = true
			case "gridSpan":
				if inTcPr {
					for _, a := range t.Attr {
						if a.Name.Local == "val" {
							fmt.Sscan(a.Value, &cell.span)
						}
					}
				}
			case "p":
				cell.items = append(cell.items, parsePara(dec)...)
				continue
			case "tbl": // bảng lồng: làm phẳng vào ô hiện tại
				r2, n2 := parseTable(dec)
				cell.items = append(cell.items, n2...)
				for _, r := range r2 {
					for _, c := range r {
						cell.items = append(cell.items, c.items...)
					}
				}
				continue
			}
			depth++
		case xml.EndElement:
			depth--
			switch t.Name.Local {
			case "tcPr":
				inTcPr = false
			case "tc":
				if inCell {
					if cell.span < 1 {
						cell.span = 1
					}
					cur = append(cur, cell)
				}
				inCell = false
			case "tr":
				rows = append(rows, cur)
			}
		}
	}
	return rows, nil
}

var reLayoutTable = regexp.MustCompile(`(?i)(^\s*số\s*[:.]|nơi\s+nhận|^\s*(tm|kt|q|tl|tuq)\.\s|^\s*v/v|ngày\s+\S+\s+tháng\s+\S+\s+năm)`)

// isLayoutText: dòng thuộc phần tiêu đề/chữ ký (bảng chứa nó là bảng bố cục, không phải bảng số liệu).
func isLayoutText(s string) bool {
	return reQuocHieu.MatchString(s) || reTieuNgu.MatchString(s) || reLayoutTable.MatchString(s)
}

// tableToItems: bảng bố cục (tiêu đề, chữ ký) được làm phẳng thành dòng; bảng dữ liệu giữ nguyên.
func tableToItems(rows [][]dxCell) []Item {
	layout := false
	for _, r := range rows {
		for _, c := range r {
			for _, it := range c.items {
				if isLayoutText(it.Text) {
					layout = true
				}
			}
		}
	}
	if layout {
		var out []Item
		// đọc theo cột trái trước rồi cột phải (cơ quan → Số; Quốc hiệu → ngày)
		maxc := 0
		for _, r := range rows {
			if len(r) > maxc {
				maxc = len(r)
			}
		}
		for c := 0; c < maxc; c++ {
			for _, r := range rows {
				if c < len(r) {
					out = append(out, r[c].items...)
				}
			}
		}
		return out
	}
	var t [][]string
	var sp [][]int
	for _, r := range rows {
		var row []string
		var rs []int
		for _, c := range r {
			var parts []string
			for _, it := range c.items {
				parts = append(parts, strings.TrimSpace(it.Text))
			}
			row = append(row, strings.Join(parts, "\n"))
			rs = append(rs, c.span)
		}
		if len(row) > 0 {
			t = append(t, row)
			sp = append(sp, rs)
		}
	}
	if len(t) == 0 {
		return nil
	}
	return []Item{{Table: t, Spans: sp}}
}
