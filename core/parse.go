package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Block là một khối nội dung phần thân văn bản.
type Block struct {
	Kind   string // para, dieu, chuong, muc, h1, h2, dash, center, table, pagebreak
	So     string
	Ten    string
	Text   string
	Bold   bool
	Italic bool
	Rows   [][]string
	Spans  [][]int
	fixed  bool   // đã sửa số thứ tự đề mục
	Label  string // nhãn chương: "Chương", "Phần"
}

// Party là một bên ký (hợp đồng, biên bản).
type Party struct {
	Label  string // "ĐẠI DIỆN BÊN A"
	ChucVu string
	HoTen  string
}

// Signature là khối chữ ký.
type Signature struct {
	ChucDanh []string // VD: "KT. GIÁM ĐỐC", "PHÓ GIÁM ĐỐC"
	GhiChu   string   // "(Đã ký)"
	HoTen    string
}

// Doc là văn bản đã tách thành các thành phần thể thức.
type Doc struct {
	Profile       string // "vbhc" | "vbqppl"
	CongVan       bool
	CoQuanChuQuan string
	CoQuanBanHanh string // các dòng ngăn bằng \n
	SoKyHieu      string
	DiaDanhNgay   string
	TenLoai       string
	TrichYeu      string
	KinhGui       []string
	ThamQuyen     string
	CanCu         []string
	BanHanh       string
	QuyetDinhLine string
	Body          []Block
	Ky            Signature
	NoiNhan       []string
	KemTheo       []Block
	Parties       []Party // ký hai bên (hợp đồng, biên bản)
}

// Profile tùy chọn khi phân tích.
const (
	ProfileAuto   = "auto"
	ProfileVBHC   = "vbhc"
	ProfileVBQPPL = "vbqppl"
)

var (
	reQuocHieu  = regexp.MustCompile(`(?i)c[ộo]ng\s+h(?:o[àa]|[òó]a)\s+x[ãa]\s+h[ộo]i\s+ch[ủu]\s+ngh[ĩi]a\s+vi[ệe]t\s+nam`)
	reTieuNgu   = regexp.MustCompile(`(?i)[đd][ộo]c\s+l[ậa]p\s*[-–—]\s*t[ựu]\s+do\s*[-–—]\s*h[ạa]nh\s+ph[úu]c`)
	reSo        = regexp.MustCompile(`(?i)^\s*s[ốo]\s*(:|\.\s|\s+\d)`)
	reDate      = regexp.MustCompile(`(?i)([\p{L} .\-]{0,40}?)[,\s]*ngày\s*(\d{1,2}|[.…]+)\s*tháng\s*(\d{1,2}|[.…]+)\s*năm\s*(\d{4}|[.…]+)`)
	reVV        = regexp.MustCompile(`(?i)^\s*v\s*/\s*v\s*[:.]?\s*`)
	reKinhGui   = regexp.MustCompile(`(?i)^\s*kính\s+gửi\s*[:.]?\s*`)
	reCanCu     = regexp.MustCompile(`(?i)^\s*(căn cứ|theo đề nghị|xét đề nghị|theo đề xuất|xét tờ trình)(\s|$)`)
	reQDLine    = regexp.MustCompile(`^\s*(QUYẾT ĐỊNH|QUYẾT NGHỊ|NGHỊ QUYẾT|CHỈ THỊ|QUY ĐỊNH)\s*:\s*$`)
	reBanHanh   = regexp.MustCompile(`(?i)ban hành\s+(nghị định|thông tư|quyết định|nghị quyết|pháp lệnh|luật)(\s|$|,|\.)`)
	reNoiNhan   = regexp.MustCompile(`(?i)^\s*nơi\s+nhận\s*[:.]?\s*`)
	reLuu       = regexp.MustCompile(`(?i)^\s*lưu\s*[:.]?\s*`)
	reSigPrefix = regexp.MustCompile(`^\s*(TM|KT|Q|TL|TUQ)\s*\.\s*\p{Lu}`)
	reSigWord   = regexp.MustCompile(`^(TỔNG GIÁM ĐỐC|PHÓ TỔNG GIÁM ĐỐC|GIÁM ĐỐC|PHÓ GIÁM ĐỐC|CHỦ TỊCH|PHÓ CHỦ TỊCH|BỘ TRƯỞNG|THỨ TRƯỞNG|THỦ TRƯỞNG|HIỆU TRƯỞNG|VIỆN TRƯỞNG|CỤC TRƯỞNG|CHÁNH VĂN PHÒNG|TRƯỞNG PHÒNG|PHÓ TRƯỞNG PHÒNG|TRƯỞNG KHOA|TRƯỞNG BAN|BÍ THƯ|NGƯỜI LẬP|NGƯỜI ĐỀ NGHỊ|NGƯỜI TRÌNH|THỦ TƯỚNG|CHÁNH ÁN|VIỆN TRƯỞNG|TRƯỞNG ĐOÀN|CHỦ NHIỆM|THƯ KÝ|KẾ TOÁN TRƯỞNG|ĐIỀU DƯỠNG TRƯỞNG)(\s|$|,)`)
	reSigOther  = regexp.MustCompile(`^(NGƯỜI\s|XÁC NHẬN|LÃNH ĐẠO|ĐẠI DIỆN|THỦ QUỸ|KẾ TOÁN|TRƯỞNG\s|PHÓ TRƯỞNG|TỔ TRƯỞNG|BAN GIÁM ĐỐC|DUYỆT|ĐƠN VỊ|THAY MẶT)`)
	reCapDuoi   = regexp.MustCompile(`^(PHÒNG|KHOA|TRUNG TÂM|TỔ|ĐỘI|BAN|VĂN PHÒNG|ĐƠN VỊ|BỘ PHẬN|CHI NHÁNH|XƯỞNG|ĐOÀN|CÔNG ĐOÀN)(\s|$)`)
	reGhiChuKy  = regexp.MustCompile(`(?i)^\s*\((đã ký|ký|chữ ký|ký tên|ký, ghi rõ|ký và ghi)[^)]*\)\s*$`)
	reHocHam    = regexp.MustCompile(`^(?:(?:GS|PGS|TS|TSKH|ThS|Ths|THS|BS|BSCKI|BSCKII|BSCK\s?I{1,2}|BSNT|CKI|CKII|DS|DSCKI|DSCKII|KS|CN|TTƯT|TTND|NGND|NGƯT)\s*\.\s*)+`)
	reChuong    = regexp.MustCompile(`(?i)^\s*chương\s+([IVXLC]+|\d+)\s*[.:\-–]?\s*(.*)$`)
	reMuc       = regexp.MustCompile(`(?i)^\s*mục\s+(\d+)\s*[.:\-–]?\s*(.*)$`)
	reDieu      = regexp.MustCompile(`^\s*Điều\s+(\d+[a-z]?)\s*[.:]\s*(.*)$`)
	reRoman     = regexp.MustCompile(`^\s*([IVX]{1,5})\s*[.\-–]\s+(.*)$`)
	reNum       = regexp.MustCompile(`^\s*(\d{1,2})\s*[.)]\s+(.*)$`)
	reNumNum    = regexp.MustCompile(`^\s*(\d{1,2}\.\d{1,2}(\.\d{1,2})?)\.?\s+(.*)$`)
	reDiem      = regexp.MustCompile(`^\s*([a-zđ])\s*\)\s*(.*)$`)
	reNumNoDot  = regexp.MustCompile(`^\s*(\d{1,2})\s+(\p{Lu}[^.;]*:?)\s*$`)
	reDash      = regexp.MustCompile(`^\s*[-–—•+*]\s*(.*)$`)
	reSeparator = regexp.MustCompile(`^[\s\-_=–—.*/]*$`)
	reEndMark   = regexp.MustCompile(`^\s*\.?/\.\s*$`)
)

