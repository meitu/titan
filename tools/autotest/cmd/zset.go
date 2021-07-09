package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

//ExampleList the key command
//memberScores record the key and value of the operation
type ExampleZSet struct {
	memberScores map[string]map[string]float64
	conn         redis.Conn
}

//NewExampleList create list object
func NewExampleZSet(conn redis.Conn) *ExampleZSet {
	return &ExampleZSet{
		conn:         conn,
		memberScores: make(map[string]map[string]float64),
	}
}

func (ez *ExampleZSet) ZAddEqual(t *testing.T, key string, values ...string) {
	msmap, ok := ez.memberScores[key]
	if !ok {
		ez.memberScores[key] = make(map[string]float64)
		msmap = ez.memberScores[key]
	}
	oldLen := len(msmap)

	req := make([]interface{}, 0, len(values))
	req = append(req, key)
	for i := range values {
		req = append(req, values[i])
	}

	if len(values)%2 != 0 {
		reply, err := redis.Int(ez.conn.Do("zadd", req...))
		assert.Equal(t, oldLen, len(msmap))
		assert.Nil(t, reply)
		assert.NotNil(t, err)
		return
	}

	uniq_members := make(map[string]bool)
	for i := range values {
		if i%2 == 0 {
			if _, ok := uniq_members[values[i+1]]; ok {
				continue
			}
			fscore, err := strconv.ParseFloat(values[i], 64)
			if err != nil {
				reply, err := redis.Int(ez.conn.Do("zadd", req...))
				assert.Equal(t, oldLen, len(msmap))
				assert.Nil(t, reply)
				assert.NotNil(t, err)
				return
			}
			msmap[values[i+1]] = fscore
			uniq_members[values[i+1]] = true
		}
	}

	reply, err := redis.Int(ez.conn.Do("zadd", req...))
	t.Logf("reply :%v, %v, %v", reply, err, req)
	assert.Equal(t, len(msmap)-oldLen, reply)
	assert.Nil(t, err)
}

func (ez *ExampleZSet) ZAddEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ez.conn.Do("zadd", args...)
	assert.EqualError(t, err, errValue)
}

func (ez *ExampleZSet) ZRemEqual(t *testing.T, key string, members ...string) {
	req := make([]interface{}, 0, len(members))
	req = append(req, key)
	for i := range members {
		req = append(req, members[i])
	}

	deleted := 0
	msmap, ok := ez.memberScores[key]
	if ok {
		for _, member := range members {
			if _, ok := msmap[member]; !ok {
				continue
			}

			delete(msmap, member)
			deleted += 1
		}
	}

	reply, err := redis.Int(ez.conn.Do("zrem", req...))
	assert.Equal(t, deleted, reply)
	assert.Nil(t, err)
}

func (ez *ExampleZSet) ZRemEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ez.conn.Do("zrem", args...)
	assert.EqualError(t, err, errValue)
}

func (ez *ExampleZSet) ZAnyOrderRangeEqual(t *testing.T, key string, start int, stop int, positiveOrder bool, withScore bool) {
	cmd := "zrange"
	if !positiveOrder {
		cmd = "zrevrange"
	}
	msmap, ok := ez.memberScores[key]
	if !ok {
		reply, err := redis.Strings(ez.conn.Do(cmd, key, start, stop))
		assert.Equal(t, msmap, reply)
		assert.Nil(t, err)
		return
	}

	if start >= len(msmap) {
		reply, err := redis.Strings(ez.conn.Do(cmd, key, start, stop))
		assert.Equal(t, []string{}, reply)
		assert.Nil(t, err)
		return
	}

	tmp := getAllOutput(msmap, positiveOrder, withScore)
	var reply []string
	var err error
	if withScore {
		reply, err = redis.Strings(ez.conn.Do(cmd, key, start, stop, "WITHSCORES"))
	} else {
		reply, err = redis.Strings(ez.conn.Do(cmd, key, start, stop))
	}

	if start < 0 {
		if start += len(msmap); start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		if stop += len(msmap); stop < 0 {
			stop = 0
		}
	} else if stop >= len(msmap) {
		stop = len(msmap) - 1
	}
	if withScore {
		assert.Equal(t, tmp[2*start:2*stop+2], reply)
	} else {
		assert.Equal(t, tmp[start:stop+1], reply)
	}
	assert.Nil(t, err)
}

func (ez *ExampleZSet) ZRangeEqual(t *testing.T, key string, start int, stop int, withScore bool) {
	ez.ZAnyOrderRangeEqual(t, key, start, stop, true, withScore)
}

func (ez *ExampleZSet) ZRangeEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ez.conn.Do("zrange", args...)
	assert.EqualError(t, err, errValue)
}

func (ez *ExampleZSet) ZRevRangeEqual(t *testing.T, key string, start int, stop int, withScore bool) {
	ez.ZAnyOrderRangeEqual(t, key, start, stop, false, withScore)
}

func (ez *ExampleZSet) ZRevRangeEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ez.conn.Do("zrevrange", args...)
	assert.EqualError(t, err, errValue)
}

