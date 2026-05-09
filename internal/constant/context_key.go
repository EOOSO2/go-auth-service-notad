package constant

type ContextKey string

const (
	CtxUserID     ContextKey = "user_id"
	CtxEmailID    ContextKey = "email_id"
	CtxEmail      ContextKey = "email"
	CtxPermission ContextKey = "permission"
)

const (
	PermAdmin   = "ADMIN"
	PermTeacher = "TEACHER"
	PermStudent = "STUDENT"
)
