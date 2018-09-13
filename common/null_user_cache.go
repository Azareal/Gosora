package common

// NullUserCache is a user cache to be used when you don't want a cache and just want queries to passthrough to the database
type NullUserCache struct {
}

// NewNullUserCache gives you a new instance of NullUserCache
func NewNullUserCache() *NullUserCache {
	return &NullUserCache{}
}

// nolint
func (mus *NullUserCache) DeallocOverflow() {}
func (mus *NullUserCache) Get(id int) (*User, error) {
	return nil, ErrNoRows
}
func (mus *NullUserCache) BulkGet(ids []int) (list []*User) {
	return make([]*User, len(ids))
}
func (mus *NullUserCache) GetUnsafe(id int) (*User, error) {
	return nil, ErrNoRows
}
func (mus *NullUserCache) Set(_ *User) error {
	return nil
}
func (mus *NullUserCache) Add(_ *User) error {
	return nil
}
func (mus *NullUserCache) AddUnsafe(_ *User) error {
	return nil
}
func (mus *NullUserCache) Remove(id int) error {
	return nil
}
func (mus *NullUserCache) RemoveUnsafe(id int) error {
	return nil
}
func (mus *NullUserCache) BulkRemove(ids []int) {}
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
