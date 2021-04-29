package lib

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/tgbv/swarm-cache/glob"
)

// Encodes a message.
//
// Takes a map meta of type J and some binary data
func EncodeMessage(meta J, data *[]byte) (error, *[]byte) {
	out := make([]byte, 0)

	// Split json
	metaByte, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("Coult not encode message! %s", err), nil
	}

	// Replace newlines with appropriate bytes within json
	metaStr := string(metaByte)
	metaStr = strings.ReplaceAll(metaStr, "\n", glob.META_NL)
	metaByte = []byte(metaStr)

	// Bind everything together
	out = append(out, metaByte...)
	if data != nil {
		out = append(out, byte('\n'))
		out = append(out, []byte(strconv.Itoa(len(*data)))...)
		out = append(out, byte('\n'))
		out = append(out, *data...)
	}

	// Out it!
	return nil, &out
}
