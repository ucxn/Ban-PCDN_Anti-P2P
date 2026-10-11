//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	user32       = syscall.NewLazyDLL("user32.dll")
	shell32      = syscall.NewLazyDLL("shell32.dll")
	messageBoxW  = user32.NewProc("MessageBoxW")
	shellExecuteW = shell32.NewProc("ShellExecuteW")
)

func messageBox(text, caption string, flags uintptr) {
	t, _ := syscall.UTF16PtrFromString(text)
	c, _ := syscall.UTF16PtrFromString(caption)
	_, _, _ = messageBoxW.Call(0, uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(c)), flags)
}

func pickWorkbookPowerShell() string {
	script := `$d=New-Object System.Windows.Forms.OpenFileDialog;` +
		`$d.Title='选择 Excel 工作簿';` +
		`$d.Filter='Excel (*.xlsx;*.xlsm)|*.xlsx;*.xlsm|所有文件 (*.*)|*.*';` +
		`$d.Multiselect=$false;` +
		`if($d.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK){[Console]::Out.Write($d.FileName)}`
	cmd := exec.Command(
		"powershell.exe", "-NoProfile", "-STA",
		"-ExecutionPolicy", "Bypass",
		"-Command", script,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func pickWorkbookLegacy() string {
	// 保留一个极简 fallback：通过 PowerShell 的另一种启动方式再尝试一次。
	script := `[Console]::OutputEncoding=[Text.Encoding]::UTF8;` +
		`Add-Type -AssemblyName System.Windows.Forms;` +
		`$d=New-Object System.Windows.Forms.OpenFileDialog;` +
		`$d.Filter='Excel (*.xlsx;*.xlsm)|*.xlsx;*.xlsm';` +
		`if($d.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK){$d.FileName}`
	cmd := exec.Command("powershell.exe", "-NoProfile", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func pickWorkbook() string {
	if p := pickWorkbookPowerShell(); p != "" {
		return p
	}
	return pickWorkbookLegacy()
}

func openFolder(dir string) {
	p, _ := syscall.UTF16PtrFromString(dir)
	verb, _ := syscall.UTF16PtrFromString("open")
	_, _, _ = shellExecuteW.Call(0, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(p)), 0, 0, 1)
}

func main() {
	path := ""
	if len(os.Args) >= 2 {
		path = os.Args[1]
	} else {
		path = pickWorkbook()
	}
	if path == "" {
		return
	}

	stats, err := ExportWorkbook(path)
	if err != nil {
		messageBox("导出失败：\n\n"+err.Error(), "工作簿文本导出器", 0x00000010)
		return
	}

	dir := filepath.Join(filepath.Dir(path), "export")
	msg := fmt.Sprintf(
		"导出完成\n\nPCDN %d 条\nHosts %d 行\nADGHome %d 行\n\n%s",
		stats.PCDN, stats.Hosts, stats.ADGHome, dir,
	)
	messageBox(msg, "工作簿文本导出器", 0x00000040)
	openFolder(dir)
}
