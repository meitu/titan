package command

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/distributedio/titan/db"
	"github.com/distributedio/titan/encoding/resp"
)

// ZAdd adds the specified members with scores to the sorted set
func ZAdd(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	kvs := ctx.Args[1:]
	if len(kvs)%2 != 0 {
		return nil, errors.New("ERR syntax error")
	}

	uniqueMembers := make(map[string]bool)
	count := len(kvs) / 2
	members := make([][]byte, 0, count)
	scores := make([]float64, 0, count)
	for i := 0; i < len(kvs)-1; i += 2 {
		member := kvs[i+1]
		if _, ok := uniqueMembers[member]; ok {
			continue
		}

		members = append(members, []byte(member))
		score, err := strconv.ParseFloat(kvs[i], 64)
		if err != nil || math.IsNaN(score) {
			return nil, ErrFloat
		}
		scores = append(scores, score)

		uniqueMembers[member] = true
	}

	zset, err := txn.ZSet(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	added, err := zset.ZAdd(members, scores)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	return Integer(ctx.Out, added), nil
}

func ZRange(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	return zAnyOrderRange(ctx, txn, true)
}

func ZRevRange(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	return zAnyOrderRange(ctx, txn, false)
}

func zAnyOrderRange(ctx *Context, txn *db.Transaction, positiveOrder bool) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	start, err := strconv.ParseInt(ctx.Args[1], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	stop, err := strconv.ParseInt(ctx.Args[2], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	withScore := bool(false)
	if len(ctx.Args) >= 4 {
		if strings.ToUpper(ctx.Args[3]) == "WITHSCORES" {
			withScore = true
		}
	}

	zset, err := txn.ZSet(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !zset.Exist() {
		return BytesArray(ctx.Out, nil), nil
	}

	items, err := zset.ZAnyOrderRange(start, stop, withScore, positiveOrder)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	if len(items) == 0 {
		return BytesArray(ctx.Out, nil), nil
	}
	return BytesArray(ctx.Out, items), nil
}

func ZRangeByScore(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	return zAnyOrderRangeByScore(ctx, txn, true)
}

func ZRevRangeByScore(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	return zAnyOrderRangeByScore(ctx, txn, false)
}

func ZCount(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	startScore, startInclude, err := getFloatAndInclude(ctx.Args[1])
	if err != nil {
		return nil, ErrMinOrMaxNotFloat
	}
	endScore, endInclude, err := getFloatAndInclude(ctx.Args[2])
	if err != nil {
		return nil, ErrMinOrMaxNotFloat
	}
	zset, err := txn.ZSet(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !zset.Exist() {
		return Integer(ctx.Out, 0), nil
	}

	items, err := zset.ZAnyOrderRangeByScore(startScore, startInclude,
		endScore, endInclude,
		false,
		int64(0), math.MaxInt64,
		true)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	if len(items) == 0 {
		return Integer(ctx.Out, 0), nil
	}
	return Integer(ctx.Out, int64(len(items))), nil
}

func zAnyOrderRangeByScore(ctx *Context, txn *db.Transaction, positiveOrder bool) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	startScore, startInclude, err := getFloatAndInclude(ctx.Args[1])
	if err != nil {
		return nil, ErrMinOrMaxNotFloat
	}
	endScore, endInclude, err := getFloatAndInclude(ctx.Args[2])
	if err != nil {
		return nil, ErrMinOrMaxNotFloat
	}

	withScore := bool(false)
	offset := int64(0)
	count := int64(math.MaxInt64)
	for i := 3; i < len(ctx.Args); i++ {
		switch strings.ToUpper(ctx.Args[i]) {
		case "WITHSCORES":
			withScore = true
		case "LIMIT":
			if offset, count, err = getLimitParameters(ctx.Args[i+1:]); err != nil {
				return nil, err
			}
			i += 2
		default:
			return nil, ErrSyntax
		}
	}

	zset, err := txn.ZSet(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !zset.Exist() {
		return BytesArray(ctx.Out, nil), nil
	}

	items, err := zset.ZAnyOrderRangeByScore(startScore, startInclude,
		endScore, endInclude,
		withScore,
		offset, count,
		positiveOrder)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	if len(items) == 0 {
		return BytesArray(ctx.Out, nil), nil
	}
	return BytesArray(ctx.Out, items), nil
}

func ZRem(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	uniqueMembers := make(map[string]bool)
	members := make([][]byte, 0, len(ctx.Args)-1)
	for _, member := range ctx.Args[1:] {
		if _, ok := uniqueMembers[member]; ok {
			continue
		}

		members = append(members, []byte(member))
		uniqueMembers[member] = true
	}

	zset, err := txn.ZSet(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !zset.Exist() {
		return Integer(ctx.Out, 0), nil
	}

	deleted, err := zset.ZRem(members)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	return Integer(ctx.Out, deleted), nil
}

func ZCard(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	zset, err := txn.ZSet(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !zset.Exist() {
		return Integer(ctx.Out, 0), nil
	}

	return Integer(ctx.Out, zset.ZCard()), nil
}

func ZScore(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	member := []byte(ctx.Args[1])

	zset, err := txn.ZSet(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !zset.Exist() {
		return NullBulkString(ctx.Out), nil
	}

	score, err := zset.ZScore(member)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	if score == nil {
		return NullBulkString(ctx.Out), nil
	}

	return BulkString(ctx.Out, string(score)), nil
}

func ZScan(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var (
		key        []byte
		cursor     []byte
		lastCursor = []byte("0")
		count      = uint64(defaultScanCount)
		kvs        = [][]byte{}
		pattern    []byte
		isAll      bool
		err        error
	)
	key = []byte(ctx.Args[0])
	if strings.Compare(ctx.Args[1], "0") != 0 {
		cursor = []byte(ctx.Args[1])
	}

	// define return result
	result := func() {
		if _, err := resp.ReplyArray(ctx.Out, 2); err != nil {
			return
		}
		resp.ReplyBulkString(ctx.Out, string(lastCursor))
		if _, err := resp.ReplyArray(ctx.Out, len(kvs)); err != nil {
			return
		}
		for i := range kvs {
			resp.ReplyBulkString(ctx.Out, string(kvs[i]))
		}
	}
	zset, err := txn.ZSet(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !zset.Exist() {
		return result, nil
	}

	if len(ctx.Args)%2 != 0 {
		return nil, ErrSyntax
	}

	for i := 2; i < len(ctx.Args); i += 2 {
		arg := strings.ToLower(ctx.Args[i])
		next := ctx.Args[i+1]
		switch arg {
		case "count":
			if count, err = strconv.ParseUint(next, 10, 64); err != nil {
				return nil, ErrInteger
			}
			if count > ScanMaxCount {
				count = ScanMaxCount
			}
			if count == 0 {
				count = uint64(defaultScanCount)
			}
		case "match":
			pattern = []byte(next)
			isAll = (pattern[0] == '*' && len(pattern) == 1)
		}
	}

	if len(pattern) == 0 {
		isAll = true
	}
	f := func(member, score []byte) bool {
		if count <= 0 {
			lastCursor = member
			return false
		}
		if isAll || globMatch(pattern, member, false) {
			kvs = append(kvs, member)
			kvs = append(kvs, score)
			count--
		}
		return true
	}

	if err := zset.ZScan(cursor, f); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return result, nil

}
