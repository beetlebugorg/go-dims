package aws

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/beetlebugorg/go-dims/internal/commands"
	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/beetlebugorg/go-dims/internal/dims"
	v4 "github.com/beetlebugorg/go-dims/internal/v4"
	v5 "github.com/beetlebugorg/go-dims/internal/v5"
	"github.com/davidbyttow/govips/v2/vips"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	dims.RequestContext

	response *events.LambdaFunctionURLStreamingResponse
}

func NewRequest(event events.LambdaFunctionURLRequest, config core.Config) (*Request, error) {
	requestUrl, err := url.Parse(event.RawPath + "?" + event.RawQueryString)
	if err != nil {
		return nil, err
	}

	request := &Request{
		response: &events.LambdaFunctionURLStreamingResponse{
			Headers: map[string]string{},
		},
	}
	httpRequest := &http.Request{
		URL: requestUrl,
	}

	// Commands can be v4 (/dims4/...) or v5 (/v5/...)
	if strings.HasPrefix(requestUrl.Path, "/dims4/") {
		path := requestUrl.Path[7:]
		parts := strings.SplitN(path, "/", 4)
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid dims4 path format")
		}

		httpRequest.SetPathValue("clientId", parts[0])
		httpRequest.SetPathValue("signature", parts[1])
		httpRequest.SetPathValue("timestamp", parts[2])
		httpRequest.SetPathValue("commands", parts[3])

		v4Request, err := v4.NewRequest(httpRequest, nil, config)
		if err != nil {
			return nil, err
		}

		request.RequestContext = v4Request
	} else if strings.HasPrefix(requestUrl.Path, "/v5/") {
		cmds := strings.TrimLeft(requestUrl.Path, "/v5/")

		httpRequest.SetPathValue("commands", cmds)

		v5Request, err := v5.NewRequest(httpRequest, nil, config)
		if err != nil {
			return nil, err
		}

		request.RequestContext = v5Request
	} else {
		return nil, core.NewStatusError(400, "path must start with /dims4/ or /v5/: "+requestUrl.Path)
	}

	return request, nil
}

func (r *Request) sendHeaders() {
	headers := make(map[string]string)

	if r.RequestContext.CacheControl() != "" {
		headers["Cache-Control"] = r.CacheControl()
	}

	if r.RequestContext.Etag() != "" {
		headers["ETag"] = r.Etag()
	}

	if r.RequestContext.Expires() != "" {
		headers["Expires"] = r.Expires()
	}

	if r.RequestContext.LastModified() != "" {
		headers["Last-Modified"] = r.LastModified()
	}

	if r.RequestContext.ContentDisposition() != "" {
		headers["Content-Disposition"] = r.ContentDisposition()
	}

	if r.RequestContext.EdgeControl() != "" {
		headers["Edge-Control"] = r.EdgeControl()
	}

	r.response.Headers = headers
}

func (r *Request) SendImage(status int, imageFormat string, imageBlob []byte) error {
	if status == http.StatusOK {
		r.sendHeaders()
	}

	headers := r.Response().Headers
	headers["Content-Type"] = fmt.Sprintf("image/%s", strings.ToLower(imageFormat))
	headers["Content-Length"] = strconv.Itoa(len(imageBlob))

	response := &events.LambdaFunctionURLStreamingResponse{}
	response.StatusCode = status
	response.Headers = headers
	response.Body = bytes.NewReader(imageBlob)

	r.response = response

	return nil
}

func (r *Request) SendError(err error) error {
	message := err.Error()

	// Send error headers.
	headers := r.Response().Headers
	maxAge := r.Config().OriginCacheControl.Error
	if maxAge > 0 {
		headers["Cache-Control"] = fmt.Sprintf("max-age=%d, public", maxAge)
		headers["Expires"] = time.Now().Add(time.Duration(maxAge) * time.Second).UTC().Format(http.TimeFormat)
	}

	// Strip stack from vips errors.
	if strings.HasPrefix(message, "VipsOperation:") {
		message = message[0:strings.Index(message, "\n")]
	}

	slog.Error("SendError", "message", message)

	// Set status code.
	status := http.StatusInternalServerError
	var statusError *core.StatusError
	var operationError *commands.OperationError
	if errors.As(err, &statusError) {
		status = statusError.StatusCode
	} else if errors.As(err, &operationError) {
		status = operationError.StatusCode
	}

	errorImage, err := core.ErrorImage(r.Config().Error.Background)
	if err != nil {
		return err
	}

	imageType, imageBlob, err := r.ProcessImage(errorImage, true)
	if err != nil {
		// If processing failed because of a bad command then return the image as-is.
		exportOptions := vips.NewJpegExportParams()
		exportOptions.Quality = 1
		imageBytes, _, _ := errorImage.ExportJpeg(exportOptions)

		return r.SendImage(status, "jpg", imageBytes)
	}

	if imageType == "" {
		imageType = "jpg"
	}

	return r.SendImage(status, imageType, imageBlob)
}

func (r *Request) Response() *events.LambdaFunctionURLStreamingResponse {
	return r.response
}
