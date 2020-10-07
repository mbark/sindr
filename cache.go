package main

import (
	"fmt"

	"github.com/peterbourgon/diskv/v3"
)

type Cache struct {
	diskv *diskv.Diskv
}

func NewCache(file string) Cache {
	return Cache{
		diskv: diskv.New(diskv.Options{
			BasePath:     file,
			Transform:    func(s string) []string { return []string{} },
			CacheSizeMax: 1024 * 1024,
		}),
	}
}

func (c Cache) StoreVersion(name, version string) error {
	return c.diskv.Write(name, []byte(version))
}

func (c Cache) GetVersion(name string) (*string, error) {
	if !c.diskv.Has(name) {
		return nil, nil
	}

	value, err := c.diskv.Read(name)
	if err != nil {
		return nil, fmt.Errorf("cache read: %w", err)
	}

	val := string(value)
	return &val, nil
}
