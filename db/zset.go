package db

import (
	"encoding/binary"
	"strconv"
	"time"

	"github.com/pingcap/tidb/kv"

	"go.uber.org/zap"
)

// ZSetMeta is the meta data of the sorted set
type ZSetMeta struct {
	Object
	Len int64
}

// ZSet implements the the sorted set
type ZSet struct {
	meta ZSetMeta
	key  []byte
	txn  *Transaction
}

type MemberScore struct {
	Member string
	Score  float64
}

const byteScoreLen = 8

func newZSet(txn *Transaction, key []byte) *ZSet {
	now := Now()
	return &ZSet{
		txn: txn,
		key: key,
		meta: ZSetMeta{
			Object: Object{
				ID:        UUID(),
				CreatedAt: now,
				UpdatedAt: now,
				ExpireAt:  0,
				Type:      ObjectZSet,
				Encoding:  ObjectEncodingHT,
			},
			Len: 0,
		},
	}
}

// GetZSet returns a sorted set, create new one if don't exists
func GetZSet(txn *Transaction, key []byte) (*ZSet, error) {
	zset := newZSet(txn, key)

	mkey := MetaKey(txn.db, key)
	start := time.Now()
	meta, err := txn.t.Get(txn.ctx, mkey)
	zap.L().Debug("zset get metaKey", zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))
	if err != nil {
		if IsErrNotFound(err) {
			return zset, nil
		}
		return nil, err
	}

	obj, err := DecodeObject(meta)
	if err != nil {
		return nil, err
	}
	if IsExpired(obj, Now()) {
		return zset, nil
	}
	if obj.Type != ObjectZSet {
		return nil, ErrTypeMismatch
	}

	m := meta[ObjectEncodingLength:]
	if len(m) != 8 {
		return nil, ErrInvalidLength
	}
	zset.meta.Object = *obj
	zset.meta.Len = int64(binary.BigEndian.Uint64(m[:8]))

	return zset, nil
}

func (zset *ZSet) ZAdd(members [][]byte, scores []float64) (int64, error) {
	added := int64(0)

	oldValues := make([][]byte, len(members))
	var err error
	if zset.meta.Len > 0 {
		start := time.Now()
		oldValues, err = zset.MGet(members)
		zap.L().Debug("zset mget", zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))
		if err != nil {
			return 0, err
		}
	}

	dkey := DataKey(zset.txn.db, zset.meta.ID)
	var found bool
	var start time.Time
	costDel, costSetMem, costSetScore := int64(0), int64(0), int64(0)
	for i := range members {
		found = false
		if oldValues[i] != nil {
			found = true
			oldScore := DecodeFloat64(oldValues[i])
			if scores[i] == oldScore {
				continue
			}
			oldScoreKey := zsetScoreKey(dkey, oldValues[i], members[i])
			start = time.Now()
			err = zset.txn.t.Delete(oldScoreKey)
			costDel += time.Since(start).Nanoseconds()
			if err != nil {
				return added, err
			}
		}
		memberKey := zsetMemberKey(dkey, members[i])
		bytesScore, err := EncodeFloat64(scores[i])
		if err != nil {
			return 0, err
		}
		start = time.Now()
		err = zset.txn.t.Set(memberKey, bytesScore)
		costSetMem += time.Since(start).Nanoseconds()
		if err != nil {
			return added, err
		}

		scoreKey := zsetScoreKey(dkey, bytesScore, members[i])
		start = time.Now()
		err = zset.txn.t.Set(scoreKey, NilValue)
		costSetScore += time.Since(start).Nanoseconds()
		if err != nil {
			return added, err
		}

		if !found {
			added += 1
		}
	}
	zap.L().Debug("zset cost(us)", zap.Int64("del oldScoreKey", costDel/1000),
		zap.Int64("set memberKey", costSetMem/1000),
		zap.Int64("set scoreKey", costSetScore/1000))

	zset.meta.Len += added
	start = time.Now()
	if err = zset.updateMeta(); err != nil {
		return 0, err
	}
	zap.L().Debug("zset update meta key", zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))

	return added, nil
}

