package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const uploadSessionIDParam = "uploadSessionId"
const fileIDParam = "fileId"

type File struct {
	logger      telemetry.Logger
	fileService service.File
	proto.UnimplementedFileServer
}

var _ runner.Service = (*File)(nil)
var _ proto.FileServer = (*File)(nil)

func (f File) Start(rn *runner.ServiceRunner) *errs.Error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(filePathPrefix, "upload-sessions", runner.Param(uploadSessionIDParam)),
			HandlerFunc: f.webGetUploadSession,
		},
		{
			Method:      http.MethodPut,
			Pattern:     path.Join(filePathPrefix, "upload-sessions", runner.Param(uploadSessionIDParam), "init"),
			HandlerFunc: f.webInitUploadSession,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     path.Join(filePathPrefix, "upload-sessions", runner.Param(uploadSessionIDParam), "delete"),
			HandlerFunc: f.webDeleteUploadSession,
		},
		{
			Method:      http.MethodPost,
			Pattern:     path.Join(filePathPrefix, "upload-sessions", runner.Param(uploadSessionIDParam), "chunks", "add"),
			HandlerFunc: f.webAddChunk,
		},
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(filePathPrefix, "files", runner.Param(fileIDParam), "metadata"),
			HandlerFunc: f.webGetFileMetadata,
		},
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(filePathPrefix, "files", runner.Param(fileIDParam)),
			HandlerFunc: f.webGetFileByID,
		},
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(filePathPrefix, "download"),
			HandlerFunc: f.webDownloadPath,
		},
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(filePathPrefix, "files"),
			HandlerFunc: f.webGetFileByPath,
		},
	})
	rn.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterFileServer(server, f)
	})

	return nil
}

