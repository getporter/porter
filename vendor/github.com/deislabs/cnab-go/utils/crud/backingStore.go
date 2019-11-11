package crud

var _ Store = &BackingStore{}

// BackingStore wraps another store that may have Connect/Close methods that need to be called.
type BackingStore struct {
	backingStore Store
}

func NewBackingStore(store Store) *BackingStore {
	return &BackingStore{
		backingStore: store,
	}
}

func (s BackingStore) Connect() error {
	if connectable, ok := s.backingStore.(HasConnect); ok {
		return connectable.Connect()
	}

	return nil
}

func (s BackingStore) Close() error {
	var err error
	if closable, ok := s.backingStore.(HasClose); ok {
		err = closable.Close()
	}
	s.backingStore = nil
	return err
}

func (s BackingStore) List(itemType string) ([]string, error) {
	err := s.Connect()
	if err != nil {
		return nil, err
	}

	return s.backingStore.List(itemType)
}

func (s BackingStore) Save(itemType string, name string, data []byte) error {
	err := s.Connect()
	if err != nil {
		return err
	}

	return s.backingStore.Save(itemType, name, data)
}

func (s BackingStore) Read(itemType string, name string) ([]byte, error) {
	err := s.Connect()
	if err != nil {
		return nil, err
	}

	return s.backingStore.Read(itemType, name)
}

func (s BackingStore) Delete(itemType string, name string) error {
	err := s.Connect()
	if err != nil {
		return err
	}

	return s.backingStore.Delete(itemType, name)
}
