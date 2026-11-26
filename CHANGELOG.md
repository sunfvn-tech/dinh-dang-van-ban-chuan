# Nhật ký phiên bản — ĐỊNH DẠNG VĂN BẢN CHUẨN

## 1.0.10 — 26/11/2026
- Bộ cài nhỏ hơn (chương trình bên trong được nén, tự giải nén khi cài).
- Sửa tên đơn vị phát triển: "Phòng Kỹ thuật - VTTB - CNTT, Bệnh viện Hữu nghị đa khoa Nghệ An - Giai đoạn 2" (Publisher trong Control Panel/Settings › Apps, thông tin phiên bản của file .exe, cuối file hướng dẫn).

## 1.0.9 — 26/11/2026
- Đổi tên hiển thị thành "DocStd": biểu tượng Desktop, Start Menu, mục Send to và tên trong Control Panel / Settings › Apps. Tên đầy đủ "ĐỊNH DẠNG VĂN BẢN CHUẨN" vẫn giữ trên cửa sổ chương trình và ghi trong phần mô tả (Comments) của mục gỡ cài đặt.
- Khi nâng cấp từ bản cũ: tự xóa các lối tắt tên cũ "ĐỊNH DẠNG VĂN BẢN CHUẨN" để không bị trùng biểu tượng. Gỡ cài đặt xóa lối tắt cả tên mới và tên cũ.

## 1.0.8 — 26/11/2026
- Hướng dẫn sử dụng: thêm mục "Chọn loại văn bản — vì sao có 3 lựa chọn" (2 bộ quy định NĐ 30/2020 và NĐ 78/2025 sửa bởi NĐ 187/2025, bảng các điểm trình bày khác nhau, quy tắc tự nhận diện, khi nào chọn tay); cập nhật mục "Chương trình tự làm" và "Giới hạn".
- Thêm nút "Hướng dẫn" trong cửa sổ chương trình (mở file hướng dẫn, kể cả khi chạy bản không cài).

## 1.0.7 — 26/11/2026
- Chia độ rộng cột bảng chuẩn xác hơn:
  1. Nội dung các dòng dữ liệu không xuống dòng (đo theo phông Times New Roman thật, có dự phòng 3% + 1,5 mm cho chênh lệch Word/LibreOffice).
  2. Dòng tiêu đề được phép xuống dòng nhưng không cắt giữa từ; phần rộng dư ưu tiên cho tiêu đề đang xuống dòng nhiều nhất, sau đó chia cho cột chữ.
  3. Bảng quá rộng: chỉ thu hẹp cột chữ (cột dài thu nhiều hơn), cột số giữ nguyên để số không bị bẻ đôi.
- Ô trống trong bảng cao bằng các ô khác (cỡ chữ dấu đoạn bằng cỡ chữ bảng).

## 1.0.6 — 26/11/2026
- Sửa lỗi xuất PDF "Exception calling Quit... The remote procedure call failed (0x800706BE)":
  - Nếu Word đang mở sẵn (Office 365 gắn vào Word đang chạy), chương trình không ẩn và không đóng Word của người dùng — chỉ đóng tài liệu nó tự mở (tránh mất văn bản chưa lưu).
  - Lỗi khi đóng Word không còn làm hỏng kết quả: kiểm tra file PDF/DOCX đã tạo thì báo thành công.
  - Xuất PDF bằng ExportAsFixedFormat; tắt macro tự chạy khi mở file; tự thử lại 1 lần nếu Word lỗi.
  - Thông báo lỗi PowerShell (CLIXML) được rút gọn thành câu dễ đọc, tiếng Việt không bị lỗi font.

