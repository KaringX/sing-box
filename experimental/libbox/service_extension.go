//go:build with_karing

package libbox

func (w *platformInterfaceWrapper) GetAssetContent(path string) ([]byte, error) { //karing
	return w.iif.GetAssetContent(path)
}
