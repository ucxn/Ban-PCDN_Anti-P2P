// 本软件的 Windows 版为主要测试版本
// 本程序可以根据主协议授权，也可以根据 CC BY-NC-ND 授权。

package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	workbookSheetNS = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
)

type ExportStats struct {
	PCDN    int
	Hosts   int
	ADGHome int
}

type sharedStringItem struct {
	Text string
}

func (s *sharedStringItem) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var b strings.Builder
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "t" {
				var v string
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				b.WriteString(v)
			} else {
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if t.Name == start.Name {
				s.Text = b.String()
				return nil
			}
		}
	}
}

type colorSpec struct {
	RGB     string
	Theme   int
	Indexed int
	Tint    float64
	Valid   bool
}

type fontStyle struct {
	Bold  bool
	Color colorSpec
}

type fillStyle struct {
	Color colorSpec
}

type xfStyle struct {
	FontID int
	FillID int
}

type styleTable struct {
	Fonts  []fontStyle
	Fills  []fillStyle
	XFs    []xfStyle
	Themes []string
}

type cell struct {
	Ref   string
	Value string
	Style int
}

type sheetData struct {
	ByRef map[string]cell
	MaxRow int
}

type xlsxReader struct {
	zip          *zip.ReadCloser
	sheetByName  map[string]string
	sharedString []string
	styles       styleTable
}

func openXLSX(path string) (*xlsxReader, error) {
	z, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	r := &xlsxReader{
		zip:         z,
		sheetByName: map[string]string{},
	}
	if err := r.readWorkbook(); err != nil {
		z.Close()
		return nil, err
	}
	if err := r.readSharedStrings(); err != nil {
		z.Close()
		return nil, err
	}
	if err := r.readStyles(); err != nil {
		z.Close()
		return nil, err
	}
	return r, nil
}

func (r *xlsxReader) Close() error {
	return r.zip.Close()
}

func (r *xlsxReader) file(name string) ([]byte, error) {
	for _, f := range r.zip.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}
	return nil, os.ErrNotExist
}

type workbookXML struct {
	Sheets []struct {
		Name string `xml:"name,attr"`
		RID  string `xml:"id,attr"`
	} `xml:"sheets>sheet"`
}

type relsXML struct {
	Relationships []struct {
		ID     string `xml:"Id,attr"`
		Target string `xml:"Target,attr"`
		Type   string `xml:"Type,attr"`
	} `xml:"Relationship"`
}

func (r *xlsxReader) readWorkbook() error {
	data, err := r.file("xl/workbook.xml")
	if err != nil {
		return err
	}
	var wb workbookXML
	if err := xml.Unmarshal(data, &wb); err != nil {
		return err
	}
	relData, err := r.file("xl/_rels/workbook.xml.rels")
	if err != nil {
		return err
	}
	var rels relsXML
	if err := xml.Unmarshal(relData, &rels); err != nil {
		return err
	}
	relMap := map[string]string{}
	for _, rel := range rels.Relationships {
		relMap[rel.ID] = rel.Target
	}
	for _, s := range wb.Sheets {
		target := relMap[s.RID]
		target = strings.TrimPrefix(target, "/")
		target = strings.TrimPrefix(target, "xl/")
		r.sheetByName[s.Name] = filepath.ToSlash(filepath.Join("xl", target))
	}
	return nil
}

type sstXML struct {
	Items []sharedStringItem `xml:"si"`
}

func (r *xlsxReader) readSharedStrings() error {
	data, err := r.file("xl/sharedStrings.xml")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var sst sstXML
	if err := xml.Unmarshal(data, &sst); err != nil {
		return err
	}
	r.sharedString = make([]string, len(sst.Items))
	for i := range sst.Items {
		r.sharedString[i] = sst.Items[i].Text
	}
	return nil
}

