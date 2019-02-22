package command

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/meitu/titan/db"
)

// SAdd adds the specified members to the set stored at key
func SAdd(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	members := make([][]byte, len(ctx.Args[1:]))
	for i, member := range ctx.Args[1:] {
		members[i] = []byte(member)
	}
	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	added, err := set.SAdd(members...)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, added), nil
}

// SMembers returns all the members of the set value stored at key
func SMembers(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	members, err := set.SMembers()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BytesArray(ctx.Out, members), nil
}

// SCard returns the set cardinality (number of elements) of the set stored at key
func SCard(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	count, err := set.SCard()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(count)), nil
}

// SIsmember returns if member is a member of the set stored at key
func SIsmember(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	member := []byte(ctx.Args[1])
	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	count, err := set.SIsmember(member)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(count)), nil

}

// SPop removes and returns one or more random elements from the set value store at key
func SPop(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var count int
	var err error
	var members [][]byte
	var set *db.Set
	key := []byte(ctx.Args[0])

	if len(ctx.Args) == 2 {
		count, err = strconv.Atoi(ctx.Args[1])
		if err != nil {
			return nil, ErrInteger
		}
	}
	set, err = txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	members, err = set.SPop(int64(count))
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BytesArray(ctx.Out, members), nil
}

// SRem removes the specified members from the set stored at key
func SRem(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var members [][]byte
	key := []byte(ctx.Args[0])
	for _, member := range ctx.Args[1:] {
		members = append(members, []byte(member))
	}
	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	count, err := set.SRem(members)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(count)), nil
}

// SMove movies member from the set at source to the set at destination
func SMove(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	member := make([]byte, 0, len(ctx.Args[2]))
	key := []byte(ctx.Args[0])
	destkey := []byte(ctx.Args[1])
	member = []byte(ctx.Args[2])

	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	count, err := set.SMove(destkey, member)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(count)), nil
}

// SUnion returns the members of the set resulting from the union of all the given sets.
func SUnion(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var members [][]byte
	var min []byte
	var count int
	var setsIter = make([]*db.SetIter, len(ctx.Args)) //存储每个set当前的迭代器位置
	for i, key := range ctx.Args {
		set, err := txn.Set([]byte(key))
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		siter, err := set.Iter()
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		defer siter.Iter.Close()
		setsIter[i] = siter
	}
	min = setsIter[0].Value()
	for count < len(ctx.Args) {
		for i := 0; i < len(ctx.Args); i++ {
			if !setsIter[i].Valid() {
				continue
			}
			if bytes.Compare(min, setsIter[i].Value()) == 1 || bytes.Equal(setsIter[i].Value(), min) {
				min = setsIter[i].Value()
			}

		}
		for i := 0; i < len(ctx.Args); i++ {
			if !setsIter[i].Valid() {
				continue
			}
			if bytes.Equal(setsIter[i].Value(), min) {
				if err := setsIter[i].Iter.Next(); err != nil {
					return nil, errors.New("ERR " + err.Error())
				}
			}

			if !setsIter[i].Valid() {
				count++
			}

		}
		members = append(members, min)
		for i := 0; i < len(ctx.Args); i++ {
			if setsIter[i].Valid() {
				min = setsIter[i].Value()
				break
			}
		}
	}
	return BytesArray(ctx.Out, members), nil
}

