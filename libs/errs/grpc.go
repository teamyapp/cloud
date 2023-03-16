package errs

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var toGRPCErrCode = map[ErrorCode]codes.Code{
	Cancelled:         codes.Canceled,
	Unknown:           codes.Unknown,
	InvalidArgument:   codes.InvalidArgument,
	Timeout:           codes.DeadlineExceeded,
	NotFound:          codes.NotFound,
	AlreadyExists:     codes.AlreadyExists,
	PermissionDenied:  codes.PermissionDenied,
	Unauthenticated:   codes.Unauthenticated,
	ResourceExhausted: codes.ResourceExhausted,
	NotReady:          codes.FailedPrecondition,
	Aborted:           codes.Aborted,
	Unimplemented:     codes.Unimplemented,
	Unreachable:       codes.Unavailable,
}

var fromGRPCErrCode = map[codes.Code]ErrorCode{
	codes.Canceled:           Cancelled,
	codes.Unknown:            Unknown,
	codes.InvalidArgument:    InvalidArgument,
	codes.DeadlineExceeded:   Timeout,
	codes.NotFound:           NotFound,
	codes.AlreadyExists:      AlreadyExists,
	codes.PermissionDenied:   PermissionDenied,
	codes.Unauthenticated:    Unauthenticated,
	codes.ResourceExhausted:  ResourceExhausted,
	codes.FailedPrecondition: NotReady,
	codes.Aborted:            Aborted,
	codes.OutOfRange:         InvalidArgument,
	codes.Unimplemented:      Unimplemented,
	codes.Unavailable:        Unreachable,
}

func FromGRPCErr(err error) *Error {
	st := status.Convert(err)
	if st.Code() == codes.OK {
		return nil
	}

	message := st.Message()
	gRPCErrCode, ok := fromGRPCErrCode[st.Code()]
	if !ok {
		return NewErrorSkipCallers(Unknown, message, 1)
	}

	return NewErrorSkipCallers(gRPCErrCode, message, 1)
}

func ToGRPCErr(err *Error) error {
	if err == nil {
		return nil
	}

	gRPCErrCode, ok := toGRPCErrCode[err.Code]
	if !ok {
		return status.Error(codes.Unknown, err.Message)
	}

	return status.Error(gRPCErrCode, err.Message)
}