type styleXML struct {
	Fonts struct {
		Fonts []struct {
			Bold  *struct{}   `xml:"b"`
			Color *colorXML   `xml:"color"`
		} `xml:"font"`
	} `xml:"fonts"`
	Fills struct {
		Fills []struct {
			Pattern struct {
				PatternType string   `xml:"patternType,attr"`
				FG          colorXML `xml:"fgColor"`
				BG          colorXML `xml:"bgColor"`
			} `xml:"patternFill"`
		} `xml:"fill"`
	} `xml:"fills"`
	CellXFs struct {
		XFs []struct {
			FontID int `xml:"fontId,attr"`
			FillID int `xml:"fillId,attr"`
		} `xml:"xf"`
	} `xml:"cellXfs"`
}

type colorXML struct {
	RGB     string  `xml:"rgb,attr"`
	Theme   int     `xml:"theme,attr"`
	Indexed int     `xml:"indexed,attr"`
	Tint    float64 `xml:"tint,attr"`
	Type    string  `xml:"type,attr"`
}

func convertColor(c colorXML) colorSpec {
	if c.RGB == "" && c.Theme == 0 && c.Indexed == 0 && c.Type == "" && c.Tint == 0 {
		return colorSpec{}
	}
	return colorSpec{
		RGB:     strings.TrimPrefix(strings.ToUpper(c.RGB), "#"),
		Theme:   c.Theme,
		Indexed: c.Indexed,
		Tint:    c.Tint,
		Valid:   true,
	}
}

func (r *xlsxReader) readStyles() error {
	data, err := r.file("xl/styles.xml")
	if err != nil {
		return err
	}
	var s styleXML
	if err := xml.Unmarshal(data, &s); err != nil {
		return err
	}
	r.styles.Fonts = make([]fontStyle, len(s.Fonts.Fonts))
	for i, f := range s.Fonts.Fonts {
		r.styles.Fonts[i] = fontStyle{
			Bold:  f.Bold != nil,
			Color: func() colorSpec {
				if f.Color == nil {
					return colorSpec{}
				}
				return convertColor(*f.Color)
			}(),
		}
	}
	r.styles.Fills = make([]fillStyle, len(s.Fills.Fills))
	for i, f := range s.Fills.Fills {
		r.styles.Fills[i] = fillStyle{Color: convertColor(f.Pattern.FG)}
	}
	r.styles.XFs = make([]xfStyle, len(s.CellXFs.XFs))
	for i, xf := range s.CellXFs.XFs {
		r.styles.XFs[i] = xfStyle{FontID: xf.FontID, FillID: xf.FillID}
	}
	return nil
}

func colNumber(ref string) int {
	n := 0
	for _, c := range ref {
		if c >= 'A' && c <= 'Z' {
			n = n*26 + int(c-'A'+1)
			continue
		}
		break
	}
	return n
}

func rowNumber(ref string) int {
	for i, c := range ref {
		if c >= '0' && c <= '9' {
			v, _ := strconv.Atoi(ref[i:])
			return v
		}
	}
	return 0
}

func parseCellValue(start xml.StartElement, d *xml.Decoder, shared []string) (string, error) {
	type cellXML struct {
		Ref string `xml:"r,attr"`
		T   string `xml:"t,attr"`
		V   string `xml:"v"`
		IS  *struct {
			Text []string `xml:"t"`
		} `xml:"is"`
	}
	var c cellXML
	if err := d.DecodeElement(&c, &start); err != nil {
		return "", err
	}
	if c.T == "s" {
		i, err := strconv.Atoi(strings.TrimSpace(c.V))
		if err != nil || i < 0 || i >= len(shared) {
			return "", nil
		}
		return shared[i], nil
	}
	if c.T == "inlineStr" && c.IS != nil {
		return strings.Join(c.IS.Text, ""), nil
	}
	return c.V, nil
}

