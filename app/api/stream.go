package api

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
)

const fileNameParam = "fileName"

type Stream struct {
	logger        telemetry.Logger
	streamService service.Stream
}

var _ runner.Service = (*Stream)(nil)

func (s Stream) Start(rn *runner.ServiceRunner) *errs.Error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(streamPathPrefix, "files", runner.Param(fileIDParam)),
			HandlerFunc: s.webGetFile,
		},
		{
			Method:      http.MethodPost,
			Pattern:     path.Join(streamPathPrefix, "files", "upload"),
			HandlerFunc: s.webUploadFile,
		},
	})

	return nil
}

func (s Stream) webGetFile(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	fileIDRaw := chi.URLParam(request, fileIDParam)
	fileID, err := strconv.ParseUint(fileIDRaw, 10, 64)
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, err.Error())
		s.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	file, internalErr := s.streamService.GetFile(ct, fileID)
	if internalErr != nil {
		s.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	writer.Header().Set("Content-Type", file.Metadata.MIMEType)
	writer.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, file.Metadata.Name))
	if file.Metadata.LastModifiedAt != nil {
		writer.Header().Set("Last-Modified", file.Metadata.LastModifiedAt.UTC().Format(http.TimeFormat))
	}

	flusher, ok := writer.(http.Flusher)
	if !ok {
		internalErr = errs.NewError(errs.Unknown, "writer must be http.Flusher")
		s.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	_, err = io.Copy(writer, file.ChunksBuffer)
	if err != nil {
		internalErr = errs.NewError(errs.Unknown, err.Error())
		s.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	flusher.Flush()
}

func (s Stream) webUploadFile(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	fileName := request.URL.Query().Get(fileNameParam)
	if fileName == "" {
		internalErr := errs.NewError(errs.InvalidArgument, "fileName is required")
		s.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	internalErr := s.streamService.AddFile(ct, fileName, request.ContentLength, request.Body)
	if internalErr != nil {
		s.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	web.WriteJSONToResponse(writer, nil)
}

func NewStream(logger telemetry.Logger, streamService service.Stream) Stream {
	return Stream{
		logger:        logger,
		streamService: streamService,
	}
}
