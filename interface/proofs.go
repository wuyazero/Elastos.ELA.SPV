package _interface

import (
	"bytes"
	"errors"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/wuyazero/Elastos.ELA.SPV/log"

	. "github.com/wuyazero/Elastos.ELA/bloom"
	. "github.com/wuyazero/Elastos.ELA.Utility/common"
	"github.com/boltdb/bolt"
)

type Proofs interface {
	// Put a merkle proof of the block
	Put(proof *MerkleProof) error

	// Get a merkle proof of a block
	Get(blockHash *Uint256) (*MerkleProof, error)

	// Get all merkle proofs in database
	GetAll() ([]*MerkleProof, error)

	// Delete a merkle proof of a block
	Delete(blockHash *Uint256) error

	// Reset database, clear all data
	Reset() error

	// Close the proofs db
	Close()
}

type ProofsDB struct {
	*sync.RWMutex
	*bolt.DB
}

var (
	BKTProofs = []byte("Proofs")
)

func NewProofsDB() (Proofs, error) {
	db, err := bolt.Open("proofs.bin", 0644, &bolt.Options{InitialMmapSize: 5000000})
	if err != nil {
		return nil, err
	}

	db.Update(func(btx *bolt.Tx) error {
		_, err := btx.CreateBucketIfNotExists(BKTProofs)
		if err != nil {
			return err
		}
		return nil
	})

	return &ProofsDB{RWMutex: new(sync.RWMutex), DB: db}, nil
}

// Put a merkle proof of the block
func (db *ProofsDB) Put(proof *MerkleProof) error {
	db.Lock()
	defer db.Unlock()

	return db.Update(func(tx *bolt.Tx) error {

		bytes, err := serializeProof(proof)
		if err != nil {
			return err
		}

		err = tx.Bucket(BKTProofs).Put(proof.BlockHash.Bytes(), bytes)
		if err != nil {
			return err
		}

		return nil
	})
}

// Get a merkle proof of a block
func (db *ProofsDB) Get(blockHash *Uint256) (proof *MerkleProof, err error) {
	db.RLock()
	defer db.RUnlock()

	err = db.View(func(tx *bolt.Tx) error {

		proof, err = getProof(tx, blockHash.Bytes())
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return proof, err
}

// Get all merkle proofs in database
func (db *ProofsDB) GetAll() (proofs []*MerkleProof, err error) {
	db.RLock()
	defer db.RUnlock()

	err = db.View(func(tx *bolt.Tx) error {

		err := tx.Bucket(BKTProofs).ForEach(func(k, v []byte) error {

			proof, err := deserializeProof(v)
			if err != nil {
				return err
			}

			proofs = append(proofs, proof)

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	})

	return proofs, nil
}

// Delete a merkle proof of a block
func (db *ProofsDB) Delete(blockHash *Uint256) error {
	db.Lock()
	defer db.Unlock()

	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(BKTProofs).Delete(blockHash.Bytes())
	})
}

func (db *ProofsDB) Reset() error {
	db.Lock()
	defer db.Unlock()

	return db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(BKTProofs)
	})
}

// Close db
func (db *ProofsDB) Close() {
	db.Lock()
	db.DB.Close()
	log.Debug("Proofs DB closed")
}

func getProof(tx *bolt.Tx, key []byte) (*MerkleProof, error) {
	proofBytes := tx.Bucket(BKTProofs).Get(key)
	if proofBytes == nil {
		return nil, errors.New(fmt.Sprintf("MerkleProof %s does not exist in database", hex.EncodeToString(key)))
	}

	return deserializeProof(proofBytes)
}

func serializeProof(proof *MerkleProof) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := proof.Serialize(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeProof(body []byte) (*MerkleProof, error) {
	var proof MerkleProof
	err := proof.Deserialize(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return &proof, nil
}
