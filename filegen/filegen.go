package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

const (
	charset_start     = byte('!') // 0x21
	charset_stop      = byte('~') // 0x7e
	charset_size      = int(charset_stop - charset_start + 1)
	fragments         = 8  //12
	max_fragment_size = 12 //8
)

var charset [charset_size]byte

func init() {
	// Initialize charset
	for i := 0; i < charset_size; i++ {
		charset[i] = charset_start + byte(i)
	}
}

/*
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
*/

/*
interface lineGenerator {
	line() string
}

struct simpleLineGenerator {
	random *rand.Rand
}
*/

// Returns line using the specified random generator.
func line(random *rand.Rand) []byte {
	buf := make([]byte, 0, fragments*max_fragment_size)
	for f := 0; f < fragments; f++ {
		fragment_size := random.Intn(max_fragment_size + 1)
		b := charset[random.Intn(charset_size)]
		for i := 0; i < fragment_size; i++ {
			buf = append(buf, b)
		}
	}
	return buf
}

// FIXME: doc
func generate(output chan<- []byte, shutdown <-chan bool, seed int64) {
	random := rand.New(rand.NewSource(seed))

	for {
		select {
		case <-shutdown:
			return
		case output <- line(random):
		}
	}
}

// Returns inexhaustible channel of strings with the specified chan_size and
// number of concurrent workers writing to it.  Writing a single value to the
// shutdown channel stops all workers gracefully.
func lines(shutdown <-chan bool, chan_size, workers int) chan []byte {
	result := make(chan []byte, chan_size)

	shutdown_flag := make([]chan bool, workers)
	var salt int64 = 1
	for i := 0; i < workers; i++ {
		shutdown_flag[i] = make(chan bool)
		go generate(result, shutdown_flag[i], time.Now().Unix()+salt)
		salt *= 100
	}

	go func() {
		<-shutdown
		for i := 0; i < workers; i++ {
			shutdown_flag[i] <- true
		}
	}()

	return result
}

// Writes size amount of generated data to w, using specified size of I/O
// buffer, size of the lines channel, and number of workers.
func write(w io.Writer, size uint64, io_buf_size, chan_size, workers int) error {
	out := bufio.NewWriterSize(w, io_buf_size)

	shutdown := make(chan bool)
	defer func() {
		shutdown <- true
	}()

	const nl = byte('\n')

	var count uint64 = 0
	for line := range lines(shutdown, chan_size, workers) {
		if count >= size {
			break
		}
		new_count := count + uint64(len(line)+1)
		if new_count >= size {
			line = line[0 : size-count-1]
		}
		count = new_count

		_, err := out.Write(line)
		if err == nil {
			err = out.WriteByte(nl)
		}

		if err != nil {
			return err
		}
	}

	return out.Flush()
}

func main() {
	var size uint64
	var help bool

	flag.Uint64Var(&size, "size", 1024*1024, "output file size in bytes")
	flag.BoolVar(&help, "help", false, "displays this help message")

	flag.Parse()

	if (flag.NArg() > 1) || help {
		fmt.Fprintln(os.Stderr, "Usage: filegen [-size SIZE] [FILE]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var out io.Writer
	if flag.NArg() == 1 {
		file, err := os.Create(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
		defer file.Close()
		out = file

	} else {
		out = os.Stdout
	}

	// FIXME: tune performance
	const io_buf_size = 4096
	const line_buf_size = 1000
	const workers = 2

	// FIXME: too long
	if err := write(out, size, io_buf_size, line_buf_size, workers); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(3)
	}
}
