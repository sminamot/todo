package todo

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/syndtr/goleveldb/leveldb"
)

type Todo struct {
	db *leveldb.DB
}

type Items []Item

type Item struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	RegDate string `json:"reg_date"`
	Done    bool   `json:"done"`
}

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
		RegDate: fmt.Sprint(time.Now().UnixNano()),
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

func (i Items) String(verbose bool) string {
	messages := make([]string, 0, len(i))

	for _, v := range i {
		message := "- ["
		check := " "
		if v.Done {
			check = "X"
		}
		message += check + "] " + v.Title
		if verbose {
			i64, _ := strconv.ParseInt(v.RegDate, 10, 64)
			t := time.Unix(i64/int64(time.Second), 0)
			message += " " + v.ID + " " + t.Format("2006/01/02 15:04:05")
		}
		messages = append(messages, message)
	}
	return strings.Join(messages, "\n")
}

func delete(s Items, i int) Items {
	s = append(s[:i], s[i+1:]...)
	n := make(Items, len(s))
	copy(n, s)
	return n
}
