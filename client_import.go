package main

// import (
// 	"bufio"
// 	"bytes"
// 	"io"
// 	"os"
// 	"strings"
// )

// func (sf *scrub_file) readSV() (chan map[string]string, error) {
//     count := func(path string) int {
//         if fd, err := os.Open(path); err == nil {
//             defer fd.Close()
//             buf := make([]byte, 32*1024)
//             count := 0
//             lineSep := []byte{'\n'}
//             var last byte
//             for {
//                 c, err := fd.Read(buf)
//                 count += bytes.Count(buf[:c], lineSep)
//                 if c != 0 {
//                     last = buf[c-1]
//                 }
//                 switch {
//                 case err == io.EOF:
//                     if last != lineSep[0] {
//                         count++
//                     }
//                     return count
//                 case err != nil:
//                     return count
//                 default:
//                 }
//             }
//         }
//         return 0
//     }
//     readl := func(br *bufio.Reader, csep string) ([]string, error) {
//         if line, _, err := br.ReadLine(); err == nil {
//             str := string(line)
//             if len(str) > 0 {
//                 toks := strings.Split(str, csep)
//                 vals := make([]string, 0, len(toks))
//                 for _, tok := range toks {
//                     tok  = strings.ReplaceAll(tok, " ",  "")
//                     tok  = strings.ReplaceAll(tok, "\t", "")
//                     vals = append(vals, tok)
//                 }
//                 return vals, nil
//             } else {
//                 return []string{}, nil
//             }
//         } else {
//             return nil, err
//         }
//     }
//     sf.lines = count(sf.path)
//     if fd, err := os.Open(sf.path); err == nil {
//         defer fd.Close()
//         br := bufio.NewReader(fd)
//         if sf.hdrs, err = readl(br, sf.csep); err != nil {
//             return nil, err
//         }
//         for i, hdr := range sf.hdrs {
//             if prop, ok := sf.hdrm[hdr];ok {	// If this hdr is mapped to another value
//                 sf.hdri[i] = prop               // then assume the other value is the 
//             } else {						    // proper value to use.
//                 sf.hdri[i] = hdr				// If no custom mapping, could still be a
//             }								    // custom column - just keep name as is.
//         }

//         // Now that we've gotten the header stuff stored in the scrub_file we can start reading rows in the background.
//         rows := make(chan map[string]string, 1000)
//         go func() {
//             // The rest of the rows are data rows.
//             for {
//                 if toks, err := readl(br, sf.csep); err == nil {
//                     row  := map[string]string{}
//                     for i, fld := range toks {
//                         if i < len(sf.hdri) {       // Make sure the data row doesn't have a column that exceeds the header row.
//                             row[sf.hdri[i]] = fld
//                         }
//                     }
//                     rows <-row
//                 } else if err == io.EOF {
//                     close(rows)
//                     break
//                 } else {
//                     sf.rderr = err
//                     close(rows)
//                     break
//                 }
//             }
//         }()
//         return rows, nil
//     } else {
//         return nil, err
//     }
// }
