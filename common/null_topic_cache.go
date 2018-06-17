package common

// NullTopicCache is a topic cache to be used when you don't want a cache and just want queries to passthrough to the database
type NullTopicCache struct {
}

// NewNullTopicCache gives you a new instance of NullTopicCache
func NewNullTopicCache() *NullTopicCache {
	return &NullTopicCache{}
}

// nolint
func (mts *NullTopicCache) Get(id int) (*Topic, error) {
	return nil, ErrNoRows
}
func (mts *NullTopicCache) GetUnsafe(id int) (*Topic, error) {
	return nil, ErrNoRows
}
func (mts *NullTopicCache) Set(_ *Topic) error {
	return nil
}
func (mts *NullTopicCache) Add(_ *Topic) error {
	return nil
}
func (mts *NullTopicCache) AddUnsafe(_ *Topic) error {
	return nil
}
func (mts *NullTopicCache) Remove(id int) error {
	return nil
}
func (mts *NullTopicCache) RemoveUnsafe(id int) error {
	return nil
}
func (mts *NullTopicCache) Flush() {
}
func (mts *NullTopicCache) Length() int {
	return 0
}
func (mts *NullTopicCache) SetCapacity(_ int) {
}
func (mts *NullTopicCache) GetCapacity() int {
	return 0
}
