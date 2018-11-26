package site

import ()

type Asset struct {
	Mime    string
	Content *[]byte
}

type AssetManager struct {
	dir   string
	cache *Cache
}

func NewAssetManager(dir string) *AssetManager {
	return &AssetManager{
		dir:   dir,
		cache: NewCache(),
	}
}

func (p *AssetManager) Load() error {
	keys, err := getKeys(p.dir, "")
	if err != nil {
		return err
	}

	for _, key := range keys {
		asset, err := p.buildAsset(key)
		if err != nil {
			return err
		}

		p.cache.Set(key, asset)
	}

	return nil
}

func (p *AssetManager) Get(key string) *Asset {
	item := p.cache.Get(key)
	return item.(*Asset)
}

func (p *AssetManager) buildAsset(filename string) (*Asset, error) {
	asset, err := getAsset(filename)
	if err != nil {
		return nil, err
	}

	return &Asset{
		Mime:    "TODO",
		Content: asset,
	}, nil
}