func (ez *ExampleZSet) ZScanEqual(t *testing.T, key string, cursor string, pattern string, count int, expected string) {
	cmd := "zscan"
	req := make([]interface{}, 0)
	req = append(req, key)
	req = append(req, cursor)
	req = append(req, "match", pattern)
	req = append(req, "count", count)

	reply, err := redis.MultiBulk(ez.conn.Do(cmd, req...))
	lastCursor, _ := redis.String(reply[0], err)
	strs, _ := redis.Strings(reply[1], err)
	fmt.Println(lastCursor, strs)
	if expected != "" {
		expectedStrs := strings.Split(expected, " ")
		assert.Equal(t, expectedStrs[0], lastCursor)
		assert.Equal(t, expectedStrs[1:], strs)
	} else {
		assert.Equal(t, "0", lastCursor)
	}
	assert.Nil(t, err)
}

func (ez *ExampleZSet) ZCountEqual(t *testing.T, key string, start string, stop string, expected int64) {
	cmd := "zcount"
	req := make([]interface{}, 0)
	req = append(req, key)
	req = append(req, start)
	req = append(req, stop)

	reply, err := redis.Int64(ez.conn.Do(cmd, req...))
	assert.Equal(t, expected, reply)
	assert.Nil(t, err)
}

func (ez *ExampleZSet) ZCountEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ez.conn.Do("zcount", args...)
	assert.EqualError(t, err, errValue)
}

func (ez *ExampleZSet) ZRangeByScoreEqual(t *testing.T, key string, start string, stop string, withScores bool, limit string, expected string) {
	ez.ZAnyOrderRangeByScoreEqual(t, key, start, stop, withScores, true, limit, expected)
}

func (ez *ExampleZSet) ZRevRangeByScoreEqual(t *testing.T, key string, start string, stop string, withScores bool, limit string, expected string) {
	ez.ZAnyOrderRangeByScoreEqual(t, key, start, stop, withScores, false, limit, expected)
}

func (ez *ExampleZSet) ZAnyOrderRangeByScoreEqual(t *testing.T, key string, start string, stop string, withScores bool, positiveOrder bool, limit string, expected string) {
	cmd := "zrangebyscore"
	if !positiveOrder {
		cmd = "zrevrangebyscore"
	}

	var reply []string
	var err error
	req := make([]interface{}, 0)
	req = append(req, key)
	req = append(req, start)
	req = append(req, stop)
	if withScores {
		req = append(req, "WITHSCORES")
	}
	if limit != "" {
		limitArgs := strings.Split(limit, " ")
		for _, limitArg := range limitArgs {
			req = append(req, limitArg)
		}
	}

	reply, err = redis.Strings(ez.conn.Do(cmd, req...))
	if expected != "" {
		expectedStrs := strings.Split(expected, " ")
		assert.Equal(t, expectedStrs, reply)
	} else {
		assert.Equal(t, []string{}, reply)
	}

	assert.Nil(t, err)
}

func (ez *ExampleZSet) ZRangeByScoreEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ez.conn.Do("zrangebyscore", args...)
	assert.EqualError(t, err, errValue)
}

func (ez *ExampleZSet) ZRevRangeByScoreEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ez.conn.Do("zrevrangebyscore", args...)
	assert.EqualError(t, err, errValue)
}

func (ez *ExampleZSet) ZCardEqual(t *testing.T, key string) {
	reply, err := redis.Int(ez.conn.Do("zcard", key))
	assert.Equal(t, len(ez.memberScores[key]), reply)
	assert.Nil(t, err)
}

func (ez *ExampleZSet) ZCardEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ez.conn.Do("zcard", args...)
	assert.EqualError(t, err, errValue)
}

func (ez *ExampleZSet) ZScoreEqual(t *testing.T, key string, member string) {
	msmap, ok := ez.memberScores[key]
	reply, err := redis.String(ez.conn.Do("zscore", key, member))
	if !ok {
		assert.Equal(t, "", reply)
		assert.EqualError(t, err, "redigo: nil returned")
		return
	}

	score, ok := msmap[member]
	if !ok {
		assert.Equal(t, "", reply)
		assert.EqualError(t, err, "redigo: nil returned")
		return
	}

	val := strconv.FormatFloat(score, 'f', -1, 64)
	assert.Equal(t, val, reply)
	assert.Nil(t, err)
}

func (ez *ExampleZSet) ZScoreEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ez.conn.Do("zscore", args...)
	assert.EqualError(t, err, errValue)
}

func getAllOutput(msmap map[string]float64, positiveOrder bool, withScore bool) []string {
	scoreMembers := make(map[float64][]string)
	for member, score := range msmap {
		if _, ok := scoreMembers[score]; !ok {
			scoreMembers[score] = make([]string, 0, 1)
		}
		scoreMembers[score] = append(scoreMembers[score], member)

	}
	scores := make([]float64, 0)
	for score, _ := range scoreMembers {
		scores = append(scores, score)
	}
	if positiveOrder {
		sort.Float64s(scores)
	} else {
		sort.Sort(sort.Reverse(sort.Float64Slice(scores)))
	}

	fullOutput := make([]string, 0, 2*len(msmap))
	for _, score := range scores {
		members := scoreMembers[score]
		if positiveOrder {
			sort.Strings(members)
		} else {
			sort.Sort(sort.Reverse(sort.StringSlice(members)))
		}
		for _, member := range members {
			fullOutput = append(fullOutput, member)
			if withScore {
				val := strconv.FormatFloat(score, 'f', -1, 64)
				fullOutput = append(fullOutput, val)
			}
		}
	}
	return fullOutput
}
