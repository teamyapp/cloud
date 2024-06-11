package api

import (
	"time"

	"github.com/teamyapp/cloud/app/entity"
	pbcloud "github.com/teamyapp/protocol/pb/pbgo/cloud"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var toProtoUploadSessionStatus = map[entity.UploadSessionStatus]pbcloud.UploadSessionStatus{
	entity.CreatedUploadSessionStatus:         pbcloud.UploadSessionStatus_CREATED,
	entity.InitializedUploadSessionStatus:     pbcloud.UploadSessionStatus_INITIALIZED,
	entity.UploadingChunksUploadSessionStatus: pbcloud.UploadSessionStatus_UPLOADING_CHUNKS,
	entity.CompletedUploadSessionStatus:       pbcloud.UploadSessionStatus_COMPLETED,
}

func toProtoTimePtr(tm *time.Time) *timestamppb.Timestamp {
	if tm == nil {
		return nil
	}

	return timestamppb.New(*tm)
}
