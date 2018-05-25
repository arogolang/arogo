package protodef

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/arogolang/arogo/config"
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
	cfg := config.Get()

	resp := ResponseAPI{
		Status: "ok",
		Coin:   cfg.CoinName,
		Data:   data,
	}
	return WriteData(w, resp)
}

func APIError(w http.ResponseWriter, data interface{}) error {
	cfg := config.Get()

	resp := ResponseAPI{
		Status: "error",
		Coin:   cfg.CoinName,
		Data:   data,
	}

	return WriteData(w, resp)
}
