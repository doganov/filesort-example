package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
)

var (
	test_chan_sizes = []int{0, 1, 5, 10, 111, 1024}
	test_worker_groups  = []int{1, 2, 3, 4, 8}
	bench_size     = uint64(1024 * 1024 * 256) //128)
)

// FIXME: TestLine

func TestLines(t *testing.T) {
	skips := []int{0, 0, 1, 2, 10, 111, 1024, 1024 * 1024}

	const max_len_line = fragments * max_fragment_size

	for _, chan_size := range test_chan_sizes {
		for _, workers := range test_worker_groups {
			shutdown := make(chan bool)
			c := lines(shutdown, chan_size, workers)
			for _, skip := range skips {
				for i := 0; i < skip; i++ {
					<-c
				}
				line := <-c
				switch {
				case len(line) == 0:
					t.Errorf("%q is empty", line)
				case len(line) > max_len_line:
					t.Errorf("%q is larger than expected", line)
				}
			}
			shutdown <- true
		}
	}
}

func TestWrite(t *testing.T) {
	sizes := []uint64{0, 1, 5, 10, 111, 1024, 1025, 1024 * 4, 1024 * 1024}
	io_buf_sizes := []int{1, 5, 10, 111, 1024, 1025, 1024 * 4, 1024 * 1024}

	w := "write(size=%d,io_buf_size=%d,chan_size=%d,workers=%d)"

	var buf bytes.Buffer

	for _, size := range sizes {
		for _, io_buf_size := range io_buf_sizes {
			for _, chan_size := range test_chan_sizes {
				for _, workers := range test_worker_groups {
					buf.Reset()
					err := write(
						&buf, size, io_buf_size, chan_size, workers)
					if err != nil {
						t.Errorf(w+" produces error: %v",
							size, io_buf_size, chan_size, workers,
							err)
					}
					if uint64(buf.Len()) != size {
						t.Errorf(w+" writes %d byte(s)",
							size, io_buf_size, chan_size, workers,
							buf.Len())
					}
				}
			}
		}
	}
}

func benchWrite(b *testing.B, io_buf_size, chan_size, workers int) {
	f, err := ioutil.TempFile("", "filegen_bench_")
	if err != nil {
		b.Errorf("Can't create temp file: %v", err)
		return
	}

	defer func() {
		if err := f.Close(); err != nil {
			b.Errorf("Can't close temp file: %v", err)
		}
		if err := os.Remove(f.Name()); err != nil {
			b.Errorf("Can't delete temp file: %v", err)
		}

	}()

	for i := 0; i < b.N; i++ {
		if err := f.Truncate(0); err != nil {
			b.Errorf("Can't truncate temp file: %v", err)
			return
		}
		write(f, bench_size, io_buf_size, chan_size, workers)
	}
}

// io_buf_size=4K, chan_size=10

func BenchmarkWrite_IOBufSize_4K_ChanSize_10_Workers_1(b *testing.B) {
	benchWrite(b, 1024*4, 10, 1)
}

func BenchmarkWrite_IOBufSize_4K_ChanSize_10_Workers_2(b *testing.B) {
	benchWrite(b, 1024*4, 10, 2)
}

func BenchmarkWrite_IOBufSize_4K_ChanSize_10_Workers_4(b *testing.B) {
	benchWrite(b, 1024*4, 10, 4)
}

func BenchmarkWrite_IOBufSize_4K_ChanSize_10_Workers_8(b *testing.B) {
	benchWrite(b, 1024*4, 10, 8)
}

// io_buf_size=4K, chan_size=100

func BenchmarkWrite_IOBufSize_4K_ChanSize_100_Workers_1(b *testing.B) {
	benchWrite(b, 1024*4, 100, 1)
}

func BenchmarkWrite_IOBufSize_4K_ChanSize_100_Workers_2(b *testing.B) {
	benchWrite(b, 1024*4, 100, 2)
}

func BenchmarkWrite_IOBufSize_4K_ChanSize_100_Workers_4(b *testing.B) {
	benchWrite(b, 1024*4, 100, 4)
}

func BenchmarkWrite_IOBufSize_4K_ChanSize_100_Workers_8(b *testing.B) {
	benchWrite(b, 1024*4, 100, 8)
}

// io_buf_size=4K, chan_size=1000

func BenchmarkWrite_IOBufSize_4K_ChanSize_1000_Workers_1(b *testing.B) {
	benchWrite(b, 1024*4, 1000, 1)
}

func BenchmarkWrite_IOBufSize_4K_ChanSize_1000_Workers_2(b *testing.B) {
	benchWrite(b, 1024*4, 1000, 2)
}

func BenchmarkWrite_IOBufSize_4K_ChanSize_1000_Workers_4(b *testing.B) {
	benchWrite(b, 1024*4, 1000, 4)
}

func BenchmarkWrite_IOBufSize_4K_ChanSize_1000_Workers_8(b *testing.B) {
	benchWrite(b, 1024*4, 1000, 8)
}

// io_buf_size=4K, chan_size=10000

func BenchmarkWrite_IOBufSize_4K_ChanSize_10000_Workers_1(b *testing.B) {
	benchWrite(b, 1024*4, 10000, 1)
}