func (r *xlsxReader) readSheetByName(name string) (*sheetData, error) {
	path, ok := r.sheetByName[name]
	if !ok {
		return nil, fmt.Errorf("找不到工作表：%s", name)
	}
	data, err := r.file(path)
	if err != nil {
		return nil, err
	}
	d := xml.NewDecoder(bytes.NewReader(data))
	out := &sheetData{ByRef: map[string]cell{}}
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		start, ok := tok.(xml.StartElement)
		if !ok || start.Name.Local != "c" {
			continue
		}
		var attrStyle, attrRef string
		var attrType string
		for _, a := range start.Attr {
			switch a.Name.Local {
			case "r":
				attrRef = a.Value
			case "s":
				attrStyle = a.Value
			case "t":
				attrType = a.Value
			}
		}
		raw, err := parseCellValue(start, d, r.sharedString)
		if err != nil {
			return nil, err
		}
		style := 0
		if attrStyle != "" {
			style, _ = strconv.Atoi(attrStyle)
		}
		_ = attrType
		out.ByRef[attrRef] = cell{Ref: attrRef, Value: raw, Style: style}
		if row := rowNumber(attrRef); row > out.MaxRow {
			out.MaxRow = row
		}
	}
	return out, nil
}

func (r *xlsxReader) styleFor(s int) xfStyle {
	if s >= 0 && s < len(r.styles.XFs) {
		return r.styles.XFs[s]
	}
	return xfStyle{}
}

func normalizeRGB(rgb string) string {
	rgb = strings.TrimSpace(strings.ToUpper(rgb))
	rgb = strings.TrimPrefix(rgb, "#")
	if len(rgb) == 8 {
		rgb = rgb[2:]
	}
	if len(rgb) == 6 {
		return rgb
	}
	return ""
}

func applyTint(rgb string, tint float64) string {
	rgb = normalizeRGB(rgb)
	if rgb == "" || tint == 0 {
		return rgb
	}
	hex := func(s string) float64 {
		v, _ := strconv.ParseUint(s, 16, 8)
		return float64(v) / 255
	}
	pack := func(v float64) string {
		if v < 0 {
			v = 0
		}
		if v > 1 {
			v = 1
		}
		return fmt.Sprintf("%02X", int(v*255+0.5))
	}
	r, g, b := hex(rgb[0:2]), hex(rgb[2:4]), hex(rgb[4:6])
	if tint < 0 {
		f := 1 + tint
		r, g, b = r*f, g*f, b*f
	} else {
		r, g, b = r+(1-r)*tint, g+(1-g)*tint, b+(1-b)*tint
	}
	return pack(r) + pack(g) + pack(b)
}

func (r *xlsxReader) schemeRGB(theme int) string {
	if theme >= 0 && theme < len(r.styles.Themes) {
		return normalizeRGB(r.styles.Themes[theme])
	}
	switch theme {
	case 0:
		return "000000"
	case 1:
		return "FFFFFF"
	case 2:
		return "1F497D"
	case 3:
		return "EEECE1"
	case 4:
		return "4F81BD"
	case 5:
		return "C0504D"
	case 6:
		return "9BBB59"
	case 7:
		return "8064A2"
	case 8:
		return "4BACC6"
	case 9:
		return "F79646"
	case 10:
		return "0000FF"
	case 11:
		return "800080"
	default:
		return ""
	}
}

func (r *xlsxReader) resolveColor(c colorSpec) string {
	if !c.Valid {
		return ""
	}
	rgb := normalizeRGB(c.RGB)
	if rgb == "" && c.Theme >= 0 {
		rgb = r.schemeRGB(c.Theme)
	}
	if rgb == "" {
		// 常见 indexed 色。
		switch c.Indexed {
		case 0:
			rgb = "000000"
		case 1:
			rgb = "FFFFFF"
		case 5:
			rgb = "FFFF00"
		case 10:
			rgb = "FF0000"
		case 12:
			rgb = "0000FF"
		}
	}
	return applyTint(rgb, c.Tint)
}

