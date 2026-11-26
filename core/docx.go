package core

import (
	"archive/zip"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// Đơn vị: twip (1/20 pt). 1 mm = 56.6929 twip.
func mm(v float64) int { return int(v*56.6929 + 0.5) }

const (
	pageW, pageH          = 11906, 16838 // A4
	marTop, marBot        = 1134, 1134   // 20 mm
	marLeft, marRight     = 1701, 851    // 30 mm, 15 mm
	textW                 = pageW - marLeft - marRight
	cellPad               = 57 // 1 mm mỗi bên trong bảng bố cục
	firstIndent           = 720 // 1,27 cm
	defaultLine           = 276 // 1,15 dòng
	fontName              = "Times New Roman"
)

// TextWidthTwip ước tính độ rộng chuỗi theo phông Times New Roman.
func TextWidthTwip(s string, sizePt float64, bold bool) int {
	tbl := widthRegular
	if bold {
		tbl = widthBold
	}
	sum := 0
	for _, r := range s {
		w, ok := tbl[r]
		if !ok {
			w = tbl['n']
		}
		sum += int(w)
	}
	return int(float64(sum) / fontUPM * sizePt * 20)
}

func esc(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;")
	return r.Replace(s)
}

type rs struct {
	text           string
	size           float64
	bold, ital, ul bool
}

func run(r rs) string {
	var b strings.Builder
	b.WriteString("<w:r><w:rPr>")
	if r.bold {
		b.WriteString("<w:b/><w:bCs/>")
	}
	if r.ital {
		b.WriteString("<w:i/><w:iCs/>")
	}
	if r.ul {
		b.WriteString(`<w:u w:val="single"/>`)
	}
	if r.size > 0 {
		fmt.Fprintf(&b, `<w:sz w:val="%d"/><w:szCs w:val="%d"/>`, int(r.size*2), int(r.size*2))
	}
	b.WriteString("</w:rPr>")
	parts := strings.Split(r.text, "\t")
	for i, p := range parts {
		if i > 0 {
			b.WriteString("<w:tab/>")
		}
		if p != "" {
			fmt.Fprintf(&b, `<w:t xml:space="preserve">%s</w:t>`, esc(p))
		}
	}
	b.WriteString("</w:r>")
	return b.String()
}

type pp struct {
	jc                    string // left, center, right, both
	first, left, right    int
	hanging               int
	before, after         float64 // pt
	line                  int     // 240 = đơn
	exact                 bool
	keepNext, pageBefore  bool
	borderBottom          bool
	borderTop             bool
	markSize              float64
	tabs                  []int
}

func para(p pp, runs ...string) string {
	var b strings.Builder
	b.WriteString("<w:p><w:pPr>")
	if p.keepNext {
		b.WriteString("<w:keepNext/>")
	}
	if p.pageBefore {
		b.WriteString("<w:pageBreakBefore/>")
	}
	if p.borderTop {
		b.WriteString(`<w:pBdr><w:top w:val="single" w:sz="6" w:space="0" w:color="000000"/></w:pBdr>`)
	}
	if p.borderBottom {
		b.WriteString(`<w:pBdr><w:bottom w:val="single" w:sz="6" w:space="1" w:color="000000"/></w:pBdr>`)
	}
	if len(p.tabs) > 0 {
		b.WriteString("<w:tabs>")
		for _, t := range p.tabs {
			fmt.Fprintf(&b, `<w:tab w:val="left" w:pos="%d"/>`, t)
		}
		b.WriteString("</w:tabs>")
	}
	line := p.line
	if line == 0 {
		line = 240
	}
	rule := "auto"
	if p.exact {
		rule = "exact"
	}
	fmt.Fprintf(&b, `<w:spacing w:before="%d" w:after="%d" w:line="%d" w:lineRule="%s"/>`, int(p.before*20), int(p.after*20), line, rule)
	if p.first != 0 || p.left != 0 || p.right != 0 || p.hanging != 0 {
		b.WriteString("<w:ind")
		if p.left != 0 {
			fmt.Fprintf(&b, ` w:left="%d"`, p.left)
		}
		if p.right != 0 {
			fmt.Fprintf(&b, ` w:right="%d"`, p.right)
		}
		if p.hanging != 0 {
			fmt.Fprintf(&b, ` w:hanging="%d"`, p.hanging)
		} else if p.first != 0 {
			fmt.Fprintf(&b, ` w:firstLine="%d"`, p.first)
		}
		b.WriteString("/>")
	}
	if p.jc != "" {
		fmt.Fprintf(&b, `<w:jc w:val="%s"/>`, p.jc)
	}
	if p.markSize > 0 {
		fmt.Fprintf(&b, `<w:rPr><w:sz w:val="%d"/><w:szCs w:val="%d"/></w:rPr>`, int(p.markSize*2), int(p.markSize*2))
	}
	b.WriteString("</w:pPr>")
	for _, r := range runs {
		b.WriteString(r)
	}
	b.WriteString("</w:p>")
	return b.String()
}

// hline: đường kẻ ngang nét liền dài width, căn giữa trong vùng avail.
func hline(width, avail int) string {
	if width > avail {
		width = avail
	}
	ind := (avail - width) / 2
	return para(pp{left: ind, right: ind, line: 100, exact: true, borderTop: true, markSize: 1, before: 3})
}

func centerLines(text string, size float64, bold, ital bool, before, after float64) string {
	var b strings.Builder
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		p := pp{jc: "center"}
		if i == 0 {
			p.before = before
		}
		if i == len(lines)-1 {
			p.after = after
		}
		b.WriteString(para(p, run(rs{text: l, size: size, bold: bold, ital: ital})))
	}
	return b.String()
}

