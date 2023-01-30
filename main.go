package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type Chunk struct {
	offset  int64
	bufsize int64
}

type WorkerChunk struct {
	offset  int64
	bufsize int64
	chunks  []Chunk
}

type WordSequencesCount struct {
	words string
	count int
}

/*
Function to concurrently read a file by chunks using some number of workers and then send chunks to channels
(each worker has its own channel) to process them.
There are functions (CalcChunkSizeForWorker and SplitToChunk) that can split a file into chunks depending on the number of workers before starting the reading.
Also they split each chunk into smaller chunks depending on the bufferSize specified in our config.
This is step #1 of our pipeline to process a file.
Our pipeline for workers will have next steps:
Worker #1: [read chunk of file]->[procesing: split to this chunk to word sequences]->[calculate word sequences putting to a map #1]
Worker #2: [read chunk of file]->[procesing: split to this chunk to word sequences]->[calculate word sequences putting to a map #2]
Worker #N: [read chunk of file]->[procesing: split to this chunk to word sequences]->[calculate word sequences putting to a map #N]

Previously I used the regexp package to split text into words, but it is slower than using
strings.Fields and strings.Trim

Once file is read and processed all maps will be merged into one map to print the results.
*/
func ReadFromFile(f *os.File, workerChunkSizes []WorkerChunk, trimCutset string, wsn int) []map[string]int {
	workersNum := len(workerChunkSizes)
	var maps = make([]map[string]int, workersNum)
	var wg sync.WaitGroup

	wg.Add(workersNum)
	for i := 0; i < workersNum; i++ {
		chunkNum := len(workerChunkSizes[i].chunks)
		var m = make(map[string]int)

		go func(i int) {
			defer wg.Done()
			log.Println("Reading file: Srart Worker#", i)
			for j := 0; j < chunkNum; j++ {
				buffer := make([]byte, workerChunkSizes[i].chunks[j].bufsize)
				offset := workerChunkSizes[i].chunks[j].offset

				_, err := f.ReadAt(buffer, offset)
				if err != nil {
					fmt.Println(err)
					return
				}

				sWords := strings.Fields(string(buffer))
				var tmp []string

				for _, w := range sWords {
					w = strings.Trim(strings.ToLower(w), trimCutset)
					if w != "" {
						if len(tmp) < wsn-1 {
							tmp = append(tmp, w)
						} else {
							tmp = append(tmp, w)
							key := strings.Join(tmp, " ")
							m[key]++
							tmp = tmp[1:]
						}
					}
				}

			}
			maps[i] = m
			log.Println("Reading file: Stop Worker#", i)
		}(i)
	}

	wg.Wait()
	return maps
}

/*
This function helps to find the end of the chunk and the start point for the next one.
We want to make sure the chunk ends with a space. We don't want to cut words when splitting text into chunks.
And we want to make sure the next chunk begins with a two-word shift back for three word sequences.
(Formula: number of word sequences - 1). We need this words overlapping to concurrently calculate word sequences from chunks.
For example three word sequences:
File: word1 word2 word3 word4 word5 word6 word7
Our chunks will be:
Chunk #1: word1 word2 word3 word4
Chunk #2: word3 word4 word5 word6
Chunk #3: word5 word6 word7
*/
func FindChunkEnd(f *os.File, offset int64, wsn int) (startNext int64, endCurrect int64) {
	f.Seek(offset, 0)

	var sNext int64 = 0
	var eCurrect int64 = 0

	reader := bufio.NewReader(f)
	for i := 0; i < wsn; i++ {
		b, err := reader.ReadBytes(' ')
		if err != nil {
			if err == io.EOF {
				return offset + int64(len(b)), offset + int64(len(b))
			}
			log.Panicln(err)
		}
		if i == 0 {
			sNext = offset + int64(len(b))
		}
		eCurrect += int64(len(b))
	}

	return sNext, offset + eCurrect
}

/*
Function to split text into chunks.
We need to provide start and end points and the buffer size.
Based on the buffer size function will calculate the number of chunks and split text into this number of chunks.
*/
func SplitToChunk(f *os.File, bufSize int64, start int64, end int64, wsn int) []Chunk {
	var chankEnd int64 = 0
	var size int64 = 0
	var startNextChank int64 = 0
	var chunk Chunk
	var chunkSisez []Chunk

	for {
		estEnd := start + bufSize
		if estEnd >= end {
			chankEnd = end
		} else {
			startNextChank, chankEnd = FindChunkEnd(f, estEnd, wsn)
		}
		size = chankEnd - start
		chunk.bufsize = size
		chunk.offset = start
		chunkSisez = append(chunkSisez, chunk)
		start = startNextChank

		if chankEnd == end {
			break
		}
	}
	return chunkSisez
}

