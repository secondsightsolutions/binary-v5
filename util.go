package main

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

var prefDateFmts = map[string]any{}

var allDateFmts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05.000",
	"2006-01-02 15:04:05.000000",
	"2006-01-02 15:04:05.999",
	"2006-01-02 15:04:05.999999",
	"2006/01/02 15:04:05",
	"2006/01/02 15:04:05.000",
	"2006/01/02 15:04:05.000000",
	"2006/01/02 15:04:05.999",
	"2006/01/02 15:04:05.999999",

	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05.000",
	"2006-01-02T15:04:05.000000",
	"2006-01-02T15:04:05.999",
	"2006-01-02T15:04:05.999999",
	"2006/01/02T15:04:05",
	"2006/01/02T15:04:05.000",
	"2006/01/02T15:04:05.000000",
	"2006/01/02T15:04:05.999",
	"2006/01/02T15:04:05.999999",

	"2006-01-02 15:04:05 MST",
	"2006-01-02 15:04:05.000 MST",
	"2006-01-02 15:04:05.000000 MST",
	"2006-01-02 15:04:05.999 MST",
	"2006-01-02 15:04:05.999999 MST",
	"2006/01/02 15:04:05 MST",
	"2006/01/02 15:04:05.000 MST",
	"2006/01/02 15:04:05.000000 MST",
	"2006/01/02 15:04:05.999 MST",
	"2006/01/02 15:04:05.999999 MST",

	"2006-01-02T15:04:05 MST",
	"2006-01-02T15:04:05.000 MST",
	"2006-01-02T15:04:05.000000 MST",
	"2006-01-02T15:04:05.999 MST",
	"2006-01-02T15:04:05.999999 MST",
	"2006/01/02T15:04:05 MST",
	"2006/01/02T15:04:05.000 MST",
	"2006/01/02T15:04:05.000000 MST",
	"2006/01/02T15:04:05.999 MST",
	"2006/01/02T15:04:05.999999 MST",

	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05.000 -0700",
	"2006-01-02 15:04:05.000000 -0700",
	"2006-01-02 15:04:05.999 -0700",
	"2006-01-02 15:04:05.999999 -0700",
	"2006/01/02 15:04:05 -0700",
	"2006/01/02 15:04:05.000 -0700",
	"2006/01/02 15:04:05.000000 -0700",
	"2006/01/02 15:04:05.999 -0700",
	"2006/01/02 15:04:05.999999 -0700",

	"2006-01-02T15:04:05 -0700",
	"2006-01-02T15:04:05.000 -0700",
	"2006-01-02T15:04:05.000000 -0700",
	"2006-01-02T15:04:05.999 -0700",
	"2006-01-02T15:04:05.999999 -0700",
	"2006/01/02T15:04:05 -0700",
	"2006/01/02T15:04:05.000 -0700",
	"2006/01/02T15:04:05.000000 -0700",
	"2006/01/02T15:04:05.999 -0700",
	"2006/01/02T15:04:05.999999 -0700",

	"20060102",
	
	"2006-01-02",
	"2006-01-02T15:04",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04",
	"2006-01-02 15:04:05",
	"2006/01/02",
	"2006/01/02 15:04",
	"2006/01/02 15:04:05",

	"1/2/2006",

	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	time.Stamp,
	time.StampMicro,
	time.StampMilli,
	time.StampNano,
	"1/2/06",
}

func TryParseStrToUnix(val, useFmt string) (int64, string, error) {
	if useFmt != "" {
		if tm, err := time.Parse(useFmt, val); err == nil {
			prefDateFmts[useFmt] = nil
			return tm.UTC().Unix(), useFmt, nil
		}
	}
	for _fmt := range prefDateFmts {
		if tm, err := time.Parse(_fmt, val); err == nil {
			return tm.UTC().Unix(), _fmt, nil
		}
	}
	for _, _fmt := range allDateFmts {
		if tm, err := time.Parse(_fmt, val); err == nil {
			prefDateFmts[_fmt] = nil
			return tm.UTC().Unix(), _fmt, nil
		}
	}
	return 0, "", fmt.Errorf("not a date")
}

func ParseStrToTime(val string) *time.Time {
	if i64, err := strconv.ParseInt(val, 10, 64); err == nil {
		if tm := ParseI64ToTime(i64); tm != nil {
			return tm
		}
	}
	for _fmt := range prefDateFmts {
		if tm, err := time.Parse(_fmt, val); err == nil {
			return &tm
		}
	}
	for _, _fmt := range allDateFmts {
		if tm, err := time.Parse(_fmt, val); err == nil {
			prefDateFmts[_fmt] = nil
			return &tm
		}
	}
	return nil
}
func ParseI64ToTime(i64 int64) *time.Time {
	if i64 > int64(time.Nanosecond) {
		tm := time.Unix(i64 / int64(time.Nanosecond), i64 % int64(time.Nanosecond))
		return &tm
	} else if i64 > int64(time.Microsecond) {
		tm := time.Unix(i64 / int64(time.Microsecond), i64 % int64(time.Microsecond))
		return &tm
	} else if i64 > int64(time.Millisecond) {
		tm := time.Unix(i64 / int64(time.Millisecond), i64 % int64(time.Millisecond))
		return &tm
	} else {
		tm := time.Unix(i64 / int64(time.Second), i64 % int64(time.Second))
		return &tm
	}
}

