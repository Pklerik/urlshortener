File: router.test
Type: alloc_space
Time: 2025-11-23 15:21:22 MSK
Showing nodes accounting for -8955.08MB, 49.58% of 18063.26MB total
Dropped 376 nodes (cum <= 90.32MB)
      flat  flat%   sum%        cum   cum%
-6774.68MB 37.51% 37.51% -8191.90MB 45.35%  compress/flate.NewWriter (inline)
-1385.19MB  7.67% 45.17% -1385.19MB  7.67%  compress/flate.(*compressor).initDeflate (inline)
 -341.45MB  1.89% 47.06%  -341.45MB  1.89%  compress/flate.(*dictDecoder).init (inline)
 -234.68MB  1.30% 48.36%  -234.68MB  1.30%  net/http.init.func15
  -76.80MB  0.43% 48.79%  -418.25MB  2.32%  compress/flate.NewReader
  -69.65MB  0.39% 49.17%   -69.65MB  0.39%  compress/flate.(*huffmanEncoder).generate
  -61.60MB  0.34% 49.52%   -61.60MB  0.34%  sync.(*Pool).pinSlow
   -7.50MB 0.042% 49.56%  -464.90MB  2.57%  io.ReadAll
      -2MB 0.011% 49.57%  -457.40MB  2.53%  compress/gzip.NewReader (inline)
      -2MB 0.011% 49.58% -8283.02MB 45.86%  github.com/Pklerik/urlshortener/internal/handler.(*LinkHandle).PostText
       1MB 0.0055% 49.57%   -54.84MB   0.3%  github.com/Pklerik/urlshortener/internal/service/links.(*BaseLinkService).RegisterLinks
   -0.50MB 0.0028% 49.58% -8287.32MB 45.88%  github.com/Pklerik/urlshortener/internal/handler.(*LinkHandle).AuditMiddleware.func1
    0.50MB 0.0028% 49.57% -8401.96MB 46.51%  github.com/go-chi/chi/middleware.RequestID.func1
   -0.50MB 0.0028% 49.58%  -246.69MB  1.37%  net/http.(*Request).write
         0     0% 49.58%   -69.65MB  0.39%  compress/flate.(*Writer).Close (inline)
         0     0% 49.58%   -69.65MB  0.39%  compress/flate.(*compressor).close
         0     0% 49.58%   -70.15MB  0.39%  compress/flate.(*compressor).deflate
         0     0% 49.58% -1417.21MB  7.85%  compress/flate.(*compressor).init
         0     0% 49.58%   -70.15MB  0.39%  compress/flate.(*compressor).writeBlock
         0     0% 49.58%   -70.15MB  0.39%  compress/flate.(*huffmanBitWriter).writeBlock
         0     0% 49.58%  -455.40MB  2.52%  compress/gzip.(*Reader).Reset
         0     0% 49.58%  -418.25MB  2.32%  compress/gzip.(*Reader).readHeader
         0     0% 49.58%   -69.12MB  0.38%  compress/gzip.(*Writer).Close
         0     0% 49.58% -8191.39MB 45.35%  compress/gzip.(*Writer).Write
         0     0% 49.58% -8297.82MB 45.94%  github.com/Pklerik/urlshortener/internal/handler.(*LinkHandle).AuthUser.func1
         0     0% 49.58%   -69.62MB  0.39%  github.com/Pklerik/urlshortener/internal/middleware.(*compressWriter).Close
         0     0% 49.58% -8192.47MB 45.35%  github.com/Pklerik/urlshortener/internal/middleware.(*compressWriter).Write
         0     0% 49.58% -8369.94MB 46.34%  github.com/Pklerik/urlshortener/internal/middleware.GZIPMiddleware.func1
         0     0% 49.58%  -514.91MB  2.85%  github.com/Pklerik/urlshortener/internal/router.BenchmarkShortenerService.func1
         0     0% 49.58% -8293.82MB 45.92%  github.com/Pklerik/urlshortener/internal/router.ConfigureRouter.func1.Timeout.2.1
         0     0% 49.58% -8401.96MB 46.51%  github.com/go-chi/chi.(*ChainHandler).ServeHTTP
         0     0% 49.58% -8287.82MB 45.88%  github.com/go-chi/chi.(*Mux).Mount.func1
         0     0% 49.58% -8416.48MB 46.59%  github.com/go-chi/chi.(*Mux).ServeHTTP
         0     0% 49.58% -8401.46MB 46.51%  github.com/go-chi/chi.(*Mux).routeHTTP
         0     0% 49.58% -8395.96MB 46.48%  github.com/go-chi/chi/middleware.RealIP.func1
         0     0% 49.58% -8395.96MB 46.48%  github.com/go-chi/chi/middleware.init.0.RequestLogger.func1.1
         0     0% 49.58%  -497.41MB  2.75%  github.com/go-resty/resty/v2.(*Client).execute
         0     0% 49.58%  -505.42MB  2.80%  github.com/go-resty/resty/v2.(*Request).Execute
         0     0% 49.58%  -505.42MB  2.80%  github.com/go-resty/resty/v2.(*Request).Send (inline)
         0     0% 49.58%  -455.40MB  2.52%  github.com/go-resty/resty/v2.readAllWithLimit
         0     0% 49.58% -8462.03MB 46.85%  net/http.(*conn).serve
         0     0% 49.58%  -457.40MB  2.53%  net/http.(*gzipReader).Read
         0     0% 49.58%  -246.69MB  1.37%  net/http.(*persistConn).writeLoop
         0     0% 49.58%  -240.19MB  1.33%  net/http.(*transferWriter).doBodyCopy
         0     0% 49.58%  -240.69MB  1.33%  net/http.(*transferWriter).writeBody
         0     0% 49.58% -8401.46MB 46.51%  net/http.HandlerFunc.ServeHTTP
         0     0% 49.58%  -240.19MB  1.33%  net/http.getCopyBuf (inline)
         0     0% 49.58% -8416.48MB 46.59%  net/http.serverHandler.ServeHTTP
         0     0% 49.58%  -277.75MB  1.54%  sync.(*Pool).Get
         0     0% 49.58%   -61.60MB  0.34%  sync.(*Pool).pin
         0     0% 49.58%  -515.41MB  2.85%  testing.(*B).launch
         0     0% 49.58%  -514.91MB  2.85%  testing.(*B).runN
