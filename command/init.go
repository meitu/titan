package command

var commands map[string]Desc

func init() {
	// commands contains all commands that open to clients
	// exec should not be in this table to avoid 'initialization loop', and it indeed not necessary be here in fact.
	commands = map[string]Desc{
		// connections
		"auth":   Desc{Proc: Auth, Cons: Constraint{2, flags("sltF"), 0, 0, 0}},
		"echo":   Desc{Proc: Echo, Cons: Constraint{2, flags("F"), 0, 0, 0}},
		"ping":   Desc{Proc: Ping, Cons: Constraint{-1, flags("tF"), 0, 0, 0}},
		"quit":   Desc{Proc: Quit, Cons: Constraint{1, 0, 0, 0, 0}},
		"select": Desc{Proc: Select, Cons: Constraint{2, flags("lF"), 0, 0, 0}},
		"swapdb": Desc{Proc: SwapDB, Cons: Constraint{3, flags("wF"), 0, 0, 0}},

		// transactions, exec and discard should called explicitly, so they are registered here
		"multi":   Desc{Proc: Multi, Cons: Constraint{1, flags("sF"), 0, 0, 0}},
		"watch":   Desc{Proc: Watch, Cons: Constraint{-2, flags("sF"), 1, -1, 1}},
		"unwatch": Desc{Proc: Unwatch, Cons: Constraint{1, flags("sF"), 0, 0, 0}},

		// lists
		"lindex":    Desc{Proc: AutoCommit(LIndex), Txn: LIndex, Cons: Constraint{3, flags("r"), 1, 1, 1}},
		"linsert":   Desc{Proc: AutoCommit(LInsert), Txn: LInsert, Cons: Constraint{5, flags("wm"), 1, 1, 1}},
		"llen":      Desc{Proc: AutoCommit(LLen), Txn: LLen, Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"lpop":      Desc{Proc: AutoCommit(LPop), Txn: LPop, Cons: Constraint{2, flags("wF"), 1, 1, 1}},
		"lpush":     Desc{Proc: AutoCommit(LPush), Txn: LPush, Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},
		"lpushx":    Desc{Proc: AutoCommit(LPushx), Txn: LPushx, Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"lrange":    Desc{Proc: AutoCommit(LRange), Txn: LRange, Cons: Constraint{4, flags("r"), 1, 1, 1}},
		"ltrim":     Desc{Proc: AutoCommit(LTrim), Txn: LTrim, Cons: Constraint{4, flags("w"), 1, 1, 1}},
		"lrem":      Desc{Proc: AutoCommit(LRem), Txn: LRem, Cons: Constraint{4, flags("w"), 1, 1, 1}},
		"lset":      Desc{Proc: AutoCommit(LSet), Txn: LSet, Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"rpop":      Desc{Proc: AutoCommit(RPop), Txn: RPop, Cons: Constraint{2, flags("wF"), 1, 1, 1}},
		"rpoplpush": Desc{Proc: AutoCommit(RPopLPush), Txn: RPopLPush, Cons: Constraint{3, flags("wF"), 1, 2, 1}},
		"rpush":     Desc{Proc: AutoCommit(RPush), Txn: RPush, Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},
		"rpushx":    Desc{Proc: AutoCommit(RPushx), Txn: RPushx, Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},

		// strings
		"get":         Desc{Proc: AutoCommit(Get), Txn: Get, Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"set":         Desc{Proc: AutoCommit(Set), Txn: Set, Cons: Constraint{-3, flags("wm"), 1, 1, 1}},
		"setnx":       Desc{Proc: AutoCommit(SetNx), Txn: SetNx, Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"setex":       Desc{Proc: AutoCommit(SetEx), Txn: SetEx, Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"psetex":      Desc{Proc: AutoCommit(PSetEx), Txn: PSetEx, Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"mget":        Desc{Proc: AutoCommit(MGet), Txn: MGet, Cons: Constraint{-2, flags("rF"), 1, -1, 1}},
		"mset":        Desc{Proc: AutoCommit(MSet), Txn: MSet, Cons: Constraint{-3, flags("wm"), 1, -1, 2}},
		"msetnx":      Desc{Proc: AutoCommit(MSetNx), Txn: MSetNx, Cons: Constraint{-3, flags("wm"), 1, -1, 2}},
		"strlen":      Desc{Proc: AutoCommit(Strlen), Txn: Strlen, Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"append":      Desc{Proc: AutoCommit(Append), Txn: Append, Cons: Constraint{3, flags("wm"), 1, 1, 1}},
		"setrange":    Desc{Proc: AutoCommit(SetRange), Txn: SetRange, Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		"getrange":    Desc{Proc: AutoCommit(GetRange), Txn: GetRange, Cons: Constraint{4, flags("r"), 1, 1, 1}},
		"incr":        Desc{Proc: AutoCommit(Incr), Txn: Incr, Cons: Constraint{2, flags("wmF"), 1, 1, 1}},
		"decr":        Desc{Proc: AutoCommit(Decr), Txn: Decr, Cons: Constraint{2, flags("wmF"), 1, 1, 1}},
		"incrby":      Desc{Proc: AutoCommit(IncrBy), Txn: IncrBy, Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"decrby":      Desc{Proc: AutoCommit(DecrBy), Txn: DecrBy, Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"incrbyfloat": Desc{Proc: AutoCommit(IncrByFloat), Txn: IncrByFloat, Cons: Constraint{3, flags("wmF"), 1, 1, 1}},
		"setbit":      Desc{Proc: AutoCommit(SetBit), Txn: SetBit, Cons: Constraint{4, flags("wm"), 1, 1, 1}},
		// "bitop":       Desc{Proc: AutoCommit(BitOp), Cons: Constraint{-4, flags("wm"), 2, -1, 1}},
		// "bitfield":    Desc{Proc: AutoCommit(BitField), Cons: Constraint{-2, flags("wm"), 1, 1, 1}},
		"getbit":   Desc{Proc: AutoCommit(GetBit), Txn: GetBit, Cons: Constraint{3, flags("r"), 1, 1, 1}},
		"bitcount": Desc{Proc: AutoCommit(BitCount), Txn: BitCount, Cons: Constraint{-2, flags("r"), 1, 1, 1}},
		"bitpos":   Desc{Proc: AutoCommit(BitPos), Txn: BitPos, Cons: Constraint{-3, flags("r"), 1, 1, 1}},
		"getset":   Desc{Proc: AutoCommit(GetSet), Txn: GetSet, Cons: Constraint{3, flags("wm"), 1, 1, 1}},

		// keys
		"type":      Desc{Proc: AutoCommit(Type), Txn: Type, Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"exists":    Desc{Proc: AutoCommit(Exists), Txn: Exists, Cons: Constraint{-2, flags("rF"), 1, -1, 1}},
		"keys":      Desc{Proc: AutoCommit(Keys), Txn: Keys, Cons: Constraint{-2, flags("rS"), 0, 0, 0}},
		"del":       Desc{Proc: AutoCommit(Delete), Txn: Delete, Cons: Constraint{-2, flags("w"), 1, -1, 1}},
		"unlink":    Desc{Proc: AutoCommit(Delete), Txn: Delete, Cons: Constraint{-2, flags("w"), 1, -1, 1}},
		"expire":    Desc{Proc: AutoCommit(Expire), Txn: Expire, Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"expireat":  Desc{Proc: AutoCommit(ExpireAt), Txn: ExpireAt, Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"pexpire":   Desc{Proc: AutoCommit(PExpire), Txn: PExpire, Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"pexpireat": Desc{Proc: AutoCommit(PExpireAt), Txn: PExpireAt, Cons: Constraint{3, flags("wF"), 1, 1, 1}},
		"persist":   Desc{Proc: AutoCommit(Persist), Txn: Persist, Cons: Constraint{2, flags("wF"), 1, 1, 1}},
		"ttl":       Desc{Proc: AutoCommit(TTL), Txn: TTL, Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"pttl":      Desc{Proc: AutoCommit(PTTL), Txn: PTTL, Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"object":    Desc{Proc: AutoCommit(Object), Txn: Object, Cons: Constraint{-2, flags("rR"), 0, 0, 0}},
		"scan":      Desc{Proc: AutoCommit(Scan), Txn: Scan, Cons: Constraint{-2, flags("rR"), 0, 0, 0}},
		"randomkey": Desc{Proc: AutoCommit(RandomKey), Txn: RandomKey, Cons: Constraint{1, flags("rR"), 0, 0, 0}},
		"touch":     Desc{Proc: AutoCommit(Touch), Txn: Touch, Cons: Constraint{-2, flags("rF"), 1, -1, 1}},

		// server
		"monitor":  Desc{Proc: Monitor, Cons: Constraint{1, flags("as"), 0, 0, 0}},
		"client":   Desc{Proc: Client, Cons: Constraint{-2, flags("as"), 0, 0, 0}},
		"debug":    Desc{Proc: AutoCommit(Debug), Cons: Constraint{-2, flags("as"), 0, 0, 0}},
		"command":  Desc{Proc: RedisCommand, Cons: Constraint{0, flags("lt"), 0, 0, 0}},
		"flushdb":  Desc{Proc: AutoCommit(FlushDB), Cons: Constraint{-1, flags("w"), 0, 0, 0}},
		"flushall": Desc{Proc: AutoCommit(FlushAll), Cons: Constraint{-1, flags("w"), 0, 0, 0}},
		"time":     Desc{Proc: Time, Cons: Constraint{1, flags("RF"), 0, 0, 0}},
		"info":     Desc{Proc: Info, Cons: Constraint{-1, flags("lt"), 0, 0, 0}},

		// hashes
		"hdel":         Desc{Proc: AutoCommit(HDel), Txn: HDel, Cons: Constraint{-3, flags("wF"), 1, 1, 1}},
		"hset":         Desc{Proc: AutoCommit(HSet), Txn: HSet, Cons: Constraint{-4, flags("wmF"), 1, 1, 1}},
		"hget":         Desc{Proc: AutoCommit(HGet), Txn: HGet, Cons: Constraint{3, flags("rF"), 1, 1, 1}},
		"hgetall":      Desc{Proc: AutoCommit(HGetAll), Txn: HGetAll, Cons: Constraint{2, flags("r"), 1, 1, 1}},
		"hexists":      Desc{Proc: AutoCommit(HExists), Txn: HExists, Cons: Constraint{3, flags("rF"), 1, 1, 1}},
		"hincrby":      Desc{Proc: AutoCommit(HIncrBy), Txn: HIncrBy, Cons: Constraint{4, flags("wmF"), 1, 1, 1}},
		"hincrbyfloat": Desc{Proc: AutoCommit(HIncrByFloat), Txn: HIncrByFloat, Cons: Constraint{4, flags("wmF"), 1, 1, 1}},
		"hkeys":        Desc{Proc: AutoCommit(HKeys), Txn: HKeys, Cons: Constraint{2, flags("rS"), 1, 1, 1}},
		"hvals":        Desc{Proc: AutoCommit(HVals), Txn: HVals, Cons: Constraint{2, flags("rS"), 1, 1, 1}},
		"hlen":         Desc{Proc: AutoCommit(HLen), Txn: HLen, Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"hstrlen":      Desc{Proc: AutoCommit(HStrLen), Txn: HStrLen, Cons: Constraint{3, flags("rF"), 1, 1, 1}},
		"hsetnx":       Desc{Proc: AutoCommit(HSetNX), Txn: HSetNX, Cons: Constraint{4, flags("wmF"), 1, 1, 1}},
		"hmget":        Desc{Proc: AutoCommit(HMGet), Txn: HMGet, Cons: Constraint{-3, flags("rF"), 1, 1, 1}},
		"hmset":        Desc{Proc: AutoCommit(HMSet), Txn: HMSet, Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},
		"hscan":        Desc{Proc: AutoCommit(HScan), Txn: HScan, Cons: Constraint{-3, flags("rR"), 0, 0, 0}},

		// sets
		"sadd":      Desc{Proc: AutoCommit(SAdd), Txn: SAdd, Cons: Constraint{-3, flags("wmF"), 1, 1, 1}},
		"smembers":  Desc{Proc: AutoCommit(SMembers), Txn: SMembers, Cons: Constraint{2, flags("rS"), 1, 1, 1}},
		"scard":     Desc{Proc: AutoCommit(SCard), Txn: SCard, Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"sismember": Desc{Proc: AutoCommit(SIsmember), Txn: SIsmember, Cons: Constraint{3, flags("rF"), 1, 1, 1}},
		"spop":      Desc{Proc: AutoCommit(SPop), Txn: SPop, Cons: Constraint{-2, flags("wRF"), 1, 1, 1}},
		"srem":      Desc{Proc: AutoCommit(SRem), Txn: SRem, Cons: Constraint{-3, flags("wF"), 1, 1, 1}},
		"sunion":    Desc{Proc: AutoCommit(SUnion), Txn: SUnion, Cons: Constraint{-2, flags("rS"), 1, -1, 1}},
		"sinter":    Desc{Proc: AutoCommit(SInter), Txn: SInter, Cons: Constraint{-2, flags("rS"), 1, -1, 1}},
		"sdiff":     Desc{Proc: AutoCommit(SDiff), Txn: SDiff, Cons: Constraint{-2, flags("rS"), 1, -1, 1}},
		"smove":     Desc{Proc: AutoCommit(SMove), Txn: SMove, Cons: Constraint{4, flags("wF"), 1, 2, 1}},

		// zsets
		"zadd":             Desc{Proc: AutoCommit(ZAdd), Txn: ZAdd, Cons: Constraint{-4, flags("wmF"), 1, 1, 1}},
		"zrange":           Desc{Proc: AutoCommit(ZRange), Txn: ZRange, Cons: Constraint{-4, flags("rF"), 1, 1, 1}},
		"zrevrange":        Desc{Proc: AutoCommit(ZRevRange), Txn: ZRevRange, Cons: Constraint{-4, flags("rF"), 1, 1, 1}},
		"zrangebyscore":    Desc{Proc: AutoCommit(ZRangeByScore), Txn: ZRangeByScore, Cons: Constraint{-4, flags("rF"), 1, 1, 1}},
		"zrevrangebyscore": Desc{Proc: AutoCommit(ZRevRangeByScore), Txn: ZRevRangeByScore, Cons: Constraint{-4, flags("rF"), 1, 1, 1}},
		"zrem":             Desc{Proc: AutoCommit(ZRem), Txn: ZRem, Cons: Constraint{-3, flags("wF"), 1, 1, 1}},
		"zcard":            Desc{Proc: AutoCommit(ZCard), Txn: ZCard, Cons: Constraint{2, flags("rF"), 1, 1, 1}},
		"zcount":           Desc{Proc: AutoCommit(ZCount), Txn: ZCount, Cons: Constraint{-4, flags("rF"), 1, 1, 1}},
		"zscore":           Desc{Proc: AutoCommit(ZScore), Txn: ZScore, Cons: Constraint{3, flags("rF"), 1, 1, 1}},
		"zscan":            Desc{Proc: AutoCommit(ZScan), Txn: ZScan, Cons: Constraint{-3, flags("rF"), 1, 1, 1}},

		// extension commands
		"escan": Desc{Proc: AutoCommit(Escan), Txn: Escan, Cons: Constraint{-1, flags("rR"), 0, 0, 0}},
	}
}
