package main

import (
	"os"
	"bufio"
	"strconv"
	"time"
	"log"
)

func main() {
	f, err := os.OpenFile("D:\\temp\\test.txt", os.O_CREATE | os.O_RDWR, 0)
	if err != nil {
		log.Fatalln("Error:", err)
	}
	var i int
	for {
		w := bufio.NewWriter(f)
		w.WriteString(strconv.Itoa(i) + "\r\n")
		w.Flush()
		f.Sync()
		time.Sleep(1 * time.Second)
		i++
	}
}