var tenLoaiList = []string{
	"QUYẾT ĐỊNH", "TỜ TRÌNH", "PHIẾU TRÌNH", "BÁO CÁO", "KẾ HOẠCH", "THÔNG BÁO", "GIẤY MỜI", "BIÊN BẢN",
	"QUY CHẾ", "QUY ĐỊNH", "NGHỊ QUYẾT", "CHỈ THỊ", "HƯỚNG DẪN", "PHƯƠNG ÁN", "ĐỀ ÁN", "DỰ ÁN",
	"CHƯƠNG TRÌNH", "GIẤY ỦY QUYỀN", "GIẤY UỶ QUYỀN", "GIẤY GIỚI THIỆU", "GIẤY NGHỈ PHÉP", "GIẤY ĐI ĐƯỜNG",
	"GIẤY BIÊN NHẬN", "PHIẾU GỬI", "PHIẾU CHUYỂN", "PHIẾU BÁO", "HỢP ĐỒNG", "CÔNG ĐIỆN", "BẢN GHI NHỚ",
	"BẢN THỎA THUẬN", "BẢN THOẢ THUẬN", "PHIẾU ĐỀ XUẤT", "ĐƠN", "LUẬT", "PHÁP LỆNH", "NGHỊ ĐỊNH",
	"THÔNG TƯ LIÊN TỊCH", "THÔNG TƯ", "LỆNH", "QUY TRÌNH", "CÔNG VĂN",
}

var chuQuanPrefix = regexp.MustCompile(`^(SỞ|BỘ|UBND|ỦY BAN|UỶ BAN|TỔNG CÔNG TY|TẬP ĐOÀN|ĐẢNG|BAN|CỤC|TỈNH ỦY|TỈNH UỶ|HỘI ĐỒNG|ĐẠI HỌC|NGÂN HÀNG|BẢO HIỂM XÃ HỘI|CÔNG TY)(\s|$)`)

var foldMap = func() map[rune]rune {
	m := map[rune]rune{}
	groups := map[rune]string{
		'a': "àáảãạăằắẳẵặâầấẩẫậ", 'e': "èéẻẽẹêềếểễệ", 'i': "ìíỉĩị", 'o': "òóỏõọôồốổỗộơờớởỡợ",
		'u': "ùúủũụưừứửữự", 'y': "ỳýỷỹỵ", 'd': "đ",
	}
	for base, chars := range groups {
		for _, c := range chars {
			m[c] = base
		}
	}
	return m
}()

// Fold bỏ dấu tiếng Việt, viết hoa (so khớp chịu lỗi nhận dạng chữ).
func Fold(s string) string {
	r := []rune(strings.ToLower(s))
	for i, c := range r {
		if b, ok := foldMap[c]; ok {
			r[i] = b
		}
	}
	return strings.ToUpper(string(r))
}

func isUpperLine(s string) bool {
	letters, lower := 0, 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			letters++
			if unicode.IsLower(r) {
				lower++
			}
		}
	}
	return letters >= 2 && lower == 0
}

func matchTenLoai(s string) string {
	t := strings.TrimSpace(strings.TrimRight(strings.TrimSpace(s), ":"))
	if !isUpperLine(t) || len([]rune(t)) > 70 {
		return ""
	}
	ft := Fold(t)
	for _, k := range tenLoaiList {
		fk := Fold(k)
		if ft == fk || strings.HasPrefix(ft, fk+" ") {
			rest := strings.TrimSpace(string([]rune(t)[len([]rune(k)):]))
			if rest == "" {
				return k
			}
			return k + " " + rest
		}
	}
	return ""
}

// isTwoParty: hợp đồng, biên bản, thỏa thuận — ký hai bên, không có "Nơi nhận".
func isTwoParty(tenLoai string) bool {
	f := Fold(tenLoai)
	for _, k := range []string{"HOP DONG", "BIEN BAN", "BAN THOA THUAN", "BAN GHI NHO"} {
		if strings.HasPrefix(f, k) {
			return true
		}
	}
	return false
}

