package persist

type HandlerFunc func(*Context)
type Context struct {
	Errors errorMsg
}
