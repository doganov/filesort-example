# filesort-example
Large text file sorting example in Go.

## Requirements

- Go 1.6 (probably works with earlier versions, but not tested)
- GNU Make

## Download

```
$ git clone https://github.com/doganov/filesort-example.git
```

## Build

```
$ cd filesort-example
$ make
```

This creates two executables in the `bin/` directory:

- `bin/filegen`
- `bin/filesort`

## Run

Create 4GB test file named `in.txt`:

```
$ bin/filegen --size=4294967296 in.txt
```

Sort it on disk with limited memory footprint:

```
$ bin/filesort --limit=100000 in.txt out.txt
```

The `--limit` option specifies the maximum number of lines to be included in the
initial chunks. Since the initial chunks are sorted in memory (one by one), the
the memory footprint depends on the size of one chunk.  Larger chunk consumes
more memory. Smaller chunk consumes less memory, but yields more chunks
and hence -- more I/O operations to merge them.

When the output filename is omitted, `filesort` writes to the standard
output. When the inputfile is omitted, `filesort` reads from the standard input.