// Parse tách văn bản thành các thành phần thể thức. logs ghi những điều đã nhận diện/sửa.
func Parse(items []Item, profile string) (*Doc, []string) {
	d := &Doc{}
	var logs []string
	logf := func(f string, a ...interface{}) { logs = append(logs, fmt.Sprintf(f, a...)) }

	// Làm sạch, bỏ dòng phân cách
	var its []Item
	for _, it := range items {
		if it.Table != nil {
			its = append(its, it)
			continue
		}
		it.Text = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(it.Text, " ", " "), "\t", " "))
		if it.Text == "" || reSeparator.MatchString(it.Text) || reEndMark.MatchString(it.Text) {
			continue
		}
		its = append(its, it)
	}
	its = preClean(its, logf)
	n := len(its)
	i := 0

	// ---------- Phần đầu ----------
	var coQuan []Item
	stage := 0 // 0: đầu; 1: đã có tên loại/trích yếu; 2: đang căn cứ
	seenSo := false

	// Bước 1: tìm tên loại (BÁO CÁO, QUYẾT ĐỊNH...) trong 30 dòng đầu; mọi dòng phía trên là phần đầu văn bản.
	tIdx := -1
	for k := 0; k < n && k < 30; k++ {
		if its[k].Table != nil {
			break
		}
		if matchTenLoai(its[k].Text) != "" {
			tIdx = k
			break
		}
	}
	if tIdx > 0 {
		short := true
		for k := 0; k < tIdx; k++ {
			if len([]rune(its[k].Text)) > 150 || reCanCu.MatchString(its[k].Text) || reKinhGui.MatchString(its[k].Text) {
				short = false
			}
		}
		if short {
			for k := 0; k < tIdx; k++ {
				it := its[k]
				s := strings.TrimSpace(reTieuNgu.ReplaceAllString(reQuocHieu.ReplaceAllString(it.Text, " "), " "))
				if s == "" || reSeparator.MatchString(s) {
					continue
				}
				if loc := reDate.FindStringIndex(s); loc != nil {
					date := strings.TrimSpace(s[loc[0]:])
					if d.DiaDanhNgay == "" || (!hasDigit(d.DiaDanhNgay) && hasDigit(date)) {
						d.DiaDanhNgay = date
					}
					s = strings.TrimSpace(s[:loc[0]])
					if s == "" {
						continue
					}
				}
				if reSo.MatchString(s) {
					if hasDigit(s) {
						d.SoKyHieu, seenSo = s, true
					}
					continue
				}
				if reVV.MatchString(s) {
					d.TrichYeu = strings.TrimSpace(reVV.ReplaceAllString(s, ""))
					continue
				}
				letters := 0
				for _, r := range s {
					if unicode.IsLetter(r) {
						letters++
					}
				}
				if letters < 3 {
					logf("Bỏ dòng ký tự lạ ở phần đầu: \"%s\"", s)
					continue
				}
				it.Text = s
				coQuan = append(coQuan, it)
			}
			if len(coQuan) > 0 {
				logf("Phần đầu văn bản: %d dòng tên cơ quan/đơn vị; Quốc hiệu, Tiêu ngữ được dựng lại một lần duy nhất.", len(coQuan))
			}
			i = tIdx
		}
	}
	for i < n {
		it := its[i]
		if it.Table != nil {
			break
		}
		s := it.Text
		// Bỏ Quốc hiệu, Tiêu ngữ (sẽ dựng lại chuẩn)
		if reQuocHieu.MatchString(s) || reTieuNgu.MatchString(s) {
			s = strings.TrimSpace(reTieuNgu.ReplaceAllString(reQuocHieu.ReplaceAllString(s, " "), " "))
			if s == "" || reSeparator.MatchString(s) {
				i++
				continue
			}
			it.Text = s
		}
		if stage == 0 && !seenSo && reSo.MatchString(s) {
			rest := s
			if loc := reDate.FindStringIndex(s); loc != nil && loc[0] > 0 {
				d.DiaDanhNgay = strings.TrimSpace(s[loc[0]:])
				rest = strings.TrimSpace(s[:loc[0]])
			}
			d.SoKyHieu = rest
			seenSo = regexp.MustCompile(`\d`).MatchString(rest)
			if !seenSo {
				d.SoKyHieu = ""
			}
			i++
			continue
		}
		if stage <= 1 && (d.DiaDanhNgay == "" || !hasDigit(d.DiaDanhNgay)) && reDate.MatchString(s) && len([]rune(s)) < 90 {
			d.DiaDanhNgay = s
			i++
			continue
		}
		if stage == 0 && reVV.MatchString(s) {
			d.TrichYeu = strings.TrimSpace(reVV.ReplaceAllString(s, ""))
			d.CongVan = true
			i++
			for i < n && its[i].Table == nil && startsLower(its[i].Text) && len([]rune(its[i].Text)) < 120 {
				d.TrichYeu += " " + its[i].Text
				i++
			}
			continue
		}
		if stage <= 1 && reKinhGui.MatchString(s) {
			rest := strings.TrimSpace(reKinhGui.ReplaceAllString(s, ""))
			if rest != "" {
				d.KinhGui = append(d.KinhGui, rest)
			}
			i++
			for i < n && its[i].Table == nil {
				t := its[i].Text
				if m := reDash.FindStringSubmatch(t); m != nil && len([]rune(t)) < 150 {
					d.KinhGui = append(d.KinhGui, m[1])
					i++
					continue
				}
				if rest == "" && len(d.KinhGui) == 0 && len([]rune(t)) < 120 {
					d.KinhGui = append(d.KinhGui, t)
					i++
				}
				break
			}
			stage = 1
			continue
		}
		if stage == 0 && d.TenLoai == "" {
			if tl := matchTenLoai(s); tl != "" && (seenSo || len(coQuan) > 0 || i < 8) {
				d.TenLoai = tl
				stage = 1
				i++
				// Trích yếu
				var ty []string
				for i < n && its[i].Table == nil && len(ty) < 4 {
					t := its[i].Text
					if reSo.MatchString(t) || reKinhGui.MatchString(t) || reCanCu.MatchString(t) || reQDLine.MatchString(t) ||
						(isUpperLine(t) && len(ty) > 0) || reDate.MatchString(t) && len([]rune(t)) < 90 {
						break
					}
					if len([]rune(t)) > 220 || (len(ty) > 0 && strings.HasSuffix(t, ".") && len([]rune(t)) > 90) {
						break
					}
					if isUpperLine(t) && len(ty) == 0 {
						// thẩm quyền ngay sau tên loại (không có trích yếu) – dừng
						if strings.Contains(t, "GIÁM ĐỐC") || strings.Contains(t, "CHỦ TỊCH") || strings.Contains(t, "BỘ TRƯỞNG") {
							break
						}
					}
					ty = append(ty, t)
					i++
					if len(ty) == 1 && !startsLowerNext(its, i) {
						break
					}
				}
				d.TrichYeu = strings.Join(ty, " ")
				continue
			}
		}
		if stage == 0 && !seenSo && isUpperLine(s) && len([]rune(s)) < 80 {
			coQuan = append(coQuan, it)
			i++
			continue
		}
		if stage >= 1 && !seenSo && reSo.MatchString(s) && hasDigit(s) && len([]rune(s)) < 80 {
			d.SoKyHieu, seenSo = s, true
			i++
			continue
		}
		if stage >= 1 && d.ThamQuyen == "" && len(d.CanCu) == 0 && isUpperLine(s) && !reQDLine.MatchString(s) && !reRoman.MatchString(s) &&
			len([]rune(s)) < 150 && regexp.MustCompile(`^(QUYẾT ĐỊNH|NGHỊ QUYẾT|CHỈ THỊ|NGHỊ ĐỊNH|THÔNG TƯ|LỆNH|PHÁP LỆNH)`).MatchString(d.TenLoai) {
			d.ThamQuyen = s
			i++
			continue
		}
		if reCanCu.MatchString(s) && (stage >= 1 || seenSo) && len([]rune(s)) < 600 {
			d.CanCu = append(d.CanCu, s)
			stage = 2
			i++
			continue
		}
		if stage == 2 && reBanHanh.MatchString(s) && len([]rune(s)) < 300 {
			d.BanHanh = s
			i++
			break
		}
		if reQDLine.MatchString(s) && stage >= 1 {
			d.QuyetDinhLine = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), ":"))
			i++
			break
		}
		if stage == 0 && seenSo && isUpperLine(s) && len([]rune(s)) < 80 && d.TenLoai == "" {
			// dòng in hoa sau "Số" mà không phải tên loại đã biết
			d.TenLoai = s
			stage = 1
			i++
			continue
		}
		break
	}
	bodyStart := i

	// Cơ quan chủ quản / ban hành
	if len(coQuan) > 0 {
		var chu, ban []string
		hasBoldInfo := false
		for _, c := range coQuan {
			if c.Bold {
				hasBoldInfo = true
			}
		}
		if k := lastCapDuoi(coQuan); k > 0 {
			for _, c := range coQuan[:k] {
				chu = append(chu, c.Text)
			}
			for _, c := range coQuan[k:] {
				ban = append(ban, c.Text)
			}
		} else if hasBoldInfo && !coQuan[0].Bold && len(coQuan) > 1 {
			for _, c := range coQuan {
				if !c.Bold && len(ban) == 0 {
					chu = append(chu, c.Text)
				} else {
					ban = append(ban, c.Text)
				}
			}
		} else if k := lastCapDuoi(coQuan); k > 0 {
			for _, c := range coQuan[:k] {
				chu = append(chu, c.Text)
			}
			for _, c := range coQuan[k:] {
				ban = append(ban, c.Text)
			}
		} else if len(coQuan) > 1 && chuQuanPrefix.MatchString(coQuan[0].Text) {
			chu = []string{coQuan[0].Text}
			for _, c := range coQuan[1:] {
				ban = append(ban, c.Text)
			}
		} else {
			for _, c := range coQuan {
				ban = append(ban, c.Text)
			}
		}
		d.CoQuanChuQuan = strings.Join(chu, "\n")
		d.CoQuanBanHanh = strings.Join(ban, "\n")
	}

	// ---------- Phụ lục sau chữ ký ----------
	annex := n
	for k := bodyStart + 5; k < n; k++ {
		t := its[k].Text
		fw := strings.Fields(t)
		if its[k].Table != nil || len(fw) < 2 || len([]rune(t)) > 220 || !reAnnex.MatchString(Fold(t)) || !isUpperLine(fw[0]+" "+fw[1]) {
			continue
		}
		hasSig := false
		for j := k - 1; j >= bodyStart && j >= k-40; j-- {
			if its[j].Table == nil && (reDaiDien.MatchString(Fold(its[j].Text)) || reNoiNhan.MatchString(its[j].Text) || reSigPrefix.MatchString(its[j].Text)) {
				hasSig = true
				break
			}
		}
		if hasSig {
			annex = k
			break
		}
	}
	if isTwoParty(d.TenLoai) {
		parseTwoParty(d, its, bodyStart, annex, logf)
		if annex < n {
			d.KemTheo = annexBlocks(parseBody(its[annex:], false))
			logf("Tách %d phụ lục sau chữ ký, mỗi phụ lục sang trang mới.", countAnnex(d.KemTheo))
		}
		finishProfile(d, profile, logf)
		logs = append(logs, normalize(d)...)
		logs = append(logs, summary(d)...)
		return d, logs
	}
	n0 := n
	n = annex

	// ---------- Phần cuối: nơi nhận, chữ ký ----------
	nn := -1
	for k := bodyStart; k < n; k++ {
		if its[k].Table == nil && reNoiNhan.MatchString(its[k].Text) {
			nn = k
			break
		}
	}
	isSigHead := func(s string) bool {
		s = strings.TrimSpace(s)
		if !isUpperLine(s) || len([]rune(s)) > 90 || reRoman.MatchString(s) {
			return false
		}
		return reSigPrefix.MatchString(s) || reSigWord.MatchString(s) || reSigOther.MatchString(s)
	}
	sig := -1
	from := bodyStart
	if nn >= 0 {
		if nn-10 > from {
			from = nn - 10
		}
	} else if n-12 > from {
		from = n - 12
	}
	for k := from; k < n; k++ {
		if its[k].Table == nil && isSigHead(its[k].Text) {
			// chỉ nhận nếu phía sau là phần ngắn (chữ ký, nơi nhận)
			ok := true
			cnt := 0
			for j := k + 1; j < n && j < k+8; j++ {
				if its[j].Table != nil || len([]rune(its[j].Text)) > 160 {
					ok = false
					break
				}
				cnt++
			}
			if ok {
				sig = k
				break
			}
		}
	}
	bodyEnd := n
	if nn >= 0 {
		bodyEnd = nn
	}
	if sig >= 0 && sig < bodyEnd {
		bodyEnd = sig
	}
	if bodyEnd < bodyStart {
		bodyEnd = bodyStart
	}

	// Duyệt phần cuối
	k := bodyEnd
	mode := ""
	sigDone, nnDone := sig < 0, nn < 0
	for k < n && !(sigDone && nnDone) {
		it := its[k]
		if it.Table != nil {
			break
		}
		s := it.Text
		switch {
		case reNoiNhan.MatchString(s):
			mode = "nn"
			rest := strings.TrimSpace(reNoiNhan.ReplaceAllString(s, ""))
			if rest != "" {
				d.NoiNhan = append(d.NoiNhan, strings.TrimSpace(reDash.ReplaceAllString(rest, "$1")))
			}
		case !sigDone && isSigHead(s):
			mode = "sig"
			nnDone = nnDone || len(d.NoiNhan) > 0
			d.Ky.ChucDanh = append(d.Ky.ChucDanh, s)
		case mode == "sig" && isUpperLine(s) && len([]rune(s)) < 90 && d.Ky.HoTen == "":
			d.Ky.ChucDanh = append(d.Ky.ChucDanh, s)
		case reGhiChuKy.MatchString(s):
			d.Ky.GhiChu = s
		case mode == "nn":
			if m := reDash.FindStringSubmatch(s); m != nil {
				d.NoiNhan = append(d.NoiNhan, m[1])
			} else if reLuu.MatchString(s) {
				d.NoiNhan = append(d.NoiNhan, s)
				nnDone = true
				if sigDone {
					k++
					goto done
				}
				mode = ""
			} else if len(d.NoiNhan) > 0 && startsLower(s) {
				d.NoiNhan[len(d.NoiNhan)-1] += " " + s
			} else if mode == "nn" && looksLikeName(s) && len(d.Ky.ChucDanh) > 0 && d.Ky.HoTen == "" {
				d.Ky.HoTen = s
				sigDone = true
			} else {
				nnDone = true
				mode = ""
				continue
			}
			if reLuu.MatchString(strings.TrimSpace(reDash.ReplaceAllString(s, "$1"))) {
				nnDone = true
				mode = ""
			}
		case mode == "sig" || (len(d.Ky.ChucDanh) > 0 && d.Ky.HoTen == ""):
			if looksLikeName(s) {
				d.Ky.HoTen = s
				sigDone = true
				mode = ""
			} else {
				goto done
			}
		default:
			goto done
		}
		k++
	}
