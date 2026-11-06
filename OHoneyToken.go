// ============================================================
//
//	OHoneyToken — Honeytoken/Canary Forge
//	Author  : khaled.s.haddad
//	Website : owlsec.org
//	Language: Go (stdlib only) — OWL suite
//
// ============================================================
package main

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const APPNAME = "OHoneyToken"
const SUITE = "Owl Deception Suite"
const TAGLINE = "Honeytoken / Canary Forge"
const SUBLINE = "Decoy AWS/Slack/GitHub/Stripe creds · deployable files · manifest · trigger correlation"
const AUTHOR = "khaled.s.haddad"
const WEBSITE = "owlsec.org"
const BANNER = `
 ██████╗ ██╗  ██╗ ██████╗ ███╗   ██╗███████╗██╗   ██╗████████╗ ██████╗ ██╗  ██╗███████╗███╗   ██╗
██╔═══██╗██║  ██║██╔═══██╗████╗  ██║██╔════╝╚██╗ ██╔╝╚══██╔══╝██╔═══██╗██║ ██╔╝██╔════╝████╗  ██║
██║   ██║███████║██║   ██║██╔██╗ ██║█████╗   ╚████╔╝    ██║   ██║   ██║█████╔╝ █████╗  ██╔██╗ ██║
██║   ██║██╔══██║██║   ██║██║╚██╗██║██╔══╝    ╚██╔╝     ██║   ██║   ██║██╔═██╗ ██╔══╝  ██║╚██╗██║
╚██████╔╝██║  ██║╚██████╔╝██║ ╚████║███████╗   ██║      ██║   ╚██████╔╝██║  ██╗███████╗██║ ╚████║
 ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝   ╚═╝      ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═══╝
`

// ---- OWL UI KIT (cyan "owl" house style) — stdlib only. Shared across the suite ----
// CYAN is the primary brand/frame colour; red is reserved for errors and CRITICAL/HIGH.
const (
	R   = "\033[91m"
	GRN = "\033[92m"
	YLW = "\033[93m"
	CY  = "\033[96m"
	W   = "\033[97m"
	B   = "\033[1m"
	X   = "\033[0m"
	MG  = "\033[95m"
	DIM = "\033[2m"
	OR  = "\033[38;5;208m"
	GY  = "\033[90m"
	DR  = CY
)

var noColor = os.Getenv("NO_COLOR") != "" || !isatty()

func isatty() bool { fi, _ := os.Stdout.Stat(); return (fi.Mode() & os.ModeCharDevice) != 0 }
func col(s string, c ...string) string {
	if noColor {
		return s
	}
	return strings.Join(c, "") + s + X
}
func clear()           { fmt.Print("\033[H\033[2J") }
func line(w int)       { fmt.Println(col(strings.Repeat("─", w), CY)) }
func rule()            { fmt.Println(col(strings.Repeat("─", 66), CY)) }
func okMsg(s string)   { fmt.Println("  " + col("[✔] ", GRN+B) + col(s, W)) }
func errMsg(s string)  { fmt.Println("  " + col("[✘] ", R+B) + col(s, R)) }
func warnMsg(s string) { fmt.Println("  " + col("[!] ", OR+B) + col(s, YLW)) }
func info(s string)    { fmt.Println("  " + col("[*] ", CY) + col(s, DIM+W)) }