// wrapBalanced ngắt dòng quá dài thành 2 dòng gần bằng nhau (tại khoảng trắng).
func wrapBalanced(text string, size float64, bold bool, limit int) string {
	var out []string
	for _, l := range strings.Split(text, "\n") {
		if l == "" || TextWidthTwip(l, size, bold) <= limit {
			out = append(out, l)
			continue
		}
		words := strings.Fields(l)
		best, bestDiff := 1, 1<<30
		for k := 1; k < len(words); k++ {
			a := TextWidthTwip(strings.Join(words[:k], " "), size, bold)
			b := TextWidthTwip(strings.Join(words[k:], " "), size, bold)
			diff := a - b
			if diff < 0 {
				diff = -diff
			}
			if diff < bestDiff {
				best, bestDiff = k, diff
			}
		}
		out = append(out, strings.Join(words[:best], " "), strings.Join(words[best:], " "))
	}
	return strings.Join(out, "\n")
}

func maxWidth(text string, size float64, bold bool) int {
	m := 0
	for _, l := range strings.Split(text, "\n") {
		if w := TextWidthTwip(l, size, bold); w > m {
			m = w
		}
	}
	return m
}

func tblOpen(widths []int, borders bool) string {
	var b strings.Builder
	total := 0
	for _, w := range widths {
		total += w
	}
	fmt.Fprintf(&b, `<w:tbl><w:tblPr><w:tblW w:w="%d" w:type="dxa"/><w:jc w:val="center"/><w:tblLayout w:type="fixed"/>`, total)
	if borders {
		b.WriteString(`<w:tblBorders><w:top w:val="single" w:sz="4" w:color="000000"/><w:left w:val="single" w:sz="4" w:color="000000"/><w:bottom w:val="single" w:sz="4" w:color="000000"/><w:right w:val="single" w:sz="4" w:color="000000"/><w:insideH w:val="single" w:sz="4" w:color="000000"/><w:insideV w:val="single" w:sz="4" w:color="000000"/></w:tblBorders>`)
		b.WriteString(`<w:tblCellMar><w:top w:w="28" w:type="dxa"/><w:left w:w="85" w:type="dxa"/><w:bottom w:w="28" w:type="dxa"/><w:right w:w="85" w:type="dxa"/></w:tblCellMar>`)
	} else {
		b.WriteString(`<w:tblBorders><w:top w:val="nil"/><w:left w:val="nil"/><w:bottom w:val="nil"/><w:right w:val="nil"/><w:insideH w:val="nil"/><w:insideV w:val="nil"/></w:tblBorders>`)
		fmt.Fprintf(&b, `<w:tblCellMar><w:left w:w="%d" w:type="dxa"/><w:right w:w="%d" w:type="dxa"/></w:tblCellMar>`, cellPad, cellPad)
	}
	b.WriteString(`</w:tblPr><w:tblGrid>`)
	for _, w := range widths {
		fmt.Fprintf(&b, `<w:gridCol w:w="%d"/>`, w)
	}
	b.WriteString("</w:tblGrid>")
	return b.String()
}

