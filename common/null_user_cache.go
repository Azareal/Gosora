package common

// NullUserCache is a user cache to be used when you don't want a cache and just want queries to passthrough to the database
type NullUserCache struct {
}

// NewNullUserCache gives you a new instance of NullUserCache
func NewNullUserCache() *NullUserCache {
	return &NullUserCache{}
}

// nolint
func (c *NullUserCache) DeallocOverflow(evictPriority bool) (evicted int) {
	return 0
}
func (c *NullUserCache) Get(id int) (*User, error) {
	return nil, ErrNoRows
}
func (c *NullUserCache) BulkGet(ids []int) (list []*User) {
	return make([]*User, len(ids))
}
func (c *NullUserCache) GetUnsafe(id int) (*User, error) {
	return nil, ErrNoRows
}
func (c *NullUserCache) Set(_ *User) error {
	return nil
}
func (c *NullUserCache) Add(_ *User) error {
	return nil
}
func (c *NullUserCache) AddUnsafe(_ *User) error {
	return nil
}
func (c *NullUserCache) Remove(id int) error {
	return nil
}
func (c *NullUserCache) RemoveUnsafe(id int) error {
	return nil
}
func (c *NullUserCache) BulkRemove(ids []int) {}
func (c *NullUserCache) Flush() {
}
func (c *NullUserCache) Length() int {
	return 0
}
func (c *NullUserCache) SetCapacity(_ int) {
}
func (c *NullUserCache) GetCapacity() int {
	return 0
}
