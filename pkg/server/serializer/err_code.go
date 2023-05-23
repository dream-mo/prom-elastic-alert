package serializer

import "net/http"

type ErrorStr string

const (
	ErrBindJson            ErrorStr = "INVALID_ARGUMENT.1" // 二级错误吗
	ErrGetFile                      = "INVALID_ARGUMENT.2"
	ErrReqStringToInt               = "INVALID_ARGUMENT.3"
	ErrEnvironment                  = "INVALID_ARGUMENT.4"
	ErrEnvironmentCodeNull          = "UNAUTHENTICATED.1"
	ErrNoSecret                     = "UNAUTHENTICATED.2"
	ErrDBOperate                    = "INTERNAL_DB"
	ErrK8sOperate                   = "INTERNAL_K8S"
	ErrUCOperate                    = "INTERNAL_UC"
	ErrJsonMarshal                  = "INTERNAL_Other.1"
	ErrJsonUnMarshal                = "INTERNAL_Other.2"
	ErrCopier                       = "INTERNAL_Other.3"
	ErrChangeLogLevel               = "INTERNAL_Other.4"
	ErrDBNotFind                    = "NOT_FOUND_DB"
	ErrTelnetFind                   = "NOT_FOUND_TELNET"
	ErrSnapshotFind                 = "NOT_FOUND_SNAPSHOT"
	ErrDBAlreadyExist               = "ALREADY_EXISTS_DB"
	ErrK8sAlreadyExist              = "ALREADY_EXISTS_K8S"
)

var ErrCodeMap = map[ErrorStr]int{
	ErrBindJson:            http.StatusBadRequest, //400
	ErrGetFile:             http.StatusBadRequest,
	ErrReqStringToInt:      http.StatusBadRequest,
	ErrEnvironment:         http.StatusBadRequest,
	ErrEnvironmentCodeNull: http.StatusUnauthorized, //401
	ErrNoSecret:            http.StatusUnauthorized,
	ErrDBNotFind:           http.StatusNotFound, //404
	ErrTelnetFind:          http.StatusNotFound,
	ErrSnapshotFind:        http.StatusNotFound,
	ErrDBAlreadyExist:      http.StatusConflict, //409
	ErrK8sAlreadyExist:     http.StatusConflict,
	ErrDBOperate:           http.StatusInternalServerError, //500
	ErrK8sOperate:          http.StatusInternalServerError,
	ErrUCOperate:           http.StatusInternalServerError,
	ErrJsonMarshal:         http.StatusInternalServerError,
	ErrCopier:              http.StatusInternalServerError,
	ErrChangeLogLevel:      http.StatusInternalServerError,
}