func cell(width int, content string, header bool) string {
	if content == "" {
		content = "<w:p/>"
	}
	shade := ""
	_ = header
	return fmt.Sprintf(`<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/>%s</w:tcPr>%s</w:tc>`, width, shade, content)
}

// BuildDocx dựng file .docx đúng thể thức từ Doc.
func BuildDocx(d *Doc, out string) error {
	var body strings.Builder
	W := body.WriteString
	lineSp := defaultLine

	// ----- Bảng tiêu đề -----
	// Cột phải tối thiểu đủ chứa Quốc hiệu cỡ 13 (nếu không được thì cỡ 12)
	qh := "CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM"
	qhSize := 13.0
	rightMin := TextWidthTwip(qh, 13, true) + 2*cellPad + 80
	if rightMin > textW*58/100 {
		qhSize = 12
		rightMin = TextWidthTwip(qh, 12, true) + 2*cellPad + 80
	}
	leftMax := textW - rightMin
	cqSize := 13.0
	// Tên cơ quan dài quá cột trái thì ngắt thành 2 dòng cân đối
	d.CoQuanBanHanh = wrapBalanced(d.CoQuanBanHanh, cqSize, true, leftMax-2*cellPad-60)
	d.CoQuanChuQuan = wrapBalanced(d.CoQuanChuQuan, cqSize, false, leftMax-2*cellPad-60)
	need := textW * 40 / 100
	for _, w := range []int{maxWidth(d.CoQuanBanHanh, cqSize, true), maxWidth(d.CoQuanChuQuan, cqSize, false)} {
		if w+2*cellPad+60 > need {
			need = w + 2*cellPad + 60
		}
	}
	if need > leftMax {
		need = leftMax
		cqSize = 12
	}
	leftW := need
	rightW := textW - leftW
	innerL, innerR := leftW-2*cellPad, rightW-2*cellPad

	var left strings.Builder
	if d.CoQuanChuQuan != "" {
		left.WriteString(centerLines(d.CoQuanChuQuan, cqSize, false, false, 0, 0))
	}
	ban := d.CoQuanBanHanh
	if ban == "" {
		ban = "..."
	}
	left.WriteString(centerLines(ban, cqSize, true, false, 0, 0))
	bw := maxWidth(ban, cqSize, true)
	lw := bw * 2 / 5
	if lw < mm(18) {
		lw = mm(18)
	}
	left.WriteString(hline(lw, innerL))

	var right strings.Builder
	right.WriteString(centerLines(qh, qhSize, true, false, 0, 0))
	tn := "Độc lập - Tự do - Hạnh phúc"
	right.WriteString(centerLines(tn, 14, true, false, 0, 0))
	right.WriteString(hline(TextWidthTwip(tn, 14, true), innerR))

	so := d.SoKyHieu
	if so == "" {
		so = "Số:      /"
	}
	var left2 strings.Builder
	left2.WriteString(para(pp{jc: "center", before: 6}, run(rs{text: so, size: 13})))
	if d.CongVan {
		ty := d.TrichYeu
		if ty == "" {
			ty = "V/v ..."
		}
		left2.WriteString(para(pp{jc: "center", before: 3}, run(rs{text: ty, size: 12})))
	}
	date := d.DiaDanhNgay
	if date == "" {
		date = "..., ngày ... tháng ... năm ..."
	}
	right2 := para(pp{jc: "center", before: 6}, run(rs{text: date, size: 14, ital: true}))

	W(tblOpen([]int{leftW, rightW}, false))
	W("<w:tr>" + cell(leftW, left.String(), false) + cell(rightW, right.String(), false) + "</w:tr>")
	W("<w:tr>" + cell(leftW, left2.String(), false) + cell(rightW, right2, false) + "</w:tr>")
	W("</w:tbl>")

	// ----- Tên loại, trích yếu -----
	if !d.CongVan {
		tl := d.TenLoai
		if tl == "" {
			tl = "..."
		}
		W(para(pp{jc: "center", before: 18, keepNext: true}, run(rs{text: tl, size: 14, bold: true})))
		if d.TrichYeu != "" {
			W(para(pp{jc: "center", keepNext: true}, run(rs{text: d.TrichYeu, size: 14, bold: true})))
			if d.Profile == ProfileVBHC {
				tw := TextWidthTwip(d.TrichYeu, 14, true)
				if tw > textW {
					tw = textW
				}
				lw := tw * 2 / 5
				if lw < mm(25) {
					lw = mm(25)
				}
				W(hline(lw, textW))
			}
		}
	}
	// ----- Kính gửi -----
	if len(d.KinhGui) > 0 {
		lbl := "Kính gửi: "
		lblW := TextWidthTwip(lbl, 14, false)
		if len(d.KinhGui) == 1 {
			W(para(pp{jc: "center", before: 12, after: 6}, run(rs{text: lbl + d.KinhGui[0] + ".", size: 14})))
		} else {
			x := mm(20)
			for i, k := range d.KinhGui {
				end := ";"
				if i == len(d.KinhGui)-1 {
					end = "."
				}
				p := pp{left: x + lblW, line: lineSp}
				txt := "- " + k + end
				if i == 0 {
					p.hanging = lblW
					p.before = 12
					txt = lbl + txt
				}
				if i == len(d.KinhGui)-1 {
					p.after = 6
				}
				W(para(p, run(rs{text: txt, size: 14})))
			}
		}
	}
	// ----- Thẩm quyền, căn cứ -----
	if d.ThamQuyen != "" {
		W(para(pp{jc: "center", before: 12, after: 6, keepNext: true}, run(rs{text: d.ThamQuyen, size: 14, bold: true})))
	}
	for i, c := range d.CanCu {
		p := pp{jc: "both", first: firstIndent, after: 6, line: lineSp}
		if i == 0 && d.ThamQuyen == "" {
			p.before = 12
		}
		W(para(p, run(rs{text: c, size: 14, ital: true})))
	}
	if d.BanHanh != "" {
		W(para(pp{jc: "both", first: firstIndent, after: 6, line: lineSp}, run(rs{text: d.BanHanh, size: 14, ital: true})))
	}
	if d.QuyetDinhLine != "" {
		W(para(pp{jc: "center", before: 6, after: 6, keepNext: true}, run(rs{text: d.QuyetDinhLine + ":", size: 14, bold: true})))
	}
	if d.ThamQuyen == "" && len(d.CanCu) == 0 && len(d.KinhGui) == 0 {
		W(para(pp{line: 120, exact: true, markSize: 4}))
	}

	// ----- Nội dung -----
	hasSig := len(d.NoiNhan) > 0 || len(d.Ky.ChucDanh) > 0 || d.Ky.HoTen != "" || len(d.Parties) > 0
	W(blocksXML(d.Body, lineSp, hasSig))

	// ----- Ký hai bên (hợp đồng, biên bản): hàng 1 chức danh, hàng 2 họ tên → tên các bên thẳng hàng -----
	if len(d.Parties) > 0 {
		half := textW / len(d.Parties)
		var widths []int
		var top, bot strings.Builder
		for _, p := range d.Parties {
			var c strings.Builder
			c.WriteString(para(pp{jc: "center", keepNext: true}, run(rs{text: p.Label, size: 13, bold: true})))
			if p.ChucVu != "" {
				c.WriteString(para(pp{jc: "center", keepNext: true}, run(rs{text: p.ChucVu, size: 13, bold: true})))
			}
			top.WriteString(cell(half, c.String(), false))
			var n strings.Builder
			for j := 0; j < 4; j++ {
				n.WriteString(para(pp{jc: "center", keepNext: true}, run(rs{text: "", size: 14})))
			}
			ht := p.HoTen
			if ht == "" {
				ht = "..."
			}
			n.WriteString(para(pp{jc: "center"}, run(rs{text: ht, size: 14, bold: true})))
			bot.WriteString(cell(half, n.String(), false))
			widths = append(widths, half)
		}
		W(para(pp{before: 6, keepNext: true}))
		W(tblOpen(widths, false))
		W("<w:tr><w:trPr><w:cantSplit/></w:trPr>" + top.String() + "</w:tr>")
		W("<w:tr><w:trPr><w:cantSplit/></w:trPr>" + bot.String() + "</w:tr></w:tbl>")
	}

	// ----- Nơi nhận, chữ ký -----
	if len(d.NoiNhan) > 0 || len(d.Ky.ChucDanh) > 0 || d.Ky.HoTen != "" {
		lw := textW * 45 / 100
		rw := textW - lw
		var l strings.Builder
		if len(d.NoiNhan) > 0 {
			l.WriteString(para(pp{}, run(rs{text: "Nơi nhận:", size: 12, bold: true, ital: true})))
			for i, x := range d.NoiNhan {
				end := ";"
				if i == len(d.NoiNhan)-1 {
					end = "."
				}
				hw := TextWidthTwip("- ", 11, false)
				l.WriteString(para(pp{left: hw, hanging: hw}, run(rs{text: "- " + x + end, size: 11})))
			}
		}
		var r strings.Builder
		for _, c := range d.Ky.ChucDanh {
			r.WriteString(para(pp{jc: "center", keepNext: true}, run(rs{text: c, size: 13, bold: true})))
		}
		blank := 4
		if d.Ky.GhiChu != "" {
			r.WriteString(para(pp{jc: "center"}))
			r.WriteString(para(pp{jc: "center"}, run(rs{text: d.Ky.GhiChu, size: 14, ital: true})))
			blank = 2
		}
		for j := 0; j < blank; j++ {
			r.WriteString(para(pp{jc: "center", keepNext: true}, run(rs{text: "", size: 14})))
		}
		ht := d.Ky.HoTen
		if ht == "" && len(d.Ky.ChucDanh) > 0 {
			ht = "..."
		}
		if ht != "" {
			r.WriteString(para(pp{jc: "center"}, run(rs{text: ht, size: 14, bold: true})))
		}
		W(para(pp{after: 0, before: 6, keepNext: true}))
		W(tblOpen([]int{lw, rw}, false))
		W("<w:tr><w:trPr><w:cantSplit/></w:trPr>" + cell(lw, l.String(), false) + cell(rw, r.String(), false) + "</w:tr></w:tbl>")
	}

	// ----- Phần kèm theo -----
	if len(d.KemTheo) > 0 && d.KemTheo[0].Kind != "annex" {
		W(para(pp{}, `<w:r><w:br w:type="page"/></w:r>`))
	}
	if len(d.KemTheo) > 0 {
		W(blocksXML(d.KemTheo, lineSp))
	}

	return writePackage(out, body.String(), d)
}

