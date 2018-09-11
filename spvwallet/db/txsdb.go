package db

import (
	"bytes"
	"database/sql"
	"math"
	"sync"

	"github.com/wuyazero/Elastos.ELA.SPV/db"

	. "github.com/wuyazero/Elastos.ELA/core"
	. "github.com/wuyazero/Elastos.ELA.Utility/common"
)

const CreateTXNDB = `CREATE TABLE IF NOT EXISTS TXNs(
				Hash BLOB NOT NULL PRIMARY KEY,
				Height INTEGER NOT NULL,
				RawData BLOB NOT NULL
			);`

type TxsDB struct {
	*sync.RWMutex
	*sql.DB
}

func NewTxsDB(db *sql.DB, lock *sync.RWMutex) (Txs, error) {
	_, err := db.Exec(CreateTXNDB)
	if err != nil {
		return nil, err
	}
	return &TxsDB{RWMutex: lock, DB: db}, nil
}

// Put a new transaction to database
func (t *TxsDB) Put(storeTx *db.StoreTx) error {
	t.Lock()
	defer t.Unlock()

	buf := new(bytes.Buffer)
	err := storeTx.Data.SerializeUnsigned(buf)
	if err != nil {
		return err
	}

	sql := `INSERT OR REPLACE INTO TXNs(Hash, Height, RawData) VALUES(?,?,?)`
	_, err = t.Exec(sql, storeTx.TxId.Bytes(), storeTx.Height, buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// Fetch a raw tx and it's metadata given a hash
func (t *TxsDB) Get(txId *Uint256) (*db.StoreTx, error) {
	t.RLock()
	defer t.RUnlock()

	row := t.QueryRow(`SELECT Height, RawData FROM TXNs WHERE Hash=?`, txId.Bytes())
	var height uint32
	var rawData []byte
	err := row.Scan(&height, &rawData)
	if err != nil {
		return nil, err
	}
	var tx Transaction
	err = tx.DeserializeUnsigned(bytes.NewReader(rawData))
	if err != nil {
		return nil, err
	}

	return &db.StoreTx{TxId: *txId, Height: height, Data: tx}, nil
}

// Fetch all transactions from database
func (t *TxsDB) GetAll() ([]*db.StoreTx, error) {
	return t.GetAllFrom(math.MaxUint32)
}

// Fetch all transactions from the given height
func (t *TxsDB) GetAllFrom(height uint32) ([]*db.StoreTx, error) {
	t.RLock()
	defer t.RUnlock()

	sql := "SELECT Hash, Height, RawData FROM TXNs"
	if height != math.MaxUint32 {
		sql += " WHERE Height=?"
	}
	var txns []*db.StoreTx
	rows, err := t.Query(sql, height)
	if err != nil {
		return txns, err
	}
	defer rows.Close()

	for rows.Next() {
		var txIdBytes []byte
		var height uint32
		var rawData []byte
		err := rows.Scan(&txIdBytes, &height, &rawData)
		if err != nil {
			return txns, err
		}

		txId, err := Uint256FromBytes(txIdBytes)
		if err != nil {
			return txns, err
		}

		var tx Transaction
		err = tx.DeserializeUnsigned(bytes.NewReader(rawData))
		if err != nil {
			return nil, err
		}

		txns = append(txns, &db.StoreTx{TxId: *txId, Height: height, Data: tx})
	}

	return txns, nil
}

// Update the height of a transaction
func (t *TxsDB) UpdateHeight(txId *Uint256, height uint32) error {
	t.Lock()
	defer t.Unlock()

	_, err := t.Exec("UPDATE TXNs SET Height=? WHERE Hash=?", height, txId.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// Delete a transaction from the db
func (t *TxsDB) Delete(txId *Uint256) error {
	t.Lock()
	defer t.Unlock()

	_, err := t.Exec("DELETE FROM TXNs WHERE Hash=?", txId.Bytes())
	if err != nil {
		return err
	}

	return nil
}
