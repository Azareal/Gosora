package common

type NullTopicCache struct {
}

// NewNullTopicCache gives you a new instance of NullTopicCache
func NewNullTopicCache() *NullTopicCache {
	return &NullTopicCache{}
}

func (mts *NullTopicCache) Get(id int) (*Topic, error) {
	return nil, ErrNoRows
}

func (mts *NullTopicCache) GetUnsafe(id int) (*Topic, error) {
	return nil, ErrNoRows
}

func (mts *NullTopicCache) Set(_ *Topic) error {
	return nil
}

func (mts *NullTopicCache) Add(item *Topic) error {
	_ = item
	return nil
}

// TODO: Make these length increments thread-safe. Ditto for the other DataStores
func (mts *NullTopicCache) AddUnsafe(item *Topic) error {
	_ = item
	return nil
}

// TODO: Make these length decrements thread-safe. Ditto for the other DataStores
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
