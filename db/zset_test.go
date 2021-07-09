package db

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getZSet(t *testing.T, key []byte) (*ZSet, *Transaction, error) {
	txn, err := mockDB.Begin()
	assert.NotNil(t, txn)
	assert.NoError(t, err)
	zset, err := GetZSet(txn, key)
	assert.NotNil(t, zset)
	assert.NoError(t, err)

	return zset, txn, nil
}

func TestZSetZADD(t *testing.T) {
	var members [][]byte
	var score []float64

	members = append(members, []byte("a"))
	members = append(members, []byte("b"))
	members = append(members, []byte("c"))
	score = append(score, 1, 2, 3)
	type args struct {
		members [][]byte
		score   []float64
	}
	type want struct {
		score []float64
		len   int64
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "TestZSetZADD",
			args: args{
				members: members,
				score:   score,
			},
			want: want{
				score: score,
				len:   int64(3),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zset, txn, err := getZSet(t, []byte("TestZSetZADD"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, zset)
			count, err := zset.ZAdd(tt.args.members, tt.args.score)
			assert.NoError(t, err)
			assert.Equal(t, count, tt.want.len)
			txn.Commit(context.TODO())

			zset, txn, err = getZSet(t, []byte("TestZSetZADD"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, zset)

			got, err := zset.MGet(tt.args.members)
			assert.Equal(t, len(got), int(tt.want.len))
			for i, score := range got {
				wantScore, err := EncodeFloat64(tt.want.score[i])
				assert.NoError(t, err)
				assert.Equal(t, score, wantScore)
			}
			assert.NoError(t, err)
		})
	}
}

func TestZSetZScore(t *testing.T) {
	var members [][]byte
	var score []float64

	members = append(members, []byte("a"))
	members = append(members, []byte("b"))
	members = append(members, []byte("c"))
	score = append(score, 1, 2, 3)
	type args struct {
		members [][]byte
		score   []float64
	}
	type want struct {
		score []float64
		len   int64
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "TestZSetZScore",
			args: args{
				members: members,
				score:   score,
			},
			want: want{
				score: score,
				len:   int64(3),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zset, txn, err := getZSet(t, []byte("TestZSetZScore"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, zset)
			count, err := zset.ZAdd(tt.args.members, tt.args.score)
			assert.NoError(t, err)
			assert.Equal(t, count, tt.want.len)
			txn.Commit(context.TODO())

			zset, txn, err = getZSet(t, []byte("TestZSetZScore"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, zset)

			got, err := zset.ZRem(tt.args.members[len(tt.args.members)-2:])
			assert.Equal(t, got, int64(2))
			txn.Commit(context.TODO())

			zset, txn, err = getZSet(t, []byte("TestZSetZScore"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, zset)
			got1, err1 := zset.ZScore(tt.args.members[0])

			wantScore := strconv.FormatFloat(tt.want.score[0], 'f', -1, 64)
			assert.NoError(t, err1)
			assert.Equal(t, string(got1), wantScore)
		})
	}
}

func sortVals(member [][]byte, scores []float64, withScore, positiveOrder, byScore bool) [][]byte {
	dataLen := len(member)
	if withScore {
		dataLen = dataLen * 2
	}
	data := make([][]byte, dataLen)
	for i, m := range member {
		idx := i
		if withScore {
			idx = i * 2
		}
		if !positiveOrder {
			idx = dataLen - idx - 1
			if withScore {
				idx -= 1
			}
		}
		data[idx] = m
		if withScore {
			data[idx+1] = []byte(strconv.FormatFloat(scores[i], 'f', -1, 64))
		}
	}
	return data
}

func TestZSetZAnyOrderRange(t *testing.T) {
	var members [][]byte
	var score []float64

	members = append(members, []byte("a"))
	members = append(members, []byte("b"))
	members = append(members, []byte("c"))
	score = append(score, 1, 2, 3)
	type args struct {
		members       [][]byte
		score         []float64
		start         int64
		end           int64
		withScore     bool
		positiveOrder bool
	}
	type want struct {
		vals [][]byte
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "TestZSetZAnyOrderRange",
			args: args{
				members:       members,
				score:         score,
				start:         int64(0),
				end:           int64(-1),
				withScore:     true,
				positiveOrder: true,
			},
			want: want{
				vals: sortVals(members, score, true, true, false),
			},
		},
		{
			name: "TestZSetZAnyOrderRevRange",
			args: args{
				members:       members,
				score:         score,
				start:         int64(0),
				end:           int64(-1),
				withScore:     true,
				positiveOrder: false,
			},
			want: want{
				vals: sortVals(members, score, true, false, false),
			},
		},
		{
			name: "TestZSetZAnyOrderRangeNoScore",
			args: args{
				members:       members,
				score:         score,
				start:         int64(0),
				end:           int64(-1),
				withScore:     false,
				positiveOrder: true,
			},
			want: want{
				vals: sortVals(members, score, false, true, false),
			},
		},
		{
			name: "TestZSetZAnyOrderRevRangeNoScore",
			args: args{
				members:       members,
				score:         score,
				start:         int64(0),
				end:           int64(-1),
				withScore:     false,
				positiveOrder: false,
			},
			want: want{
				vals: sortVals(members, score, false, false, false),
			},
		},
		{
			name: "TestZSetZAnyOrderSliceRange",
			args: args{
				members:       members,
				score:         score,
				start:         int64(0),
				end:           int64(1),
				withScore:     true,
				positiveOrder: true,
			},
			want: want{
				vals: sortVals(members[0:2], score[0:2], true, true, false),
			},
		},
		{
			name: "TestZSetZAnyOrderSliceRevRange",
			args: args{
				members:       members,
				score:         score,
				start:         int64(0),
				end:           int64(1),
				withScore:     true,
				positiveOrder: false,
			},
			want: want{
				vals: sortVals(members[1:], score[1:], true, false, false),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zset, txn, err := getZSet(t, []byte(tt.name))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, zset)
			count, err := zset.ZAdd(tt.args.members, tt.args.score)
			assert.NoError(t, err)
			assert.Equal(t, count, int64(len(tt.args.members)))
			txn.Commit(context.TODO())

			zset, txn, err = getZSet(t, []byte(tt.name))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, zset)
			got, err := zset.ZAnyOrderRange(tt.args.start, tt.args.end, tt.args.withScore, tt.args.positiveOrder)
			assert.NoError(t, err)
			for i, val := range got {
				assert.Equal(t, val, tt.want.vals[i])
			}
		})
	}
}

