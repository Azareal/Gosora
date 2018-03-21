package adventure

type AdventureStore interface {
	Create() (int, error)
}

type DefaultAdventureStore struct {
}