/*
The function creates a chunk for each worker and if this chunk is greater than the buffer size for file reading
(buffer Size in config) it splits this chunk into smaller chunks ≈ buffer size.
The result will be a list of chunks with a start point (a byte in a file) and with the number of bytes to read for each worker.
*/
func CalcChunkSizeForWorker(f *os.File, workNum int, bufSize int64, wsn int) []WorkerChunk {
	finfo, err := f.Stat()
	if err != nil {
		log.Panicln(err)
	}

	fileSize := finfo.Size()
	if fileSize <= bufSize {
		workNum = 1
	}

	workerChunkSizes := make([]WorkerChunk, workNum)
	estWrkReadSize := fileSize / int64(workNum)

	wchunks := SplitToChunk(f, estWrkReadSize, 0, fileSize, wsn)

	for i, chunk := range wchunks {
		workerChunkSizes[i].offset = chunk.offset
		workerChunkSizes[i].bufsize = chunk.bufsize
		if workerChunkSizes[i].bufsize > bufSize {
			workerChunkSizes[i].chunks = SplitToChunk(f, bufSize, workerChunkSizes[i].offset, workerChunkSizes[i].offset+workerChunkSizes[i].bufsize, wsn)
		} else {

			workerChunkSizes[i].chunks = append(workerChunkSizes[i].chunks, (Chunk{offset: workerChunkSizes[i].offset, bufsize: workerChunkSizes[i].bufsize}))
		}
	}

	return workerChunkSizes
}

/*
Function to read text from os.Stdin
Only 1 worker is used to process os.Stdin.
*/
func readFromStdin(file *os.File, trimCutset string, wsn int) []map[string]int {
	var maps = make([]map[string]int, 1)
	var m = make(map[string]int)

	reader := bufio.NewReader(file)
	for {
		b, err := reader.ReadBytes('\n') //reading line be line
		if err != nil {
			if err != io.EOF {
				log.Panicln(err)
			}
			break
		}

		if string(b) != "\n" {
			sWords := strings.Fields(string(b))
			var tmp []string

			//This code can be a separate function to use it here and in ReadFromFile to follow DRY principle.
			for _, w := range sWords {
				w = strings.Trim(strings.ToLower(w), trimCutset)
				if w != "" {
					if len(tmp) < wsn-1 {
						tmp = append(tmp, w)
					} else {
						tmp = append(tmp, w)
						key := strings.Join(tmp, " ")
						m[key]++
						tmp = tmp[1:]
					}
				}
			}
		}
	}

	maps[0] = m

	return maps
}

func main() {
	mres := make(map[string]int)
	var smap [][]map[string]int

	//Config
	var workersNum int
	//pJobNum := 3                                       //Number of jobs to process chunks for each worker
	var bufferSize int64 = 1 * 1024 * 1024             //a buffer size for file reading. For example 1MB
	trimCutset := ".,:;<>'!?\"”“‘’()-_{}[]*=+$%^&@#`~" //cutset for trim
	mcn := 100                                         //number of the most common word sequences to output.
	wsn := 3                                           //number of word sequences that need to be calculated. We can calculate 1,2,3,4,5,6 etc. word sequences

	/*
	   Calculate the number of workers to process one file.
	   We are using some numbers of workers (depending on CPU cores) to process one file concurrently
	   (split files to chunk and read and process them at the same time).

	*/
	if runtime.NumCPU() <= 2 {
		workersNum = 1
	} else {
		//workersNum = runtime.NumCPU() - 1
		workersNum = runtime.NumCPU()
	}

	//Stdin
	if len(os.Args[1:]) == 0 {
		smap = append(smap, readFromStdin(os.Stdin, trimCutset, wsn))
	}

	/*
		Files.
		At this moment this is the one-by-one file processing realization.
		Concurrency can also be used to process files at the same time, but we need to realize worker pool manager.
		The number of workers should be calculated depending on how many files should be proceed and how many CPU cores we have.
		But I don't think it will be faster than processing files one-by-one using all CPU cores for each file.
	*/
	if len(os.Args[1:]) > 0 {
		for _, arg := range os.Args[1:] {
			f, err := os.Open(arg)
			if err != nil {
				log.Panicln("Error:", err)
			}
			defer f.Close()

			workerChunkSizes := CalcChunkSizeForWorker(f, workersNum, bufferSize, wsn)
			smap = append(smap, ReadFromFile(f, workerChunkSizes, trimCutset, wsn))
		}
	}

	//Merge results to one map
	for _, sm := range smap {
		for _, m := range sm {
			for k, v := range m {
				mres[k] = mres[k] + v
			}
		}
	}

	//Add info from map to dict
	wsc := make([]WordSequencesCount, 0, len(mres))
	for k, v := range mres {
		wsc = append(wsc, WordSequencesCount{words: k, count: v})
	}

	//Sort dict
	sort.Slice(wsc, func(i, j int) bool {
		return wsc[i].count > wsc[j].count
	})

	//Print results
	for i := 0; i < len(wsc) && i < mcn; i++ {
		fmt.Println(wsc[i].words, "-", wsc[i].count)
	}
}
