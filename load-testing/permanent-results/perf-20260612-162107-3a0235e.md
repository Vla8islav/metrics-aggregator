I've located two issues:

### My server establishing brand-new postgres connections mid-flight. 

It probably indicates that it's bad when it comes to the connection pool

Namely, this one should be so small so it doesn't appear on the graph.

Possible fix:

db.SetMaxOpenConns(20)
db.SetMaxIdleConns(20) // match MaxOpenConns
db.SetConnMaxIdleTime(5 * time.Minute)

### GZIP allocation issue

go tool pprof -text -sample_index=alloc_space -nodecount=15
load-testing/perf-20260612-142008-5acd7aa/heap.pprof

3549.53MB 53.54% 53.54% 6466.57MB 97.54% compress/flate.NewWriter (inline)
1782.11MB 26.88% 80.42% 2917.05MB 44.00% compress/flate.(*compressor).init
1109.42MB 16.73% 97.16% 1109.42MB 16.73% compress/flate.newDeflateFast (inline)
40.55MB 0.61% 97.77% 40.55MB 0.61% sync.(*Pool).pinSlow

we're creating a new compressor for each request.
It should be pooled

var gzipPool = sync.Pool{
New: func() any {
return gzip.NewWriter(io.Discard)
},
}

gz := gzipPool.Get().(*gzip.Writer)
gz.Reset(w) // rebind to current response; reuses the 800KB tables
// ... stream response through gz ...
gz.Close() // flushes gzip trailer — MUST happen before Put
gzipPool.Put(gz)

also it shouldn't be run on tiny requests and responses at all

const gzipMinSize = 1024