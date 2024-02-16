package dims

type Image struct {
	Bytes        []byte // The downloaded image.
	Size         int    // The original image size in bytes.
	Format       string // The original image format.
	Status       int    // The HTTP status code of the downloaded image.
	CacheControl string // The cache headers from the downloaded image.
	EdgeControl  string // The edge control headers from the downloaded image.
	LastModified string // The last modified header from the downloaded image.
	Etag         string // The etag header from the downloaded image.
}