func banner() {
	fmt.Println(col(BANNER, CY))
	fmt.Println("  " + col(SUITE+"  ", B+MG) + col("v"+VERSION, YLW))
	fmt.Println("  " + col(SUBLINE, W))
	fmt.Println()
	fmt.Println("  " + col("Developer : ", DIM) + col(AUTHOR, CY) +
		col("   |   Website : ", DIM) + col(WEBSITE, CY))
	rule()
	fmt.Println()
}
func prompt(label string) string {
	fmt.Print("  " + col("[»] ", YLW) + col(label+": ", W))
	s, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(s)
}
func pause() {
	fmt.Print(col("\n  [↩] Press Enter to continue...", YLW))
	bufio.NewReader(os.Stdin).ReadString('\n')
}
func menu(title string, items [][2]string) string {
	clear()
	banner()
	fmt.Println("  " + col("⬡  "+title, B+CY))
	fmt.Println()
	for i, it := range items {
		fmt.Println("  " + col(pad(fmt.Sprintf("[%d]", i+1), 4), CY+B) + "  " +
			col(pad(it[0], 24), W) + col("─ "+it[1], DIM))
	}
	fmt.Println()
	fmt.Println("  " + col(pad("[H]", 4), YLW+B) + "  " + col(pad("Help & Usage", 24), W) + col("─ How to use, every option, CLI flags & examples", DIM))
	fmt.Println("  " + col(pad("[0]", 4), R+B) + "  " + col(pad("Exit", 24), W) + col("─ Quit "+APPNAME, DIM))
	fmt.Println()
	rule()
	return prompt("Select option")
}
func pad(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
func table(headers []string, rows [][]string) {
	w := make([]int, len(headers))
	for i, h := range headers {
		w[i] = len(h)
	}
	for _, r := range rows {
		for i, c := range r {
			if i < len(w) && len(c) > w[i] {
				w[i] = len(c)
			}
		}
	}
	var hb strings.Builder
	for i, h := range headers {
		hb.WriteString(col(pad(h, w[i]), CY+B))
		if i < len(headers)-1 {
			hb.WriteString(col(" │ ", DR))
		}
	}
	fmt.Println("  " + hb.String())
	seg := make([]string, len(w))
	for i := range w {
		seg[i] = strings.Repeat("─", w[i])
	}
	fmt.Println("  " + col(strings.Join(seg, "─┼─"), DR))
	for _, r := range rows {
		var rb strings.Builder
		for i := range headers {
			v := ""
			if i < len(r) {
				v = r[i]
			}
			rb.WriteString(col(pad(v, w[i]), W))
			if i < len(headers)-1 {
				rb.WriteString(col(" │ ", DR))
			}
		}
		fmt.Println("  " + rb.String())
	}
}

// ---- Common finding / report types ----
type Finding struct {
	ID          string            `json:"id"`
	Severity    string            `json:"severity"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Evidence    string            `json:"evidence"`
	Mitre       []string          `json:"mitre_attack"`
	Category    string            `json:"category"`
	Score       float64           `json:"score"`
	Tags        []string          `json:"tags"`
	Location    string            `json:"location"`
	Timestamp   string            `json:"timestamp"`
	Extra       map[string]string `json:"extra,omitempty"`
}

type Report struct {
	Tool        string            `json:"tool"`
	Version     string            `json:"version"`
	GeneratedAt string            `json:"generated_at"`
	Subject     string            `json:"subject"`
	Grade       string            `json:"grade"`
	Score       float64           `json:"score"`
	Summary     string            `json:"summary"`
	Findings    []Finding         `json:"findings"`
	Stats       map[string]int    `json:"stats"`
	Meta        map[string]string `json:"meta"`
}

const VERSION = "1.0"

func newReport(subject string) *Report {
	return &Report{
		Tool: APPNAME, Version: VERSION,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Subject:     subject, Grade: "N/A", Score: 0,
		Findings: []Finding{}, Stats: map[string]int{}, Meta: map[string]string{},
	}
}

func (r *Report) add(f Finding) {
	if f.ID == "" {
		f.ID = fmt.Sprintf("F-%04d", len(r.Findings)+1)
	}
	if f.Timestamp == "" {
		f.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if f.Extra == nil {
		f.Extra = map[string]string{}
	}
	r.Findings = append(r.Findings, f)
	r.Stats[f.Severity]++
	r.Stats["total"]++
}

func gradeFromScore(score float64) string {
	switch {
	case score >= 97:
		return "A+"
	case score >= 93:
		return "A"
	case score >= 90:
		return "A-"
	case score >= 87:
		return "B+"
	case score >= 83:
		return "B"
	case score >= 80:
		return "B-"
	case score >= 77:
		return "C+"
	case score >= 73:
		return "C"
	case score >= 70:
		return "C-"
	case score >= 67:
		return "D+"
	case score >= 63:
		return "D"
	case score >= 60:
		return "D-"
	default:
		return "F"
	}
}

func severityWeight(s string) float64 {
	switch strings.ToLower(s) {
	case "critical":
		return 25
	case "high":
		return 15
	case "medium":
		return 8
	case "low":
		return 3
	case "info":
		return 0.5
	default:
		return 1
	}
}

func (r *Report) finalize() {
	penalty := 0.0
	for _, f := range r.Findings {
		penalty += severityWeight(f.Severity)
	}
	score := 100.0 - penalty
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	r.Score = score
	r.Grade = gradeFromScore(score)
	r.Summary = fmt.Sprintf("%d findings · grade %s (%.1f/100)", len(r.Findings), r.Grade, r.Score)
}

func exportTXT(r *Report) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("=== %s v%s Report ===\n", r.Tool, r.Version))
	b.WriteString(fmt.Sprintf("Generated : %s\n", r.GeneratedAt))
	b.WriteString(fmt.Sprintf("Subject   : %s\n", r.Subject))
	b.WriteString(fmt.Sprintf("Grade     : %s (%.1f)\n", r.Grade, r.Score))
	b.WriteString(fmt.Sprintf("Summary   : %s\n", r.Summary))
	b.WriteString(strings.Repeat("=", 64) + "\n\n")
	for _, f := range r.Findings {
		b.WriteString(fmt.Sprintf("[%s] %s — %s\n", f.Severity, f.ID, f.Title))
		b.WriteString(fmt.Sprintf("  Category : %s\n", f.Category))
		b.WriteString(fmt.Sprintf("  Location : %s\n", f.Location))
		b.WriteString(fmt.Sprintf("  Evidence : %s\n", f.Evidence))
		b.WriteString(fmt.Sprintf("  Detail   : %s\n", f.Description))
		if len(f.Mitre) > 0 {
			b.WriteString(fmt.Sprintf("  MITRE    : %s\n", strings.Join(f.Mitre, ", ")))
		}
		if len(f.Tags) > 0 {
			b.WriteString(fmt.Sprintf("  Tags     : %s\n", strings.Join(f.Tags, ", ")))
		}
		b.WriteString("\n")
	}
	b.WriteString(fmt.Sprintf("\nStats: %v\n", r.Stats))
	return b.String()
}

func exportMD(r *Report) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s Report\n\n", r.Tool))
	b.WriteString("| Field | Value |\n|---|---|\n")
	b.WriteString(fmt.Sprintf("| Version | %s |\n", r.Version))
	b.WriteString(fmt.Sprintf("| Generated | %s |\n", r.GeneratedAt))
	b.WriteString(fmt.Sprintf("| Subject | `%s` |\n", r.Subject))
	b.WriteString(fmt.Sprintf("| Grade | **%s** (%.1f) |\n", r.Grade, r.Score))
	b.WriteString(fmt.Sprintf("| Summary | %s |\n\n", r.Summary))
	b.WriteString("## Findings\n\n")
	for _, f := range r.Findings {
		b.WriteString(fmt.Sprintf("### %s — %s (%s)\n\n", f.ID, f.Title, f.Severity))
		b.WriteString(fmt.Sprintf("- **Category:** %s\n", f.Category))
		b.WriteString(fmt.Sprintf("- **Location:** `%s`\n", f.Location))
		b.WriteString(fmt.Sprintf("- **Evidence:** %s\n", f.Evidence))
		b.WriteString(fmt.Sprintf("- **Description:** %s\n", f.Description))
		if len(f.Mitre) > 0 {
			b.WriteString(fmt.Sprintf("- **MITRE ATT&CK:** %s\n", strings.Join(f.Mitre, ", ")))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func exportJSON(r *Report) string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	return string(data)
}

func exportCSV(r *Report) string {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "severity", "title", "category", "location", "evidence", "mitre", "score", "tags"})
	for _, f := range r.Findings {
		_ = w.Write([]string{f.ID, f.Severity, f.Title, f.Category, f.Location, f.Evidence,
			strings.Join(f.Mitre, ";"), fmt.Sprintf("%.2f", f.Score), strings.Join(f.Tags, ";")})
	}
	w.Flush()
	return buf.String()
}

func exportSARIF(r *Report) string {
	type sarifMsg struct {
		Text string `json:"text"`
	}
	type artLoc struct {
		URI string `json:"uri"`
	}
	type physLoc struct {
		ArtifactLocation artLoc `json:"artifactLocation"`
	}
	type sarifLoc struct {
		PhysicalLocation physLoc `json:"physicalLocation"`
	}
	type sarifResult struct {
		RuleID     string                 `json:"ruleId"`
		Level      string                 `json:"level"`
		Message    sarifMsg               `json:"message"`
		Locations  []sarifLoc             `json:"locations"`
		Properties map[string]interface{} `json:"properties,omitempty"`
	}
	type sarifRule struct {
		ID               string                 `json:"id"`
		Name             string                 `json:"name"`
		ShortDescription sarifMsg               `json:"shortDescription"`
		FullDescription  sarifMsg               `json:"fullDescription"`
		HelpURI          string                 `json:"helpUri,omitempty"`
		Properties       map[string]interface{} `json:"properties,omitempty"`
	}
	type sarifDriver struct {
		Name           string      `json:"name"`
		Version        string      `json:"version"`
		InformationURI string      `json:"informationUri"`
		Rules          []sarifRule `json:"rules"`
	}
	type sarifTool struct {
		Driver sarifDriver `json:"driver"`
	}
	type sarifRun struct {
		Tool    sarifTool     `json:"tool"`
		Results []sarifResult `json:"results"`
	}
	type sarifDoc struct {
		Schema  string     `json:"$schema"`
		Version string     `json:"version"`
		Runs    []sarifRun `json:"runs"`
	}
	levelMap := map[string]string{"critical": "error", "high": "error", "medium": "warning", "low": "note", "info": "note"}
	rules := []sarifRule{}
	results := []sarifResult{}
	seen := map[string]bool{}
	for _, f := range r.Findings {
		rid := f.Category
		if rid == "" {
			rid = "generic"
		}
		if !seen[rid] {
			seen[rid] = true
			rules = append(rules, sarifRule{
				ID: rid, Name: rid,
				ShortDescription: sarifMsg{Text: rid},
				FullDescription:  sarifMsg{Text: f.Description},
				HelpURI:          "https://owlsec.org",
				Properties:       map[string]interface{}{"tags": f.Mitre},
			})
		}
		uri := f.Location
		if uri == "" {
			uri = r.Subject
		}
		lvl := levelMap[strings.ToLower(f.Severity)]
		if lvl == "" {
			lvl = "warning"
		}
		results = append(results, sarifResult{
			RuleID: rid, Level: lvl,
			Message:   sarifMsg{Text: fmt.Sprintf("%s: %s", f.Title, f.Evidence)},
			Locations: []sarifLoc{{PhysicalLocation: physLoc{ArtifactLocation: artLoc{URI: uri}}}},
			Properties: map[string]interface{}{
				"mitre": f.Mitre, "grade_score": f.Score, "tags": f.Tags,
			},
		})
	}
	doc := sarifDoc{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{{
			Tool: sarifTool{Driver: sarifDriver{
				Name: APPNAME, Version: VERSION, InformationURI: "https://owlsec.org", Rules: rules,
			}},
			Results: results,
		}},
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	return string(data)
}

func exportReport(r *Report, format, outPath string) error {
	r.finalize()
	var body string
	switch strings.ToLower(format) {
	case "txt", "text":
		body = exportTXT(r)
	case "md", "markdown":
		body = exportMD(r)
	case "json":
		body = exportJSON(r)
	case "csv":
		body = exportCSV(r)
	case "sarif":
		body = exportSARIF(r)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
	if outPath == "" || outPath == "-" {
		fmt.Print(body)
		return nil
	}
	return os.WriteFile(outPath, []byte(body), 0644)
}

func printReportSummary(r *Report) {
	r.finalize()
	okMsg(fmt.Sprintf("Grade %s (%.1f/100) — %d findings", r.Grade, r.Score, len(r.Findings)))
	rows := [][]string{}
	for _, f := range r.Findings {
		mitre := ""
		if len(f.Mitre) > 0 {
			mitre = f.Mitre[0]
		}
		rows = append(rows, []string{f.ID, f.Severity, truncate(f.Title, 40), mitre})
		if len(rows) >= 25 {
			break
		}
	}
	if len(rows) > 0 {
		table([]string{"ID", "Sev", "Title", "MITRE"}, rows)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func sha256hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func sha1hex(b []byte) string {
	h := sha1.Sum(b)
	return hex.EncodeToString(h[:])
}
func md5hex(b []byte) string {
	h := md5.Sum(b)
	return hex.EncodeToString(h[:])
}

func entropyOf(b []byte) float64 {
	if len(b) == 0 {
		return 0
	}
	var freq [256]float64
	for _, c := range b {
		freq[c]++
	}
	n := float64(len(b))
	ent := 0.0
	for _, f := range freq {
		if f == 0 {
			continue
		}
		p := f / n
		ent -= p * math.Log2(p)
	}
	return ent
}

func readFileLimited(path string, max int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if max <= 0 {
		max = 64 << 20
	}
	return io.ReadAll(io.LimitReader(f, max))
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func mustAtoi(s string, def int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func pct(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return 100.0 * float64(n) / float64(d)
}

func nowStamp() string { return time.Now().Format("20060102-150405") }

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}

func u16le(b []byte, off int) uint16 {
	if off+2 > len(b) {
		return 0
	}
	return binary.LittleEndian.Uint16(b[off:])
}
func u32le(b []byte, off int) uint32 {
	if off+4 > len(b) {
		return 0
	}
	return binary.LittleEndian.Uint32(b[off:])
}
func u64le(b []byte, off int) uint64 {
	if off+8 > len(b) {
		return 0
	}
	return binary.LittleEndian.Uint64(b[off:])
}
func i64le(b []byte, off int) int64 {
	return int64(u64le(b, off))
}

func utf16leToString(b []byte) string {
	if len(b)%2 != 0 {
		b = b[:len(b)-1]
	}
	u := make([]uint16, len(b)/2)
	for i := 0; i < len(u); i++ {
		u[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	// decode until null
	n := 0
	for n < len(u) && u[n] != 0 {
		n++
	}
	runes := make([]rune, 0, n)
	for i := 0; i < n; {
		r := rune(u[i])
		if r >= 0xD800 && r <= 0xDBFF && i+1 < n {
			r2 := rune(u[i+1])
			if r2 >= 0xDC00 && r2 <= 0xDFFF {
				r = 0x10000 + ((r - 0xD800) << 10) + (r2 - 0xDC00)
				i += 2
				runes = append(runes, r)
				continue
			}
		}
		runes = append(runes, r)
		i++
	}
	return string(runes)
}

func filetimeToTime(ft uint64) time.Time {
	// Windows FILETIME: 100-ns intervals since 1601-01-01
	if ft == 0 {
		return time.Time{}
	}
	unix := int64(ft/10) - 11644473600000000
	return time.Unix(0, unix*1000).UTC()
}

func writeUsageHeader(w io.Writer, extra string) {
	fmt.Fprintf(w, "%s — %s\n", APPNAME, TAGLINE)
	fmt.Fprintf(w, "Author: khaled.s.haddad  |  https://owlsec.org  |  v%s\n\n", VERSION)
	fmt.Fprintf(w, "USAGE:\n  %s [options]\n  %s --demo\n  %s  (no args → interactive TUI)\n\n", APPNAME, APPNAME, APPNAME)
	fmt.Fprintf(w, "COMMON OPTIONS:\n")
	fmt.Fprintf(w, "  -h, --help            Show this help\n")
	fmt.Fprintf(w, "  --version             Print version\n")
	fmt.Fprintf(w, "  --demo                Run self-contained demo with synthetic data\n")
	fmt.Fprintf(w, "  --format FMT          Output format: txt|md|json|csv|sarif (default txt)\n")
	fmt.Fprintf(w, "  --output FILE         Write report to FILE (default stdout)\n")
	fmt.Fprintf(w, "  --no-color            Disable ANSI colors\n")
	if extra != "" {
		fmt.Fprintf(w, "\nTOOL OPTIONS:\n%s\n", extra)
	}
	fmt.Fprintf(w, "\nEXAMPLES:\n")
	fmt.Fprintf(w, "  %s --forge all --out-dir ./canaries\n", APPNAME)
	fmt.Fprintf(w, "  %s --forge aws-akid,slack,github-pat\n", APPNAME)
	fmt.Fprintf(w, "  %s --analyze access.log --manifest ./canaries/manifest.json\n", APPNAME)
	fmt.Fprintf(w, "  %s --analyze audit.log --format sarif --output triggers.sarif\n", APPNAME)
	fmt.Fprintf(w, "  %s --demo\n\n", APPNAME)
	fmt.Fprintf(w, "EXPORTS: TXT, Markdown, JSON, CSV, SARIF 2.1.0\n")
	fmt.Fprintf(w, "MITRE ATT&CK tags included on findings where applicable.\n")
	fmt.Fprintf(w, "Grading scale: A+ … F based on finding severity weights.\n")
}

// ---- Shared depth: MITRE ATT&CK technique database (subset) ----
type AttckTech struct {
	ID, Tactic, Name, Description string
}

var attckDB = []AttckTech{
	{"T1001", "Command and Control", "Data Obfuscation", "Obfuscate C2 traffic"},
	{"T1003", "Credential Access", "OS Credential Dumping", "Dump credentials from OS"},
	{"T1005", "Collection", "Data from Local System", "Collect local data"},
	{"T1027", "Defense Evasion", "Obfuscated Files or Information", "Obfuscate files/info"},
	{"T1036", "Defense Evasion", "Masquerading", "Masquerade as legitimate"},
	{"T1041", "Exfiltration", "Exfiltration Over C2 Channel", "Exfil over C2"},
	{"T1046", "Discovery", "Network Service Discovery", "Scan network services"},
	{"T1048", "Exfiltration", "Exfiltration Over Alternative Protocol", "Alt protocol exfil"},
	{"T1055", "Defense Evasion", "Process Injection", "Inject into processes"},
	{"T1059", "Execution", "Command and Scripting Interpreter", "Script interpreters"},
	{"T1059.001", "Execution", "PowerShell", "PowerShell execution"},
	{"T1059.003", "Execution", "Windows Command Shell", "cmd.exe"},
	{"T1059.005", "Execution", "Visual Basic", "VBScript/VBA"},
	{"T1059.007", "Execution", "JavaScript", "JS execution"},
	{"T1068", "Privilege Escalation", "Exploitation for Privilege Escalation", "PrivEsc exploit"},
	{"T1070", "Defense Evasion", "Indicator Removal", "Remove indicators"},
	{"T1070.004", "Defense Evasion", "File Deletion", "Delete files"},
	{"T1071", "Command and Control", "Application Layer Protocol", "App-layer C2"},
	{"T1082", "Discovery", "System Information Discovery", "System info discovery"},
	{"T1083", "Discovery", "File and Directory Discovery", "File discovery"},
	{"T1105", "Command and Control", "Ingress Tool Transfer", "Transfer tools in"},
	{"T1112", "Defense Evasion", "Modify Registry", "Registry modification"},
	{"T1129", "Execution", "Shared Modules", "Load shared modules"},
	{"T1140", "Defense Evasion", "Deobfuscate/Decode Files or Information", "Decode files"},
	{"T1204", "Execution", "User Execution", "User opens malicious file"},
	{"T1204.002", "Execution", "Malicious File", "User opens malicious file"},
	{"T1218", "Defense Evasion", "System Binary Proxy Execution", "Lolbins"},
	{"T1486", "Impact", "Data Encrypted for Impact", "Ransomware encryption"},
	{"T1543", "Persistence", "Create or Modify System Process", "Service/daemon persist"},
	{"T1547", "Persistence", "Boot or Logon Autostart Execution", "Autostart"},
	{"T1547.001", "Persistence", "Registry Run Keys / Startup Folder", "Run keys"},
	{"T1552", "Credential Access", "Unsecured Credentials", "Find credentials"},
	{"T1553", "Defense Evasion", "Subvert Trust Controls", "Trust subversion"},
	{"T1564", "Defense Evasion", "Hide Artifacts", "Hide artifacts"},
	{"T1566", "Initial Access", "Phishing", "Phishing"},
	{"T1566.001", "Initial Access", "Spearphishing Attachment", "Malicious attachment"},
	{"T1574", "Persistence", "Hijack Execution Flow", "Hijack flow"},
}

func attckLookup(id string) *AttckTech {
	for i := range attckDB {
		if attckDB[i].ID == id {
			return &attckDB[i]
		}
	}
	return nil
}

func attckRows() [][]string {
	rows := [][]string{}
	for _, t := range attckDB {
		rows = append(rows, []string{t.ID, t.Tactic, t.Name})
	}
	return rows
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

func containsFold(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

func indexByte(b []byte, c byte) int {
	for i, x := range b {
		if x == c {
			return i
		}
	}
	return -1
}

func splitLines(s string) []string {
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func absInt64(a int64) int64 {
	if a < 0 {
		return -a
	}
	return a
}

// ============================================================
//  Honeytoken / canary forge — authentic decoy credentials,
//  deployable files + manifest, and manifest-correlated alerts.
//  All values are SYNTHETIC decoys pointing at example infra.
// ============================================================

type HoneyToken struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Value       string `json:"value"`
	TrackURL    string `json:"track_url"`
	Fingerprint string `json:"fingerprint"`
	File        string `json:"file,omitempty"`
	Notes       string `json:"notes"`
	Created     string `json:"created"`
}

type Manifest struct {
	Tool    string       `json:"tool"`
	Version string       `json:"version"`
	Created string       `json:"created"`
	Tokens  []HoneyToken `json:"tokens"`
}

const canaryDomain = "canary.owlsec.example"

var tokenKinds = []string{
	"aws-akid", "aws-secret", "aws-creds", "slack", "github-pat",
	"stripe", "gcp-key", "api-bearer", "db-conn", "url", "dns", "doc",
}

func base32Upper(n int) string {
	const a = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	b := randomBytes(n)
	out := make([]byte, n)
	for i := range out {
		out[i] = a[int(b[i])%len(a)]
	}
	return string(out)
}

func base62(n int) string {
	const a = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := randomBytes(n)
	out := make([]byte, n)
	for i := range out {
		out[i] = a[int(b[i])%len(a)]
	}
	return string(out)
}

func base64ish(n int) string {
	const alpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	b := randomBytes(n)
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = alpha[int(b[i])%len(alpha)]
	}
	return string(out)
}

// forgeToken builds one synthetic canary of the requested kind.
func forgeToken(kind string) HoneyToken {
	id := randomHex(8)
	track := fmt.Sprintf("https://%s/t/%s", canaryDomain, id)
	val := ""
	switch strings.ToLower(kind) {
	case "aws-akid", "aws":
		val = "AKIA" + base32Upper(16) // 20 chars, AWS access-key-id shape
	case "aws-secret":
		val = base64ish(40) // 40-char secret shape
	case "aws-creds":
		val = fmt.Sprintf("[default]\naws_access_key_id = AKIA%s\naws_secret_access_key = %s\nregion = us-east-1",
			base32Upper(16), base64ish(40))
	case "slack":
		val = fmt.Sprintf("xoxb-%s-%s-%s", randDigits(12), randDigits(13), base62(24))
	case "github-pat":
		val = "ghp_" + base62(36)
	case "stripe":
		val = "sk_live_" + base62(24)
	case "gcp-key":
		val = fmt.Sprintf(`{"type":"service_account","project_id":"owlsec-canary","private_key_id":"%s","client_email":"canary@%s"}`,
			randomHex(20), canaryDomain)
	case "api-bearer":
		val = "Bearer " + base62(40)
	case "db-conn":
		val = fmt.Sprintf("postgres://canary:%s@db.%s:5432/app", base62(12), canaryDomain)
	case "url", "url-http":
		val = track
	case "dns":
		val = fmt.Sprintf("%s.%s", id, canaryDomain)
	case "doc", "doc-docx", "doc-pdf", "doc-xlsx":
		val = fmt.Sprintf("OWL-CANARY-%s (beacon: %s)", id, track)
	default:
		val = "OWLTOKEN-" + id
	}
	fp := hmacFingerprint(id, val)
	return HoneyToken{
		ID: id, Kind: kind, Value: val, TrackURL: track, Fingerprint: fp,
		Notes:   "synthetic canary — decoy only, do not use on real infrastructure",
		Created: time.Now().UTC().Format(time.RFC3339),
	}
}

func randDigits(n int) string {
	b := randomBytes(n)
	out := make([]byte, n)
	for i := range out {
		out[i] = '0' + b[i]%10
	}
	return string(out)
}

// hmacFingerprint is a short stable tag used to correlate a trigger log line
// back to a specific issued token in the manifest.
func hmacFingerprint(key, msg string) string {
	m := hmac.New(sha256.New, []byte(key))
	m.Write([]byte(msg))
	return hex.EncodeToString(m.Sum(nil))[:12]
}

// forgeSet issues one token per requested kind.
func forgeSet(kinds []string) []HoneyToken {
	var out []HoneyToken
	for _, k := range kinds {
		out = append(out, forgeToken(k))
	}
	return out
}

// fileContent renders the on-disk decoy for a token kind.
func fileContent(t HoneyToken) (string, string) {
	switch strings.ToLower(t.Kind) {
	case "aws-creds", "aws-akid", "aws", "aws-secret":
		return "credentials", t.Value + "\n"
	case "gcp-key":
		return t.ID + "-sa.json", t.Value + "\n"
	case "db-conn":
		return ".env", "DATABASE_URL=" + t.Value + "\n"
	case "url", "dns", "doc":
		return "README-backup.txt", "Backup endpoint: " + t.Value + "\n"
	default:
		return "token-" + t.Kind + "-" + t.ID + ".txt", t.Value + "\n"
	}
}

// writeCanaries forges and writes deployable decoy files + a manifest.
func writeCanaries(outDir string, kinds []string) (*Manifest, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	man := &Manifest{Tool: APPNAME, Version: VERSION, Created: time.Now().UTC().Format(time.RFC3339)}
	for _, k := range kinds {
		t := forgeToken(k)
		base, content := fileContent(t)
		fp := filepath.Join(outDir, fmt.Sprintf("%s-%s", t.ID, base))
		if err := os.WriteFile(fp, []byte(content), 0o600); err != nil {
			return nil, err
		}
		t.File = fp
		man.Tokens = append(man.Tokens, t)
	}
	mb, _ := json.MarshalIndent(man, "", "  ")
	if err := os.WriteFile(filepath.Join(outDir, "manifest.json"), mb, 0o600); err != nil {
		return nil, err
	}
	readme := "OWL Honeytoken canaries — SYNTHETIC DECOYS ONLY.\n" +
		"Place these files where an intruder would find them (home dirs, repos, shares).\n" +
		"Any read/use of a value here is suspicious. Correlate hits against manifest.json\n" +
		"with:  " + APPNAME + " --analyze <log> --manifest manifest.json\n"
	_ = os.WriteFile(filepath.Join(outDir, "README.txt"), []byte(readme), 0o644)
	return man, nil
}

func loadManifest(path string) (*Manifest, error) {
	b, err := readFileLimited(path, 8<<20)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// analyzeLog scans a log/text file for honeytoken triggers. With a manifest,
// an exact token/ID/track match is a confirmed CRITICAL alert; otherwise
// heuristic canary patterns raise HIGH findings.
func analyzeLog(r *Report, data []byte, man *Manifest) {
	lines := strings.Split(string(data), "\n")
	hits := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if man != nil {
			for _, t := range man.Tokens {
				if (t.ID != "" && strings.Contains(line, t.ID)) ||
					(len(t.Value) >= 8 && strings.Contains(line, firstLine(t.Value))) ||
					strings.Contains(line, t.TrackURL) {
					hits++
					r.add(Finding{Severity: "critical", Title: "Confirmed honeytoken trigger",
						Evidence: fmt.Sprintf("%s (%s) seen: %s", t.ID, t.Kind, truncate(strings.TrimSpace(line), 100)),
						Category: "trigger", Mitre: []string{"T1552", "T1530"}, Location: "log",
						Tags: []string{t.Kind}, Extra: map[string]string{"token_id": t.ID, "track": t.TrackURL}})
					break
				}
			}
		}
		for _, f := range heuristicTrigger(line) {
			hits++
			r.add(f)
		}
	}
	if hits == 0 {
		r.add(Finding{Severity: "info", Title: "No honeytoken triggers detected", Category: "trigger", Mitre: []string{"T1552"}, Location: "log"})
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func heuristicTrigger(line string) []Finding {
	needles := []struct{ pat, title string }{
		{canaryDomain, "Canary domain contacted"},
		{"OWLTOKEN-", "OWL token string observed"},
		{"OWL-CANARY-", "OWL canary document marker observed"},
	}
	var out []Finding
	for _, n := range needles {
		if strings.Contains(line, n.pat) {
			out = append(out, Finding{Severity: "high", Title: n.title,
				Evidence: truncate(strings.TrimSpace(line), 110), Category: "trigger",
				Mitre: []string{"T1552", "T1530"}, Location: "log"})
		}
	}
	return out
}

func placementPlan(root string, kinds []string) []Finding {
	spots := []string{
		"~/.aws/credentials", "~/.ssh/known-backup", "/srv/shares/finance/",
		"C:\\Users\\Public\\Documents\\", "repo/.env", "/var/backups/",
	}
	out := []Finding{}
	for i, k := range kinds {
		t := forgeToken(k)
		spot := spots[i%len(spots)]
		out = append(out, Finding{Severity: "info", Title: "Placement suggestion: " + k,
			Description: "deploy near " + spot, Evidence: truncate(firstLine(t.Value), 60),
			Category: "placement", Mitre: []string{"T1552"}, Location: spot, Tags: []string{k},
			Extra: map[string]string{"track": t.TrackURL}})
	}
	return out
}

func runDemo() *Report {
	r := newReport("demo")
	kinds := []string{"aws-akid", "aws-secret", "slack", "github-pat", "stripe", "gcp-key", "db-conn", "url", "dns"}
	for _, t := range forgeSet(kinds) {
		r.add(Finding{Severity: "info", Title: "Forged " + t.Kind, Evidence: truncate(firstLine(t.Value), 64),
			Category: "forge", Mitre: []string{"T1552"}, Location: t.TrackURL, Tags: []string{t.Kind},
			Extra: map[string]string{"fingerprint": t.Fingerprint}})
	}
	for _, f := range placementPlan("~", kinds) {
		r.add(f)
	}
	// Simulated trigger with a manifest match.
	man := &Manifest{Tokens: []HoneyToken{{ID: "deadbeef", Kind: "url", Value: "x", TrackURL: "https://" + canaryDomain + "/t/deadbeef"}}}
	analyzeLog(r, []byte("GET https://"+canaryDomain+"/t/deadbeef from 203.0.113.10\nnormal line\n"), man)
	r.finalize()
	return r
}

func runAnalyze(path, format, output string) int {
	return runAnalyzeM(path, "", format, output)
}

func runAnalyzeM(path, manifestPath, format, output string) int {
	r := newReport(path)
	var man *Manifest
	if manifestPath != "" {
		m, err := loadManifest(manifestPath)
		if err != nil {
			errMsg("manifest: " + err.Error())
			return 1
		}
		man = m
		r.Meta["manifest_tokens"] = strconv.Itoa(len(m.Tokens))
	}
	data, err := readFileLimited(path, 64<<20)
	if err != nil {
		errMsg(err.Error())
		return 1
	}
	analyzeLog(r, data, man)
	r.finalize()
	_ = exportReport(r, format, output)
	if output != "" || format == "txt" {
		printReportSummary(r)
	}
	return 0
}

func printHelp() {
	extra := `  --forge KINDS         Forge canary tokens (comma list, or "all")
  --out-dir DIR         Write deployable canary files + manifest.json to DIR
  --analyze LOG         Scan a log/text file for honeytoken triggers
  --manifest FILE       Correlate --analyze hits against this manifest.json
  --list-kinds          List supported token kinds
  --list-attck          Show ATT&CK techniques
  --selftest            Run internal self-tests

Kinds: ` + strings.Join(tokenKinds, ", ") + `
`
	writeUsageHeader(os.Stdout, extra)
}

func kindsFromArg(s string) []string {
	if s == "" || strings.EqualFold(s, "all") {
		return tokenKinds
	}
	var out []string
	for _, k := range strings.Split(s, ",") {
		if k = strings.TrimSpace(k); k != "" {
			out = append(out, k)
		}
	}
	return out
}

func runCLI(args []string) int {
	fs := flag.NewFlagSet(APPNAME, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var showHelp, showVersion, demo, noColorFlag, listAttck, listKinds, selfTest bool
	var format, output, forge, outDir, analyze, manifest string
	fs.BoolVar(&showHelp, "help", false, "")
	fs.BoolVar(&showHelp, "h", false, "")
	fs.BoolVar(&showVersion, "version", false, "")
	fs.BoolVar(&demo, "demo", false, "")
	fs.BoolVar(&noColorFlag, "no-color", false, "")
	fs.StringVar(&format, "format", "txt", "")
	fs.StringVar(&output, "output", "", "")
	fs.StringVar(&forge, "forge", "", "")
	fs.StringVar(&outDir, "out-dir", "", "")
	fs.StringVar(&analyze, "analyze", "", "")
	fs.StringVar(&analyze, "input", "", "")
	fs.StringVar(&analyze, "path", "", "")
	fs.StringVar(&manifest, "manifest", "", "")
	fs.BoolVar(&listAttck, "list-attck", false, "")
	fs.BoolVar(&listKinds, "list-kinds", false, "")
	fs.BoolVar(&selfTest, "selftest", false, "")
	if err := fs.Parse(args); err != nil {
		errMsg(err.Error())
		printHelp()
		return 2
	}
	if noColorFlag {
		noColor = true
	}
	switch {
	case showHelp:
		printHelp()
		return 0
	case showVersion:
		fmt.Printf("%s %s\n", APPNAME, VERSION)
		return 0
	case selfTest:
		return runSelfTest()
	case listKinds:
		for _, k := range tokenKinds {
			fmt.Println("  " + k)
		}
		return 0
	case listAttck:
		table([]string{"ID", "Tactic", "Name"}, attckRows())
		return 0
	case demo:
		r := runDemo()
		_ = exportReport(r, format, output)
		if output != "" || format == "txt" {
			printReportSummary(r)
		}
		return 0
	case analyze != "":
		return runAnalyzeM(analyze, manifest, format, output)
	case forge != "" || outDir != "":
		kinds := kindsFromArg(forge)
		if outDir != "" {
			man, err := writeCanaries(outDir, kinds)
			if err != nil {
				errMsg(err.Error())
				return 1
			}
			okMsg(fmt.Sprintf("Wrote %d canary files + manifest.json to %s", len(man.Tokens), outDir))
			table([]string{"Kind", "ID", "File"}, manifestRows(man))
			return 0
		}
		rows := [][]string{}
		for _, t := range forgeSet(kinds) {
			rows = append(rows, []string{t.Kind, truncate(firstLine(t.Value), 56), t.TrackURL})
		}
		table([]string{"Kind", "Value", "Track URL"}, rows)
		return 0
	}
	printHelp()
	return 2
}

func manifestRows(m *Manifest) [][]string {
	rows := [][]string{}
	for _, t := range m.Tokens {
		rows = append(rows, []string{t.Kind, t.ID, t.File})
	}
	return rows
}

func runTUI() {
	for {
		choice := menu("Main Menu", [][2]string{
			{"Forge tokens (preview)", "generate decoy tokens, show only"},
			{"Forge & write canaries", "write canary files + manifest.json"},
			{"Analyze trigger log", "scan a log for honeytoken use"},
			{"ATT&CK techniques", "techniques this detection supports"},
			{"Run demo", "forge + scan a synthetic log"},
		})
		switch choice {
		case "0", "00", "q", "Q", "":
			okMsg("Goodbye.")
			return
		case "h", "H", "help", "?":
			tuiHelp()
		case "1":
			k := prompt("kinds (comma list, blank=all)")
			rows := [][]string{}
			for _, t := range forgeSet(kindsFromArg(k)) {
				rows = append(rows, []string{t.Kind, truncate(firstLine(t.Value), 52), t.Fingerprint})
			}
			table([]string{"Kind", "Value", "Fingerprint"}, rows)
			pause()
		case "2":
			dir := prompt("output directory")
			if dir == "" {
				warnMsg("cancelled")
				pause()
				continue
			}
			man, err := writeCanaries(dir, tokenKinds)
			if err != nil {
				errMsg(err.Error())
			} else {
				okMsg(fmt.Sprintf("Wrote %d canaries + manifest.json", len(man.Tokens)))
				table([]string{"Kind", "ID", "File"}, manifestRows(man))
			}
			pause()
		case "3":
			p := prompt("log file")
			m := prompt("manifest.json (blank = heuristics only)")
			runAnalyzeM(p, m, "txt", "")
			pause()
		case "4":
			table([]string{"ID", "Tactic", "Name"}, attckRows())
			pause()
		case "5":
			printReportSummary(runDemo())
			pause()
		default:
			warnMsg("Invalid selection")
			pause()
		}
	}
}

func main() {
	if len(os.Args) <= 1 {
		runTUI()
		return
	}
	for _, a := range os.Args[1:] {
		if a == "--no-color" {
			noColor = true
		}
	}
	os.Exit(runCLI(os.Args[1:]))
}

// ============================================================
//  Self-test — token formats, manifest round-trip, correlation
// ============================================================

func runSelfTest() int {
	pass, fail := 0, 0
	check := func(name string, cond bool, detail string) {
		if cond {
			pass++
			okMsg(name)
		} else {
			fail++
			errMsg(name + " — " + detail)
		}
	}
	fmt.Println()
	line(66)
	fmt.Println("  " + col("⬡  OHoneyToken self-test", B+CY))
	line(66)
	fmt.Println()

	// Token formats.
	ak := forgeToken("aws-akid")
	check("AWS AKID starts with AKIA", strings.HasPrefix(ak.Value, "AKIA"), ak.Value)
	check("AWS AKID length 20", len(ak.Value) == 20, strconv.Itoa(len(ak.Value)))
	sec := forgeToken("aws-secret")
	check("AWS secret length 40", len(sec.Value) == 40, strconv.Itoa(len(sec.Value)))
	gh := forgeToken("github-pat")
	check("GitHub PAT ghp_ + 36", strings.HasPrefix(gh.Value, "ghp_") && len(gh.Value) == 40, gh.Value)
	st := forgeToken("stripe")
	check("Stripe sk_live_ prefix", strings.HasPrefix(st.Value, "sk_live_"), st.Value)
	sl := forgeToken("slack")
	check("Slack xoxb- prefix", strings.HasPrefix(sl.Value, "xoxb-"), sl.Value)
	dns := forgeToken("dns")
	check("DNS token under canary domain", strings.HasSuffix(dns.Value, canaryDomain), dns.Value)

	// Uniqueness + fingerprint.
	a, b := forgeToken("api-bearer"), forgeToken("api-bearer")
	check("tokens are unique", a.Value != b.Value && a.ID != b.ID, "")
	check("fingerprint is stable", hmacFingerprint(a.ID, a.Value) == a.Fingerprint, a.Fingerprint)

	// Manifest write + reload round-trip.
	dir, _ := os.MkdirTemp("", "ohoney")
	defer os.RemoveAll(dir)
	man, err := writeCanaries(dir, []string{"aws-creds", "db-conn", "url"})
	check("writeCanaries succeeds", err == nil && man != nil, errStr(err))
	if man != nil {
		check("3 tokens written", len(man.Tokens) == 3, strconv.Itoa(len(man.Tokens)))
		for _, t := range man.Tokens {
			if _, e := os.Stat(t.File); e != nil {
				check("canary file exists: "+t.Kind, false, t.File)
			}
		}
		check("all canary files on disk", allFilesExist(man), "")
		rl, e := loadManifest(dir + "/manifest.json")
		check("manifest reloads", e == nil && rl != nil && len(rl.Tokens) == 3, errStr(e))

		// Correlation: a log line containing an issued token id → CRITICAL.
		r := newReport("t")
		analyzeLog(r, []byte("GET "+man.Tokens[2].TrackURL+" 1.2.3.4\n"), rl)
		got := false
		for _, f := range r.Findings {
			if f.Severity == "critical" && f.Title == "Confirmed honeytoken trigger" {
				got = true
			}
		}
		check("manifest correlation fires CRITICAL", got, "")

		// Clean log → no trigger.
		r2 := newReport("t")
		analyzeLog(r2, []byte("just a normal boring log line\n"), rl)
		clean := true
		for _, f := range r2.Findings {
			if f.Category == "trigger" && f.Severity != "info" {
				clean = false
			}
		}
		check("clean log → no false trigger", clean, "")
	}

	// Heuristic (no manifest).
	r3 := newReport("t")
	analyzeLog(r3, []byte("curl https://"+canaryDomain+"/t/abc\n"), nil)
	h := false
	for _, f := range r3.Findings {
		if f.Severity == "high" {
			h = true
		}
	}
	check("heuristic domain hit → HIGH", h, "")

	fmt.Println()
	if fail == 0 {
		okMsg(fmt.Sprintf("All %d checks passed.", pass))
		return 0
	}
	errMsg(fmt.Sprintf("%d passed, %d failed.", pass, fail))
	return 1
}

func allFilesExist(m *Manifest) bool {
	for _, t := range m.Tokens {
		if _, e := os.Stat(t.File); e != nil {
			return false
		}
	}
	return true
}

func errStr(e error) string {
	if e == nil {
		return "nil"
	}
	return e.Error()
}

// ---------------------------------------------------------------------------
// Help & Usage  (interactive menu [H])
// Page 1: in-app guide (what the tool does, every menu option, tips).
// Page 2: the full command-line reference (same text as --help).
// ---------------------------------------------------------------------------

type owlHelpItem struct{ key, text string }

func owlHelpWrap(s string, width int) []string {
	var out []string
	for _, para := range strings.Split(s, "\n") {
		words := strings.Fields(para)
		if len(words) == 0 {
			out = append(out, "")
			continue
		}
		cur := ""
		for _, w := range words {
			if cur != "" && len([]rune(cur))+1+len([]rune(w)) > width {
				out = append(out, cur)
				cur = w
			} else if cur == "" {
				cur = w
			} else {
				cur += " " + w
			}
		}
		out = append(out, cur)
	}
	return out
}

func owlHelpTitle(t string) {
	fmt.Println()
	fmt.Println("  " + col("▸ "+t, B+CY))
}

func owlHelpPara(lines []string) {
	for _, l := range lines {
		for _, w := range owlHelpWrap(l, 60) {
			fmt.Println("    " + col(w, W))
		}
	}
}

func owlHelpRows(items []owlHelpItem) {
	kw := 0
	for _, it := range items {
		if n := len([]rune(it.key)); n > kw {
			kw = n
		}
	}
	for _, it := range items {
		k := "[" + it.key + "]" + strings.Repeat(" ", kw-len([]rune(it.key)))
		for i, w := range owlHelpWrap(it.text, 56-kw) {
			if i == 0 {
				fmt.Println("    " + col(k, YLW+B) + "  " + col(w, W))
			} else {
				fmt.Println("    " + strings.Repeat(" ", kw+2) + "  " + col(w, W))
			}
		}
	}
}

func owlHelpRule() { fmt.Println(col(strings.Repeat("─", 66), CY)) }

func tuiHelp() {
	clear()
	banner()
	owlHelpRule()
	fmt.Println("  " + col("⬡  Help & Usage — "+APPNAME, B+CY))
	owlHelpRule()
	owlHelpTitle("WHAT IT DOES")
	owlHelpPara(owlHelpAbout)
	owlHelpTitle("HOW TO USE THE MENU")
	owlHelpPara(owlHelpFlow)
	owlHelpTitle("MENU OPTIONS")
	owlHelpRows(owlHelpMenu)
	owlHelpTitle("TIPS")
	owlHelpRows(owlHelpTips)
	fmt.Println()
	owlHelpRule()
	fmt.Println("  " + col("Next page: command-line flags & copy-paste examples", DIM))
	pause()
	clear()
	owlHelpCLI()
	fmt.Println()
	fmt.Println("  " + col("Tip: the same reference is printed by  "+APPNAME+" --help", DIM))
	pause()
}

var owlHelpAbout = []string{
	"OHoneyToken forges honeytokens (canary credentials) and detects when they are used. A honeytoken is a fake secret you plant where an intruder would look — a decoy AWS key, Slack token, GitHub PAT, Stripe key, database connection string, URL, DNS name or document. It is never real, so any use of it is a reliable sign of intrusion.",
	"It generates realistic-looking tokens (correct prefixes and formats), writes deployable canary files with a manifest, and scans log or text files for triggers. With the manifest it tells you exactly which planted token fired and where it was placed.",
}

var owlHelpFlow = []string{
	"Type the number of an option and press Enter. Use [1] to preview tokens, [2] to write real canary files, and [3] to scan a log for triggers. [4] and [5] are reference / demo. [H] opens this help, [0] or q quits.",
	"Workflow: forge canaries with [2], plant the files where an attacker would snoop (config dirs, repos, shares, wikis), keep the manifest safe, then feed your access / audit logs to [3] to catch any use.",
}

var owlHelpMenu = []owlHelpItem{
	{"1", "Forge tokens (preview): enter the kinds as a comma list (blank = all). Shows the generated values and their fingerprints without writing any files."},
	{"2", "Forge & write canaries: enter an output directory. Writes one deployable canary per kind plus a manifest.json. Keep the manifest private; it links each token to where you planted it."},
	{"3", "Analyze trigger log: enter a log / text file and, optionally, the manifest.json. Reports which honeytokens appear in the log. With the manifest it names the exact token and its placement; without it, heuristics only."},
	{"4", "ATT&CK techniques: the MITRE techniques this detection supports."},
	{"5", "Run demo: forges a set, then scans a synthetic log that triggers some of them."},
}

var owlHelpTips = []owlHelpItem{
	{"*", "Kinds: " + strings.Join(tokenKinds, ", ") + "."},
	{"*", "A honeytoken only works if it is never used legitimately. Plant it, record it in the manifest, and leave it alone."},
	{"*", "Feed the logs of the system where you planted the tokens (cloud audit log, web server log, DNS log). A hit means someone found and tried your decoy."},
	{"*", "On the command line: --forge all --out-dir ./canaries, then --analyze access.log --manifest ./canaries/manifest.json."},
}

func owlHelpCLI() { printHelp() }
