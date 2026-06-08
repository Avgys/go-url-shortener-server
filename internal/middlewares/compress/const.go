package compress

const gzipType = "gzip"
const noResult = "no result"

// minGzipResponseBytes avoids gzip overhead on small responses.
const minGzipResponseBytes = 1400
const acceptEncodingHeader = "Accept-Encoding"
const contentTypeHeader = "Content-Type"
const contentEncodingHeader = "Content-Encoding"

var typesToDecompress = []string{"application/json", "text/plain", "application/x-gzip"}
var supportedEncodings = []string{gzipType}
var noEncoding = "no-encoding"
