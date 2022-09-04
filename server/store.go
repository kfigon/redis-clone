package server

import "sync"

type inmemoryDatastore struct {
	d map[string]string
	mut sync.RWMutex
}

func newDataStore() *inmemoryDatastore {
	return &inmemoryDatastore{
		d: map[string]string{},
	}
}

func (d *inmemoryDatastore) get(key string) (string, bool) {
	d.mut.Lock()
	defer d.mut.Unlock()
	v, ok := d.d[key]
	return v, ok
}

func (d *inmemoryDatastore) store(key string, value string) {
	d.mut.Lock()
	defer d.mut.Unlock()
	d.d[key] = value
}

func (d *inmemoryDatastore) delete(key string) {
	d.mut.Lock()
	defer d.mut.Unlock()
	delete(d.d, key)
}

