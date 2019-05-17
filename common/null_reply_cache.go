package common

// NullReplyCache is a reply cache to be used when you don't want a cache and just want queries to passthrough to the database
type NullReplyCache struct {
}

// NewNullReplyCache gives you a new instance of NullReplyCache
func NewNullReplyCache() *NullReplyCache {
	return &NullReplyCache{}
}

// nolint
func (c *NullReplyCache) Get(id int) (*Reply, error) {
	return nil, ErrNoRows
}
func (c *NullReplyCache) GetUnsafe(id int) (*Reply, error) {
	return nil, ErrNoRows
}
func (c *NullReplyCache) BulkGet(ids []int) (list []*Reply) {
	return make([]*Reply, len(ids))
}
func (c *NullReplyCache) Set(_ *Reply) error {
	return nil
}
func (c *NullReplyCache) Add(_ *Reply) error {
	return nil
}
func (c *NullReplyCache) AddUnsafe(_ *Reply) error {
	return nil
}
func (c *NullReplyCache) Remove(id int) error {
	return nil
}
func (c *NullReplyCache) RemoveUnsafe(id int) error {
	return nil
}
func (c *NullReplyCache) Flush() {
}
func (c *NullReplyCache) Length() int {
	return 0
}
func (c *NullReplyCache) SetCapacity(_ int) {
}
func (c *NullReplyCache) GetCapacity() int {
	return 0
}
