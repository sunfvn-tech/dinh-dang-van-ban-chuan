# DocStd — ĐỊNH DẠNG VĂN BẢN CHUẨN

Ứng dụng Windows định dạng lại / soạn văn bản **đúng thể thức pháp luật Việt Nam**:
- **Văn bản hành chính** theo Nghị định 30/2020/NĐ-CP;
- **Văn bản quy phạm pháp luật (VBQPPL)** theo Nghị định 78/2025/NĐ-CP, sửa đổi bởi Nghị định 187/2025/NĐ-CP.

Đầu vào: `.docx`, `.doc`, `.txt`, `.pdf` (thêm `.rtf`, `.odt`). Đầu ra: `<tên gốc>_chuanhoa.docx`, tùy chọn thêm `.pdf`. File gốc giữ nguyên.

> Phát triển bởi Phòng Kỹ thuật - VTTB - CNTT, Bệnh viện Hữu nghị đa khoa Nghệ An - Giai đoạn 2.

## Tải về và cài đặt
Vào mục **[Releases](../../releases)**, tải `DinhDangVanBanChuan_Setup_vX.Y.Z.exe` và chạy.
- Cài cho người dùng hiện tại, **không cần quyền Admin**; tạo biểu tượng **DocStd** ở Desktop, Start Menu và mục *Send to*.
- Cài im lặng hàng loạt: `setup.exe /S` (đổi thư mục: `/D=C:\DuongDan`).
- Gỡ: Settings › Apps (Control Panel › Programs) › DocStd › Uninstall.
- Không muốn cài: dùng `DinhDangVanBanChuan.exe` (chạy luôn) hoặc `ddvbc-cli.exe` (dòng lệnh).

Yêu cầu: Windows 10/11 64-bit. File `.doc`, `.pdf` và xuất PDF cần Microsoft Word 2010+ hoặc LibreOffice; `.docx`, `.txt` xử lý trực tiếp.

## Cách dùng
1. Chọn file (hoặc kéo thả vào cửa sổ / chuột phải › Send to › DocStd). Chọn được nhiều file.
2. Chọn loại văn bản: **Tự nhận diện** / **Văn bản hành chính (NĐ 30/2020)** / **VBQPPL (NĐ 78/2025 sửa bởi NĐ 187/2025)**.
3. Chọn kiểu xuất: chỉ Word, hoặc Word + PDF.
4. Bấm **ĐỊNH DẠNG VĂN BẢN**, đọc *Nhật ký xử lý* (thành phần nhận diện được, điểm đã sửa, chỗ còn `...` cần điền).

Dòng lệnh: `ddvbc-cli.exe [-p auto|vbhc|vbqppl] [-pdf] [-o thư_mục] file...`

### Vì sao có 3 lựa chọn loại văn bản
| Điểm | Văn bản hành chính (NĐ 30) | VBQPPL (NĐ 78 sửa bởi NĐ 187) |
|---|---|---|
| Số, ký hiệu | `07/QĐ-BV` (không có năm) | `03/2026/QĐ-UBND` (có năm) |
| Đường kẻ dưới trích yếu | Có | Không |
| Dòng căn cứ cuối | Dấu `.` | Dấu `;` rồi dòng "… ban hành …" dấu `.` |
| Học hàm, học vị trước tên người ký | Bỏ | Giữ |

*Tự nhận diện* coi là VBQPPL nếu số ký hiệu có năm (`…/2026/…`), có dòng "… ban hành Quyết định/Nghị định/Thông tư/Nghị quyết…" sau căn cứ, hoặc tên loại là Luật, Pháp lệnh, Nghị định, Thông tư. Chọn tay khi chương trình nhận nhầm hoặc khi soạn dự thảo VBQPPL chưa có số.

## Chương trình tự làm
- A4; lề trên/dưới 20 mm, trái 30 mm, phải 15 mm; Times New Roman; số trang giữa lề trên, ẩn ở trang 1.
- Phần đầu dựng lại một lần duy nhất: cơ quan chủ quản / cơ quan ban hành (đậm, kẻ dưới), Quốc hiệu – Tiêu ngữ (kẻ dưới), ô Số và Địa danh – ngày cùng hàng. Nhận diện Quốc hiệu ở mọi kiểu chữ, kể cả khi nằm trong bảng có viền.
- Chuẩn hóa: số < 10 thêm 0; ngày < 10, tháng 1–2 thêm 0; ký hiệu viết hoa liền; trích yếu viết HOA → chữ thường đậm (giữ chữ viết tắt); căn cứ in nghiêng dấu `;`; đề mục thiếu dấu chấm.
- Chương / Mục / Điều / Phần; mục I., II.; khối Nơi nhận + chữ ký hai cột (bổ sung `- Lưu: VT`); bỏ học hàm, học vị người ký (VBHC).
- Hợp đồng, biên bản: khối "ĐẠI DIỆN BÊN A" (trái) / "ĐẠI DIỆN BÊN B" (phải), tên hai bên thẳng hàng; phụ lục sau chữ ký sang trang mới.
- Bảng số liệu: độ rộng cột theo nội dung thật (dữ liệu không xuống dòng, cột số không bị bẻ đôi, tiêu đề lặp khi sang trang, dòng "Tổng" đậm, hỗ trợ ô gộp).
- Văn bản chuyển từ bản scan: sửa dấu đầu dòng nhận sai, bỏ dòng bị nhân đôi, nhận diện chịu lỗi dấu ("HỢP ĐÒNG", "ĐIÊU 1:"), cảnh báo OCR.

**Giới hạn:** nhận diện theo quy tắc — luôn soát lại trước khi trình ký; không sửa lỗi chính tả do nhận dạng chữ; văn bản nhiều người ký (không phải hợp đồng), đánh số tự động của Word có thể cần chỉnh tay.

## Mã nguồn
Go thuần (không thư viện ngoài), giao diện Win32 gốc.

| Thư mục | Nội dung |
|---|---|
| `core/` | Đọc .docx/.txt, nhận diện + chuẩn hóa thành phần (`parse.go`), ghi .docx đúng thể thức (`docx.go`), chuyển .doc/.pdf và xuất PDF qua Word COM/LibreOffice (`convert_windows.go`) |
| `win/` | Lớp gọi Win32 API: cửa sổ, hộp thoại, registry, lối tắt (IShellLinkW) |
| `cmd/app/` | Ứng dụng giao diện `DinhDangVanBanChuan.exe` |
| `cmd/setup/` | Bộ cài `setup.exe` (nhúng app, lối tắt, gỡ cài đặt, `/S`) |
| `cmd/cli/` | Bản dòng lệnh `ddvbc-cli.exe` |
| `tools/` | Sinh bảng độ rộng phông, icon, tài nguyên Windows (icon/manifest/version) |
| `vidu/` | Văn bản mẫu để thử |

### Build
Cần Go ≥ 1.21, Python 3 + Pillow (+ fontTools nếu sinh lại bảng phông). Chạy được trên Linux/macOS/Windows (cross-compile):
```
./build.sh     # ra dist/: DinhDangVanBanChuan.exe, DinhDangVanBanChuan_Setup_vX.exe, ddvbc-cli.exe
```
Đổi phiên bản: sửa `VERSION`, thêm mục trong `CHANGELOG.md`, chạy lại `build.sh`.
Thử lõi trên Linux: `go run ./cmd/cli -pdf vidu/QD_test.docx` (cần LibreOffice để xuất PDF).

## Lịch sử phiên bản
Xem [CHANGELOG.md](CHANGELOG.md).
