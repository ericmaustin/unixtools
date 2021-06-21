package capacity

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	c, err := Parse("1.23TB")
	if err != nil {
		panic(err)
	}

	fmt.Println(c)

	if c.TB() != 1.23 {
		panic(fmt.Errorf("parsing failed"))
	}

	c, err = Parse("1.23MB")
	if err != nil {
		panic(err)
	}

	fmt.Println(c)

	if c.MB() != 1.23 {
		panic(fmt.Errorf("parsing failed"))
	}

	c, err = Parse("1.23KB")
	if err != nil {
		panic(err)
	}

	fmt.Println(c)

	if c.KB() != 1.23 {
		panic(fmt.Errorf("parsing failed"))
	}

	c, err = Parse("1.23GB")
	if err != nil {
		panic(err)
	}

	fmt.Println(c)

	if c.GB() != 1.23 {
		panic(fmt.Errorf("parsing failed"))
	}

	c, _ = Parse("38.2T")

	fmt.Println(c.B())

	fmt.Println(NewCapacity(41970420416512, B).TiB())
}