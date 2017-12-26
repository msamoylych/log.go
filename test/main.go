package main

import (
	"os"
	"bufio"
	"strconv"
	"time"
)

func main() {
	f, _ := os.OpenFile("D:\\temp\\test.txt", os.O_CREATE | os.O_RDWR, 0)
	var i int
	for {
		w := bufio.NewWriter(f)
		w.WriteString(strconv.Itoa(i) + "\r\n")
		w.Flush()
		f.Sync()
		time.Sleep(1 * time.Millisecond)
		i++
	}
}
