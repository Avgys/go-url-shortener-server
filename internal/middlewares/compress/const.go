package compress

const gzipType = "gzip"
const noResult = "no result"
const acceptEncodingHeader = "Accept-Encoding"
const contentTypeHeader = "Content-Type"
const contentEncodingHeader = "Content-Encoding"

var typesToDecompress = []string{"application/json", "text/plain", "application/x-gzip"}
var supportedEncodings = []string{gzipType}
