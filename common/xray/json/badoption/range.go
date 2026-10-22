package badoption

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/sagernet/sing-box/common/xray/crypto"
	E "github.com/sagernet/sing/common/exceptions"
)

type Range struct {
	From int `json:"from"`
	To   int `json:"to"`
}

func (c *Range) Build() *Range {
	return (*Range)(c)
}

func (c *Range) MarshalJSON() ([]byte, error) {
	if c.From == c.To {
		return json.Marshal(c.From)
	}
	return json.Marshal(fmt.Sprintf("%d-%d", c.From, c.To))
}

func (c *Range) UnmarshalJSON(content []byte) error {
	var rangeValue struct {
		From int `json:"from"`
		To   int `json:"to"`
	}
	var stringValue string
	err := json.Unmarshal(content, &stringValue)
	if err == nil {
		parts := strings.Split(stringValue, "-")
		if len(parts) != 2 {
			from, err := strconv.Atoi(parts[0])
			if err != nil {
				return err
			}
			rangeValue.From, rangeValue.To = from, from
		} else {
			from, err := strconv.Atoi(parts[0])
			if err != nil {
				return err
			}
			to, err := strconv.Atoi(parts[1])
			if err != nil {
				return err
			}
			rangeValue.From, rangeValue.To = from, to
		}
	} else {
		var intValue int
		err := json.Unmarshal(content, &intValue)
		if err == nil {
			rangeValue.From, rangeValue.To = intValue, intValue
		} else {
			err := json.Unmarshal(content, &rangeValue)
			if err != nil {
				return err
			}
		}
	}
	if rangeValue.From > rangeValue.To {
		return E.New("invalid range")
	}
	*c = Range{rangeValue.From, rangeValue.To}
	return nil
}

func (c *Range) String() string {
	if c.From == c.To {
		return strconv.FormatInt(int64(c.From), 10)
	}
	return fmt.Sprintf("%d-%d", c.From, c.To)
}
func (c Range) Rand() int {
	return int(crypto.RandBetween(int64(c.From), int64(c.To)))
}
