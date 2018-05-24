package pool

import (
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/arogolang/arogo/errlog"
	ml "github.com/arogolang/arogo/middleware"
	"github.com/arogolang/arogo/protodef"
	"github.com/arogolang/arogo/util"
	"github.com/arogolang/arogo/vars"

	"github.com/nytimes/gziphandler"
)

const VERSION = "0.3.0"

var (
	tplIndex      []byte
	tplBlocks     []byte
	tplFootor     []byte
	tplBenchmarks []byte
	tplHeader     []byte
	tplInfo       []byte
	tplPayment    []byte
)

func readTplFile(tplPath string) (data []byte, err error) {
	file, err := os.Open(tplPath)
	if err != nil {
		errlog.Errorf("Open error %s", err)
		return
	}

	data, err = ioutil.ReadAll(file)
	if err != nil {
		errlog.Errorf("read error %s", err)
		file.Close()
		return
	}

	file.Close()
	return
}

func init() {
	tplIndex, _ = readTplFile("./template/index.html")
	tplBlocks, _ = readTplFile("./template/blocks.html")
	tplFootor, _ = readTplFile("./template/footer.html")
	tplBenchmarks, _ = readTplFile("./template/benchmarks.html")
	tplHeader, _ = readTplFile("./template/header.html")
	tplInfo, _ = readTplFile("./template/info.html")
	tplPayment, _ = readTplFile("./template/payments.html")
}

func handleGetAddress(w http.ResponseWriter, r *http.Request) {
}

func HandleMine(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("q")

	switch q {
	case "info":
		mineInfo := &vars.CurrentMineInfo{}
		mineInfo.CurrentBlockInfo = vars.GlobalBlockInfo
		mineInfo.PublicKey = ""
		mineInfo.Limit = 10000

		protodef.APIEcho(w, mineInfo)

	case "submitNonce":
		// TODO
		protodef.APIEcho(w, "submitNonce")

	default:
		protodef.APIError(w, "invalid command")
	}
}

type DataAjaxShare struct {
	Shares   int     `json:"shares"`
	Historic int     `json:"historic"`
	Percent  float64 `json:"percent"`
	Bestdl   int     `json:"bestdl"`
	Id       string  `json:"id"`
}

type DataAjaxShareHis struct {
	DataAjaxShare
	Hashrate  int     `json:"hashrate"`
	Pending   float64 `json:"pending"`
	Totalpaid float64 `json:"totalpaid"`
}

type DataAjaxIndex struct {
	Height        int `json:"height"`
	Lastwon       int `json:"lastwon"`
	Miners        int `json:"miners"`
	Difficulty    int `json:"difficulty"`
	Totalpaid     int `json:"totalpaid"`
	Totalhr       int `json:"totalhr"`
	Avghr         int `json:"avghr"`
	Totalshares   int `json:"totalshares"`
	Totalhistoric int `json:"totalhistoric"`

	Shares    []DataAjaxShare    `json:"shares"`
	Historics []DataAjaxShareHis `json:"historics"`
}

type DataAjaxPayment struct {
	Id      string  `json:"id"`
	Address string  `json:"address"`
	Val     float64 `json:"val"`
	Txn     string  `json:"txn"`
}

type DataAjaxBlock struct {
	Id     string  `json:"id"`
	Height int     `json:"height"`
	Reward float64 `json:"reward"`
	Miner  string  `json:"miner"`
}

func HandleAjax(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.FormValue("q")
	switch q {
	case "blocks":
		blocks := []DataAjaxBlock{}
		protodef.WriteData(w, blocks)

	case "payments":
		payments := []DataAjaxPayment{}
		protodef.WriteData(w, payments)

	case "index":
		info := DataAjaxIndex{}
		protodef.WriteData(w, info)

	default:
		protodef.APIError(w, "invalid command")
	}
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("q")

	w.Write(tplHeader)

	switch q {
	case "blocks":
		w.Write(tplBlocks)

	case "payments":
		w.Write(tplPayment)

	case "benchmarks":
		w.Write(tplBenchmarks)

	case "info":
		w.Write(tplInfo)

	default:
		w.Write(tplIndex)
	}

	qData := `<script>var q="{QQ}";</script>`

	qData = strings.Replace(qData, "{QQ}", q, -1)
	w.Write([]byte(qData))

	w.Write(tplFootor)
}

func NewPoolServer(poolAddr string) {
	errlog.Info("Starting web listener on : ", poolAddr)

	mux := ml.NewMuxWrap()

	// for profiling
	mux.Mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.Mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.Mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.Mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.Mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.Get("/index.php", HandleIndex)
	mux.Mux.HandleFunc("/api.php", HandleAjax)
	mux.Mux.HandleFunc("/mine.php", HandleMine)

	mux.Handle("/assets/", util.NoDirListing(http.FileServer(http.Dir("./"))))
	mux.Mux.HandleFunc("/", HandleIndex)

	srv := &http.Server{
		Handler:      gziphandler.GzipHandler(mux),
		Addr:         poolAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go srv.ListenAndServe()
}
