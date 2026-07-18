⚡ Optimize `os.ReadDir` in TUI model

💡 **What:** Replaced `os.ReadDir(dir)` with `os.Open(dir)` followed by `f.ReadDir(-1)`. We now filter out hidden files *before* sorting, and sort `dirs` and `files` individually using `slices.SortFunc`.

🎯 **Why:** `os.ReadDir` internally reads all entries and immediately sorts them. For large directories containing many hidden files (like `.` files in version control or large caches), sorting everything first only to discard many elements is wasteful sync I/O in the event loop.

📊 **Measured Improvement:**
A benchmark simulating a directory with 10,000 hidden files and 1,000 visible files showed an almost 2x performance increase.

```
goos: linux
goarch: amd64
cpu: Intel(R) Xeon(R) Processor @ 2.30GHz
BenchmarkReadDir-4            	     144	   8260646 ns/op
BenchmarkReadDirOptimized-4   	     242	   4932681 ns/op
```
