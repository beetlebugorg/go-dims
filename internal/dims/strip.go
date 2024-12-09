package dims

func StripMetadataCommand(request *Request, args string) error {
	request.exportJpegParams.StripMetadata = true
	request.exportWebpParams.StripMetadata = true
	request.exportPngParams.StripMetadata = true

	return nil
}
