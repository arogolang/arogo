package pool

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/arogolang/arogo/config"
	"github.com/arogolang/arogo/errlog"
	"github.com/arogolang/arogo/util"
)

var (
	currentHeight int64
)

func (s *PoolWebService) CheckBlocks() error {
	height, err := s.dbMgr.NodeDB.GetLastBlockHeight()
	if err != nil {
		return err
	}

	if height > 0 && currentHeight != height {
		currentHeight = height
		_, err = s.dbMgr.PoolDB.Exec("UPDATE miners SET historic=historic*0.95+shares, shares=0,bestdl=1000000")
		_, err = s.dbMgr.PoolDB.Exec("TRUNCATE table nonces")

		var totalHR int64

		r, err := s.dbMgr.PoolDB.GetMinersWithHistoric()
		if err != nil {
			return err
		}

		for _, x := range r {
			thr, err := s.dbMgr.PoolDB.GetMinerHashRate(x.ID)
			if err == nil {
				if x.Historic/thr < 2 {
					thr = 0
				}

				totalHR += thr
			}
		}

		s.dbMgr.PoolDB.Exec("UPDATE info SET val=? WHERE id='total_hash_rate'", totalHR)
	}

	return err
}

func (s *PoolWebService) PayOnce() error {
	now := time.Now()
	hour := now.Hour()
	min := now.Minute()

	var blockPaid int64 = 500
	if hour == 10 && min < 20 {
		blockPaid = 5000
	}

	current, err := s.dbMgr.NodeDB.GetLastBlockHeight()
	if err != nil {
		return err
	}

	s.dbMgr.PoolDB.Exec("DELETE FROM miners WHERE historic+shares<=20")
	s.dbMgr.PoolDB.Exec("UPDATE miners SET hashrate=(SELECT SUM(hashrate) FROM workers WHERE miner=miners.id AND updated>UNIX_TIMESTAMP()-3600)")
	s.dbMgr.PoolDB.Exec("UPDATE miners SET pending=(SELECT SUM(val) FROM payments WHERE done=0 AND payments.address=miners.id AND height>=?", current-blockPaid)

	r, err := s.dbMgr.PoolDB.GetPendingPayBlocks(current-blockPaid, current-10)
	if r == nil || len(r) == 0 {
		errlog.Error("No payments pending")
		return err
	}

	for _, x := range r {
		bExist, _ := s.dbMgr.NodeDB.GetBlockWithId(x)
		if bExist == 0 {
			s.dbMgr.PoolDB.Exec("DELETE FROM blocks WHERE id=?", x)
			s.dbMgr.PoolDB.Exec("DELETE FROM payments WHERE block=?", x)
		}
	}

	cfg := config.Get()

	var totalPaid float64
	pays, err := s.dbMgr.PoolDB.GetPendingPayments(current-blockPaid, current-10)
	for _, x := range pays {
		if x.Val < cfg.MinPayout {
			continue
		}

		fee := x.Val * 0.0025
		if fee < 0.00000001 {
			fee = 0.00000001
		}

		if fee > 10 {
			fee = 10
		}

		val := fmt.Sprintf("%.8f", x.Val-fee)

		form := url.Values{}
		form.Add("coin", "arionum")
		form.Add("dst", x.Address)
		form.Add("val", val)
		form.Add("private_key", cfg.PrivateKey)
		form.Add("public_key", cfg.PublicKey)

		ok, data, err := util.PostDataToNode(cfg.NodeAddrURL+"/api.php?q=send", form.Encode())
		if !ok {
			errlog.Errorf("post node payment error %v", err)
		} else {
			totalPaid += x.Val
		}

		s.dbMgr.PoolDB.Exec("UPDATE payments SET txn=?, done=1 WHERE address=? AND height<? AND done=0 AND height>=?",
			data, x.Address, current-blockPaid, current-10)

		s.dbMgr.PoolDB.Exec("UPDATE miners  SET total_paid=total_paid + ? WHERE id=?", x.Val, x.Address)

	}

	oldTotalPaidStr, err := s.dbMgr.PoolDB.GetInfoVal("total_paid")
	oldTotalPaid, err := strconv.ParseFloat(oldTotalPaidStr, 64)
	oldTotalPaid += totalPaid

	s.dbMgr.PoolDB.Exec("UPDATE info SET val=? WHERE id='total_paid'", oldTotalPaid)

	s.dbMgr.PoolDB.Exec("UPDATE miners SET pending=(SELECT SUM(val) FROM payments WHERE done=0 AND payments.address=miners.id AND height>=?)",
		current-blockPaid)

	s.dbMgr.PoolDB.Exec("DELETE FROM payments WHERE done=1 AND height = ?", current-1000)

	return err
}

func (s *PoolWebService) MonitorBlocks() {
	statUpdate := time.NewTicker(6 * time.Second)
	statPayment := time.NewTicker(60 * 4 * time.Second)

	defer statUpdate.Stop()
	defer statPayment.Stop()

	for {
		select {
		case <-statUpdate.C:
			s.CheckBlocks()
		case <-statPayment.C:
			s.PayOnce()
		}
	}
}
