# Nhật ký dự án DocStd — ĐỊNH DẠNG VĂN BẢN CHUẨN

Tổng hợp phiên làm việc ngày 26/11/2026 (anh Hồ Viết Thắng và Claude). Dùng làm tài liệu nền cho project "App - Định dạng văn bản chuẩn".

## 1. Mục tiêu
Ứng dụng Windows nhận văn bản (.docx, .doc, .txt, .pdf) và xuất .docx (tùy chọn thêm .pdf) đúng thể thức:
- Văn bản hành chính: NĐ 30/2020/NĐ-CP.
- VBQPPL: NĐ 78/2025/NĐ-CP, Phụ lục I thay bởi NĐ 187/2025/NĐ-CP.

Yêu cầu giao diện: tên "ĐỊNH DẠNG VĂN BẢN CHUẨN"; nút thu nhỏ/phóng to/tắt; thanh trạng thái dưới ghi chữ nhỏ "Định dạng lại / soạn văn bản đúng thể thức pháp luật Việt Nam (NĐ 30/2020 cho văn bản hành chính; NĐ 78/2025 sửa bởi NĐ 187/2025 cho VBQPPL)"; góc dưới phải là phiên bản; hai lựa chọn xuất (chỉ Word / Word + PDF); có bộ cài setup.exe để cài máy khác; đóng gói code kèm version.

Tên hiển thị trên biểu tượng Desktop/Start Menu/Control Panel: **DocStd**.
Đơn vị: Phòng Kỹ thuật - VTTB - CNTT, Bệnh viện Hữu nghị đa khoa Nghệ An - Giai đoạn 2.

## 2. Vị trí
- Mã nguồn, bộ cài, gói zip theo phiên bản: `D:\ClaudeTN\ClaudeProject\DinhDangVanBanChuan\` (mã nguồn trong `src\`).
- Kết quả chạy thử văn bản thật: `D:\ClaudeTN\ClaudeData\DataTest\`.
- Skill cũ (giữ nguyên, không dùng chung): `D:\ClaudeTN\ClaudeProject\DinhDangVanBan\`.
- GitHub: repo public `dinh-dang-van-ban-chuan`; mỗi phiên bản có Release kèm setup.exe.

## 3. Quyết định kỹ thuật
- **Go thuần, Win32 gốc, không thư viện ngoài**: môi trường build không tải được thư viện GitHub; cross-compile ra .exe từ Linux.
- Tài nguyên Windows (icon, manifest Common Controls v6 + DPI aware, thông tin phiên bản) tự sinh bằng `tools/mkrsrc.py` thành file `.syso`.
- Đọc .docx trực tiếp (XML); .doc/.pdf chuyển qua **Microsoft Word COM** (PowerShell), dự phòng LibreOffice. Xuất PDF bằng `ExportAsFixedFormat`.
- Bộ cài tự viết (Go): cài cho người dùng hiện tại vào `%LOCALAPPDATA%\Programs\DinhDangVanBanChuan`, không cần Admin; lối tắt tạo bằng **IShellLinkW** (WScript.Shell không lưu được tên tiếng Việt); đăng ký gỡ cài đặt ở HKCU; `/S` cài im lặng.
- Độ rộng chữ đo theo bảng số đo Times New Roman (Liberation Serif tương thích) để dựng đường kẻ và chia cột bảng.

## 4. Lịch sử sự cố và cách sửa
| Bản | Sự cố anh báo | Nguyên nhân | Cách sửa |
|---|---|---|---|
| 1.0.1 | Cài báo lỗi "không tạo được lối tắt" (FileNotFoundException) | WScript.Shell không lưu được tên .lnk có Ị, Ạ, Ă | Tạo lối tắt bằng IShellLinkW; thay icon theo ảnh anh gửi |
| 1.0.2 | Còn viền khung, bảng không vừa ô, "Người làm báo cáo" bên trái, lặp Quốc hiệu | Regex không khớp "Cộng **hòa**" chữ thường → bảng tiêu đề bị coi là bảng số liệu | Nhận diện Quốc hiệu mọi kiểu chữ; tách cấp trên/phòng; nhận "NGƯỜI LÀM BÁO CÁO"; chia cột theo chữ thật |
| 1.0.3 | Chạy lại trên file đã định dạng | Phần đầu giả "...", "Số: /" làm hỏng nhận diện | Bỏ qua phần đầu giả; trích yếu HOA → thường; "2 Đánh giá" → "2. Đánh giá" |
| 1.0.4 | Vẫn lặp Quốc hiệu | Dòng "Bộ phận CNTT" viết thường xen giữa làm dừng nhận diện phần đầu | Xác định phần đầu theo vị trí tên loại; chặn Quốc hiệu lọt vào thân; đoạn cuối đi cùng chữ ký |
| 1.0.5 | Hợp đồng FPT (bản scan) không nhận được gì | Hợp đồng ký 2 bên, phụ lục sau chữ ký, lỗi OCR | Khối ký Bên A/Bên B, phụ lục sang trang, so khớp không dấu, dọn dòng lặp/dấu đầu dòng OCR |
| 1.0.6 | Xuất PDF lỗi 0x800706BE ở `$w.Quit()` | Office 365 gắn vào Word đang mở; Quit cố đóng Word của người dùng | Không ẩn/đóng Word đang mở sẵn; kiểm tra file PDF đã tạo; thông báo lỗi gọn |
| 1.0.7 | Ô bảng xuống dòng làm lệch | Cột chia theo tiêu đề dài, cột chữ bị ép | Ưu tiên dữ liệu không xuống dòng, tiêu đề được xuống dòng, chỉ thu hẹp cột chữ; ô trống cùng chiều cao |
| 1.0.8 | Giải thích 3 lựa chọn loại văn bản | — | Ghi vào file hướng dẫn; thêm nút "Hướng dẫn" |
| 1.0.9 | Đổi tên hiển thị | — | "DocStd" ở lối tắt và Control Panel; xóa lối tắt tên cũ khi nâng cấp |
| 1.0.10 | Sửa tên đơn vị | — | "…Bệnh viện Hữu nghị đa khoa Nghệ An - Giai đoạn 2" |

## 5. Đã kiểm chứng / chưa kiểm chứng
- **Đã kiểm chứng (trên Linux + LibreOffice):** lõi nhận diện và file .docx/.pdf với 9 văn bản thử (quyết định, công văn, báo cáo, VBQPPL, 3 báo cáo máy in của anh, hợp đồng FPT); tài nguyên .exe (icon, version, manifest) đọc bằng pefile; mã Windows qua `go vet`.
- **Chưa kiểm chứng (cần anh thử trên Windows):** giao diện, bộ cài/gỡ, lối tắt DocStd, nút Hướng dẫn, xuất PDF qua Word bản 1.0.6+, hiển thị bảng trong Word (đo chữ có thể lệch nhẹ so với LibreOffice — đã chừa 3% + 1,5 mm).

## 6. Việc có thể làm tiếp
- Đánh số trang riêng cho phụ lục/văn bản kèm theo.
- Văn bản nhiều người ký (không phải hợp đồng).
- Đọc đánh số tự động (numbering) của Word.
- Tùy chọn bỏ ô "Số" và "Nơi nhận" cho báo cáo nội bộ.
- Ký số file .exe để Windows SmartScreen không cảnh báo.