done:
	tailEnd := k

	// ---------- Thân ----------
	hasDieu := false
	for _, it := range its[bodyStart:bodyEnd] {
		if _, _, ok := matchDieu(it.Text); it.Table == nil && ok {
			hasDieu = true
			break
		}
	}
	d.Body = parseBody(its[bodyStart:bodyEnd], hasDieu)
	if annex < n0 {
		d.KemTheo = annexBlocks(parseBody(its[annex:n0], false))
	}
	if tailEnd < n {
		d.KemTheo = append(parseBody(its[tailEnd:n], false), d.KemTheo...)
		for idx := range d.KemTheo {
			b := &d.KemTheo[idx]
			if idx < 4 && b.Kind == "para" && isUpperLine(b.Text) && len([]rune(b.Text)) < 120 {
				b.Kind, b.Bold = "center", true
			} else if idx < 5 && b.Kind == "para" && strings.HasPrefix(b.Text, "(") {
				b.Kind, b.Italic = "center", true
			}
		}
	}

	finishProfile(d, profile, logf)
	logs = append(logs, normalize(d)...)
	logs = append(logs, summary(d)...)
	return d, logs
}

// lastCapDuoi: vị trí dòng đơn vị cấp dưới (PHÒNG, KHOA...) đứng sau tên cơ quan cấp trên.
func lastCapDuoi(c []Item) int {
	for i := len(c) - 1; i > 0; i-- {
		if reCapDuoi.MatchString(strings.ToUpper(strings.TrimSpace(c[i].Text))) {
			return i
		}
	}
	return -1
}