func StrDecToInt(val string) int {
	if d, err := strconv.ParseInt(val, 10, 64); err == nil {
		return int(d)
	}
	return 0
}
func StrDecToInt64(val string) int64 {
	if d, err := strconv.ParseInt(val, 10, 64); err == nil {
		return d
	}
	return 0
}

var ScreenLevel = struct {
	None int
	Text int
	Bars int
	All  int
}{ 0, 1, 2, 3}
/*
// screen() displays progress on the screen based on the running identity plus parameters [mfr] and [prc] (brg always ScreenLevel.All).
// By putting the level information (mfr, proc) into the place of the call, the code here can decide whether to display, and how much.
// There are no conditional calls to screen(); the calling code always calls screen(). Make the decision here what to display (or not).
// mfr - if the caller is a manufacturer, this is the ScreenLevel to use
// prc - if the caller is a processor, this is the ScreenLevel to use
func screen(start time.Time, bar *int, cur, max int, mfr, prc int, nl bool, text string, args ...any) {
	// Rows that are "hidden" actually just have the N of M hidden. Still display the row, along with the continually updating times (all platforms) and memory usages (on linux).
	_brg := strings.EqualFold("brg", name)
	_mnu := strings.EqualFold(Type, "manu")
	_lvl := 0						// Start with ScreenLevel.None
	bars := 20						// The gauge will display N of 20 bars (max 20).
	if _brg {
		_lvl = ScreenLevel.All		// BRG always displays all.
	} else if _mnu {				// If we are a manufacturer...
		_lvl = mfr					// ...the screen level is whatever was passed in in the [mfr] param.
	} else {						// If we are a processor...
		_lvl = prc					// ...the screen level is whatever was passed in in the [prc] param.
	}
	if _lvl == ScreenLevel.None {	// If we're still ScreenLevel.None then nothing to show. Return.
		return
	}
	
	MemUse.Lock()
	valid := MemUse.valid
	ready := MemUse.ready
	inuse := MemUse.inuse
	avail := MemUse.avail
	total := MemUse.total
	used  := 0
	MemUse.Unlock()
	if total > 0 {
		used = 100 - (avail*100)/total
	}

	dur := time.Since(start)
	mil := dur.Milliseconds() % 1000
	sec := dur.Milliseconds() / 1000 % 60
	min := dur.Milliseconds() / 1000 / 60
	den := dur.Milliseconds()
	if den < 1000 {
		den = 1000
	}
	per := (int64(cur+1) * 1000) / den
	if cur < 0 {
		per = 0
	}
	cnt := 0
	if max > 0 {
		fcr := float32(cur+1)
		fmx := float32(max)
		pct := fcr/fmx
		cnt = int(float32(bars) * pct)
		if cnt == 0 {
			cnt = 1
		} else if cnt > bars {	// Should never happen.
			cnt = bars
		}

	} else if bar != nil {
		if nl {
			cnt = (*bar%bars)
		} else {
			cnt = (*bar%bars)+1
		}
		*bar++
	}
	fmt.Printf("\r%100s", " ")

	timS := fmt.Sprintf("%s (%02dm.%02ds.%03dms)", start.Format("2006-01-02 15:04:05"), min, sec, mil)
	memS := fmt.Sprintf(" (%5dM of %dM: %2d%% used)", inuse, total, used)
	perS := fmt.Sprintf(" (%6d/sec) ", per)
	barS := ""
	text = fmt.Sprintf(text, args...)
	txtS := fmt.Sprintf(" %-30s", text)
	curS := fmt.Sprintf(" %d", cur+1)
	maxS := fmt.Sprintf(" %d of %d", cur+1, max)
	prtS := "\r" + timS
	if !ready || !valid {
		memS = ""
	}
	if valid {
		prtS += memS
	}
	if bar != nil {
		barS = "[" + fmt.Sprintf("%-*s", bars, strings.Repeat("#", cnt))  + "]"
	}
	if cur >= 0 {
		if _lvl == ScreenLevel.All {
			prtS += perS
		}
		if _lvl >= ScreenLevel.Bars {
			prtS += barS
		}
	}
	if _lvl >= ScreenLevel.Text {
		prtS += txtS
	}
	if cur >= 0 {
		if _lvl == ScreenLevel.All {
			if max > 0 {
				prtS += maxS
			} else {
				prtS += curS
			}
		}
	}
	fmt.Print(prtS)
	if nl {
		fmt.Println()
	}
}
*/
/*
func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}
	var last byte

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)
		if c != 0 {
			last = buf[c-1]
		}
		switch {
		case err == io.EOF:
			if last != lineSep[0] {
				count++
			}
			return count, nil
		case err != nil:
			return count, err
		}
	}
}

func renderCols(hdrs []string, row map[string]string) string {
	var sb bytes.Buffer
	for i, hdr := range hdrs {
		sb.WriteString(row[hdr])
		if i < len(hdrs)-1 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}

func lastTok(str, sep string) string {
	toks := strings.Split(str, sep)
	return toks[len(toks)-1]
}
*/

