#!/bin/sh
# Build ĐỊNH DẠNG VĂN BẢN CHUẨN cho Windows 64-bit (chạy được trên Linux/macOS/Windows có Go >= 1.21, Python 3 + Pillow).
set -e
cd "$(dirname "$0")"
V=$(cat VERSION)
sed -i "s/const Version = \".*\"/const Version = \"$V\"/" core/version.go
python3 tools/gen_widths.py core/widths_gen.go 2>/dev/null || true
python3 tools/mkrsrc.py assets/app.ico "$V" "ĐỊNH DẠNG VĂN BẢN CHUẨN" DinhDangVanBanChuan.exe cmd/app/rsrc_windows_amd64.syso
python3 tools/mkrsrc.py assets/app.ico "$V" "Cài đặt ĐỊNH DẠNG VĂN BẢN CHUẨN" setup.exe cmd/setup/rsrc_windows_amd64.syso
python3 tools/mkrsrc.py assets/app.ico "$V" "ĐỊNH DẠNG VĂN BẢN CHUẨN - dòng lệnh" ddvbc-cli.exe cmd/cli/rsrc_windows_amd64.syso
mkdir -p dist
cp cmd/setup/payload/HuongDanSuDung.txt cmd/app/HuongDanSuDung.txt
export GOOS=windows GOARCH=amd64 CGO_ENABLED=0
go build -trimpath -ldflags "-H windowsgui -s -w" -o dist/DinhDangVanBanChuan.exe ./cmd/app
gzip -9 -c dist/DinhDangVanBanChuan.exe > cmd/setup/payload/DinhDangVanBanChuan.exe.gz
cp cmd/setup/payload/HuongDanSuDung.txt dist/
go build -trimpath -ldflags "-H windowsgui -s -w" -o "dist/DinhDangVanBanChuan_Setup_v$V.exe" ./cmd/setup
go build -trimpath -ldflags "-s -w" -o dist/ddvbc-cli.exe ./cmd/cli
ls -la dist
