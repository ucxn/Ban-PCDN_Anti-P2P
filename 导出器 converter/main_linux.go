//go:build !windows

package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法：pcdnexport <工作簿.xlsx/.xlsm>")
		return
	}
	if _, err := ExportWorkbook(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "导出失败：", err)
		os.Exit(1)
	}
}
