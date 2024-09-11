package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

var MemUse = struct {
    sync.Mutex
    inuse int
    avail int
    total int
    valid bool
    ready bool
}{sync.Mutex{}, 0, 0, 0, false, false}

func memoryWatch(done chan interface{}) {
    rgxMem := regexp.MustCompile(`(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+).*`)
    durSec := 0
    for {
        select {
        case _,ok := <-done:
            if !ok {
                return
            }

        case <-time.After(time.Duration(durSec)*time.Second):
            durSec = 2
            memUse := 0
            memTot := 0

            // Linux
            cmd := exec.Command("free", "-m")
            if out, err := cmd.StdoutPipe(); err == nil {
                if err := cmd.Start(); err == nil {
                    scan := bufio.NewScanner(out)
                    if scan.Scan() {	    // Header line
                        if scan.Scan() {	// Memory line
                            mems := rgxMem.FindStringSubmatch(scan.Text())
                            if len(mems) == 8 {
                                totl := StrDecToInt(mems[2])
                                free := StrDecToInt(mems[7])
                                if totl > 0 {
                                    pct := (free*100)/totl
                                    if pct < 5 {
										exit(nil, 20, "\nOut of memory you have %d of %d megabytes of memory remaining, exiting", free, totl)
                                    }
                                }
                                memUse = totl - free
                                memTot = totl					
                            }
                        }
                    }
                }
            }
            if memUse == 0 {
                // Mac
                cmd = exec.Command("ps", "-p", fmt.Sprintf("%d", os.Getpid()), "-o", "rss")
                if out, err := cmd.StdoutPipe(); err == nil {
                    if err := cmd.Start(); err == nil {
                        scan := bufio.NewScanner(out)
                        if scan.Scan() {	    // Header line
                            if scan.Scan() {	// Memory line
                                memUse = StrDecToInt(strings.TrimSpace(scan.Text())) / 1024		
                            }
                        }
                    }
                }
                cmd = exec.Command("sysctl",  "-n", "hw.memsize_usable")
                if out, err := cmd.StdoutPipe(); err == nil {
                    if err := cmd.Start(); err == nil {
                        scan := bufio.NewScanner(out)
                        if scan.Scan() {
                            line := scan.Text()
                            memTot = StrDecToInt(strings.TrimSpace(line))
                            memTot /= 1024 * 1024
                        }
                    }
                }
            }
            if memUse > 0 {
                MemUse.Lock()
                MemUse.ready = true
                MemUse.valid = true
                MemUse.avail = memTot - memUse
                MemUse.inuse = memUse
                MemUse.total = memTot
                MemUse.Unlock()
                continue
            }

            // Unable to run linux nor mac memory check
            MemUse.Lock()
            MemUse.ready = true
            MemUse.valid = false
            MemUse.Unlock()
            return
        }
    }
}
