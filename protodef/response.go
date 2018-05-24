package protodef

import (
	"encoding/json"
	"io"
	"net/http"
)

func WriteData(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(data)
}

type ResponseAPI struct {
	Status string      `json:"status"`
	Coin   string      `json:"coin"`
	Data   interface{} `json:"data"`
}

func APIEcho(w http.ResponseWriter, data interface{}) error {
	resp := ResponseAPI{
		Status: "ok",
		Coin:   "arionum",
		Data:   data,
	}
	return WriteData(w, resp)
}

func APIError(w http.ResponseWriter, data interface{}) error {
	resp := ResponseAPI{
		Status: "error",
		Coin:   "arionum",
		Data:   data,
	}

	return WriteData(w, resp)
}