var (
	reOcrBullet = regexp.MustCompile(`^(?:[sS]|[vV]'|[vV]’|[✓✔■□●▪➢►o>])\s+(\p{Lu})`)
	reAnnex     = regexp.MustCompile(`^PHU LUC(\s|:|$|\d)`)
	reDaiDien   = regexp.MustCompile(`^DAI DIEN BEN ([AB])\b`)
	reDieuFold  = regexp.MustCompile(`^\s*DIEU\s+(\d+[A-Z]?)\s*[.:]\s*`)
	rePhanFold  = regexp.MustCompile(`^\s*PHAN\s+(\d+|[IVX]+)\s*[-–.:]\s*`)
	reBenFold   = regexp.MustCompile(`\(BEN ([AB])\)`)
	reDaiDienL  = regexp.MustCompile(`(?i)^\s*đại\s+diện(?:\s+bởi)?\s*:\s*(?:ông|bà)?\s*(.+)$`)
	reChucVuL   = regexp.MustCompile(`(?i)^\s*chức\s+vụ\s*:\s*(.+)$`)
)

// matchDieu nhận "Điều 1.", "ĐIỀU 1:", cả khi sai dấu do nhận dạng chữ ("ĐIÊU 1:").
func matchDieu(s string) (so, rest string, ok bool) {
	f := Fold(s)
	m := reDieuFold.FindStringSubmatchIndex(f)
	if m == nil {
		return "", "", false
	}
	r := []rune(s)
	cut := len([]rune(f[:m[1]]))
	if cut > len(r) {
		return "", "", false
	}
	return f[m[2]:m[3]], strings.TrimSpace(string(r[cut:])), true
}

func mostlyUpper(s string) bool {
	up, low := 0, 0
	for _, r := range s {
		if unicode.IsUpper(r) {
			up++
		} else if unicode.IsLower(r) {
			low++
		}
	}
	return up >= 3 && up*10 >= (up+low)*7
}

func titleName(s string) string {
	w := strings.Fields(strings.ToLower(strings.TrimSpace(s)))
	for i := range w {
		w[i] = capFirst(w[i])
	}
	return strings.Join(w, " ")
}

// preClean: sửa dấu đầu dòng bị nhận dạng sai, bỏ khối dòng bị nhân đôi khi chuyển từ PDF.
func preClean(its []Item, logf func(string, ...interface{})) []Item {
	bul := 0
	for k := range its {
		if its[k].Table == nil {
			if m := reOcrBullet.FindStringSubmatchIndex(its[k].Text); m != nil {
				its[k].Text = "- " + its[k].Text[m[2]:]
				bul++
			}
		}
	}
	if bul > 0 {
		logf("Sửa %d dấu đầu dòng bị nhận dạng sai (s, v', o...) thành \"-\".", bul)
	}
	same := func(a, b Item) bool {
		return a.Table == nil && b.Table == nil && strings.TrimSpace(a.Text) == strings.TrimSpace(b.Text)
	}
	var out []Item
	dup := 0
	for k := 0; k < len(its); {
		removed := false
		for m := 8; m >= 1; m-- {
			if k+m <= len(its) && len(out) >= m {
				ok := true
				chars := 0
				for j := 0; j < m; j++ {
					if !same(out[len(out)-m+j], its[k+j]) {
						ok = false
						break
					}
					chars += len([]rune(its[k+j].Text))
				}
				if ok && (m >= 2 || chars >= 12) {
					k += m
					dup += m
					removed = true
					break
				}
			}
		}
		if !removed {
			out = append(out, its[k])
			k++
		}
	}
	if dup > 0 {
		logf("Bỏ %d dòng bị nhân đôi liền nhau (lỗi khi chuyển từ PDF/scan).", dup)
	}
	weird := 0
	reWeird := regexp.MustCompile(`[\^~|]|[a-zà-ỹ][A-ZĐ][a-zà-ỹ]`)
	for _, it := range out {
		if it.Table == nil && reWeird.MatchString(it.Text) {
			weird++
		}
	}
	if len(out) > 30 && weird*100/len(out) >= 2 {
		logf("⚠ Văn bản có dấu hiệu được nhận dạng từ ảnh scan (%d dòng có ký tự lạ). Chương trình chỉ sửa thể thức, không sửa lỗi chính tả do nhận dạng — nên dùng bản Word gốc nếu có.", weird)
	}
	return out
}

// annexBlocks: mỗi "PHỤ LỤC ..." sang trang mới, căn giữa đậm; dòng "(Kèm theo ...)" nghiêng.
func annexBlocks(bs []Block) []Block {
	for i := range bs {
		b := &bs[i]
		if b.Kind != "table" && len([]rune(b.Text)) < 220 && reAnnex.MatchString(Fold(b.Text)) {
			b.Kind, b.Bold = "annex", true
			b.Text = strings.ToUpper(b.Text)
			if i+1 < len(bs) && strings.HasPrefix(strings.TrimSpace(bs[i+1].Text), "(") && bs[i+1].Kind != "table" {
				bs[i+1].Kind, bs[i+1].Italic, bs[i+1].Bold = "center", true, false
			}
		}
	}
	return bs
}

