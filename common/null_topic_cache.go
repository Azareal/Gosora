package common

// NullTopicCache is a topic cache to be used when you don't want a cache and just want queries to passthrough to the database
type NullTopicCache struct {
}

// NewNullTopicCache gives you a new instance of NullTopicCache
func NewNullTopicCache() *NullTopicCache {
	return &NullTopicCache{}
}

// nolint
func (c *NullTopicCache) Get(id int) (*Topic, error) {
	return nil, ErrNoRows
}
func (c *NullTopicCache) GetUnsafe(id int) (*Topic, error) {
	return nil, ErrNoRows
}
func (c *NullTopicCache) BulkGet(ids []int) (list []*Topic) {
	return make([]*Topic, len(ids))
}
func (c *NullTopicCache) Set(_ *Topic) error {
	return nil
}
func (c *NullTopicCache) Add(_ *Topic) error {
	return nil
}
func (c *NullTopicCache) AddUnsafe(_ *Topic) error {
	return nil
}
func (c *NullTopicCache) Remove(id int) error {
	return nil
}
func (c *NullTopicCache) RemoveUnsafe(id int) error {
	return nil
}
func (c *NullTopicCache) Flush() {
}
func (c *NullTopicCache) Length() int {
	return 0
}
func (c *NullTopicCache) SetCapacity(_ int) {
}
func (c *NullTopicCache) GetCapacity() int {
	return 0
}
