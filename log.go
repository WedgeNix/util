package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// GetLogFile returns the standard logging file for outputting.
func GetLogFile() *os.File {
	return f
}

var (
	f         *os.File
	w         io.Writer
	gids      []uint64
	gidsState sync.Mutex
	logLock   sync.Mutex

	uniN = map[rune]string{
		'0': "₀",
		'1': "₁",
		'2': "₂",
		'3': "₃",
		'4': "₄",
		'5': "₅",
		'6': "₆",
		'7': "₇",
		'8': "₈",
		'9': "₉",
	}
)

// Log calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Log(v ...interface{}) {
	fnm := LANow().Format("Mon Jan 2, 2006 (3∶04 PM)") + ".log"
	if w == nil {
		f, _ = os.Create(fnm)
		w = io.MultiWriter(os.Stdout, f)
		// ready <- true
	}

	logLock.Lock()
	defer logLock.Unlock()
	t := LANow().Format("║Mon Jan 2, 3:04:05 PM║ ")
	tabCnt, gidCnt := tabs()
	tabs := strings.Repeat("│", tabCnt)
	num := strconv.Itoa(tabCnt + 1)
	for _, r := range num {
		num = strings.Replace(num, string(r), uniN[r], -1)
	}
	if tabCnt == 0 && gidCnt == 0 {
		num = ""
	}
	ln := tabs + num
	args := fmt.Sprint(v...)
	end := strings.Repeat("│", max(gidCnt-1-(len(ln)/3+len(args)), 0))
	fmt.Fprintln(w, t+ln+args+end)
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func exists() bool {
	gidsState.Lock()
	defer gidsState.Unlock()
	myGID := getGID()
	for _, gid := range gids {
		if gid == myGID {
			return true
		}
	}
	return false
}

// Visualization allows goroutine towers to display.
var Visualization bool

func tabs() (int, int) {
	if !Visualization {
		return 0, 0
	}

	gidsState.Lock()
	defer gidsState.Unlock()

	gid := getGID()

	index := -1
	for i, id := range gids {
		if id == gid {
			index = i
		}
	}

	if index != -1 {
		return index, len(gids)
	}

	gids = append(gids, gid)

	return len(gids) - 1, len(gids)
}

func getGID() uint64 {
	// find gid (CAUTION; anti-pattern)
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	gid, _ := strconv.ParseUint(string(b), 10, 64)
	return gid
}
