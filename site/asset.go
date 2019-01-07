package site

//log "github.com/sirupsen/logrus"

type Asset struct {
	Mime    string
	Content *[]byte
	Etag    string
}

func LoadAssets(site *Site, assetsDir string) error {
	keys, err := getKeys(site.rootDir+assetsDir, "")
	if err != nil {
		return err
	}

	for _, key := range keys {
		asset, err := buildAsset(site.rootDir + assetsDir + key)
		if err != nil {
			return err
		}

		site.cache.Set(assetsDir+key, asset)
	}

	return nil
}

func buildAsset(filename string) (*Content, error) {
	buffer, mime, err := getAsset(filename)
	if err != nil {
		return nil, err
	}

	return &Content{
		Mime:         mime,
		Content:      buffer,
		Etag:         getEtag(buffer),
		CacheControl: "public, max-age=2419200",
	}, nil
}
