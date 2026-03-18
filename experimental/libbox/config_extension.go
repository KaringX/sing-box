package libbox

import (
	"os"
)

func (s *platformInterfaceStub) GetAssetContent(path string) ([]byte, error) { //karing
	return nil, os.ErrInvalid
}