func blocksXML(bs []Block, lineSp int, keepLast ...bool) string {
	var b strings.Builder
	W := b.WriteString
	for idx, k := range bs {
		base := pp{jc: "both", first: firstIndent, after: 6, line: lineSp}
		if len(keepLast) > 0 && keepLast[0] && idx == len(bs)-1 {
			base.keepNext = true // đoạn cuối đi cùng khối chữ ký, không để chữ ký đứng riêng một trang
		}
		switch k.Kind {
		case "chuong":
			lbl := k.Label
			if lbl == "" {
				lbl = "Chương"
			}
			W(para(pp{jc: "center", before: 12, keepNext: true, line: lineSp}, run(rs{text: lbl + " " + k.So, size: 14, bold: true})))
			if k.Ten != "" {
				W(para(pp{jc: "center", after: 6, keepNext: true, line: lineSp}, run(rs{text: strings.ToUpper(k.Ten), size: 14, bold: true})))
			}
		case "muc":
			W(para(pp{jc: "center", before: 6, keepNext: true, line: lineSp}, run(rs{text: "Mục " + k.So, size: 14, bold: true})))
			if k.Ten != "" {
				W(para(pp{jc: "center", after: 6, keepNext: true, line: lineSp}, run(rs{text: strings.ToUpper(k.Ten), size: 14, bold: true})))
			}
		case "dieu":
			runs := []string{run(rs{text: "Điều " + k.So + ". ", size: 14, bold: true})}
			if k.Ten != "" {
				runs = append(runs, run(rs{text: k.Ten, size: 14, bold: true}))
				base.keepNext = true
			} else {
				runs = append(runs, run(rs{text: k.Text, size: 14}))
			}
			W(para(base, runs...))
		case "h1":
			base.keepNext = true
			W(para(base, run(rs{text: k.So + ". " + k.Text, size: 14, bold: true})))
		case "dash":
			W(para(base, run(rs{text: "- " + k.Text, size: 14, bold: k.Bold})))
		case "center":
			W(para(pp{jc: "center", after: 6, line: lineSp}, run(rs{text: k.Text, size: 14, bold: k.Bold, ital: k.Italic})))
		case "annex":
			W(para(pp{jc: "center", pageBefore: true, after: 6, keepNext: true, line: lineSp}, run(rs{text: k.Text, size: 14, bold: true})))
		case "pagebreak":
			W(para(pp{}, `<w:r><w:br w:type="page"/></w:r>`))
		case "table":
			W(tableXML(k.Rows, k.Spans))
			W(para(pp{after: 6}))
		default:
			if k.Bold {
				base.keepNext = len([]rune(k.Text)) < 120
			}
			W(para(base, run(rs{text: k.Text, size: 14, bold: k.Bold, ital: k.Italic})))
		}
	}
	return b.String()
}

