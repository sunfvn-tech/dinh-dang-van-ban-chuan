// Công cụ dòng lệnh: ddvbc-cli [-p auto|vbhc|vbqppl] [-pdf] [-o thư_mục] file...
package main

import (
	"flag"
	"fmt"
	"os"

	"ddvbc/core"
)

func main() {
	prof := flag.String("p", "auto", "bộ quy định: auto | vbhc | vbqppl")
	pdf := flag.Bool("pdf", false, "xuất thêm PDF")
	out := flag.String("o", "", "thư mục kết quả (mặc định cùng thư mục file gốc)")
	flag.Parse()
	fmt.Println(core.AppName, "v"+core.Version)
	code := 0
	for _, f := range flag.Args() {
		_, err := core.Process(core.Options{Input: f, OutDir: *out, Profile: *prof, PDF: *pdf}, func(s string) { fmt.Println(s) })
		if err != nil {
			fmt.Println("LỖI:", err)
			code = 1
		}
	}
	os.Exit(code)
}