func (zset *ZSet) MGet(members [][]byte) ([][]byte, error) {
	ikeys := make([][]byte, len(members))
	dkey := DataKey(zset.txn.db, zset.meta.ID)
	for i := range members {
		ikeys[i] = zsetMemberKey(dkey, members[i])
	}

	return BatchGetValues(zset.txn, ikeys)
}

func (zset *ZSet) updateMeta() error {
	meta := zset.encodeMeta(zset.meta)
	return zset.txn.t.Set(MetaKey(zset.txn.db, zset.key), meta)
}

func (zset *ZSet) encodeMeta(meta ZSetMeta) []byte {
	b := EncodeObject(&meta.Object)
	m := make([]byte, 8)
	binary.BigEndian.PutUint64(m[:8], uint64(meta.Len))
	return append(b, m...)
}

func (zset *ZSet) Exist() bool {
	return zset.meta.Len != 0
}

func (zset *ZSet) ZAnyOrderRange(start int64, stop int64, withScore bool, positiveOrder bool) ([][]byte, error) {
	if stop < 0 {
		if stop = zset.meta.Len + stop; stop < 0 {
			return [][]byte{}, nil
		}
	} else if stop >= zset.meta.Len {
		stop = zset.meta.Len - 1
	}
	if start < 0 {
		if start = zset.meta.Len + start; start < 0 {
			start = 0
		}
	}
	// return 0 elements
	if start > stop || start >= zset.meta.Len {
		return [][]byte{}, nil
	}
	dkey := DataKey(zset.txn.db, zset.meta.ID)
	scorePrefix := ZSetScorePrefix(dkey)
	var iter Iterator
	var err error
	startTime := time.Now()

	upperBoundKey := kv.Key(scorePrefix).PrefixNext()
	if positiveOrder {
		iter, err = zset.txn.t.Iter(scorePrefix, upperBoundKey)
	} else {
		iter, err = zset.txn.t.IterReverse(upperBoundKey)
	}
	zap.L().Debug("zset seek", zap.Int64("cost(us)", time.Since(startTime).Nanoseconds()/1000))

	if err != nil {
		return nil, err
	}

	var items [][]byte
	cost := int64(0)
	for i := int64(0); err == nil && i <= stop && iter.Valid() && iter.Key().HasPrefix(scorePrefix); i++ {
		if i >= start {
			if len(iter.Key()) <= len(scorePrefix)+byteScoreLen+len(":") {
				zap.L().Error("score&member's length isn't enough to be decoded",
					zap.ByteString("meta key", zset.key), zap.ByteString("data key", iter.Key()))
				startTime = time.Now()
				err = iter.Next()
				cost += time.Since(startTime).Nanoseconds()
				continue
			}

			scoreAndMember := iter.Key()[len(scorePrefix):]
			score := scoreAndMember[0:byteScoreLen]
			member := scoreAndMember[byteScoreLen+len(":"):]
			items = append(items, member)
			if withScore {
				val := []byte(strconv.FormatFloat(DecodeFloat64(score), 'f', -1, 64))
				items = append(items, val)
			}
		}

		startTime = time.Now()
		err = iter.Next()
		cost += time.Since(startTime).Nanoseconds()
	}
	zap.L().Debug("zset all next", zap.Int64("cost(us)", cost/1000))

	return items, nil
}

