package tests

import (
	"fmt"

	"github.com/tgbv/swarm-cache/lib"
)

func TestLibDecodeMessage() {

	message := []byte(`{"freestyler":"rap rap pac pac", "tupac": 22}` + "\n" + "11" + "\n" + "I love you.")

	fmt.Println(lib.DecodeMessage(&message))
}
