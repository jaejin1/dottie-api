package errors

import "fmt"

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

var (
	ErrUnauthorized    = New("UNAUTHORIZED", "인증이 필요합니다")
	ErrForbidden       = New("FORBIDDEN", "권한이 없습니다")
	ErrNotFound        = New("NOT_FOUND", "리소스를 찾을 수 없습니다")
	ErrBadRequest      = New("BAD_REQUEST", "잘못된 요청입니다")
	ErrInternalServer  = New("INTERNAL_SERVER_ERROR", "서버 오류가 발생했습니다")
	ErrRoomNotFound    = New("ROOM_NOT_FOUND", "방을 찾을 수 없습니다")
	ErrUserNotFound    = New("USER_NOT_FOUND", "사용자를 찾을 수 없습니다")
	ErrDayLogNotFound  = New("DAYLOG_NOT_FOUND", "기록을 찾을 수 없습니다")
	ErrDotNotFound     = New("DOT_NOT_FOUND", "dot을 찾을 수 없습니다")
	ErrRoomFull        = New("ROOM_FULL", "방이 가득 찼습니다")
	ErrInvalidInvite   = New("INVALID_INVITE_CODE", "유효하지 않은 초대 코드입니다")
	ErrAlreadyMember   = New("ALREADY_MEMBER", "이미 방의 멤버입니다")
	ErrFileTooLarge    = New("FILE_TOO_LARGE", "파일 크기가 너무 큽니다 (최대 10MB)")
	ErrInvalidFileType = New("INVALID_FILE_TYPE", "지원하지 않는 파일 형식입니다")
)