// ZAnyOrderRangeByScore returns the items of a zset in specific order
func (zset *ZSet) ZAnyOrderRangeByScore(startScore float64, startInclude bool,
	stopScore float64, stopInclude bool,
	withScore bool,
	offset int64, count int64,
	positiveOrder bool) ([][]byte, error) {
	if positiveOrder && startScore > stopScore {
		return nil, nil
	}
	if !positiveOrder && startScore < stopScore {
		return nil, nil
	}
	if startScore == stopScore && (!startInclude || !stopInclude) {
		return nil, nil
	}
	if offset < 0 || count == 0 {
		return nil, nil
	}

	dkey := DataKey(zset.txn.db, zset.meta.ID)
	scorePrefix := ZSetScorePrefix(dkey)

	startPrefix := make([]byte, len(scorePrefix)+byteScoreLen)
	copy(startPrefix, scorePrefix)
	byteStartScore, err := EncodeFloat64(startScore)
	if err != nil {
		return nil, err
	}
	copy(startPrefix[len(scorePrefix):], byteStartScore)

	stopPrefix := make([]byte, len(scorePrefix)+byteScoreLen)
	copy(stopPrefix, scorePrefix)
	byteStopScore, err := EncodeFloat64(stopScore)
	if err != nil {
		return nil, err
	}
	copy(stopPrefix[len(scorePrefix):], byteStopScore)

	var iter Iterator

	if positiveOrder {
		upperBoundKey := kv.Key(stopPrefix).PrefixNext()
		iter, err = zset.txn.t.Iter(startPrefix, upperBoundKey)
	} else {
		upperBoundKey := kv.Key(startPrefix).PrefixNext()
		iter, err = zset.txn.t.IterReverse(upperBoundKey)
	}
	if err != nil {
		return nil, err
	}

	var items [][]byte
	countN := int64(0)
	startComFinished := false
	for i := int64(0); err == nil && iter.Valid() && iter.Key().HasPrefix(scorePrefix); i, err = i+1, iter.Next() {
		key := iter.Key()
		if len(key) <= len(scorePrefix)+byteScoreLen+len(":") {
			zap.L().Error("score&member's length isn't enough to be decoded",
				zap.ByteString("meta key", zset.key), zap.ByteString("data key", iter.Key()))
			continue
		}

		curPrefix := key[:len(scorePrefix)+byteScoreLen]
		if !startInclude && !startComFinished {
			if curPrefix.Cmp(startPrefix) == 0 {
				offset += 1
				continue
			} else {
				startComFinished = true
			}
		}

		comWithStop := curPrefix.Cmp(stopPrefix)
		if (!stopInclude && comWithStop == 0) ||
			(positiveOrder && comWithStop > 0) ||
			(!positiveOrder && comWithStop < 0) {
			break
		}

		if i < offset {
			continue
		}
		countN += 1
		if count > 0 && countN > count {
			break
		}

		scoreAndMember := key[len(scorePrefix):]
		score := scoreAndMember[0:byteScoreLen]
		member := scoreAndMember[byteScoreLen+len(":"):]
		items = append(items, member)
		if withScore {
			val := []byte(strconv.FormatFloat(DecodeFloat64(score), 'f', -1, 64))
			items = append(items, val)
		}
	}

	return items, nil
}

func (zset *ZSet) ZRem(members [][]byte) (int64, error) {
	deleted := int64(0)

	start := time.Now()
	scores, err := zset.MGet(members)
	zap.L().Debug("zrem mget", zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))
	if err != nil {
		return 0, err
	}

	dkey := DataKey(zset.txn.db, zset.meta.ID)
	costDelMem, costDelScore := int64(0), int64(0)
	for i := range members {
		if scores[i] == nil {
			continue
		}

		scoreKey := zsetScoreKey(dkey, scores[i], members[i])
		start = time.Now()
		err = zset.txn.t.Delete(scoreKey)
		costDelScore += time.Since(start).Nanoseconds()
		if err != nil {
			return deleted, err
		}

		memberKey := zsetMemberKey(dkey, members[i])
		start = time.Now()
		err = zset.txn.t.Delete(memberKey)
		costDelMem += time.Since(start).Nanoseconds()
		if err != nil {
			return deleted, err
		}

		deleted += 1
	}
	zap.L().Debug("zrem cost(us)", zap.Int64("del memberKey", costDelMem/1000),
		zap.Int64("del scoreKey", costDelScore/1000))
	zset.meta.Len -= deleted

	if zset.meta.Len == 0 {
		mkey := MetaKey(zset.txn.db, zset.key)
		start = time.Now()
		err = zset.txn.t.Delete(mkey)
		zap.L().Debug("zrem delete meta key", zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))
		if err != nil {
			return deleted, err
		}
		if zset.meta.Object.ExpireAt > 0 {
			start = time.Now()
			err := unExpireAt(zset.txn.t, mkey, zset.meta.Object.ExpireAt)
			zap.L().Debug("zrem delete expire key", zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))
			if err != nil {
				return deleted, err
			}
		}
		return deleted, nil
	}

	start = time.Now()
	err = zset.updateMeta()
	zap.L().Debug("zrem update meta key", zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))
	return deleted, err
}
func (zset *ZSet) ZCard() int64 {
	return zset.meta.Len
}