// parseTwoParty: hợp đồng/biên bản — thân văn bản, khối ký hai bên.
func parseTwoParty(d *Doc, its []Item, bodyStart, end int, logf func(string, ...interface{})) {
	sigStart := end
	for k := bodyStart; k < end; k++ {
		if its[k].Table == nil && len([]rune(its[k].Text)) < 60 && reDaiDien.MatchString(Fold(its[k].Text)) {
			sigStart = k
			break
		}
	}
	hasDieu := false
	for _, it := range its[bodyStart:sigStart] {
		if it.Table == nil {
			if _, _, ok := matchDieu(it.Text); ok {
				hasDieu = true
				break
			}
		}
	}
	d.Body = parseBody(its[bodyStart:sigStart], hasDieu)
	// Thông tin đại diện các bên trong phần đầu hợp đồng
	info := map[string]*Party{"A": {Label: "ĐẠI DIỆN BÊN A"}, "B": {Label: "ĐẠI DIỆN BÊN B"}}
	cur := ""
	for _, it := range its[bodyStart:sigStart] {
		if it.Table != nil {
			continue
		}
		f := Fold(it.Text)
		if m := reBenFold.FindStringSubmatch(f); m != nil {
			cur = m[1]
			continue
		}
		if cur == "" {
			continue
		}
		if m := reDaiDienL.FindStringSubmatch(it.Text); m != nil && info[cur].HoTen == "" {
			info[cur].HoTen = titleName(m[1])
		} else if m := reChucVuL.FindStringSubmatch(it.Text); m != nil && info[cur].ChucVu == "" {
			info[cur].ChucVu = strings.ToUpper(trimEndPunct(m[1]))
		}
	}
	// Dự phòng: đọc trong vùng chữ ký
	var region []string
	for k := sigStart; k < end; k++ {
		if its[k].Table == nil {
			region = append(region, its[k].Text)
		}
	}
	cur = ""
	for _, t := range region {
		f := Fold(t)
		if m := reDaiDien.FindStringSubmatch(f); m != nil {
			cur = m[1]
			continue
		}
		if cur == "" {
			continue
		}
		if looksLikeName(titleName(t)) && info[cur].HoTen == "" && len(strings.Fields(t)) >= 2 {
			info[cur].HoTen = titleName(t)
		} else if isUpperLine(t) && reSigWord.MatchString(t) && info[cur].ChucVu == "" {
			info[cur].ChucVu = t
		}
	}
	if sigStart < end {
		d.Parties = []Party{*info["A"], *info["B"]}
		if len(region) > 2 {
			show := region
			if len(show) > 8 {
				show = show[:8]
			}
			logf("Dựng lại khối ký hai bên (Bên A trái, Bên B phải) từ thông tin \"Đại diện\", \"Chức vụ\" của mỗi bên; bỏ %d dòng trong vùng chữ ký/con dấu: %s", len(region), strings.Join(show, " | "))
		}
	}
}

func countAnnex(bs []Block) int {
	c := 0
	for _, b := range bs {
		if b.Kind == "annex" {
			c++
		}
	}
	return c
}

func hasDigit(s string) bool { return strings.IndexAny(s, "0123456789") >= 0 }

var reVietSyl = regexp.MustCompile(`^(ngh|ng|nh|ch|gh|gi|kh|ph|qu|th|tr|[bcdđghklmnpqrstvx])?[aàáảãạăằắẳẵặâầấẩẫậeèéẻẽẹêềếểễệiìíỉĩịoòóỏõọôồốổỗộơờớởỡợuùúủũụưừứửữựyỳýỷỹỵ]+(ng|nh|ch|[cmnpt])?$`)

// SentenceCase chuyển dòng viết HOA toàn bộ sang chữ thường (giữ chữ viết tắt như BV, HNĐK, GĐ2, UBND).
func SentenceCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		core := strings.Trim(w, ",.;:()\"'“”–-")
		low := strings.ToLower(core)
		if core == "" || hasDigit(core) || !reVietSyl.MatchString(low) {
			continue // chữ viết tắt hoặc ký hiệu: giữ nguyên
		}
		words[i] = strings.Replace(w, core, low, 1)
	}
	return capFirst(strings.Join(words, " "))
}

func finishProfile(d *Doc, profile string, logf func(string, ...interface{})) {
	d.Profile = profile
	if profile == ProfileAuto || profile == "" {
		d.Profile = ProfileVBHC
		if regexp.MustCompile(`\d+/\d{4}/`).MatchString(d.SoKyHieu) || d.BanHanh != "" {
			d.Profile = ProfileVBQPPL
		}
		switch d.TenLoai {
		case "LUẬT", "PHÁP LỆNH", "NGHỊ ĐỊNH", "THÔNG TƯ", "THÔNG TƯ LIÊN TỊCH":
			d.Profile = ProfileVBQPPL
		}
		if d.Profile == ProfileVBQPPL {
			logf("Tự nhận diện: VBQPPL → áp dụng NĐ 78/2025 (sửa bởi NĐ 187/2025).")
		} else {
			logf("Tự nhận diện: văn bản hành chính → áp dụng NĐ 30/2020/NĐ-CP.")
		}
	}
	if d.TenLoai == "" && (d.CongVan || len(d.KinhGui) > 0) {
		d.CongVan = true
	}
	if d.TenLoai == "CÔNG VĂN" {
		d.TenLoai = ""
		d.CongVan = true
	}

}

func startsLower(s string) bool {
	for _, r := range strings.TrimSpace(s) {
		return unicode.IsLower(r)
	}
	return false
}

func startsLowerNext(its []Item, i int) bool {
	return i < len(its) && its[i].Table == nil && startsLower(its[i].Text) && len([]rune(its[i].Text)) < 150
}

func looksLikeName(s string) bool {
	s = strings.TrimSpace(reHocHam.ReplaceAllString(strings.TrimSpace(s), ""))
	w := strings.Fields(s)
	if len(w) < 2 || len(w) > 6 || strings.ContainsAny(s, ";:,()") {
		return false
	}
	for _, x := range w {
		r := []rune(x)
		if !unicode.IsUpper(r[0]) {
			return false
		}
	}
	return true
}

func parseBody(its []Item, hasDieu bool) []Block {
	var out []Block
	for i := 0; i < len(its); i++ {
		it := its[i]
		if it.Table != nil {
			out = append(out, Block{Kind: "table", Rows: it.Table, Spans: it.Spans})
			continue
		}
		s := it.Text
		if (reQuocHieu.MatchString(s) || reTieuNgu.MatchString(s)) && len([]rune(s)) < 80 {
			continue // Quốc hiệu/Tiêu ngữ chỉ có ở phần đầu (đã dựng lại), không lặp trong thân
		}
		if m := reChuong.FindStringSubmatch(s); m != nil && len([]rune(s)) < 150 {
			b := Block{Kind: "chuong", So: strings.ToUpper(m[1]), Ten: m[2]}
			if b.Ten == "" && i+1 < len(its) && its[i+1].Table == nil && len([]rune(its[i+1].Text)) < 150 && !reDieu.MatchString(its[i+1].Text) {
				b.Ten = its[i+1].Text
				i++
			}
			out = append(out, b)
			continue
		}
		if m := reMuc.FindStringSubmatch(s); m != nil && hasDieu && len([]rune(s)) < 150 {
			b := Block{Kind: "muc", So: m[1], Ten: m[2]}
			if b.Ten == "" && i+1 < len(its) && its[i+1].Table == nil && len([]rune(its[i+1].Text)) < 150 && !reDieu.MatchString(its[i+1].Text) {
				b.Ten = its[i+1].Text
				i++
			}
			out = append(out, b)
			continue
		}
		if m := rePhanFold.FindStringSubmatchIndex(Fold(s)); m != nil && len([]rune(s)) < 160 {
			f := Fold(s)
			ten := strings.TrimSpace(string([]rune(s)[len([]rune(f[:m[1]])):]))
			out = append(out, Block{Kind: "chuong", Label: "Phần", So: f[m[2]:m[3]], Ten: ten})
			continue
		}
		if so, rest0, ok := matchDieu(s); ok {
			b := Block{Kind: "dieu", So: strings.ToLower(so)}
			rest := rest0
			if mostlyUpper(rest) {
				rest = SentenceCase(rest)
			}
			if len([]rune(rest)) < 110 && !strings.HasSuffix(rest, ".") && !strings.HasSuffix(rest, ":") && !strings.HasSuffix(rest, ";") {
				b.Ten = rest
			} else {
				b.Text = rest
			}
			out = append(out, b)
			continue
		}
		if m := reRoman.FindStringSubmatch(s); m != nil && len([]rune(s)) < 160 {
			out = append(out, Block{Kind: "h1", So: m[1], Text: m[2], Bold: true})
			continue
		}
		short := len([]rune(s)) < 110 && !strings.HasSuffix(s, ".") && !strings.HasSuffix(s, ";") && !strings.HasSuffix(s, ",")
		if m := reNumNum.FindStringSubmatch(s); m != nil {
			out = append(out, Block{Kind: "para", Text: s, Bold: short && !hasDieu, Italic: short && !hasDieu && strings.Count(m[1], ".") >= 1 && m[2] != ""})
			continue
		}
		if m := reNumNoDot.FindStringSubmatch(s); m != nil && len([]rune(s)) < 80 {
			s = m[1] + ". " + m[2]
			out = append(out, Block{Kind: "para", Text: s, Bold: !hasDieu, fixed: true})
			continue
		}
		if reNum.MatchString(s) {
			out = append(out, Block{Kind: "para", Text: s, Bold: (short && !hasDieu) || it.Bold})
			continue
		}
		if reDiem.MatchString(s) {
			out = append(out, Block{Kind: "para", Text: s, Bold: it.Bold})
			continue
		}
		if m := reDash.FindStringSubmatch(s); m != nil && !strings.HasPrefix(strings.TrimSpace(s), "--") {
			out = append(out, Block{Kind: "dash", Text: m[1], Bold: it.Bold})
			continue
		}
		if isUpperLine(s) && len([]rune(s)) < 140 {
			if !strings.Contains(s, ":") && len([]rune(s)) < 100 {
				out = append(out, Block{Kind: "center", Text: s, Bold: true})
			} else {
				out = append(out, Block{Kind: "para", Text: s, Bold: true})
			}
			continue
		}
		out = append(out, Block{Kind: "para", Text: s, Bold: it.Bold, Italic: it.Italic})
	}
	return out
}