// SInter returns the members of the set resulting from the intersection of all the given sets.
func SInter(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var members [][]byte
	var max []byte
	var mkeys = make([][]byte, len(ctx.Args))
	var setsIter = make([]*db.SetIter, len(ctx.Args)) //存储每个set当前的迭代器位置
	for i, key := range ctx.Args {
		set, err := txn.Set([]byte(key))
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		mkey := db.GetMetaKey(txn, []byte(key))
		siter, err := set.Iter()
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		defer siter.Iter.Close()
		setsIter[i] = siter
		mkeys[i] = mkey
	}

	// Batch get meta information
	// If the set corresponding to key does not exist, it is processed as an empty set
	mval, err := db.BatchGetValues(txn, mkeys)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	for _, val := range mval {
		if val == nil {
			return nil, nil
		}
		smeta, _ := db.DecodeSetMeta(val)
		if smeta.Len == 0 {
			return nil, nil
		}
	}

	max = setsIter[0].Value()
	for {
		i := 0
	Loop:
		for ; i < len(ctx.Args); i++ {
			for ; setsIter[i].Valid(); setsIter[i].Iter.Next() {
				if bytes.Compare(setsIter[i].Value(), max) == 1 {
					max = setsIter[i].Value()
					break Loop
				} else if bytes.Equal(setsIter[i].Value(), max) {
					break
				}
			}
			if !setsIter[i].Valid() {
				return BytesArray(ctx.Out, members), nil
			}
		}
		if i == len(ctx.Args) {
			members = append(members, max)
			if err := setsIter[0].Iter.Next(); err != nil {
				return nil, errors.New("ERR " + err.Error())
			}
			if !setsIter[0].Valid() {
				return BytesArray(ctx.Out, members), nil
			}
			max = setsIter[0].Value()
		}

	}

	return BytesArray(ctx.Out, members), nil
}

// SDiff returns the members of the set resulting from the difference between the first set and all the successive sets.
func SDiff(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var members [][]byte
	var min []byte
	var count int
	var setsIter = make([]*db.SetIter, len(ctx.Args)) //存储每个set当前的迭代器位置
	for i, key := range ctx.Args {
		set, err := txn.Set([]byte(key))
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		siter, err := set.Iter()
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		defer siter.Iter.Close()
		setsIter[i] = siter
	}

	min = setsIter[0].Value()
	for {
	Loop:
		// check to see if the same element exists as the current membet for the benchmark key
		for i := 0; i < len(ctx.Args); i++ {
			if !setsIter[i].Valid() {
				continue
			}
			if bytes.Equal(min, setsIter[i].Value()) {
				if i == 0 || !setsIter[i].Valid() {
					continue
				}
				if err := setsIter[i].Iter.Next(); err != nil {
					return nil, errors.New("ERR " + err.Error())
				}
				if err := setsIter[0].Iter.Next(); err != nil {
					return nil, errors.New("ERR " + err.Error())
				}
				if !setsIter[0].Valid() {
					return BytesArray(ctx.Out, members), nil
				}
				min = setsIter[0].Value()
				goto Loop
			}
		}
		//find min in members
		for i := 0; i < len(ctx.Args); i++ {
			if !setsIter[i].Valid() {
				if i == 0 {
					return BytesArray(ctx.Out, members), nil
				}
				continue
			}
			if bytes.Compare(min, setsIter[i].Value()) == 1 {
				min = setsIter[i].Value()
			}
		}
		//Find the smallest element in the current member and move the pointer back
		for i := 0; i < len(ctx.Args); i++ {
			if !setsIter[i].Valid() {
				continue
			}
			if bytes.Equal(min, setsIter[i].Value()) {
				if i == 0 {
					members = append(members, min)
					if err := setsIter[0].Iter.Next(); err != nil {
						return nil, errors.New("ERR " + err.Error())
					}
					for bytes.Equal(min, setsIter[0].Value()) {
						if err := setsIter[0].Iter.Next(); err != nil {
							return nil, errors.New("ERR " + err.Error())
						}
					}
				} else if setsIter[i].Valid() {
					if err := setsIter[i].Iter.Next(); err != nil {
						return nil, errors.New("ERR " + err.Error())
					}
				}
			}
		}
		if setsIter[0].Valid() {
			min = setsIter[0].Value()
		}

		var j int
		for i := 0; i < len(ctx.Args); i++ {
			if !setsIter[i].Valid() {
				j++
			}
		}
		count = j
		if count == len(ctx.Args) {
			break
		}

	}
	return BytesArray(ctx.Out, members), nil
}
