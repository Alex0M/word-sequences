package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

// dictionary to sort and print info about word sequences and their numbers
type WordSequencesCount struct {
	words string
	count int
}

// map to store info about word sequences and their numbers
var m = make(map[string]int)

// Function to split text into words using regular expression
func splitTextToWords(t string) []string {
	re := regexp.MustCompile(`(\b[^\s]+\b)`)
	words := re.FindAllString(t, -1)

	return words
}

// Function to count word sequences and add them to map
func countWordSequences(sWords []string, swfpl []string, wsn int) []string {
	s := append(swfpl, sWords...)

	if len(s) < wsn {
		swfpl = s
	} else {
		swfpl = s[(len(s)-wsn)+1:]
	}

	for i := 0; i <= len(s)-wsn; i++ {
		keyWord := strings.ToLower(strings.Join(s[i:i+wsn], " "))
		m[keyWord]++
	}
	return swfpl
}

// Function to read a file in chunks of some size, process them, and repeat until the end. We don't want to load a large file into memory.
func readFromFile(file *os.File, bufferSize int, wsn int) {
	var swfpl []string
	buffer := make([]byte, bufferSize)
	var totalReadBytes int64 = 0
	var readBytes int = 0

	for {
		b, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Panicln(err)
			}

			break
		}

		//make sure that the end of chunk is a space or '\n' if not then we need to read only until last space or last \n
		readBytes = b
		for i := b - 1; i > 1; i-- {
			if string(buffer[i]) == "\n" || string(buffer[i]) == " " {
				readBytes = i
				break
			}
		}

		sWords := splitTextToWords(string(buffer[:readBytes])) //split chunk to words
		swfpl = countWordSequences(sWords, swfpl, wsn)         //count word sequences

		totalReadBytes = totalReadBytes + int64(readBytes) //calculate how many bytes we already read in file
		_, errSeek := file.Seek(totalReadBytes, 0)         //set the offset for next reading
		if errSeek != nil {
			log.Panicln(err)
		}
	}
}

// Function to read text from os.Stdin
func readFromStdin(file *os.File, wsn int) {
	var swfpl []string

	reader := bufio.NewReader(file)
	for {
		b, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Panicln(err)
			}

			break
		}

		if string(b) != "\n" {
			sWords := splitTextToWords(string(b))
			swfpl = countWordSequences(sWords, swfpl, wsn)
		}
	}
}

func main() {
	wsc := make([]WordSequencesCount, 0, len(m))
	bufferSize := 1 * 1024 * 1024 //a buffer size for fole reading. For example 1MB
	wsn := 3                      //number of word sequences that need to be calculated. We can calucate 1,2,3,4,5 etc. word sequences
	mcn := 100                    //number of the most common word sequences to output.

	if len(os.Args[1:]) == 0 {
		readFromStdin(os.Stdin, wsn)
	}

	if len(os.Args[1:]) > 0 {
		for _, arg := range os.Args[1:] {
			f, err := os.Open(arg)
			if err != nil {
				log.Panicln("Error:", err)
			}
			defer f.Close()
			readFromFile(f, bufferSize, wsn)
		}
	}

	//add info from map to dict
	for k, v := range m {
		wsc = append(wsc, WordSequencesCount{words: k, count: v})
	}

	//sort dict
	sort.Slice(wsc, func(i, j int) bool {
		return wsc[i].count > wsc[j].count
	})

	//print
	for i := 0; i < len(wsc) && i < mcn; i++ {
		fmt.Println(wsc[i].words, "-", wsc[i].count)
	}
}
