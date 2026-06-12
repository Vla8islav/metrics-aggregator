Рзультат оптимизаии
```
-> % go tool pprof -text -sample_index=alloc_space -nodecount=15 load-testing/permanent-results/perf-20260612-162107-3a0235e/heap.pprof
File: ___11go_build_github_com_Vla8islav_metrics_aggregator_cmd_server
Type: alloc_space
Time: 2026-06-12 16:21:53 MSK
Showing nodes accounting for 33314.95MB, 97.93% of 34020.17MB total
Dropped 421 nodes (cum <= 170.10MB)
Showing top 15 nodes out of 25
flat  flat%   sum%        cum   cum%
18214.79MB 53.54% 53.54% 32956.12MB 96.87%  compress/flate.NewWriter (inline)
8998.35MB 26.45% 79.99% 14741.33MB 43.33%  compress/flate.(*compressor).init
5631.40MB 16.55% 96.54%  5631.40MB 16.55%  compress/flate.newDeflateFast (inline)
210.28MB  0.62% 97.16%   210.28MB  0.62%  sync.(*Pool).pinSlow
194.45MB  0.57% 97.73%   194.45MB  0.57%  compress/flate.(*dictDecoder).init (inline)
48.68MB  0.14% 97.88%   243.14MB  0.71%  compress/flate.NewReader
10.50MB 0.031% 97.91%   495.26MB  1.46%  main.main.WithLogging.func2.1
4MB 0.012% 97.92%   276.75MB  0.81%  compress/gzip.NewReader
1.50MB 0.0044% 97.92% 33512.51MB 98.51%  github.com/Vla8islav/metrics-aggregator/internal/middlewares.handleOutgoingCompression
1MB 0.0029% 97.93%   297.59MB  0.87%  github.com/gorilla/mux.(*Router).ServeHTTP
0     0% 97.93%   272.75MB   0.8%  compress/gzip.(*Reader).Reset
0     0% 97.93%   243.14MB  0.71%  compress/gzip.(*Reader).readHeader
0     0% 97.93% 32956.12MB 96.87%  compress/gzip.(*Writer).Write
0     0% 97.93% 32976.19MB 96.93%  github.com/Vla8islav/metrics-aggregator/internal/middlewares.(*capturingSignResponseWriter).FlushToOriginal
0     0% 97.93% 32947.19MB 96.85%  github.com/Vla8islav/metrics-aggregator/internal/middlewares.gzipWriter.Write
(⎈|yc-kuber-cluster:metrics-aggregator)vla8islav@MacBook-Pro-4 [16:54:31] [~/repos/github/Vla8islav/metrics-aggregator] [iter17 *]
-> % go tool pprof -text -sample_index=alloc_space -nodecount=15 load-testing/permanent-results/perf-20260612-164306-3a0235e/heap.pprof
File: ___11go_build_github_com_Vla8islav_metrics_aggregator_cmd_server
Type: alloc_space
Time: 2026-06-12 16:43:51 MSK
Showing nodes accounting for 140.75MB, 78.00% of 180.45MB total
Dropped 109 nodes (cum <= 0.90MB)
Showing top 15 nodes out of 166
flat  flat%   sum%        cum   cum%
55.19MB 30.58% 30.58%    55.19MB 30.58%  compress/flate.(*dictDecoder).init (inline)
19.39MB 10.75% 41.33%    31.44MB 17.43%  compress/flate.NewWriter
14.56MB  8.07% 49.40%    69.74MB 38.65%  compress/flate.NewReader
13.05MB  7.23% 56.63%    13.05MB  7.23%  bufio.NewReaderSize
6.64MB  3.68% 60.31%     6.64MB  3.68%  compress/flate.newDeflateFast (inline)
5.41MB  3.00% 63.31%    12.05MB  6.68%  compress/flate.(*compressor).init
4.50MB  2.49% 65.80%     4.50MB  2.49%  net/http.(*Request).WithContext
4MB  2.22% 68.02%        4MB  2.22%  net/textproto.MIMEHeader.Set
3MB  1.66% 69.68%        8MB  4.43%  github.com/jackc/pgx/v5/pgconn/ctxwatch.(*ContextWatcher).Watch
2.50MB  1.39% 71.07%     6.50MB  3.60%  net/http.readRequest
2.50MB  1.39% 72.45%     2.50MB  1.39%  net/textproto.readMIMEHeader
2.50MB  1.39% 73.84%     2.50MB  1.39%  net/http.Header.Clone
2.50MB  1.39% 75.22%     2.50MB  1.39%  context.(*cancelCtx).propagateCancel
2.50MB  1.39% 76.61%    43.04MB 23.85%  main.main.WithLogging.func2.1
2.50MB  1.39% 78.00%        5MB  2.77%  context.AfterFunc
```