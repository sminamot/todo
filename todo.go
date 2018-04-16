package todo

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/rs/xid"
	"github.com/syndtr/goleveldb/leveldb"
)

type Todo struct {
	db *leveldb.DB
}

type Items []Item

type Item struct {
	ID      string  `json:"id"`
	Title   string  `json:"title"`
	RegDate RegDate `json:"reg_date"`
	Done    bool    `json:"done"`
}

type RegDate string

func New(f string) (*Todo, error) {
	db, err := leveldb.OpenFile(f, nil)
	if err != nil {
		return nil, err
	}
	return &Todo{db}, nil
}

func (t *Todo) Close() {
	t.db.Close()
}

func (t *Todo) Add(key, title string) error {
	items, err := t.Get(key)
	if err != nil {
		return err
	}

	guid := xid.New()

	item := Item{
		ID:      guid.String(),
		Title:   title,
		RegDate: RegDate(fmt.Sprint(time.Now().UnixNano())),
		Done:    false,
	}
	items = append(items, item)

	err = t.register(key, items)

	return nil
}

func (t *Todo) Get(key string) (Items, error) {
	ret, err := t.db.Get([]byte(key), nil)
	if err == leveldb.ErrNotFound {
		return Items{}, nil
	}
	if err != nil {
		return nil, err
	}

	items := Items{}
	err = json.Unmarshal(ret, &items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (t *Todo) Done(key, id string, done bool) error {
	items, err := t.Get(key)
	if err != nil {
		return err
	}

	for i, v := range items {
		if v.ID == id {
			items[i].Done = done
			if err := t.register(key, items); err != nil {
				return err
			}
			break
		}
		if i == len(items)-1 {
			return nil
		}
	}

	return nil
}

func (t *Todo) Delete(key, id string) error {
	items, err := t.Get(key)
	if err != nil {
		return err
	}

	for i, v := range items {
		if v.ID == id {
			items = delete(items, i)
			if err := t.register(key, items); err != nil {
				return err
			}
			break
		}
		if i == len(items)-1 {
			return nil
		}
	}

	return nil
}

func (t *Todo) register(key string, items Items) error {
	items.doneSort()
	jsonItems, err := json.Marshal(items)
	if err != nil {
		return err
	}

	err = t.db.Put([]byte(key), jsonItems, nil)
	if err != nil {
		return err
	}

	return nil
}

func (items *Items) doneSort() {
	df := make(Items, 0, len(*items)) // Done is false
	dt := make(Items, 0, len(*items)) // Done is true
	for _, v := range *items {
		if v.Done {
			dt = append(dt, v)
			continue
		}
		df = append(df, v)
	}
	// sort slice by RegDate
	sort.Slice(df, func(i, j int) bool {
		return df[i].RegDate < df[j].RegDate
	})
	sort.Slice(dt, func(i, j int) bool {
		return dt[i].RegDate < dt[j].RegDate
	})
	*items = append(df, dt...)
}

func (rd RegDate) Time() (time.Time, error) {
	i64, err := strconv.ParseInt(string(rd), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(i64/int64(time.Second), 0), nil
}

func delete(s Items, i int) Items {
	s = append(s[:i], s[i+1:]...)
	n := make(Items, len(s))
	copy(n, s)
	return n
}
