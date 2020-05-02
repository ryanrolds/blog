package site

//log "github.com/sirupsen/logrus"

type Asset struct {
	Mime    string
	Content *[]byte
	Etag    string
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
	if item == nil {
		return nil
	}

	return item.(*Asset)
}

func (p *AssetManager) buildAsset(filename string) (*Asset, error) {
	buffer, mime, err := getAsset(filename)
	if err != nil {
		return nil, err
	}

	return &Asset{
		Mime:    mime,
		Content: buffer,
		Etag:    getEtag(buffer),
	}, nil
}

func (p *AssetManager) GetHashes() *Hashes {
	hashes := Hashes{}

	keys := p.cache.GetKeys()
	for _, key := range keys {
		value := p.cache.Get(key)
		hashes[key] = value.(*Asset).Etag
	}

	return &hashes
}
