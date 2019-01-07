package site

//log "github.com/sirupsen/logrus"

type Asset struct {
	Mime    string
	Content *[]byte
	Etag    string
}

func LoadAssets(dir string, cache *ContentCache) error {
	keys, err := getKeys(dir, "")
	if err != nil {
		return err
	}

	for _, key := range keys {
		asset, err := buildAsset(key)
		if err != nil {
			return err
		}

		cache.Set("/static/"+key, asset)
	}

	// TODO robots.txt
	//robotsFile := "allow.txt"
	//if s.Env != "production" {
	//	robotsFile = "disallow.txt"
	//}

	return nil
}

func buildAsset(filename string) (*Asset, error) {
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