var reNumCell = regexp.MustCompile(`^[\d\s.,%/()+\-–:]*$`)

func isTotalRow(r []string) bool {
	for i, v := range r {
		if i > 1 {
			break
		}
		t := strings.ToLower(strings.TrimSpace(v))
		if strings.HasPrefix(t, "tổng") || strings.HasPrefix(t, "cộng") {
			return true
		}
	}
	return false
}

func longestWord(s string, size float64, bold bool) int {
	m := 0
	for _, w := range strings.Fields(s) {
		if x := TextWidthTwip(w, size, bold); x > m {
			m = x
		}
	}
	return m
}

// tableXML dựng bảng số liệu: độ rộng cột theo nội dung thực, không bẻ đôi số/từ.
func tableXML(rows [][]string, spans [][]int) string {
	spanOf := func(r, c int) int {
		if r < len(spans) && c < len(spans[r]) && spans[r][c] > 1 {
			return spans[r][c]
		}
		return 1
	}
	cols := 0
	for r, row := range rows {
		n := 0
		for c := range row {
			n += spanOf(r, c)
		}
		if n > cols {
			cols = n
		}
	}
	if cols == 0 {
		return ""
	}
	size := 13.0
	if cols >= 6 {
		size = 12
	}
	const cellMar = 85 // 1,5 mm mỗi bên
	pad := 2*cellMar + mm(1.5) // lề ô + dự phòng chênh lệch đo chữ giữa Word và LibreOffice
	// Đo từng cột: bodyNat = chữ dài nhất của các ô dữ liệu (không xuống dòng),
	// hdrNat = tiêu đề trên một dòng, minW = từ dài nhất (không bao giờ cắt giữa từ).
	minW := make([]int, cols)
	bodyNat := make([]int, cols)
	hdrNat := make([]int, cols)
	numeric := make([]bool, cols)
	for c := range numeric {
		numeric[c] = true
	}
	for r, row := range rows {
		g := 0
		total := isTotalRow(row)
		for c, v := range row {
			sp := spanOf(r, c)
			if sp == 1 && g < cols {
				bold := r == 0 || total
				for _, ln := range strings.Split(v, "\n") {
					if w := longestWord(ln, size, bold)*103/100 + pad; w > minW[g] {
						minW[g] = w
					}
					w := TextWidthTwip(ln, size, bold)*103/100 + pad
					if r == 0 {
						if w > hdrNat[g] {
							hdrNat[g] = w
						}
					} else if w > bodyNat[g] {
						bodyNat[g] = w
					}
				}
				if r > 0 && !total && !reNumCell.MatchString(strings.TrimSpace(v)) {
					numeric[g] = false
				}
			}
			g += sp
		}
	}
	need := make([]int, cols)
	sumNeed := 0
	for c := range need {
		if minW[c] < mm(8) {
			minW[c] = mm(8)
		}
		need[c] = bodyNat[c]
		if minW[c] > need[c] {
			need[c] = minW[c]
		}
		sumNeed += need[c]
	}
	widths := make([]int, cols)
	if sumNeed <= textW {
		// Đủ chỗ: dữ liệu không xuống dòng; phần dư ưu tiên cho tiêu đề đang bị xuống dòng nhiều nhất
		copy(widths, need)
		extra := textW - sumNeed
		for extra > 0 {
			best, bestLines := -1, 1.0
			for c := range widths {
				if hdrNat[c] > widths[c] {
					lines := float64(hdrNat[c]-pad) / float64(widths[c]-pad)
					if lines > bestLines+0.01 {
						best, bestLines = c, lines
					}
				}
			}
			if best < 0 {
				break
			}
			step := mm(1)
			if step > extra {
				step = extra
			}
			widths[best] += step
			extra -= step
		}
		// Còn dư: chia cho các cột chữ theo tỉ lệ (cột số giữ gọn)
		if extra > 0 {
			sumText := 0
			for c := range widths {
				if !numeric[c] {
					sumText += widths[c]
				}
			}
			for c := range widths {
				if sumText > 0 && !numeric[c] {
					widths[c] += extra * widths[c] / sumText
				} else if sumText == 0 {
					widths[c] += extra / cols
				}
			}
		}
	} else {
		// Thiếu chỗ: thu hẹp cột chữ (cột dài thu nhiều hơn), cột số giữ nguyên
		over := sumNeed - textW
		flex := 0
		for c := range need {
			if !numeric[c] {
				flex += need[c] - minW[c]
			}
		}
		if flex >= over && flex > 0 {
			for c := range need {
				widths[c] = need[c]
				if !numeric[c] {
					widths[c] -= over * (need[c] - minW[c]) / flex
				}
			}
		} else {
			sumMin := 0
			for _, w := range minW {
				sumMin += w
			}
			for c := range widths {
				widths[c] = minW[c] * textW / sumMin
			}
		}
	}
	acc := 0
	for _, w := range widths {
		acc += w
	}
	widths[len(widths)-1] += textW - acc

	var b strings.Builder
	fmt.Fprintf(&b, `<w:tbl><w:tblPr><w:tblW w:w="%d" w:type="dxa"/><w:jc w:val="center"/><w:tblLayout w:type="fixed"/>`, textW)
	b.WriteString(`<w:tblBorders><w:top w:val="single" w:sz="4" w:color="000000"/><w:left w:val="single" w:sz="4" w:color="000000"/><w:bottom w:val="single" w:sz="4" w:color="000000"/><w:right w:val="single" w:sz="4" w:color="000000"/><w:insideH w:val="single" w:sz="4" w:color="000000"/><w:insideV w:val="single" w:sz="4" w:color="000000"/></w:tblBorders>`)
	fmt.Fprintf(&b, `<w:tblCellMar><w:top w:w="40" w:type="dxa"/><w:left w:w="%d" w:type="dxa"/><w:bottom w:w="40" w:type="dxa"/><w:right w:w="%d" w:type="dxa"/></w:tblCellMar></w:tblPr><w:tblGrid>`, cellMar, cellMar)
	for _, w := range widths {
		fmt.Fprintf(&b, `<w:gridCol w:w="%d"/>`, w)
	}
	b.WriteString("</w:tblGrid>")
	for r, row := range rows {
		header := r == 0
		total := isTotalRow(row)
		b.WriteString("<w:tr><w:trPr><w:cantSplit/>")
		if header {
			b.WriteString("<w:tblHeader/>")
		}
		b.WriteString("</w:trPr>")
		g := 0
		for c := 0; g < cols; c++ {
			v, sp := "", 1
			if c < len(row) {
				v, sp = row[c], spanOf(r, c)
			} else {
				sp = cols - g
			}
			w := 0
			for k := g; k < g+sp && k < cols; k++ {
				w += widths[k]
			}
			jc := "left"
			if header || (sp == 1 && numeric[g]) {
				jc = "center"
			}
			fmt.Fprintf(&b, `<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/>`, w)
			if sp > 1 {
				fmt.Fprintf(&b, `<w:gridSpan w:val="%d"/>`, sp)
			}
			b.WriteString(`<w:vAlign w:val="center"/></w:tcPr>`)
			for _, ln := range strings.Split(v, "\n") {
				b.WriteString(para(pp{jc: jc, markSize: size}, run(rs{text: ln, size: size, bold: header || total})))
			}
			b.WriteString("</w:tc>")
			g += sp
		}
		b.WriteString("</w:tr>")
	}
	b.WriteString("</w:tbl>")
	return b.String()
}

