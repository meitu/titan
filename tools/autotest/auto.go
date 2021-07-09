package autotest

import (
	"strconv"
	"testing"
	"time"

	"github.com/distributedio/titan/tools/autotest/cmd"

	"github.com/gomodule/redigo/redis"
)

//AutoClient check redis comman
type AutoClient struct {
	es *cmd.ExampleString
	el *cmd.ExampleList
	ek *cmd.ExampleKey
	ez *cmd.ExampleZSet
	*cmd.ExampleSystem
	em *cmd.ExampleMulti
	// addr string
	pool *redis.Pool
	conn redis.Conn
}

//NewAutoClient creat auto client
func NewAutoClient() *AutoClient {
	return &AutoClient{}
}

//Start run client
func (ac *AutoClient) Start(addr string) {
	// ac.pool = newPool(addr)
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	_, err = redis.String(conn.Do("auth", "test-1542098935-1-7ca41bda4efc2a1889c04e"))
	if err != nil {
		panic(err)
	}
	ac.conn = conn
	ac.es = cmd.NewExampleString(conn)
	ac.ek = cmd.NewExampleKey(conn)
	ac.el = cmd.NewExampleList(conn)
	ac.ez = cmd.NewExampleZSet(conn)
	ac.ExampleSystem = cmd.NewExampleSystem(conn)
	ac.em = cmd.NewExampleMulti(conn)
}

//Close shut client
func (ac *AutoClient) Close() {
	// ac.pool.Close()
	ac.conn.Close()
}

//StringCase check string case
func (ac *AutoClient) StringCase(t *testing.T) {
	ac.es.SetNxEqual(t, "key-setx", "v1")
	ac.es.SetExEqual(t, "key-set", "v2", 1)
	ac.es.PSetexEqual(t, "key-set", "v3", 3000)

	ac.es.SetEqual(t, "key-set", "value")
	ac.es.AppendEqual(t, "key-set", "value")
	ac.es.AppendEqual(t, "append", "value")
	ac.es.StrlenEqual(t, "key-set")
	// ac.es.MSetNxEqual(t, 1, "key-setm", "value", "key-set", "value")

	ac.es.MSetEqual(t, "key-set", "value")
	ac.es.MGetEqual(t, "key-not-exist")
	ac.es.IncrEqual(t, "incr")
	ac.es.IncrEqual(t, "incr")
	ac.es.IncrByEqual(t, "incr", 19)
	ac.es.DecrEqual(t, "incr")
	ac.es.DecrByEqual(t, "incr", 19)
	ac.es.IncrByFloatEqual(t, "incr", 1.111111)
	ac.es.IncrByFloatEqual(t, "incr", 2.1e2)
	ac.es.StrlenEqual(t, "heng")

}

//ListCase check list case
//TODO
func (ac *AutoClient) ListCase(t *testing.T) {
	ac.el.LpushEqual(t, "key-list", "v1", "v2", "v3", "v4")
	ac.el.LlenEqual(t, "key-list")
	ac.el.LsetEqual(t, "key-list", 3, "v0")
	ac.el.LindexEqual(t, "key-list", 3)
	ac.el.LrangeEqual(t, "key-list", 0, 10)
	ac.el.LrangeEqual(t, "key-list", 99, 100)
	ac.el.LpopEqual(t, "key-list")
	ac.el.LpopEqual(t, "key-list-l")

	var keys []string
	for i := 0; i < 4000; i++ {
		num := strconv.Itoa(i)
		keys = append(keys, "v", num)
	}
	ac.el.LpushEqual(t, "zkey-list", keys...)
	ac.el.LlenEqual(t, "zkey-list")
	ac.el.LsetEqual(t, "zkey-list", 3, "v0")
	ac.el.LindexEqual(t, "zkey-list", 3)
	ac.el.LrangeEqual(t, "zkey-list", 0, 10)
	ac.el.LrangeEqual(t, "zkey-list", 99, 100)
	ac.el.LpopEqual(t, "zkey-list")
}

