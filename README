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

- Reading file chunks concurrently. We can use multiple go routines to read file and it will help to speed up the processing of the chunks.
- Using multiple go routines to processing list of files. It can be one go-routine per file for example.
- Support unicode characters.