const nsMain = `xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"`

func writePackage(out, body string, d *Doc) error {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(f)
	put := func(name, content string) {
		w, _ := zw.Create(name)
		w.Write([]byte(content))
	}
	put("[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/><Override PartName="/word/settings.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"/><Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/><Override PartName="/word/header2.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/><Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/><Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/></Types>`)
	put("_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/><Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/></Relationships>`)
	put("word/_rels/document.xml.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings" Target="settings.xml"/><Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header1.xml"/><Relationship Id="rId4" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header2.xml"/></Relationships>`)
	put("word/styles.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles `+nsMain+`><w:docDefaults><w:rPrDefault><w:rPr><w:rFonts w:ascii="`+fontName+`" w:hAnsi="`+fontName+`" w:eastAsia="`+fontName+`" w:cs="`+fontName+`"/><w:color w:val="000000"/><w:sz w:val="28"/><w:szCs w:val="28"/><w:lang w:val="vi-VN" w:eastAsia="en-US" w:bidi="ar-SA"/></w:rPr></w:rPrDefault><w:pPrDefault><w:pPr><w:spacing w:after="0" w:line="240" w:lineRule="auto"/></w:pPr></w:pPrDefault></w:docDefaults><w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/><w:qFormat/></w:style><w:style w:type="table" w:default="1" w:styleId="TableNormal"><w:name w:val="Normal Table"/><w:tblPr><w:tblInd w:w="0" w:type="dxa"/><w:tblCellMar><w:top w:w="0" w:type="dxa"/><w:left w:w="108" w:type="dxa"/><w:bottom w:w="0" w:type="dxa"/><w:right w:w="108" w:type="dxa"/></w:tblCellMar></w:tblPr></w:style><w:style w:type="paragraph" w:styleId="Header"><w:name w:val="header"/><w:basedOn w:val="Normal"/></w:style></w:styles>`)
	put("word/settings.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:settings `+nsMain+`><w:zoom w:percent="100"/><w:defaultTabStop w:val="720"/><w:characterSpacingControl w:val="doNotCompress"/><w:compat><w:compatSetting w:name="compatibilityMode" w:uri="http://schemas.microsoft.com/office/word" w:val="15"/></w:compat></w:settings>`)
	put("word/header1.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:hdr `+nsMain+`><w:p><w:pPr><w:pStyle w:val="Header"/><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:sz w:val="26"/><w:szCs w:val="26"/></w:rPr><w:fldChar w:fldCharType="begin"/></w:r><w:r><w:rPr><w:sz w:val="26"/><w:szCs w:val="26"/></w:rPr><w:instrText xml:space="preserve"> PAGE </w:instrText></w:r><w:r><w:rPr><w:sz w:val="26"/><w:szCs w:val="26"/></w:rPr><w:fldChar w:fldCharType="separate"/></w:r><w:r><w:rPr><w:sz w:val="26"/><w:szCs w:val="26"/></w:rPr><w:t>2</w:t></w:r><w:r><w:rPr><w:sz w:val="26"/><w:szCs w:val="26"/></w:rPr><w:fldChar w:fldCharType="end"/></w:r></w:p></w:hdr>`)
	put("word/header2.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:hdr `+nsMain+`><w:p><w:pPr><w:pStyle w:val="Header"/></w:pPr></w:p></w:hdr>`)
	sect := fmt.Sprintf(`<w:sectPr><w:headerReference w:type="default" r:id="rId3"/><w:headerReference w:type="first" r:id="rId4"/><w:pgSz w:w="%d" w:h="%d"/><w:pgMar w:top="%d" w:right="%d" w:bottom="%d" w:left="%d" w:header="567" w:footer="567" w:gutter="0"/><w:cols w:space="720"/><w:titlePg/></w:sectPr>`,
		pageW, pageH, marTop, marRight, marBot, marLeft)
	put("word/document.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document `+nsMain+`><w:body>`+body+sect+`</w:body></w:document>`)
	title := d.TenLoai
	if d.CongVan {
		title = "Công văn"
	}
	if d.TrichYeu != "" {
		title += " " + d.TrichYeu
	}
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	put("docProps/core.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><dc:title>`+esc(strings.TrimSpace(title))+`</dc:title><dc:creator>`+esc(strings.ReplaceAll(d.CoQuanBanHanh, "\n", " "))+`</dc:creator><dcterms:created xsi:type="dcterms:W3CDTF">`+now+`</dcterms:created><dcterms:modified xsi:type="dcterms:W3CDTF">`+now+`</dcterms:modified></cp:coreProperties>`)
	put("docProps/app.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"><Application>DINH DANG VAN BAN CHUAN `+Version+`</Application></Properties>`)
	if err := zw.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