//ZSetCase check zset case
func (ac *AutoClient) ZSetCase(t *testing.T) {
	ac.ez.ZAddEqual(t, "key-zset", "2.0", "member1", "-1.5", "member2", "3.6", "member3", "-3.5", "member4", "2.5", "member1")
	ac.ez.ZCardEqual(t, "key-zset")
	ac.ez.ZCardEqual(t, "key-zset1")
	ac.ez.ZScoreEqual(t, "key-zset1", "member1")
	ac.ez.ZScoreEqual(t, "key-zset", "member1")
	ac.ez.ZScoreEqual(t, "key-zset", "member5")
	ac.ez.ZRangeEqual(t, "key-zset", 0, -1, true)
	ac.ez.ZRangeEqual(t, "key-zset", 0, -1, false)
	ac.ez.ZRangeEqual(t, "key-zset", 1, 4, true)
	ac.ez.ZRangeEqual(t, "key-zset", -4, 5, true)
	ac.ez.ZRangeEqual(t, "key-zset", -6, 5, true)
	ac.ez.ZRangeEqual(t, "key-zset", 4, 1, true)
	ac.ez.ZRangeEqual(t, "key-zset", 6, 10, true)
	ac.ez.ZRevRangeEqual(t, "key-zset", 0, -1, true)
	ac.ez.ZRevRangeEqual(t, "key-zset", 0, -1, false)
	ac.ez.ZRevRangeEqual(t, "key-zset", 1, 4, true)
	ac.ez.ZRevRangeEqual(t, "key-zset", -4, 5, true)
	ac.ez.ZRevRangeEqual(t, "key-zset", -6, 5, true)
	ac.ez.ZRevRangeEqual(t, "key-zset", 4, 1, true)
	ac.ez.ZRevRangeEqual(t, "key-zset", 6, 10, true)

	ac.ez.ZAddEqual(t, "key-zset", "0.0", "member5", "1.5", "member2")
	ac.ez.ZRangeEqual(t, "key-zset", 0, -1, true)

	ac.ez.ZAddEqual(t, "key-zset", "3.6", "member3", "0.0", "member5")
	ac.ez.ZRangeEqual(t, "key-zset", 0, -1, true)

	ac.ez.ZAddEqual(t, "key-zset", "2.0", "member11", "2.05", "member6")
	ac.ez.ZRangeEqual(t, "key-zset", 0, -1, true)

	ac.ez.ZRangeByScoreEqual(t, "key-zset", "-inf", "+inf", true, "", "member4 -3.5 member5 0 member2 1.5 member1 2 member11 2 member6 2.05 member3 3.6")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(-inf", "+inf", true, "", "member4 -3.5 member5 0 member2 1.5 member1 2 member11 2 member6 2.05 member3 3.6")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "-inf", "(+inf", true, "", "member4 -3.5 member5 0 member2 1.5 member1 2 member11 2 member6 2.05 member3 3.6")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "-inf", "inf", true, "", "member4 -3.5 member5 0 member2 1.5 member1 2 member11 2 member6 2.05 member3 3.6")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "-inf", "inf", false, "", "member4 member5 member2 member1 member11 member6 member3")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "-3.5", "inf", true, "", "member4 -3.5 member5 0 member2 1.5 member1 2 member11 2 member6 2.05 member3 3.6")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(-3.5", "inf", true, "", "member5 0 member2 1.5 member1 2 member11 2 member6 2.05 member3 3.6")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "0.0", "inf", true, "", "member5 0 member2 1.5 member1 2 member11 2 member6 2.05 member3 3.6")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(0.0", "inf", true, "", "member2 1.5 member1 2 member11 2 member6 2.05 member3 3.6")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(0.0", "3.6", true, "", "member2 1.5 member1 2 member11 2 member6 2.05 member3 3.6")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(0.0", "+3.6", true, "", "member2 1.5 member1 2 member11 2 member6 2.05 member3 3.6")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(0.0", "(3.6", true, "", "member2 1.5 member1 2 member11 2 member6 2.05")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(0.0", "2.05", true, "", "member2 1.5 member1 2 member11 2 member6 2.05")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(0.0", "2.05", true, "LIMIT -1 1", "")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(0.0", "2.05", true, "limit 0 -1", "member2 1.5 member1 2 member11 2 member6 2.05")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(0.0", "2.05", true, "LIMIT 0 0", "")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(0.0", "2.05", true, "LIMIT 0 2", "member2 1.5 member1 2")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(0.0", "2.05", true, "LIMIT 0 4", "member2 1.5 member1 2 member11 2 member6 2.05")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(0.0", "2.05", true, "LIMIT 0 5", "member2 1.5 member1 2 member11 2 member6 2.05")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(0.0", "2.05", true, "LIMIT 1 2", "member1 2 member11 2")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(0.0", "2.05", true, "LIMIT 3 2", "member6 2.05")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(0.0", "2.05", true, "LIMIT 4 2", "")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "(2", "3.6", true, "", "member6 2.05 member3 3.6")
	ac.ez.ZRangeByScoreEqual(t, "key-zset", "0", "(2", true, "", "member5 0 member2 1.5")

	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "+inf", "-inf", true, "", "member3 3.6 member6 2.05 member11 2 member1 2 member2 1.5 member5 0 member4 -3.5")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "(+inf", "-inf", true, "", "member3 3.6 member6 2.05 member11 2 member1 2 member2 1.5 member5 0 member4 -3.5")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "+inf", "(-inf", true, "", "member3 3.6 member6 2.05 member11 2 member1 2 member2 1.5 member5 0 member4 -3.5")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "inf", "-inf", true, "", "member3 3.6 member6 2.05 member11 2 member1 2 member2 1.5 member5 0 member4 -3.5")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "inf", "-inf", false, "", "member3 member6 member11 member1 member2 member5 member4")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "inf", "-3.5", true, "", "member3 3.6 member6 2.05 member11 2 member1 2 member2 1.5 member5 0 member4 -3.5")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "inf", "(-3.5", true, "", "member3 3.6 member6 2.05 member11 2 member1 2 member2 1.5 member5 0")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "inf", "0.0", true, "", "member3 3.6 member6 2.05 member11 2 member1 2 member2 1.5 member5 0")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "inf", "(0.0", true, "", "member3 3.6 member6 2.05 member11 2 member1 2 member2 1.5")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "3.6", "(0.0", true, "", "member3 3.6 member6 2.05 member11 2 member1 2 member2 1.5")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "+3.6", "(0.0", true, "", "member3 3.6 member6 2.05 member11 2 member1 2 member2 1.5")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "(3.6", "(0.0", true, "", "member6 2.05 member11 2 member1 2 member2 1.5")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "2.05", "(0.0", true, "", "member6 2.05 member11 2 member1 2 member2 1.5")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "2.05", "(0.0", true, "LIMIT -1 1", "")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "2.05", "(0.0", true, "limit 0 -1", "member6 2.05 member11 2 member1 2 member2 1.5")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "2.05", "(0.0", true, "LIMIT 0 0", "")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "2.05", "(0.0", true, "LIMIT 0 2", "member6 2.05 member11 2")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "2.05", "(0.0", true, "LIMIT 0 4", "member6 2.05 member11 2 member1 2 member2 1.5")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "2.05", "(0.0", true, "LIMIT 0 5", "member6 2.05 member11 2 member1 2 member2 1.5")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "2.05", "(0.0", true, "LIMIT 1 2", "member11 2 member1 2")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "2.05", "(0.0", true, "LIMIT 3 2", "member2 1.5")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "2.05", "(0.0", true, "LIMIT 4 2", "")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "3.6", "(2", true, "", "member3 3.6 member6 2.05")
	ac.ez.ZRevRangeByScoreEqual(t, "key-zset", "(2", "0", true, "", "member2 1.5 member5 0")

	ac.ez.ZCountEqual(t, "key-zset", "0", "(2", int64(2))

	ac.ez.ZScanEqual(t, "key-zset", "0", "*", 2, "member2 member4 -3.5 member5 0")
	ac.ez.ZScanEqual(t, "key-zset", "member2", "member*", 2, "member11 member2 1.5 member1 2")

	ac.ez.ZRemEqual(t, "key-zset", "member2", "member1", "member3", "member4", "member1")
	ac.ez.ZRangeEqual(t, "key-zset", 0, -1, true)

	ac.ek.ExpireEqual(t, "key-zset", 5, 1)
	ac.ez.ZRemEqual(t, "key-zset", "member5", "member11", "member6", "member7")
	ac.ez.ZRangeEqual(t, "key-zset", 0, -1, true)
	ac.ek.ExistsEqual(t, 0, "key-zset")
	ac.ek.TTLEqual(t, "key-set", -2)
	//check expire won't delete old key(obj id is different with new one)
	ac.ez.ZAddEqual(t, "key-zset", "1.0", "m1")
	time.Sleep(time.Second * 10)
	ac.ek.ExistsEqual(t, 1, "key-zset")
	ac.ek.DelEqual(t, 1, "key-zset")

	ac.ez.ZAddEqual(t, "key-zset-short-key", "2.0", "a", "2.05", "b")
	ac.ez.ZRemEqual(t, "key-zset-short-key", "a", "e")
	ac.ek.DelEqual(t, 1, "key-zset-short-key")
}

