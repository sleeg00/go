package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type httpServer struct {
	Log *Log
}

func newHTTPServer() *httpServer {
	return &httpServer{
		Log: NewLog(),
	}
}

type ProduceRequest struct { //Clien가 record -> 배열 DB 한 줄이라고 생각
	Record Record `'json:"record"`
}

type ProduceResponse struct { //응답으로는 offset -> DB에 행열 번호를 준다고 생각
	Offset uint64 `json:"offset"`
}

type ConsumeRequest struct {
	Offset uint64 `json:"offset"`
}

type CosumeResponse struct {
	Record Record `json:"record"`
}

func NewHTTPServer(addr string) *http.Server {
	httpsrv := newHTTPServer()
	r := mux.NewRouter() //라우터 생성
	r.HandleFunc("/", httpsrv.handleProduce).Methods("POST")
	//HandleFunc(요청된 "/"-> Path로 어떤 Reuqest 핸들러 적용시킬지)
	r.HandleFunc("/", httpsrv.handleConsume).Methods("GET")
	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

func (s *httpServer) handleProduce(w http.ResponseWriter, r *http.Request) {
	//ResponseWriter -> Response 즉 응답을 남길수 있는 파라미터
	//Request는 당연히 요청을 볼 수 있는 것

	var req ProduceRequest                      //객체를 만든다는 느낌
	err := json.NewDecoder(r.Body).Decode(&req) //참조라서 역직렬화한 것을 저장하지 않아도 되는 듯
	//과정 설명 NewDecoder(json -> 디코더 생성) -> Decoder(Json -> Dto 생성)
	//요청 Json을 디코더 생성 그리고 역직렬화
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	off, err := s.Log.Append(req.Record)
	//Log에 응답추가

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ProduceResponse{Offset: off}  //Offset을 구조체에 담아 Response에 저장
	err = json.NewEncoder(w).Encode(res) //직렬화
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *httpServer) handleConsume(w http.ResponseWriter, r *http.Request) {
	var req ConsumeRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	record, err := s.Log.Read(req.Offset)

	if err == ErrOffsetNotFound {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := CosumeResponse{Record: record}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
