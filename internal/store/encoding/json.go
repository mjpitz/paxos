package encoding

import "encoding/json"

var JSON = &Encoding{
	Marshal: func(entry interface{}) ([]byte, error) {
		return json.Marshal(entry)
	},
	Unmarshal: func(bytes []byte, entry interface{}) error {
		return json.Unmarshal(bytes, entry)
	},
}