func sleep(durn time.Duration, stop chan any) bool {
	select {
	case <-stop:
		return true
	case <-time.After(durn):
		if durn < time.Duration(32) * time.Second {
			durn *= 2
		}
		return false
	}
}

func findAny(srch string, list []string) string {
	for _, str := range list {
		if strings.EqualFold(str, srch) {
			return str
		}
	}
	return ""
}

func split(str, sep string, dflt string) []string {
	if str == "" {
		if dflt == "" {
			return []string{}
		}
		str = dflt
	}
	toks := strings.Split(str, sep)
	list := make([]string, 0, len(toks))
	for _, tok := range toks {
		tok  = strings.Trim(tok, " ")
		tok  = strings.Trim(tok, "\t")
		list = append(list, tok)
	}
	return list
}

func Log(app, fcn, tgt, msg string, dur time.Duration, vals map[string]any, err error, args ...any) {
	mesg := fmt.Sprintf(msg, args...)
	mil  := dur.Milliseconds() % 1000
	sec  := dur.Milliseconds() / 1000 % 60
	min  := dur.Milliseconds() / 1000 / 60
	curT := time.Now().Format("2006-01-02 15:04:05")
	durn := fmt.Sprintf("(%02dm.%02ds.%03dms)", min, sec, mil)
	errs := ""
	if err != nil {
		erm := err.Error()
		if strings.Contains(erm, "connection refused") {
			erm = "connection refused"
		} else if strings.Contains(erm, "not authorized") {
			erm = "not authorized"
		} else if strings.Contains(erm, "rpc error: code = Unknown desc = ") {
			toks := strings.Split(erm, "rpc error: code = Unknown desc = ")
			erm = toks[1]
		}
		errs = "::" + erm
	}
	if len(vals) > 0 {
		vs := []string{}
		for k := range vals {
			vs = append(vs, k)
		}
		svs := sort.StringSlice(vs)
		svs.Sort()
		str := ""
		if len(svs) > 0 {
			str = " ("
			cnt := 0
			for _, k := range svs {
				cnt++
				if cnt < len(svs) {
					str = fmt.Sprintf("%s%s=%v ", str, k, vals[k])
				} else {
					str = fmt.Sprintf("%s%s=%v", str, k, vals[k])
				}
			}
			str += ")"
		}
		mesg += str
	}
	fmt.Printf("%s [%-5s] %s %-15s %-26s %s %s\n", curT, app, durn, fcn, tgt, mesg, errs)
}

func getCreds(tlsInfo credentials.TLSInfo) (cn, ou, o string) {
	for _, chain := range tlsInfo.State.VerifiedChains {
		if len(chain) > 0 {
			cn = chain[0].Subject.CommonName
			ou = chain[0].Subject.OrganizationalUnit[0]
			o  = chain[0].Subject.Organization[0]
			cn = strings.ToLower(cn)
			ou = strings.ToLower(ou)
			o  = strings.ToLower(o)
			if len(cn) > 0 {
				break
			}
		}
	}
	return
}

func getPublicAddr(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok && p != nil {
		return p.Addr.String()
	}
	return ""
}
func getLocalAddr() string {
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return ""
}

func metaGet(ctx context.Context, key string) string {
	vals := metadata.ValueFromIncomingContext(ctx, key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}
func metaGetI64(ctx context.Context, key string) int64 {
	vals := metadata.ValueFromIncomingContext(ctx, key)
	if len(vals) > 0 {
		if i64, err := strconv.ParseInt(vals[0], 10, 64); err == nil {
			return i64
		}
	}
	return -1
}
func metaManu(ctx context.Context) string {
	_,_,_,manu,_,_ := getMetaGRPC(ctx)
	return manu
}
func metaValue(md metadata.MD, key string) string {
	if vals := md.Get(key); len(vals) > 0 {
		return vals[0]
	}
	return ""
}
func metaValueInt64(md metadata.MD, key string) int64 {
	if val := metaValue(md, key); val != "" {
		if iVal, err := strconv.ParseInt(val, 10, 64); err == nil {
			return iVal
		}
	}
	return 0
}