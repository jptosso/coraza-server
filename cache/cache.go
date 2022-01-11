package cache

import (
	"fmt"
	"time"

	"github.com/ReneKroon/ttlcache/v2"
	"github.com/jptosso/coraza-waf/v2"
)

type TxCache struct {
	transactions *ttlcache.Cache
}

func (ts *TxCache) Get(id string) *coraza.Transaction {
	tx, err := ts.transactions.Get(id)
	if err != nil {
		return nil
	}
	return tx.(*coraza.Transaction)
}

func (ts *TxCache) Store(tx *coraza.Transaction) error {
	return ts.transactions.Set(tx.ID, tx)
}

func (ts *TxCache) SetTransactionTtl(ttl int) {
	ts.transactions.SetTTL(time.Duration(ttl) * time.Second)
}

func (ts *TxCache) Expire(txid string) error {
	if err := ts.transactions.Remove(txid); err != nil {
		return err
	}
	return nil
}

func NewTxCache(ttl int, limit int) *TxCache {
	c := ttlcache.NewCache()
	c.SetCacheSizeLimit(limit)
	c.SetTTL(time.Duration(ttl) * time.Second)
	c.SetExpirationCallback(expirationCallback)
	return &TxCache{
		transactions: c,
	}
}

func expirationCallback(key string, value interface{}) {
	tx, ok := value.(*coraza.Transaction)
	if !ok {
		fmt.Println("Failed to expire transaction " + key)
		return
	}
	tx.ProcessLogging()
	if err := tx.Clean(); err != nil {
		fmt.Println("Failed to clean transaction " + key)
	}
	fmt.Printf("Transaction %s forced to die\n", tx.ID)
}
