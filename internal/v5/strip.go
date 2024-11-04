package v5

func StripMetadataCommand(request *RequestV5, args string) error {
	request.exportJpegParams.StripMetadata = true
	request.exportWebpParams.StripMetadata = true
	request.exportPngParams.StripMetadata = true

	return nil
}