type visualStyle struct {
	Bold     bool
	FontRGB  string
	FillRGB  string
	FontCode string
	FillCode string
}

func (r *xlsxReader) styleInfo(index int) visualStyle {
	xf := r.styleFor(index)
	var out visualStyle
	if xf.FontID >= 0 && xf.FontID < len(r.styles.Fonts) {
		f := r.styles.Fonts[xf.FontID]
		out.Bold = f.Bold
		out.FontRGB = r.resolveColor(f.Color)
	}
	if xf.FillID >= 0 && xf.FillID < len(r.styles.Fills) {
		out.FillRGB = r.resolveColor(r.styles.Fills[xf.FillID].Color)
	}
	out.FontCode = fontColorCode(out.FontRGB)
	out.FillCode = fillColorCode(out.FillRGB)
	return out
}

func rgbTuple(rgb string) (int, int, int) {
	rgb = normalizeRGB(rgb)
	if len(rgb) != 6 {
		return 0, 0, 0
	}
	r, _ := strconv.ParseInt(rgb[0:2], 16, 64)
	g, _ := strconv.ParseInt(rgb[2:4], 16, 64)
	b, _ := strconv.ParseInt(rgb[4:6], 16, 64)
	return int(r), int(g), int(b)
}

func fontColorCode(rgb string) string {
	rgb = normalizeRGB(rgb)
	switch rgb {
	case "", "000000":
		return "黑"
	case "FF0000":
		return "红"
	case "C00000":
		return "深红"
	case "FF0066":
		return "粉"
	case "0070C0", "0000FF":
		return "蓝"
	case "00B050", "008000", "70AD47", "92D050":
		return "绿"
	default:
		r, g, b := rgbTuple(rgb)
		if float64(b) > float64(r)*1.2 && float64(b) > float64(g)*1.1 {
			return "蓝"
		}
		if float64(r) > float64(g)*1.6 && float64(r) > float64(b)*1.2 {
			return "红"
		}
		return "其它"
	}
}

func fillColorCode(rgb string) string {
	rgb = normalizeRGB(rgb)
	switch rgb {
	case "FFFF00", "FFF2CC", "FFD966":
		return "y"
	case "00B050", "C6E0B4", "E2F0D9", "92D050":
		return "g"
	default:
		return "n"
	}
}

type PCDNItem struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
	Note    string `json:"note"`
	Font    string `json:"font"`
	Bold    string `json:"bold"`
	Fill    string `json:"fill"`
	FontRGB string `json:"font_rgb"`
	FillRGB string `json:"fill_rgb"`
}

func parseIntegerID(v string) (int, bool) {
	i, err := strconv.Atoi(strings.TrimSpace(v))
	return i, err == nil
}

func (r *xlsxReader) extractPCDN(s *sheetData) []PCDNItem {
	var items []PCDNItem
	appendItem := func(idCell, nameCell, patternCell, noteCell string) {
		id, ok := parseIntegerID(s.ByRef[idCell].Value)
		if !ok {
			return
		}
		p := strings.TrimSpace(s.ByRef[patternCell].Value)
		if p == "" {
			return
		}
		st := r.styleInfo(s.ByRef[patternCell].Style)
		bold := "N"
		if st.Bold {
			bold = "B"
		}
		items = append(items, PCDNItem{
			ID:      id,
			Name:    s.ByRef[nameCell].Value,
			Pattern: p,
			Note:    s.ByRef[noteCell].Value,
			Font:    st.FontCode,
			Bold:    bold,
			Fill:    st.FillCode,
			FontRGB: st.FontRGB,
			FillRGB: st.FillRGB,
		})
	}
	for row := 1; row <= s.MaxRow; row++ {
		appendItem(
			fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row),
			fmt.Sprintf("C%d", row), fmt.Sprintf("D%d", row),
		)
		appendItem(
			fmt.Sprintf("E%d", row), fmt.Sprintf("F%d", row),
			fmt.Sprintf("G%d", row), fmt.Sprintf("H%d", row),
		)
	}
	// 以序号排序，得到 1..N 的机器顺序。
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].ID < items[i].ID {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	return items
}