//KeyCase check key case
//TODO
func (ac *AutoClient) KeyCase(t *testing.T) {
	ac.ek.TTLEqual(t, "key-set", -1)
	ac.ek.RandomKeyEqual(t)
	ac.ek.ScanEqual(t, "", 4)
	ac.ek.ExistsEqual(t, 3, "key-set", "incr", "foo", "append")
	ac.ek.DelEqual(t, 4, "key-set", "incr", "foo", "append", "key-setx")
	ac.ek.TTLEqual(t, "key-set", -2)
	ac.ek.RandomKeyEqual(t)
	ac.ek.ScanEqual(t, "", 0)

	ac.es.SetEqual(t, "key-set", "value")
	ac.ek.TypeEqual(t, "key-set", "string")
	ac.ek.ObjectEqual(t, "key-set", "raw")
	ac.ek.ExpireEqual(t, "key-set", 2, 1)
	time.Sleep(time.Millisecond)
	ac.ek.TTLEqual(t, "key-set", 1)
	time.Sleep(time.Second * 2)
	ac.ek.ExpireEqual(t, "key-set", 1, 0)
	ac.ek.ExpireEqual(t, "key-set", 0, 0)

	//test PExpire
	ac.el.LpushEqual(t, "key-set", "value")
	ac.ek.TypeEqual(t, "key-set", "list")
	ac.ek.PExpireEqual(t, "key-set", 2000, 1)
	time.Sleep(time.Millisecond)
	ac.ek.TTLEqual(t, "key-set", 1)
	time.Sleep(time.Second * 2)
	ac.ek.PExpireEqual(t, "key-set", 1, 0)
	ac.ek.PExpireEqual(t, "key-set", 0, 0)

	at := time.Now().Unix() + 1
	var key []string
	for i := 0; i < 4000; i++ {
		num := strconv.Itoa(i)
		key = append(key, "v", num)
	}
	ac.el.LpushEqual(t, "zkey-listx", key...)
	ac.ek.TypeEqual(t, "zkey-listx", "list")
	ac.ek.ObjectEqual(t, "zkey-listx", "ziplist")
	ac.ek.ExpireAtEqual(t, "zkey-listx", int(at), 1)
	time.Sleep(time.Second * 1)
	ac.ek.ExpireAtEqual(t, "zkey-listx", int(at), 0)

	//test PExpire
	at = (time.Now().Unix() + 1) * 1000
	ac.el.LpushEqual(t, "key-setz", "value")
	ac.ek.TypeEqual(t, "key-setz", "list")
	ac.ek.ObjectEqual(t, "key-setz", "linkedlist")
	ac.ek.PExpireAtEqual(t, "key-setz", int(at), 1)
	ac.ek.TTLEqual(t, "key-setz", 0)
	time.Sleep(time.Second * 2)
	ac.ek.PExpireAtEqual(t, "key-setz", int(at), 0)

	at = time.Now().Unix() + 1
	ac.es.SetEqual(t, "zkey-listx", "value")
	ac.ek.ExpireAtEqual(t, "zkey-listx", int(at), 1)
	ac.ek.PersistEqual(t, "zkey-listx", 1)
	ac.ek.PersistEqual(t, "zkey-listx", 0)

	//test zset
	ac.ez.ZAddEqual(t, "key-zset", "2.0", "member1")
	ac.ek.ExistsEqual(t, 1, "key-zset")
	ac.ek.TypeEqual(t, "key-zset", "zset")
	ac.ek.ObjectEqual(t, "key-zset", "hashtable")
	ac.ek.TTLEqual(t, "key-zset", -1)
	ac.ek.ExpireEqual(t, "key-zset", 2, 1)
	time.Sleep(time.Millisecond)
	ac.ek.TTLEqual(t, "key-zset", 1)
	time.Sleep(time.Second * 2)
	ac.ek.ExpireEqual(t, "key-zset", 1, 0)

	ac.ez.ZAddEqual(t, "key-zset1", "2.0", "member1")
	ac.ek.DelEqual(t, 1, "key-zset1")
	ac.ek.ExistsEqual(t, 0, "key-zset1")
}

//SystemCase check system case
func (ac *AutoClient) SystemCase(t *testing.T) {
	//auth
	ac.AuthEqual(t, "test-1542098935-1-7ca41bda4efc2a1889c04e")
	//ping
	ac.PingEqual(t)
}

//MultiCase bug mutil exec repley msg is error
func (ac *AutoClient) MultiCase(t *testing.T) {
	//multi
	ac.em.MultiEqual(t)
	ac.em.Cmd(t)
	// exec
	ac.em.ExecEqual(t)
}
