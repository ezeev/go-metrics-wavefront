package tags

// Tag simple struct to hold a tag
type Tag struct {
	Key   string
	Value string
}

// EncodeTags provides a metric key containing tags
func EncodeTags(key string, tags ...Tag) string {
	keyAppend := "["
	for _, tag := range tags {
		keyAppend += tag.Key + "=\"" + tag.Value + "\""
	}
	keyAppend += "]"
	key += keyAppend
	return key
}