func (f File) CreateUploadSession(ct context.Context, empty *emptypb.Empty) (*proto.CreateUploadSessionResponse, error) {
	uploadSessionID, err := f.fileService.CreateUploadSession(ct)
	if err != nil {
		f.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &proto.CreateUploadSessionResponse{
		UploadSessionId: uploadSessionID,
	}, nil
}

func (f File) FindUploadSession(ct context.Context, req *proto.FindUploadSessionRequest) (*proto.UploadSession, error) {
	uploadSession, err := f.fileService.GetUploadSession(ct, req.UploadSessionId)
	if err != nil {
		f.logger.ErrorWithContext(ct, err)
		return &proto.UploadSession{}, errs.ToGRPCErr(err)
	}

	return &proto.UploadSession{
		Id:                     uploadSession.ID,
		Status:                 toProtoUploadSessionStatus[uploadSession.Status],
		UploadedSizeInBytes:    uploadSession.UploadedSizeInBytes,
		FileId:                 uploadSession.FileID,
		FileName:               uploadSession.FileName,
		MimeType:               uploadSession.MIMEType,
		TotalSizeInBytes:       uploadSession.TotalSizeInBytes,
		TotalNumOfChunks:       int32(uploadSession.TotalNumOfChunks),
		ChunkIDs:               uploadSession.ChunkIDs,
		NextChunkIndexToUpload: int32(uploadSession.NextChunkIndexToUpload),
		ActualContentHash:      uploadSession.ActualContentHash,
		ExpectedContentHash:    uploadSession.ExpectedContentHash,
		CreatedAt:              timestamppb.New(uploadSession.CreatedAt),
		UpdatedAt:              toProtoTimePtr(uploadSession.UpdatedAt),
	}, nil
}

func (f File) webGetUploadSession(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	uploadSessionIDRaw := chi.URLParam(request, uploadSessionIDParam)
	uploadSessionID, err := strconv.ParseUint(uploadSessionIDRaw, 10, 64)
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, err.Error())
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	uploadSession, internalErr := f.fileService.GetUploadSession(request.Context(), uploadSessionID)
	if err != nil {
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	web.WriteJSONToResponse(writer, uploadSession)
}

func (f File) webInitUploadSession(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	uploadSessionIDRaw := chi.URLParam(request, uploadSessionIDParam)
	uploadSessionID, err := strconv.ParseUint(uploadSessionIDRaw, 10, 64)
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, err.Error())
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	buf, err := io.ReadAll(request.Body)
	if err != nil {
		internalErr := errs.NewError(errs.IO, err.Error())
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	var body struct {
		FileName            string `json:"fileName"`
		MIMEType            string `json:"mimeType"`
		ExpectedContentHash string `json:"expectedContentHash"`
		TotalSizeInBytes    uint64 `json:"totalSizeInBytes"`
		TotalNumOfChunks    int    `json:"totalNumOfChunks"`
	}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		internalErr := errs.NewError(errs.Deserialization, err.Error())
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	uploadSession, internalErr := f.fileService.InitUploadSession(
		request.Context(),
		uploadSessionID,
		body.FileName,
		body.MIMEType,
		body.ExpectedContentHash,
		body.TotalSizeInBytes,
		body.TotalNumOfChunks)
	if internalErr != nil {
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	web.WriteJSONToResponse(writer, uploadSession)
}

func (f File) webDeleteUploadSession(writer http.ResponseWriter, request *http.Request) {
	panic("not implemented")
}

func (f File) webAddChunk(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	uploadSessionIDRaw := chi.URLParam(request, uploadSessionIDParam)
	uploadSessionID, err := strconv.ParseUint(uploadSessionIDRaw, 10, 64)
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, err.Error())
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	uploadSession, internalErr := f.fileService.AddChunk(ct, uploadSessionID, request.Body, request.ContentLength)
	if internalErr != nil {
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	web.WriteJSONToResponse(writer, uploadSession)
}

func (f File) webGetFileMetadata(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	fileIDRaw := chi.URLParam(request, fileIDParam)
	fileID, err := strconv.ParseUint(fileIDRaw, 10, 64)
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, err.Error())
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	fileMetadata, internalErr := f.fileService.GetFileMetadata(request.Context(), fileID)
	if internalErr != nil {
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	web.WriteJSONToResponse(writer, fileMetadata)
}

func (f File) webDownloadPath(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	encodedFilePath := request.URL.Query().Get("path")
	if encodedFilePath == "" {
		internalErr := errs.NewError(errs.InvalidArgument, "path query param is required")
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	filePath, err := url.QueryUnescape(encodedFilePath)
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, err.Error())
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	fileStream, internalErr := f.fileService.GetCompressedFileStream(request.Context(), filePath)
	if internalErr != nil {
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	writer.Header().Set("Content-Type", fileStream.MIMEContentType)
	writer.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileStream.Name))

	flusher, ok := writer.(http.Flusher)
	if !ok {
		internalErr = errs.NewError(errs.Unknown, "writer must be http.Flusher")
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	_, err = io.Copy(writer, fileStream.ContentReader)
	if err != nil {
		internalErr = errs.NewError(errs.Unknown, err.Error())
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	flusher.Flush()
}

func (f File) webGetFileByPath(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	encodedFilePath := request.URL.Query().Get("path")
	if encodedFilePath == "" {
		internalErr := errs.NewError(errs.InvalidArgument, "path query param is required")
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	filePath, err := url.QueryUnescape(encodedFilePath)
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, err.Error())
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	metadata, internalErr := f.fileService.GetFileMetadataFromPath(ct, filePath)
	if internalErr != nil {
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	writer.Header().Set("Content-Type", metadata.ContentType)
	writer.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, metadata.Name))
	writer.Header().Set("ETag", metadata.ETag)
	if eTag := request.Header.Get("If-None-Match"); len(eTag) > 0 {
		if eTag == metadata.ETag {
			writer.WriteHeader(http.StatusNotModified)
			return
		}
	}

	fileReader, internalErr := f.fileService.GetFileFromPath(request.Context(), filePath)
	if internalErr != nil {
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	flusher, ok := writer.(http.Flusher)
	if !ok {
		internalErr = errs.NewError(errs.Unknown, "writer must be http.Flusher")
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	_, err = io.Copy(writer, fileReader)
	if err != nil {
		internalErr = errs.NewError(errs.Unknown, err.Error())
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	flusher.Flush()
}

func (f File) webGetFileByID(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	fileIDRaw := chi.URLParam(request, fileIDParam)
	fileID, err := strconv.ParseUint(fileIDRaw, 10, 64)
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, err.Error())
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	file, internalErr := f.fileService.GetFile(request.Context(), fileID)
	if internalErr != nil {
		f.logger.ErrorWithContext(ct, internalErr)
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
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	_, err = io.Copy(writer, file.ChunksBuffer)
	if err != nil {
		internalErr = errs.NewError(errs.Unknown, err.Error())
		f.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	flusher.Flush()
}

func NewFile(logger telemetry.Logger, fileService service.File) File {
	return File{
		logger:      logger,
		fileService: fileService,
	}
}
