package model

type Block struct {
	ID        string  `db:"id" json:"id"`
	Height    int64   `db:"height" json:"height"`
	Miner     string  `db:"miner" json:"miner"`
	Reward    float64 `db:"reward"`
	RewardStr string  `json:"reward"`
}

type Miners struct {
	ID        string  `db:"id"`
	Shares    int64   `db:"shares"`
	Historic  int64   `db:"historic"`
	TotalPaid float64 `db:"total_paid"`
	Updated   int64   `db:"updated"`
	BestDL    int64   `db:"bestdl"`
	Pending   float64 `db:"pending"`
	HashRate  int64   `db:"hashrate"`
}

type Payment struct {
	ID      int64   `db:"id" json:"id"`
	Address string  `db:"address" json:"address"`
	Val     float64 `db:"val" json:"val"`
	Txn     string  `db:"txn" json:"txn"`
	Block   string  `db:"block"`
	Done    int64   `db:"done"`
	Height  int64   `db:"height"`
}

type NodeBlock struct {
	ID           string `db:"id"`
	Generator    string `db:"generator"`
	Height       int64  `db:"height"`
	Date         int64  `db:"date"`
	Nonce        string `db:"nonce"`
	Signature    string `db:"signature"`
	Difficulty   string `db:"difficulty"`
	Argon        string `db:"argon"`
	Transactions int    `db:"transactions"`
}
