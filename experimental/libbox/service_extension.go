// karing
package libbox

var contextId int //karing
var restartFlag bool

func SetRestart(restart bool) {
	restartFlag = restart
}

func GetRestart() bool {
	return restartFlag
}