func TestZSetZAnyOrderRangeScore(t *testing.T) {
	var members [][]byte
	var score []float64

	members = append(members, []byte("a"))
	members = append(members, []byte("b"))
	members = append(members, []byte("c"))
	score = append(score, 1, 2, 3)
	type args struct {
		members       [][]byte
		score         []float64
		start         float64
		startInclude  bool
		end           float64
		endInclude    bool
		withScore     bool
		offset        int64
		limit         int64
		positiveOrder bool
	}
	type want struct {
		vals [][]byte
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "TestZSetZAnyOrderRangeScore",
			args: args{
				members:       members,
				score:         score,
				start:         float64(0),
				startInclude:  false,
				end:           float64(3),
				endInclude:    false,
				withScore:     true,
				offset:        int64(0),
				limit:         int64(3),
				positiveOrder: true,
			},
			want: want{
				vals: sortVals(members, score, true, true, true),
			},
		},
		{
			name: "TestZSetZAnyOrderRevRangeScore",
			args: args{
				members:       members,
				score:         score,
				start:         float64(1),
				startInclude:  true,
				end:           float64(3),
				endInclude:    false,
				withScore:     true,
				offset:        int64(0),
				limit:         int64(3),
				positiveOrder: true,
			},
			want: want{
				vals: sortVals(members[:2], score[:2], true, true, true),
			},
		},
		{
			name: "TestZSetZAnyOrderRangeByScoreDisableWithScore",
			args: args{
				members:       members,
				score:         score,
				start:         float64(1),
				startInclude:  true,
				end:           float64(3),
				endInclude:    true,
				withScore:     true,
				offset:        int64(0),
				limit:         int64(3),
				positiveOrder: true,
			},
			want: want{
				vals: sortVals(members, score, true, true, true),
			},
		},
		{
			name: "TestZSetZAnyOrderRevRangeByScoreDisalbeWithScore",
			args: args{
				members:       members,
				score:         score,
				start:         float64(3),
				startInclude:  true,
				end:           float64(2),
				endInclude:    true,
				withScore:     true,
				offset:        int64(0),
				limit:         int64(3),
				positiveOrder: false,
			},
			want: want{
				vals: sortVals(members[1:], score[1:], true, false, true),
			},
		},
		{
			name: "TestZSetZAnyOrderSliceRangeByScore",
			args: args{
				members:       members,
				score:         score,
				start:         float64(1),
				startInclude:  true,
				end:           float64(3),
				endInclude:    true,
				offset:        int64(1),
				limit:         int64(2),
				withScore:     true,
				positiveOrder: true,
			},
			want: want{
				vals: sortVals(members[1:], score[1:], true, true, true),
			},
		},
		{
			name: "TestZSetZAnyOrderSliceRevRangeByScore",
			args: args{
				members:       members,
				score:         score,
				start:         float64(3),
				startInclude:  true,
				end:           float64(1),
				endInclude:    true,
				offset:        int64(1),
				limit:         int64(2),
				withScore:     true,
				positiveOrder: false,
			},
			want: want{
				vals: sortVals(members[:2], score[:2], true, false, true),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zset, txn, err := getZSet(t, []byte(tt.name))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, zset)
			count, err := zset.ZAdd(tt.args.members, tt.args.score)
			assert.NoError(t, err)
			assert.Equal(t, count, int64(len(tt.args.members)))
			txn.Commit(context.TODO())

			zset, txn, err = getZSet(t, []byte(tt.name))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, zset)
			got, err := zset.ZAnyOrderRangeByScore(tt.args.start, tt.args.startInclude, tt.args.end, tt.args.endInclude, tt.args.withScore, tt.args.offset, tt.args.limit, tt.args.positiveOrder)
			assert.NoError(t, err)
			for i, val := range got {
				assert.Equal(t, val, tt.want.vals[i])
			}
		})
	}
}

