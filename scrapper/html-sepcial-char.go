package scrapper

var specialChars map[string]byte = map[string]byte{}

// TODO: autogenerate this code for a more serious HTML scrapper
func init() {
	specialChars["&amp;"] = '&'
	specialChars["&quot;"] = '"'
	specialChars["&nbsp;"] = ' '
	specialChars["&lt;"] = '<'
	specialChars["&gt;"] = '>'
	specialChars["&middot;"] = '·'
}
