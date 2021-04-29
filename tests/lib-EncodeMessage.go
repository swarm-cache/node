package tests

import (
	"fmt"

	"github.com/tgbv/swarm-cache/lib"
)

func TestLibEncodeMessage() {
	meta := lib.J{
		"freestyle": 22,
		"love":      "whatever",
	}
	data := []byte("Some raw binary data here..")

	err, out := lib.EncodeMessage(meta, &data)
	if err != nil {
		panic(err)
	}

	err, metaDec, dataDec := lib.DecodeMessage(out)
	if err != nil {
		panic(err)
	}

	fmt.Println(metaDec, string(*dataDec))
}
