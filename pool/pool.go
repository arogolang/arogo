package pool

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/arogolang/arogo/config"
	"github.com/arogolang/arogo/vars"

	"github.com/arogolang/arogo/errlog"
	ml "github.com/arogolang/arogo/middleware"
	"github.com/arogolang/arogo/protodef"
	"github.com/arogolang/arogo/util"

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
	tplHeader, _ = readTplFile("./web/template/header.html")
	tplFootor, _ = readTplFile("./web/template/footer.html")

	tplIndex, _ = readTplFile("./web/template/index.html")
	tplBlocks, _ = readTplFile("./web/template/blocks.html")
	tplBenchmarks, _ = readTplFile("./web/template/benchmarks.html")
	tplInfo, _ = readTplFile("./web/template/info.html")
	tplPayment, _ = readTplFile("./web/template/payments.html")
}

type DataAjaxShare struct {
	Shares   int64  `json:"shares"`
	Historic int64  `json:"historic"`
	Percent  string `json:"percent"`
	Bestdl   int64  `json:"bestdl"`
	Id       string `json:"id"`
}

type DataAjaxShareHis struct {
	DataAjaxShare
	Hashrate  int64  `json:"hashrate"`
	Pending   string `json:"pending"`
	Totalpaid string `json:"totalpaid"`
}

type DataAjaxIndex struct {
	Height        int64  `json:"height"`
	Lastwon       int64  `json:"lastwon"`
	Miners        int    `json:"miners"`
	Difficulty    int64  `json:"difficulty"`
	Totalpaid     string `json:"totalpaid"`
	Totalhr       string `json:"totalhr"`
	Avghr         string `json:"avghr"`
	Totalshares   int64  `json:"totalshares"`
	Totalhistoric int64  `json:"totalhistoric"`

	Shares    []DataAjaxShare    `json:"shares"`
	Historics []DataAjaxShareHis `json:"historics"`
}

type PoolWebService struct {
	dbMgr *vars.PoolDBMgr
}

func (s *PoolWebService) HandleAjax(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.FormValue("q")
	switch q {
	case "blocks":
		blocks, err := s.dbMgr.PoolDB.GetLatestBlocks(100)
		if err != nil {
			errlog.Error(err)
		}

		for i, _ := range blocks {
			blocks[i].RewardStr = fmt.Sprintf("%.2f", blocks[i].Reward)
		}

		protodef.WriteData(w, blocks)

	case "payments":
		pays, err := s.dbMgr.PoolDB.GetLatestPayments(1000)
		if err != nil {
			errlog.Error(err)
		}

		for i, _ := range pays {
			if pays[i].Done == 0 {
				pays[i].Txn = "Pending"
			}
		}

		protodef.WriteData(w, pays)

	case "index":
		info := DataAjaxIndex{}

		r, err := s.dbMgr.PoolDB.GetAllMiners()
		if err != nil {
			errlog.Error(err)
		}

		current, err := s.dbMgr.NodeDB.GetCurrentBlock()
		if err != nil {
			errlog.Error(err)
		}

		lastWon, err := s.dbMgr.NodeDB.GetLastBlockHeight()
		if err != nil {
			errlog.Error(err)
		}

		info.Miners = len(r)

		var totalShares, totalHistoric int64

		for _, x := range r {
			totalHistoric += x.Historic
			totalShares += x.Shares
		}

		for _, x := range r {
			share := DataAjaxShare{}
			shareHis := DataAjaxShareHis{}

			share.Id = x.ID
			share.Shares = x.Shares
			share.Historic = x.Historic
			share.Bestdl = x.BestDL
			shareHis.DataAjaxShare = share

			share.Percent = fmt.Sprintf("%.2f", float64(x.Shares*100)/float64(totalShares))
			info.Shares = append(info.Shares, share)

			shareHis.Percent = fmt.Sprintf("%.2f", float64(x.Historic*100)/float64(totalHistoric))
			shareHis.Pending = fmt.Sprintf("%.2f", x.Pending)
			shareHis.Totalpaid = fmt.Sprintf("%.2f", x.TotalPaid)
			shareHis.Hashrate = x.HashRate
			info.Historics = append(info.Historics, shareHis)
		}

		totalHrStr, err := s.dbMgr.PoolDB.GetInfoVal("total_hash_rate")
		totalHr, err := strconv.Atoi(totalHrStr)

		if info.Miners > 0 {
			info.Avghr = fmt.Sprintf("%v", totalHr/info.Miners)
		}

		var total_hr_float float64

		if totalHr >= 1000000 {
			total_hr_float = float64(totalHr) / 1000000
			info.Totalhr = fmt.Sprintf("%.2f MH/s", total_hr_float)
		} else if totalHr >= 1000 {
			total_hr_float = float64(totalHr) / 1000
			info.Totalhr = fmt.Sprintf("%.2f KH/s", total_hr_float)
		} else {
			total_hr_float = float64(totalHr)
			info.Totalhr = fmt.Sprintf("%.2f H/s", total_hr_float)
		}

		info.Totalshares = totalShares
		info.Totalhistoric = totalHistoric
		info.Height = current.Height
		info.Difficulty, err = strconv.ParseInt(current.Difficulty, 10, 0)
		info.Difficulty = 200000000 - info.Difficulty
		info.Lastwon = lastWon

		totalPaiedStr, err := s.dbMgr.PoolDB.GetInfoVal("total_paid")
		totalPaid, err := strconv.ParseFloat(totalPaiedStr, 64)
		info.Totalpaid = fmt.Sprintf("%.2f", totalPaid/1000000)

		protodef.WriteData(w, info)

	default:
		protodef.APIError(w, "invalid command")
	}
}

