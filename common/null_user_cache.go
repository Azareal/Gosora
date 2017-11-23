package common

type NullUserCache struct {
}

// NewNullUserCache gives you a new instance of NullUserCache
func NewNullUserCache() *NullUserCache {
	return &NullUserCache{}
}

func (mus *NullUserCache) Get(id int) (*User, error) {
	return nil, ErrNoRows
}

func (mus *NullUserCache) BulkGet(ids []int) (list []*User) {
	return list
}

func (mus *NullUserCache) GetUnsafe(id int) (*User, error) {
	return nil, ErrNoRows
}

func (mus *NullUserCache) Set(_ *User) error {
	return nil
}

func (mus *NullUserCache) Add(item *User) error {
	_ = item
	return nil
}

func (mus *NullUserCache) AddUnsafe(item *User) error {
	_ = item
	return nil
}

func (mus *NullUserCache) Remove(id int) error {
	return nil
}

func (mus *NullUserCache) RemoveUnsafe(id int) error {
	return nil
}

func (mus *NullUserCache) Flush() {
}

func (mus *NullUserCache) Length() int {
	return 0
}

func (mus *NullUserCache) SetCapacity(_ int) {
}

func (mus *NullUserCache) GetCapacity() int {
	return 0
}
