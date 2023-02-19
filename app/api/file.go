package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
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

type File struct {
	dataCollector telemetry.DataCollector
	fileService   service.File
	proto.UnimplementedFileServer
}

var _ runner.Service = (*File)(nil)
var _ proto.FileServer = (*File)(nil)

func (f File) Start(rn *runner.ServiceRunner) *errs.Error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Path:        path.Join(filePathPrefix, "upload-sessions", "{uploadSessionId}"),
			Method:      http.MethodGet,
			HandlerFunc: f.webGetUploadSession,
		},
		{
			Path:        path.Join(filePathPrefix, "upload-sessions", "{uploadSessionId}", "init"),
			Method:      http.MethodPut,
			HandlerFunc: f.webInitUploadSession,
		},
		{
			Path:        path.Join(filePathPrefix, "upload-sessions", "{uploadSessionId}", "delete"),
			Method:      http.MethodDelete,
			HandlerFunc: f.webDeleteUploadSession,
		},
		{
			Path:        path.Join(filePathPrefix, "upload-sessions", "{uploadSessionId}", "chunks", "add"),
			Method:      http.MethodPost,
			HandlerFunc: f.webAddChunk,
		},
		{
			Path:        path.Join(filePathPrefix, "files", "{fileId}", "metadata"),
			Method:      http.MethodGet,
			HandlerFunc: f.webGetFileMetadata,
		},
		{
			Path:        path.Join(filePathPrefix, "files", "{fileId}"),
			Method:      http.MethodGet,
			HandlerFunc: f.webGetFile,
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
		f.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &proto.CreateUploadSessionResponse{
		UploadSessionId: uploadSessionID,
	}, nil
}

func (f File) FindUploadSession(ct context.Context, req *proto.FindUploadSessionRequest) (*proto.UploadSession, error) {
	uploadSession, err := f.fileService.GetUploadSession(ct, req.UploadSessionId)
	if err != nil {
		f.dataCollector.Logger.ErrorWithContext(ct, err)
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
	uploadSessionIDParam := mux.Vars(request)["uploadSessionId"]
	uploadSessionID, err := strconv.ParseUint(uploadSessionIDParam, 10, 64)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: err,
		}
		f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	uploadSession, internalErr := f.fileService.GetUploadSession(request.Context(), uploadSessionID)
	if err != nil {
		f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	web.WriteJSONToResponse(ct, f.dataCollector, writer, uploadSession)
}

func (f File) webInitUploadSession(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	uploadSessionIDParam := mux.Vars(request)["uploadSessionId"]
	uploadSessionID, err := strconv.ParseUint(uploadSessionIDParam, 10, 64)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: err,
		}
		f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	buf, err := io.ReadAll(request.Body)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.IO,
			EmbedErr: err,
		}
		f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
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
		internalErr := &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
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
		f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	web.WriteJSONToResponse(ct, f.dataCollector, writer, uploadSession)
}

func (f File) webDeleteUploadSession(writer http.ResponseWriter, request *http.Request) {
	panic("not implemented")
}

func (f File) webAddChunk(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	uploadSessionIDParam := mux.Vars(request)["uploadSessionId"]
	uploadSessionID, err := strconv.ParseUint(uploadSessionIDParam, 10, 64)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: err,
		}
		f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	data, err := io.ReadAll(request.Body)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	uploadSession, internalErr := f.fileService.AddChunk(request.Context(), uploadSessionID, data)
	if internalErr != nil {
		f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	web.WriteJSONToResponse(ct, f.dataCollector, writer, uploadSession)
}

func (f File) webGetFileMetadata(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	fileIDParam := mux.Vars(request)["fileId"]
	fileID, err := strconv.ParseUint(fileIDParam, 10, 64)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: err,
		}
		f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	fileMetadata, internalErr := f.fileService.GetFileMetadata(request.Context(), fileID)
	if internalErr != nil {
		f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	web.WriteJSONToResponse(ct, f.dataCollector, writer, fileMetadata)
}

func (f File) webGetFile(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	fileIDParam := mux.Vars(request)["fileId"]
	fileID, err := strconv.ParseUint(fileIDParam, 10, 64)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: err,
		}
		f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	file, internalErr := f.fileService.GetFile(request.Context(), fileID)
	if internalErr != nil {
		f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
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
		internalErr = &errs.Error{
			Code:    errs.Unknown,
			Message: "writer must be http.Flusher",
		}
		f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	for chunkResult := range file.ChunksBuffer {
		if chunkResult.Error != nil {
			f.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: chunkResult.Error})
			errs.SetHTTPErr(chunkResult.Error, writer)
			return
		}

		_, err = writer.Write(chunkResult.Value)
		if err != nil {
			internalErr = &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}
			f.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			errs.SetHTTPErr(internalErr, writer)
			return
		}

		flusher.Flush()
	}
}

func NewFile(dataCollector telemetry.DataCollector, fileService service.File) File {
	return File{
		dataCollector: dataCollector,
		fileService:   fileService,
	}
}