## 1.0.5 — 26/11/2026
- Hỗ trợ HỢP ĐỒNG, BIÊN BẢN, BẢN THỎA THUẬN (ký hai bên): khối "ĐẠI DIỆN BÊN A" (trái) / "ĐẠI DIỆN BÊN B" (phải), chức vụ và họ tên lấy từ mục "Đại diện:", "Chức vụ:" của mỗi bên; tên hai bên thẳng hàng; không tự thêm "Nơi nhận".
- Số ký hiệu đặt dưới tên loại (kiểu hợp đồng) được đưa lên ô Số.
- Phụ lục sau chữ ký ("PHỤ LỤC 01...") tách riêng, mỗi phụ lục sang trang mới, dòng "(Kèm theo...)" in nghiêng.
- Nhận diện chịu lỗi dấu do nhận dạng chữ (OCR): "HỢP ĐÒNG", "ĐIÊU 1:", "ĐIÈU 2:" → "HỢP ĐỒNG", "Điều 1.", "Điều 2."; tên điều viết hoa toàn bộ → chữ thường đậm.
- "PHẦN 1 - ..." trình bày như chương (Phần 1 / tên phần căn giữa, đậm); dòng in hoa ngắn (GIỮA, VÀ, tên các bên) căn giữa đậm.
- Văn bản chuyển từ bản scan: sửa dấu đầu dòng bị nhận sai (s, v', o → -), bỏ khối dòng bị nhân đôi liền nhau, bỏ ký tự rác ở phần đầu, cảnh báo văn bản OCR.

## 1.0.4 — 26/11/2026
- Sửa lặp Quốc hiệu: phần đầu văn bản nay được xác định theo vị trí tên loại (BÁO CÁO, QUYẾT ĐỊNH...) — mọi dòng phía trên là phần đầu, Quốc hiệu/Tiêu ngữ/Số/ngày được nhận ra ở bất kỳ thứ tự, kiểu chữ nào; các dòng còn lại là tên cơ quan, đơn vị (kể cả dòng viết thường như "Bộ phận CNTT").
- Lớp an toàn: dòng Quốc hiệu/Tiêu ngữ đứng riêng không bao giờ được in lại trong thân văn bản.
- Cơ quan nhiều cấp: mỗi cấp một dòng (cấp trên chữ thường, đơn vị ban hành in đậm có kẻ dưới).
- Đoạn cuối luôn đi cùng khối chữ ký (không để chữ ký đứng riêng một trang).

## 1.0.3 — 26/11/2026
- Chạy được cả trên file đã qua định dạng trước đó (bỏ qua phần đầu giả "...", "Số:      /", "..., ngày ... tháng ... năm ...").
- Trích yếu viết HOA toàn bộ → chữ thường đứng đậm theo NĐ 30, giữ nguyên chữ viết tắt (BV, HNĐK, GĐ2, UBND, CNTT...).
- Sửa đề mục thiếu dấu chấm sau số thứ tự ("2 Đánh giá:" → "2. Đánh giá:").

## 1.0.2 — 26/11/2026
- Sửa: bảng tiêu đề (cơ quan | Quốc hiệu) không được nhận ra khi Quốc hiệu viết thường "Cộng hòa xã hội chủ nghĩa Việt Nam" → trước đây giữ nguyên bảng có viền và in Quốc hiệu hai lần. Nay nhận diện theo Quốc hiệu/Tiêu ngữ ở mọi kiểu chữ, dựng lại phần đầu không viền đúng NĐ 30.
- Tách đúng cơ quan cấp trên (chữ thường) và đơn vị ban hành cấp phòng/khoa (in đậm, kẻ dưới), VD "CT CP BỆNH VIỆN HNĐK GĐ2" / "PHÒNG KT-VTTB-CNTT".
- Nhận diện khối ký dạng "NGƯỜI LÀM BÁO CÁO", "NGƯỜI LẬP", "XÁC NHẬN CỦA...", "TRƯỞNG PHÒNG..." → đặt ở góc dưới bên phải.
- Bảng số liệu: độ rộng cột tính theo chữ thật (không bẻ đôi "TT", "10", "23"), căn giữa theo chiều dọc, cột số căn giữa, cột chữ căn trái, dòng tiêu đề đậm và lặp lại khi sang trang, dòng "Tổng/Cộng" in đậm, không tách dòng qua trang, hỗ trợ ô gộp cột.
- Bỏ khoảng trống thừa dưới trích yếu.

## 1.0.1 — 26/11/2026
- Sửa lỗi bộ cài báo "không tạo được lối tắt" (FileNotFoundException): lối tắt giờ được tạo trực tiếp bằng Windows API (IShellLinkW), không qua PowerShell/WScript.Shell, nên lưu được tên tiếng Việt có dấu "ĐỊNH DẠNG VĂN BẢN CHUẨN.lnk".
- Lối tắt lỗi không còn làm hỏng cả quá trình cài (chỉ cảnh báo).
- Gỡ cài đặt xóa lối tắt trực tiếp, không cần PowerShell.
- Icon mới (Quốc huy và tập văn bản).
- Thông báo lỗi chuyển đổi Word gọn hơn (tắt thanh tiến trình PowerShell).

## 1.0.0 — 26/11/2026
- Phát hành đầu tiên.
- Đầu vào: .docx, .doc, .txt, .pdf (thêm .rtf, .odt); chọn nhiều file hoặc kéo thả vào cửa sổ, kéo thả lên biểu tượng, "Gửi tới (Send to)".
- Hai bộ quy định: văn bản hành chính (NĐ 30/2020/NĐ-CP) và VBQPPL (NĐ 78/2025/NĐ-CP sửa bởi NĐ 187/2025/NĐ-CP); có chế độ tự nhận diện.
- Hai lựa chọn xuất: chỉ Word (.docx) hoặc Word + PDF.
- Tự nhận diện và trình bày lại: cơ quan chủ quản/ban hành, Quốc hiệu - Tiêu ngữ, số ký hiệu, địa danh ngày tháng, tên loại, trích yếu (V/v cho công văn), kính gửi, thẩm quyền, căn cứ, chương/mục/điều, mục La Mã, bảng, khối chữ ký, nơi nhận, phần kèm theo.
- Chuẩn hóa: số < 10 thêm 0, ngày < 10 và tháng 1, 2 thêm 0, ký hiệu viết hoa, dấu câu dòng căn cứ và nơi nhận, bổ sung "Lưu: VT", bỏ học hàm học vị người ký (VBHC), khoảng trắng và dấu câu.
- Trang A4, lề 20-20-30-15 mm, Times New Roman, số trang giữa lề trên và ẩn ở trang 1.
- Bộ cài setup.exe: cài cho người dùng hiện tại (không cần Admin), lối tắt Desktop/Start Menu/Send to, gỡ cài đặt trong Settings › Apps, cài im lặng /S.