func (s *PoolWebService) HandleMine(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("q")

	switch q {
	case "info":
		hashrate := r.FormValue("hashrate")
		if hashrate != "" {
			hashrateInt, err := strconv.ParseInt(hashrate, 10, 0)
			if err == nil && hashrateInt > 0 {
				miner := r.FormValue("address")
				worker := r.FormValue("worker")
				ip, _, _ := net.SplitHostPort(r.RemoteAddr)

				worker = util.MD5String(miner + worker + ip)
				s.dbMgr.PoolDB.UpdateWorkerHashRate(worker, miner, hashrateInt, ip)
			}
		}

		cfg := config.Get()

		mineInfo := &vars.CurrentMineInfo{}
		mineInfo.CurrentBlockInfo = vars.GlobalBlockInfo
		mineInfo.PublicKey = cfg.PublicKey
		mineInfo.Limit = cfg.Limit

		protodef.APIEcho(w, mineInfo)

	case "submitNonce":
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		reject, _ := s.dbMgr.PoolDB.GetIPRejectCount(ip)
		if reject > 0 {
			protodef.APIError(w, "rejected")
			break
		}

		nonce := r.FormValue("nonce")
		address := r.FormValue("address")

		chk, _ := s.dbMgr.PoolDB.GetNonceCount(nonce)
		if chk > 0 {
			s.dbMgr.PoolDB.InsertAbUser(address, nonce)
			protodef.APIError(w, "duplicate")
			break
		}

		s.dbMgr.PoolDB.InsertNonce(nonce)

		height, _ := strconv.ParseInt(r.FormValue("height"), 10, 0)
		if height > 1 && vars.GlobalBlockInfo.Height != height {
			protodef.APIError(w, "stale block")
			break
		}

		cfg := config.Get()

		argon := r.FormValue("argon")
		argon2 := `$argon2i$v=19$m=524288,t=1,p=1` + argon
		base := fmt.Sprintf("%v-%v-%v-%v", cfg.PublicKey, nonce, vars.GlobalBlockInfo.Block, vars.GlobalBlockInfo.Diffculty)

		if !util.Argon2Verify(base, argon) {
			protodef.APIError(w, "Invalid argon "+base+argon2)
			break
		}

		hash := base + argon2
		hash = util.GetAroHash(hash)
		m := strings.SplitN(hash, "", 2)

		m10, _ := strconv.ParseInt(m[10], 16, 0)
		m15, _ := strconv.ParseInt(m[15], 16, 0)
		m20, _ := strconv.ParseInt(m[20], 16, 0)
		m23, _ := strconv.ParseInt(m[23], 16, 0)
		m31, _ := strconv.ParseInt(m[31], 16, 0)
		m40, _ := strconv.ParseInt(m[40], 16, 0)
		m45, _ := strconv.ParseInt(m[45], 16, 0)
		m55, _ := strconv.ParseInt(m[55], 16, 0)

		duration := fmt.Sprintf("%v%v%v%v%v%v%v%v",
			m10, m15, m20, m23, m31, m40, m45, m55)
		mdiff, _ := strconv.ParseInt(vars.GlobalBlockInfo.Diffculty, 10, 0)

		if len(duration) > 15 || mdiff <= 0 {
			s.dbMgr.PoolDB.UpdateDBWithRejectNonce(nonce, ip)

			protodef.APIError(w, "Invalid argon "+base+argon2)
			break
		}

		mresult, _ := strconv.ParseInt(duration, 10, 0)
		result := mresult / mdiff

		if result > 0 && result <= 240 {
			ok, err := util.SubmitNonceToNode(cfg.NodeAddrURL, argon, nonce, cfg.PublicKey, cfg.PrivateKey)
			if err != nil {
				errlog.Error(err)
			} else if ok {
				bl, err := s.dbMgr.NodeDB.GetCurrentBlock()
				if err != nil {
					errlog.Error(err)
				}

				added, err := s.dbMgr.PoolDB.GetBlockWithId(bl.ID)
				if err != nil {
					errlog.Error(err)
				}

				if bl.Generator == cfg.Address && added == 0 {
					reward, err := s.dbMgr.NodeDB.GetBlockValWithId(bl.ID)
					if err != nil {
						errlog.Error(err)
					}

					if reward <= 0.001 {
						protodef.APIError(w, "something went wrong")
						break
					}

					original_reward := reward
					r, err := s.dbMgr.PoolDB.GetMiners()

					var total_shares, total_historic int64

					for _, x := range r {
						total_shares += x.Shares
						total_historic += x.Historic
					}

					reward = reward * (1 - cfg.Fee)
					miner_reward := reward * cfg.MinerReward
					his_reward := reward * cfg.HisReward
					cur_reward := reward * cfg.CurrentReward

					for _, x := range r {
						var crw float64

						if x.Shares > 0 {
							crw += cur_reward * float64(x.Shares) / float64(total_shares)
						}

						if x.Historic > 0 {
							crw += his_reward * float64(x.Historic) / float64(total_historic)
						}

						if x.ID == address {
							crw += miner_reward
						}

						s.dbMgr.PoolDB.InsertPayment(bl.ID, x.ID, crw, bl.Height)
					}

					s.dbMgr.PoolDB.InsertPayment(bl.ID, cfg.FeeAddress, original_reward*cfg.Fee, bl.Height)

					s.dbMgr.PoolDB.InsertPoolBlock(bl.ID, address, bl.Height, original_reward)

					protodef.APIEcho(w, "accepted")
				} else {
					protodef.APIError(w, "rejected - block changed - 2")
					break
				}
			} else {
				protodef.APIError(w, "rejected - block changed - 1")
				break
			}
		} else if result > 0 && result < cfg.Limit {
			share := (cfg.Limit - result) / 100
			s.dbMgr.PoolDB.InsertMinersNewShare(address, share, result)

			protodef.APIEcho(w, "accepted")
		} else {
			s.dbMgr.PoolDB.UpdateDBWithRejectNonce(nonce, ip)

			protodef.APIError(w, fmt.Sprintf("rejected - %v", result))
		}

	default:
		protodef.APIError(w, "invalid command")
	}
}

func (s *PoolWebService) HandleIndex(w http.ResponseWriter, r *http.Request) {
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

func NewPoolServer(poolAddr string, dbMgr *vars.PoolDBMgr) {
	errlog.Info("Starting web listener on : ", poolAddr)

	webService := &PoolWebService{
		dbMgr: dbMgr,
	}

	mux := ml.NewMuxWrap()

	// for profiling
	mux.Mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.Mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.Mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.Mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.Mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.Handle("/assets/", util.NoDirListing(http.FileServer(http.Dir("./web"))))

	mux.Mux.HandleFunc("/api.php", webService.HandleAjax)
	mux.Mux.HandleFunc("/mine.php", webService.HandleMine)
	mux.Get("/index.php", webService.HandleIndex)
	mux.Get("/", webService.HandleIndex)

	srv := &http.Server{
		Handler:      gziphandler.GzipHandler(mux),
		Addr:         poolAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go srv.ListenAndServe()
}