func (zset *ZSet) ZScore(member []byte) ([]byte, error) {
	bScore, err := zset.zScoreBytes(member)
	if err != nil {
		if IsErrNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	fscore := DecodeFloat64(bScore)
	sscore := strconv.FormatFloat(fscore, 'f', -1, 64)
	return []byte(sscore), nil
}

func (zset *ZSet) zScoreBytes(member []byte) ([]byte, error) {
	dkey := DataKey(zset.txn.db, zset.meta.ID)
	memberKey := zsetMemberKey(dkey, member)
	bScore, err := zset.txn.t.Get(zset.txn.ctx, memberKey)
	if err != nil {
		return nil, err
	}
	return bScore, nil
}

func (zset *ZSet) ZScan(cursor []byte, f func(key, val []byte) bool) error {
	if !zset.Exist() {
		return nil
	}
	dkey := DataKey(zset.txn.db, zset.meta.ID)
	prefix := ZSetScorePrefix(dkey)
	endPrefix := kv.Key(prefix).PrefixNext()
	ikey := prefix
	if len(cursor) > 0 {
		bScore, err := zset.zScoreBytes(cursor)
		if err != nil {
			if IsErrNotFound(err) {
				return nil
			}
			return err
		}
		if len(bScore) > 0 {
			ikey = append(ikey, bScore...)
		}
	}
	iter, err := zset.txn.t.Iter(ikey, endPrefix)
	if err != nil {
		return err
	}
	for iter.Valid() && iter.Key().HasPrefix(prefix) {
		scoreAndMember := iter.Key()[len(prefix):]
		member := scoreAndMember[byteScoreLen+len(":"):]
		byteScore := scoreAndMember[0:byteScoreLen]
		score := []byte(strconv.FormatFloat(DecodeFloat64(byteScore), 'f', -1, 64))
		if !f(member, score) {
			break
		}
		if err := iter.Next(); err != nil {
			return err
		}
	}
	return nil
}

func zsetMemberKey(dkey []byte, member []byte) []byte {
	var memberKey []byte
	memberKey = append(memberKey, dkey...)
	memberKey = append(memberKey, ':', 'M', ':')
	memberKey = append(memberKey, member...)
	return memberKey
}

// ZSetScorePrefix builds a score key prefix from a redis key
func ZSetScorePrefix(dkey []byte) []byte {
	var sPrefix []byte
	sPrefix = append(sPrefix, dkey...)
	sPrefix = append(sPrefix, ':', 'S', ':')
	return sPrefix
}

func zsetScoreKey(dkey []byte, score []byte, member []byte) []byte {
	var scoreKey []byte
	scoreKey = append(scoreKey, dkey...)
	scoreKey = append(scoreKey, ':', 'S', ':')
	scoreKey = append(scoreKey, score...)
	scoreKey = append(scoreKey, ':')
	scoreKey = append(scoreKey, member...)
	return scoreKey
}
