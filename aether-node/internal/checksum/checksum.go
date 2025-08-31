package checksum

import (
	"bytes"
	"encoding/hex"
	"encoding/json"

	"github.com/sirupsen/logrus"
)

type Checksum []byte

func (c *Checksum) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		logrus.Errorf("Error Unmarshalling checksum: %s", err)
		return err
	}
	*c = b
	return nil
}

func (c *Checksum) MarshalJSON() ([]byte, error) {
	s := hex.EncodeToString(*c)
	return json.Marshal(s)
}

func (c *Checksum) IsSame(another *Checksum) bool {
	return bytes.Compare(*c, *another) == 0
}
