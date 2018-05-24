package middleware

import (
	"expvar"
	"net/http"
)

var (
	concNum = expvar.NewInt("concurrentNum")
)

type MiddleWareHandle interface {
	WrapHandler(handle func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request)
}

type MuxWrap struct {
	Mls []MiddleWareHandle
	Mux *http.ServeMux
}

func NewMuxWrap() *MuxWrap {
	return &MuxWrap{
		Mls: []MiddleWareHandle{},
		Mux: http.NewServeMux(),
	}
}

func (e *MuxWrap) Post(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	h := e.processMethods([]string{"POST"}, handler)
	e.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		h(writer, request)
		request.Body.Close()
	})
}

func (e *MuxWrap) Get(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	h := e.processMethods([]string{"GET"}, handler)
	e.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		h(writer, request)
		request.Body.Close()
	})
}

func (e *MuxWrap) Delete(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	h := e.processMethods([]string{"DELETE"}, handler)
	e.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		h(writer, request)
		request.Body.Close()
	})
}

func (e *MuxWrap) HandleFuncMethods(methods []string, pattern string, handler func(http.ResponseWriter, *http.Request)) {
	h := e.processMethods(methods, handler)
	e.HandleFunc(pattern, h)
}

func (e *MuxWrap) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	h := e.processMiddleWare(handler)
	e.Mux.HandleFunc(pattern, h)
}

func (e *MuxWrap) WSHandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	e.Mux.HandleFunc(pattern, handler)
}

func (e *MuxWrap) Handle(pattern string, handler http.Handler) {
	e.Mux.Handle(pattern, handler)
}

func (e *MuxWrap) AddMiddleWare(m MiddleWareHandle) {
	e.Mls = append(e.Mls, m)
}

func (e *MuxWrap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.Mux.ServeHTTP(w, r)
}

func (e *MuxWrap) processMiddleWare(handle func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	next := handle
	for _, m := range e.Mls {
		next = m.WrapHandler(next)
	}

	return next
}

func (e *MuxWrap) processMethods(methods []string, handle func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Write([]byte{})
			return
		}
		allow := false
		for _, m := range methods {
			if r.Method == m {
				allow = true
				break
			}
		}
		if allow {
			handle(w, r)
		} else {
			methodNotAllowResp(w, r)
		}
	}
}

func methodNotAllowResp(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "method not allow", http.StatusMethodNotAllowed)
}