func TestZSet_ZScan(t *testing.T) {
	var members [][]byte
	var score []float64

	members = append(members, []byte("abc"))
	members = append(members, []byte("aec"))
	members = append(members, []byte("acc"))
	members = append(members, []byte("bc"))
	score = append(score, -1.1, -1, 1, 2.1)

	zset, txn, err := getZSet(t, []byte("TestZSet_ZScan"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, zset)
	count, err := zset.ZAdd(members, score)
	assert.NoError(t, err)
	assert.Equal(t, count, int64(len(members)))
	txn.Commit(context.TODO())

	type args struct {
		cursor []byte
		f      func(key, val []byte) bool
	}
	var value [][]byte
	count = 2

	tests := []struct {
		name string
		args args
		want [][]byte
	}{
		{
			name: "TestZSet_ZScan",
			args: args{
				f: func(member, score []byte) bool {
					if count == 0 {
						return false
					}
					value = append(value, member, score)
					count--
					return true

				},
			},
			want: append([][]byte{}, []byte("abc"), []byte("-1.1"), []byte("aec"), []byte("-1")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			zset, txn, err := getZSet(t, []byte("TestZSet_ZScan"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, zset)

			err = zset.ZScan(tt.args.cursor, tt.args.f)
			txn.Commit(context.TODO())

			assert.Equal(t, value, tt.want)
			assert.NoError(t, err)
		})
	}
}
