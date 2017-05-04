package structs

import (
	"encoding/hex"
	"encoding/json"
)

// MarshalJSON is a custom Location json marshaller
func (loc *Location) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Id":    hex.EncodeToString(loc.Id),
		"Vnode": loc.Vnode,
	})
}
