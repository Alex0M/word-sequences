## Version 2. Concurrency

This version implemented the approach to read a file by chunks concurrently using some number of workers and then send chunks to channels (each worker has its own channel) to process them.

Our pipeline for workers will have next steps:  
Worker #1: [read chunks of file] -> [procesing (can be 1,2,3,etc jobs): split chunks to word sequences] -> [calculate word sequences putting to a map #1]  
Worker #2: [read chunks of file] -> [procesing (can be 1,2,3,etc jobs): split chunks to word sequences] -> [calculate word sequences putting to a map #2]  
Worker #N: [read chunks of file] -> [procesing (can be 1,2,3,etc jobs): split chunks to word sequences] -> [calculate word sequences putting to a map #N]  

Once file is read and processed all maps will be merged (sum of all maps) into one map to print the results.

There are new functions (CalcChunkSizeForWorker and SplitToChunk) have beed added.
These functions can split a file into chunks depending on the number of workers before starting the reading.
They aslo split each chunk into smaller chunks depending on the bufferSize specified in our config.

Also was added function FindChunkEnd.
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


Previously I used the regexp package to split text into words, but it is slower than using
strings.Fields and strings.Trim  
Also using strings.Fields and strings.Trim to split into words is supporting unicode characters. Needs to test more deeper.

This approach is showing how to use channels. Using channels we can run all steps at the same time and also add additional jobs for some steps if these steps are taking longer than for example previous one.

Here is the link to the same approach with multiple workers but without channels.
Workers in doing all steps in our pipeline consecutively.

https://github.com/Alex0M/word-sequences/tree/workers



## Run the program

### Prerequisites

go 1.19 must be installed. Run `go version` to check.

### Clone gir repo and Build

```bash
#Clone repo
git clone https://github.com/Alex0M/word-sequences.git

#Build adn Run
cd ./word-sequences
go build -o word-sequences .
#files
./word-sequences <file paths>  #./word-sequences ./text-examples/moby-dick.txt

#stdin
cat ./text-examples/moby-dick.txt | ./word-sequences


#Or run using go run
go run main.go <file paths>  #./go run main.go ./text-examples/moby-dick.txt

cat ./text-examples/moby-dick.txt | go run main.go
```

## Docker

```bash
#Clone repo
git clone https://github.com/Alex0M/word-sequences.git

#Build adn Run
cd ./word-sequences
docker build -t <iamge name and tag> . #(e.g. docker build -t word-sequences:v0.0.1 .)
docker run -v /path/to/files/dir:/dir/in/container <image name> /dir/in/container/<file name> 
#(e.g.  docker run -v /tmp/word-sequences/text-examples:/tmp word-sequences:v0.0.1 /tmp/moby-dick.txt )
```
## How to run the tests.
Here is the helpful bash script to run one docker container per file in a directory and save the result in <file_name>.log file in the log directory.
This script will run one container for each file at a time. 
The script will create a logs directory automatically. 

```bash
#Run script
./processing-files-dir.sh [Docker image] [absolute path to local folder with files] [absolute path to local log folder]
#(e.g. ./processing-files-dir.sh word-sequences:v0.0.1 /tmp/word-sequences/text-example /tmp/word-sequences/logs)
```
Go to a log directory to checl result.

## What would you do next?

- Write unit tests
- Deeper testing support unicode characters.
- Create worker pool manager.
- Add config file to configure workersNum, bufferSize, trimCutset, etc.
