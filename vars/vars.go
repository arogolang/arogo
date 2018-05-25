package vars

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/arogolang/arogo/mysqldb"
)

type CurrentBlockInfo struct {
	Diffculty string `json:"difficulty"`
	Block     string `json:"block"`
	Height    int64  `json:"height"`
}

type CurrentMineInfo struct {
	CurrentBlockInfo

	PublicKey string `json:"public_jkey"`
	Limit     int64  `json:"limit"`
}

type PoolDBMgr struct {
	PoolDB *mysqldb.MySqlDB
	NodeDB *mysqldb.MySqlDB
}

var GlobalBlockInfo CurrentBlockInfo

func UpdateCurrentBlock() error {
	hc := http.Client{}
	req, err := http.NewRequest("GET", "http://91.92.108.149:9090/api.php?q=currentBlock", nil)
	if err != nil {
		return err
	}

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}

	if resp.Body == nil {
		return nil
	}

	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return readErr
	}

	var o = make(map[string]*json.RawMessage)
	if err := json.Unmarshal(body, &o); err != nil {
		return errors.New("bad request")
	}

	_, okData := o["data"]
	if okData && o["data"] != nil {

		oe := make(map[string]*json.RawMessage)
		if err := json.Unmarshal(*o["data"], &oe); err != nil {
			return errors.New("bad response: " + string(*o["data"]))
		}

		if oe["difficulty"] == nil || oe["id"] == nil || oe["height"] == nil {
			return errors.New("bad response: " + string(*o["data"]))
		}

		GlobalBlockInfo.Block = string(*oe["id"])
		GlobalBlockInfo.Diffculty = string(*oe["difficulty"])
		GlobalBlockInfo.Height, _ = strconv.ParseInt(string(*oe["height"]), 10, 0)
	}

	return nil
}