// ---------- Chuẩn hóa ----------

var (
	reSpaces      = regexp.MustCompile(`[ \t\x{00a0}]+`)
	reSpaceBefore = regexp.MustCompile(`\s+([,.;:!?)\]])`)
	reSpaceAfterO = regexp.MustCompile(`([(\[])\s+`)
	reNoSpaceAft  = regexp.MustCompile(`([,;:])(\p{L})`)
)

// CleanText chuẩn hóa khoảng trắng và dấu câu (không đổi nội dung).
func CleanText(s string) string {
	s = reSpaces.ReplaceAllString(strings.TrimSpace(s), " ")
	s = reSpaceBefore.ReplaceAllString(s, "$1")
	s = reSpaceAfterO.ReplaceAllString(s, "$1")
	s = reNoSpaceAft.ReplaceAllStringFunc(s, func(m string) string {
		r := []rune(m)
		return string(r[0]) + " " + string(r[1:])
	})
	// trả lại các viết tắt dạng V/v, Q.Đ không bị ảnh hưởng (chỉ tác động , ; :)
	return s
}

func capFirst(s string) string {
	r := []rune(s)
	if len(r) > 0 {
		r[0] = unicode.ToUpper(r[0])
	}
	return string(r)
}

func trimEndPunct(s string) string {
	return strings.TrimSpace(strings.TrimRight(strings.TrimSpace(s), ";.,:"))
}

// NormSo chuẩn hóa "Số: 07/QĐ-BV".
func NormSo(s string) string {
	m := regexp.MustCompile(`(?i)^\s*s[ốo]\s*[:.]?\s*`).ReplaceAllString(s, "")
	m = strings.ReplaceAll(strings.TrimSpace(m), " ", "")
	if m == "" {
		return "Số:      /"
	}
	parts := strings.SplitN(m, "/", 2)
	num := parts[0]
	if v, err := strconv.Atoi(num); err == nil && v < 10 && len(num) == 1 {
		num = "0" + num
	}
	if strings.Trim(num, ".…") == "" {
		num = "     "
	}
	if len(parts) == 2 {
		return "Số: " + num + "/" + strings.ToUpper(parts[1])
	}
	return "Số: " + num
}

// NormDate chuẩn hóa "Địa danh, ngày 05 tháng 02 năm 2026".
func NormDate(s string) string {
	m := reDate.FindStringSubmatch(s)
	if m == nil {
		return strings.TrimSpace(s)
	}
	place := strings.TrimSpace(strings.Trim(m[1], " ,"))
	if place == "" {
		place = "..."
	}
	day, mon := m[2], m[3]
	if v, err := strconv.Atoi(day); err == nil {
		day = fmt.Sprintf("%02d", v)
	} else {
		day = "..."
	}
	if v, err := strconv.Atoi(mon); err == nil {
		if v < 3 {
			mon = fmt.Sprintf("%02d", v)
		} else {
			mon = strconv.Itoa(v)
		}
	} else {
		mon = "..."
	}
	year := m[4]
	if strings.Trim(year, ".…") == "" {
		year = "..."
	}
	return fmt.Sprintf("%s, ngày %s tháng %s năm %s", capFirst(place), day, mon, year)
}