func BenchmarkWrite_IOBufSize_4K_ChanSize_10000_Workers_2(b *testing.B) {
	benchWrite(b, 1024*4, 10000, 2)
}

func BenchmarkWrite_IOBufSize_4K_ChanSize_10000_Workers_4(b *testing.B) {
	benchWrite(b, 1024*4, 10000, 4)
}

func BenchmarkWrite_IOBufSize_4K_ChanSize_10000_Workers_8(b *testing.B) {
	benchWrite(b, 1024*4, 10000, 8)
}

// io_buf_size=8K, chan_size=1000

func BenchmarkWrite_IOBufSize_8K_ChanSize_1000_Workers_NumCPU(b *testing.B) {
	benchWrite(b, 1024*8, 1000, runtime.NumCPU())
}

func BenchmarkWrite_IOBufSize_8K_ChanSize_1000_Workers_NumCPUx2(b *testing.B) {
	benchWrite(b, 1024*8, 1000, runtime.NumCPU()*2)
}

// io_buf_size=8K, chan_size=10000
func BenchmarkWrite_IOBufSize_8K_ChanSize_10000_Workers_NumCPU(b *testing.B) {
	benchWrite(b, 1024*8, 10000, runtime.NumCPU())
}

func BenchmarkWrite_IOBufSize_8K_ChanSize_10000_Workers_NumCPUx2(b *testing.B) {
	benchWrite(b, 1024*8, 10000, runtime.NumCPU()*2)
}

// io_buf_size=64K, chan_size=1000

func BenchmarkWrite_IOBufSize_64K_ChanSize_1000_Workers_NumCPU(b *testing.B) {
	benchWrite(b, 1024*64, 1000, runtime.NumCPU())
}

func BenchmarkWrite_IOBufSize_64K_ChanSize_1000_Workers_NumCPUx2(b *testing.B) {
	benchWrite(b, 1024*64, 1000, runtime.NumCPU()*2)
}

// io_buf_size=64K, chan_size=10000
func BenchmarkWrite_IOBufSize_64K_ChanSize_10000_Workers_NumCPU(b *testing.B) {
	benchWrite(b, 1024*64, 10000, runtime.NumCPU())
}

func BenchmarkWrite_IOBufSize_64K_ChanSize_10000_Workers_NumCPUx2(b *testing.B) {
	benchWrite(b, 1024*64, 10000, runtime.NumCPU()*2)
}

// io_buf_size=512K, chan_size=1000

func BenchmarkWrite_IOBufSize_512K_ChanSize_1000_Workers_NumCPU(b *testing.B) {
	benchWrite(b, 1024*512, 1000, runtime.NumCPU())
}

func BenchmarkWrite_IOBufSize_512K_ChanSize_1000_Workers_NumCPUx2(b *testing.B) {
	benchWrite(b, 1024*512, 1000, runtime.NumCPU()*2)
}

// io_buf_size=512K, chan_size=10000
func BenchmarkWrite_IOBufSize_512K_ChanSize_10000_Workers_NumCPU(b *testing.B) {
	benchWrite(b, 1024*512, 10000, runtime.NumCPU())
}

func BenchmarkWrite_IOBufSize_512K_ChanSize_10000_Workers_NumCPUx2(b *testing.B) {
	benchWrite(b, 1024*512, 10000, runtime.NumCPU()*2)
}

// io_buf_size=4096K, chan_size=1000

func BenchmarkWrite_IOBufSize_4096K_ChanSize_1000_Workers_NumCPU(b *testing.B) {
	benchWrite(b, 1024*1024*4, 1000, runtime.NumCPU())
}

func BenchmarkWrite_IOBufSize_4096K_ChanSize_1000_Workers_NumCPUx2(b *testing.B) {
	benchWrite(b, 1024*1024*4, 1000, runtime.NumCPU()*2)
}

// io_buf_size=4096K, chan_size=10000

func BenchmarkWrite_IOBufSize_4096K_ChanSize_10000_Workers_NumCPU(b *testing.B) {
	benchWrite(b, 1024*1024*4, 10000, runtime.NumCPU())
}

func BenchmarkWrite_IOBufSize_4096K_ChanSize_10000_Workers_NumCPUx2(b *testing.B) {
	benchWrite(b, 1024*1024*4, 10000, runtime.NumCPU()*2)
}

// io_buf_size=32768K, chan_size=1000

func BenchmarkWrite_IOBufSize_32768K_ChanSize_1000_Workers_NumCPU(b *testing.B) {
	benchWrite(b, 1024*1024*32, 1000, runtime.NumCPU())
}

func BenchmarkWrite_IOBufSize_32768K_ChanSize_1000_Workers_NumCPUx2(b *testing.B) {
	benchWrite(b, 1024*1024*32, 1000, runtime.NumCPU()*2)
}

// io_buf_size=32768K, chan_size=10000

func BenchmarkWrite_IOBufSize_32768K_ChanSize_10000_Workers_NumCPU(b *testing.B) {
	benchWrite(b, 1024*1024*32, 10000, runtime.NumCPU())
}

func BenchmarkWrite_IOBufSize_32768K_ChanSize_10000_Workers_NumCPUx2(b *testing.B) {
	benchWrite(b, 1024*1024*32, 10000, runtime.NumCPU()*2)
}
