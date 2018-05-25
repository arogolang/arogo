package mysqldb

import (
	"fmt"
	"strings"

	"github.com/arogolang/arogo/model"
	"github.com/arogolang/arogo/util"
)

func (p *MySqlDB) InitTables() (err error) {
	sqlAll, err := util.ReadFileToString("pool.sql")
	if err != nil {
		return err
	}

	sqlLine := strings.Split(sqlAll, ";")
	for _, sql := range sqlLine {
		_, err = p.db.Exec(sql)
		if err != nil {
			break
		}
	}

	return
}

func (p *MySqlDB) CheckTables(db string, table string) (exists bool, err error) {
	var count int64
	sql := fmt.Sprintf(`SELECT count(*) FROM information_schema.tables WHERE table_schema = '%v' AND table_name = '%v' LIMIT 1;`, db, table)

	err = p.db.Get(&count, sql)
	if err == nil && count > 0 {
		exists = true
	}

	return
}

func (p *MySqlDB) UpdateWorkerHashRate(worker string, addr string, hr int64, ip string) (err error) {
	_, err = p.db.Exec("INSERT into workers SET id=?, hashrate=?,updated=UNIX_TIMESTAMP(), miner=?, ip=? ON DUPLICATE KEY UPDATE updated=UNIX_TIMESTAMP(), hashrate=?, ip=?",
		worker, hr, addr, ip, hr, ip)
	return
}

func (p *MySqlDB) GetIPRejectCount(ip string) (count int64, err error) {
	err = p.db.Get(&count, "SELECT COUNT(1) FROM rejects WHERE ip=? AND data>UNIX_TIMESTAMP()-20", ip)
	return
}

func (p *MySqlDB) GetNonceCount(nonce string) (cound int64, err error) {
	err = p.db.Get(&count, "SELECT count(1) FROM nonces WHERE nonce=?", nonce)
	return
}

func (p *MySqlDB) InsertAbUser(miner string, nonce string) (err error) {
	_, err = p.db.Exec("INSERT into abusers SET miner=?, nonce=?", miner, nonce)

	return
}

func (p *MySqlDB) InsertNonce(nonce string) (err error) {
	_, err = p.db.Exec("INSERT IGNORE into nonces SET nonce=?", nonce)

	return
}

func (p *MySqlDB) UpdateDBWithRejectNonce(nonce string, ip string) (err error) {
	_, err = p.db.Exec("DELETE FROM nonces WHERE nonce=?", nonce)
	_, err = p.db.Exec("INSERT into rejects SET ip=?, data=UNIX_TIMESTAMP() ON DUPLICATE KEY update data=UNIX_TIMESTAMP()", ip)

	return
}

func (p *MySqlDB) InsertMinersNewShare(addr string, share int64, bdl int64) (err error) {
	_, err = p.db.Exec("INSERT INTO miners SET  id=?, shares=shares+?, updated=UNIX_TIMESTAMP(),bestdl=? ON DUPLICATE KEY UPDATE shares=shares+?, updated=UNIX_TIMESTAMP()",
		addr, share, bdl, share)

	_, err = p.db.Exec("UPDATE miners SET bestdl=? WHERE id=? AND bestdl>?",
		bdl, addr, bdl)

	return
}

func (p *MySqlDB) GetCurrentBlock() (block model.NodeBlock, err error) {
	err = p.db.Get(&block, "SELECT * FROM blocks ORDER by height DESC LIMIT 1")
	return
}

func (p *MySqlDB) GetBlockWithId(id string) (count int64, err error) {
	err = p.db.Get(&count, "SELECT COUNT(1) FROM blocks WHERE id=?", id)
	return
}

func (p *MySqlDB) GetLastBlockHeight() (count int64, err error) {
	err = p.db.Get(&count, "SELECT height FROM blocks ORDER by height DESC LIMIT 1")
	return
}

func (p *MySqlDB) GetBlockValWithId(id string) (count float64, err error) {
	err = p.db.Get(&count, "SELECT val FROM transactions WHERE block=? AND version=0", id)
	return
}

func (p *MySqlDB) GetMiners() (miners []model.Miners, err error) {
	err = p.db.Select(&miners, "SELECT * FROM miners WHERE shares>0 OR historic>0")
	return
}

func (p *MySqlDB) GetAllMiners() (miners []model.Miners, err error) {
	err = p.db.Select(&miners, "SELECT * FROM miners")
	return
}

func (p *MySqlDB) GetLatestBlocks(limit int) (blocks []model.Block, err error) {
	err = p.db.Select(&blocks, "SELECT * FROM blocks ORDER by height DESC LIMIT ?", limit)
	return
}

func (p *MySqlDB) GetLatestPayments(limit int) (payments []model.Payment, err error) {
	err = p.db.Select(&payments, "SELECT id,address,val,done,txn FROM payments ORDER by id DESC LIMIT ?", limit)
	return
}

func (p *MySqlDB) InsertPayment(block string, addr string, val float64, height int64) (err error) {
	_, err = p.db.Exec("INSERT into payments SET address=?, block=?, height=?, val=?, txn='',done=0", addr, block, height, val)
	return
}

func (p *MySqlDB) InsertPoolBlock(block string, addr string, height int64, reward float64) (err error) {
	_, err = p.db.Exec("INSERT IGNORE into blocks SET reward=?, id=?, height=?, miner=?", reward, block, height, addr)
	return
}

func (p *MySqlDB) GetInfoVal(name string) (val string, err error) {
	err = p.db.Get(&val, "SELECT val FROM info WHERE id=?);", name)
	return
}
