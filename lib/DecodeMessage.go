package lib

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/tgbv/swarm-cache/glob"
)

// The general purpose json type
type J = glob.J

// Splits an incoming message into meta and data (binary).
//
// Input must be of format: [valid json]\n[data length]\n[binary data here]
//
// Verified the message as well. If it's length is the same as within boundary returns error.
func DecodeMessage(input *[]byte) (error, J, *[]byte) {
	meta := make([]byte, 0)
	data := make([]byte, 0)
	providedLength := make([]byte, 0)

	// split
	nCount := 0
	for i := range *input {
		if (*input)[i] == byte('\n') {
			nCount++
			continue
		}

		if nCount == 0 {
			meta = append(meta, (*input)[i])
		}

		if nCount == 1 {
			providedLength = append(providedLength, (*input)[i])
		}

		if nCount > 1 {
			data = append(data, (*input)[i])
		}
	}

	// check length if there is any
	if len(providedLength) > 0 {
		actualProvLength, err := strconv.Atoi(string(providedLength))
		if err != nil || len(data) != actualProvLength {
			return fmt.Errorf("Corrupted input! (length/data)"), J{}, nil
		}
	}

	// deserialize meta
	metaStr := string(meta)
	metaStr = strings.ReplaceAll(metaStr, glob.META_NL, "\n")
	meta = []byte(metaStr)

	// serialize meta into map
	metaOut := J{}
	err := json.Unmarshal(meta, &metaOut)
	if err != nil {
		return fmt.Errorf("Corrupted input! (json)"), J{}, nil
	}

	return nil, metaOut, &data
}