func normalize(d *Doc) []string {
	var logs []string
	logf := func(f string, a ...interface{}) { logs = append(logs, fmt.Sprintf(f, a...)) }

	var cq []string
	for _, l := range strings.Split(d.CoQuanChuQuan, "\n") {
		if l = strings.ToUpper(CleanText(l)); l != "" {
			cq = append(cq, l)
		}
	}
	d.CoQuanChuQuan = strings.Join(cq, "\n")
	var lines []string
	for _, l := range strings.Split(d.CoQuanBanHanh, "\n") {
		if l = strings.ToUpper(CleanText(l)); l != "" {
			lines = append(lines, l)
		}
	}
	d.CoQuanBanHanh = strings.Join(lines, "\n")

	if d.SoKyHieu != "" {
		ns := NormSo(d.SoKyHieu)
		if ns != strings.TrimSpace(d.SoKyHieu) {
			logf("Sửa số, ký hiệu: \"%s\" → \"%s\"", strings.TrimSpace(d.SoKyHieu), ns)
		}
		d.SoKyHieu = ns
	}
	if d.DiaDanhNgay != "" {
		nd := NormDate(d.DiaDanhNgay)
		if nd != strings.TrimSpace(d.DiaDanhNgay) {
			logf("Sửa địa danh, ngày tháng: \"%s\" → \"%s\"", strings.TrimSpace(d.DiaDanhNgay), nd)
		}
		d.DiaDanhNgay = nd
	}
	d.TenLoai = strings.ToUpper(CleanText(d.TenLoai))
	if d.TrichYeu != "" {
		t := trimEndPunct(CleanText(d.TrichYeu))
		if d.CongVan {
			t = "V/v " + strings.TrimSpace(reVV.ReplaceAllString(t, ""))
		} else {
			if isUpperLine(t) {
				t2 := SentenceCase(t)
				logf("Trích yếu viết HOA toàn bộ → chữ thường đứng đậm (NĐ 30): \"%s\" → \"%s\" (kiểm tra lại tên riêng nếu có).", t, t2)
				t = t2
			}
			t = capFirst(t)
		}
		d.TrichYeu = t
	}
	for i, k := range d.KinhGui {
		d.KinhGui[i] = capFirst(trimEndPunct(CleanText(k)))
	}
	d.ThamQuyen = strings.ToUpper(CleanText(d.ThamQuyen))
	for i, c := range d.CanCu {
		c = capFirst(trimEndPunct(CleanText(c)))
		if i == len(d.CanCu)-1 && d.Profile == ProfileVBHC && d.BanHanh == "" {
			c += "."
		} else {
			c += ";"
		}
		d.CanCu[i] = c
	}
	if len(d.CanCu) > 0 {
		logf("Chuẩn hóa %d dòng căn cứ (in nghiêng, kết thúc bằng dấu \";\"%s).", len(d.CanCu), map[bool]string{true: ", dòng cuối dấu \".\"", false: ""}[d.Profile == ProfileVBHC && d.BanHanh == ""])
	}
	if d.BanHanh != "" {
		d.BanHanh = capFirst(trimEndPunct(CleanText(d.BanHanh))) + "."
	}
	cleanBlocks := func(bs []Block) {
		for i := range bs {
			b := &bs[i]
			b.Text = CleanText(b.Text)
			b.Ten = CleanText(b.Ten)
			for r := range b.Rows {
				for c := range b.Rows[r] {
					b.Rows[r][c] = CleanText(strings.ReplaceAll(b.Rows[r][c], "\n", " \n "))
					b.Rows[r][c] = strings.ReplaceAll(b.Rows[r][c], " \n ", "\n")
				}
			}
			if b.Kind == "h1" {
				b.Text = strings.ToUpper(b.Text)
			}
		}
	}
	for _, b := range d.Body {
		if b.fixed {
			logf("Sửa đề mục thiếu dấu chấm sau số thứ tự: \"%s\".", b.Text)
		}
	}
	cleanBlocks(d.Body)
	cleanBlocks(d.KemTheo)

	// Nơi nhận
	if len(d.NoiNhan) > 0 || len(d.Ky.ChucDanh) > 0 {
		var nn []string
		hasLuu := false
		for _, x := range d.NoiNhan {
			x = trimEndPunct(CleanText(strings.TrimLeft(x, "-–+• ")))
			if x == "" {
				continue
			}
			if reLuu.MatchString(x) {
				hasLuu = true
				x = "Lưu: " + strings.TrimSpace(reLuu.ReplaceAllString(x, ""))
				if strings.TrimSpace(strings.TrimPrefix(x, "Lưu:")) == "" {
					x = "Lưu: VT"
				}
			}
			nn = append(nn, capFirst(x))
		}
		if !hasLuu {
			nn = append(nn, "Lưu: VT")
			logf("Bổ sung dòng \"- Lưu: VT\" ở Nơi nhận (cần ghi thêm đơn vị soạn thảo nếu có).")
		}
		// đưa dòng Lưu xuống cuối
		var norm, luu []string
		for _, x := range nn {
			if strings.HasPrefix(x, "Lưu:") {
				luu = append(luu, x)
			} else {
				norm = append(norm, x)
			}
		}
		d.NoiNhan = append(norm, luu[:1]...)
	}
	// Chữ ký
	for i, c := range d.Ky.ChucDanh {
		c = strings.ToUpper(CleanText(c))
		c = regexp.MustCompile(`^(TM|KT|Q|TL|TUQ)\s*\.\s*`).ReplaceAllString(c, "$1. ")
		d.Ky.ChucDanh[i] = c
	}
	if d.Ky.HoTen != "" {
		h := CleanText(d.Ky.HoTen)
		if d.Profile == ProfileVBHC {
			if h2 := strings.TrimSpace(reHocHam.ReplaceAllString(h, "")); h2 != h {
				logf("Bỏ học hàm, học vị trước họ tên người ký: \"%s\" → \"%s\" (NĐ 30/2020).", h, h2)
				h = h2
			}
		}
		d.Ky.HoTen = h
	}
	if d.Ky.GhiChu != "" {
		d.Ky.GhiChu = CleanText(d.Ky.GhiChu)
	}
	return logs
}

func summary(d *Doc) []string {
	var l []string
	add := func(name, v string) {
		if strings.TrimSpace(v) == "" {
			l = append(l, "  • "+name+": (không thấy — để trống \"...\", cần điền)")
		} else {
			l = append(l, "  • "+name+": "+strings.ReplaceAll(v, "\n", " / "))
		}
	}
	l = append(l, "Các thành phần đã nhận diện:")
	if d.CoQuanChuQuan == "" {
		l = append(l, "  • Cơ quan chủ quản: (không có — bỏ qua nếu cơ quan ban hành không có cấp trên trực tiếp)")
	} else {
		add("Cơ quan chủ quản", d.CoQuanChuQuan)
	}
	add("Cơ quan ban hành", d.CoQuanBanHanh)
	add("Số, ký hiệu", d.SoKyHieu)
	add("Địa danh, ngày", d.DiaDanhNgay)
	if d.CongVan {
		l = append(l, "  • Loại: Công văn")
	} else {
		add("Tên loại", d.TenLoai)
	}
	if len(d.Parties) > 0 && d.TrichYeu == "" {
		l = append(l, "  • Trích yếu: (nằm trong tên văn bản)")
	} else {
		add("Trích yếu", d.TrichYeu)
	}
	if len(d.KinhGui) > 0 {
		l = append(l, "  • Kính gửi: "+strings.Join(d.KinhGui, "; "))
	}
	if d.ThamQuyen != "" {
		l = append(l, "  • Thẩm quyền: "+d.ThamQuyen)
	}
	if len(d.CanCu) > 0 {
		l = append(l, fmt.Sprintf("  • Căn cứ: %d dòng", len(d.CanCu)))
	}
	dieu := 0
	for _, b := range d.Body {
		if b.Kind == "dieu" {
			dieu++
		}
	}
	l = append(l, fmt.Sprintf("  • Nội dung: %d khối%s", len(d.Body), map[bool]string{true: fmt.Sprintf(" (%d điều)", dieu), false: ""}[dieu > 0]))
	if len(d.Parties) > 0 {
		for _, p := range d.Parties {
			add(p.Label, strings.TrimSpace(p.ChucVu+" — "+p.HoTen))
		}
	} else {
		add("Chức vụ người ký", strings.Join(d.Ky.ChucDanh, " / "))
		add("Họ tên người ký", d.Ky.HoTen)
	}
	if len(d.Parties) == 0 {
		l = append(l, fmt.Sprintf("  • Nơi nhận: %d mục", len(d.NoiNhan)))
	}
	if len(d.KemTheo) > 0 {
		l = append(l, fmt.Sprintf("  • Phần kèm theo/phụ lục: %d khối (trình bày sang trang mới)", len(d.KemTheo)))
	}
	return l
}