func extractColumnA(s *sheetData, preserveBlank bool) []string {
	out := make([]string, 0, s.MaxRow)
	for row := 1; row <= s.MaxRow; row++ {
		v := s.ByRef[fmt.Sprintf("A%d", row)].Value
		if preserveBlank || v != "" {
			out = append(out, v)
		}
	}
	return out
}

func tsvEscape(v string) string {
	return strings.NewReplacer("\t", " ", "\r\n", "\\n", "\n", "\\n", "\r", "\\n").Replace(v)
}

func exportTimeUTC7621() time.Time {
	return time.Now().UTC().Add(7*time.Hour + 37*time.Minute + 15*time.Second + 600*time.Millisecond)
}

func writeJSONL(path string, items []PCDNItem) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, item := range items {
		if err := enc.Encode(item); err != nil {
			return err
		}
	}
	return nil
}

func writeTSV(path string, items []PCDNItem) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	if _, err := fmt.Fprintln(w, "id\tname\tpattern\tnote\tfont\tbold\tfill\tfont_rgb\tfill_rgb"); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := fmt.Fprintf(
			w,
			"%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			tsvEscape(item.Name),
			tsvEscape(item.Pattern),
			tsvEscape(item.Note),
			item.Font,
			item.Bold,
			item.Fill,
			item.FontRGB,
			item.FillRGB,
		); err != nil {
			return err
		}
	}
	return w.Flush()
}

func buildReadme(path string, stats ExportStats) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	t := exportTimeUTC7621().Format("2006-01-02 15:04:05")
	_, err = fmt.Fprintf(
		f,
		"时间：%s（UTC+7.621）\n\n\"PCDN域名\"按当前工作簿版式自动把 A:D 与 E:H 两组记录摊平成 1..N。\n\nPCDN：%d 条\nHosts：%d 行\nADGHome：%d 行\n",
		t, stats.PCDN, stats.Hosts, stats.ADGHome,
	)
	return err
}

func ExportWorkbook(path string) (ExportStats, error) {
	var stats ExportStats
	r, err := openXLSX(path)
	if err != nil {
		return stats, err
	}
	defer r.Close()

	pcdnSheet, err := r.readSheetByName("PCDN域名")
	if err != nil {
		return stats, err
	}
	hostsSheet, err := r.readSheetByName("Hosts")
	if err != nil {
		return stats, err
	}
	adgSheet, err := r.readSheetByName("ADGHome")
	if err != nil {
		return stats, err
	}

	pcdn := r.extractPCDN(pcdnSheet)
	hosts := extractColumnA(hostsSheet, false)
	adg := extractColumnA(adgSheet, true)

	outDir := filepath.Join(filepath.Dir(path), "export")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return stats, err
	}

	if err := writeJSONL(filepath.Join(outDir, "PCDN域名.jsonl"), pcdn); err != nil {
		return stats, err
	}
	if err := writeTSV(filepath.Join(outDir, "PCDN域名.tsv"), pcdn); err != nil {
		return stats, err
	}
	if err := writeLines(filepath.Join(outDir, "Hosts.txt"), hosts); err != nil {
		return stats, err
	}
	if err := writeLines(filepath.Join(outDir, "ADGHome.txt"), adg); err != nil {
		return stats, err
	}

	stats.PCDN = len(pcdn)
	stats.Hosts = len(hosts)
	stats.ADGHome = len(adg)
	if err := buildReadme(filepath.Join(outDir, "导出说明.txt"), stats); err != nil {
		return stats, err
	}
	return stats, nil
}

func writeLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, line := range lines {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return err
		}
	}
	return nil
}
