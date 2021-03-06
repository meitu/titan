package db

import (
	"context"
	"testing"

	"github.com/distributedio/titan/db/store"
	"github.com/pingcap/tidb/kv"
	"github.com/stretchr/testify/assert"
)

func TestGC(t *testing.T) {
	hashCall := func(t *testing.T, key []byte, count int64) []byte {
		hash, txn, err := getHash(t, []byte(key))
		assert.NoError(t, err)
		assert.NotNil(t, txn)
		assert.NotNil(t, hash)
		for count > 0 {
			encode, _ := EncodeInt64(count)
			hash.HSet(encode, []byte("val"))
			count--
		}
		kv := GetKv(txn)
		assert.NotNil(t, kv)
		c, err := kv.Delete([][]byte{key})
		assert.NoError(t, err)
		assert.Equal(t, c, int64(1))
		txn.Commit(context.TODO())
		return hash.meta.ID
	}

	type args struct {
		key        []byte
		fieldCount int64
		gcCount    int
		call       func(*testing.T, []byte, int64) []byte
	}
	type want struct {
		keyExists bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "TestGCIterVal",
			args: args{
				key:        []byte("TestGCHash1"),
				fieldCount: 10,
				gcCount:    5,
				call:       hashCall,
			},
			want: want{
				keyExists: true,
			},
		},
		{
			name: "TestGCAll",
			args: args{
				key:        []byte("TestGCHash2"),
				fieldCount: 10,
				gcCount:    17,
				call:       hashCall,
			},
			want: want{
				keyExists: false,
			},
		},
		{
			name: "TestLimitZero",
			args: args{
				key:        []byte("TestGCZeroHash3"),
				fieldCount: 10,
				gcCount:    0,
				call:       hashCall,
			},
			want: want{
				keyExists: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := clearGCData(t); err != nil {
				assert.NoError(t, err)
			}
			id := tt.args.call(t, tt.args.key, tt.args.fieldCount)
			txn := getTxn(t)
			doGC(txn.db, tt.args.gcCount)

			txn = getTxn(t)
			gcKey := toTiKVGCKey(toTiKVDataKey([]byte(txn.db.Namespace), txn.db.ID, id))

			_, err := txn.t.Get(txn.ctx, gcKey)
			txn.Commit(context.TODO())
			if tt.want.keyExists {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, true, store.IsErrNotFound(err))
			}
		})
	}

}

func clearGCData(t *testing.T) error {
	gcPrefix := toTiKVGCKey(nil)
	endGCPrefix := kv.Key(gcPrefix).PrefixNext()

	txn := getTxn(t)
	itr, err := txn.t.Iter(gcPrefix, endGCPrefix)
	if err != nil {
		return err
	}
	defer itr.Close()
	if !itr.Valid() || !itr.Key().HasPrefix(gcPrefix) {
		return nil
	}
	call := func(k kv.Key) bool {
		if resultErr := txn.t.Delete(k); resultErr != nil {
			return true
		}
		return false
	}
	if err := kv.NextUntil(itr, call); err != nil {
		return err
	}
	if err := txn.Commit(context.TODO()); err != nil {
		return err
	}
	return nil
}
