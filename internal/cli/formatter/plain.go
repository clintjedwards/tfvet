package formatter

import (
	"encoding/json"
	"fmt"

	"github.com/TylerBrock/colorjson"
)

type plainPrinter struct{}

func newPlainPrinter() plainPrinter {
	return plainPrinter{}
}

// print outputs the object as a pretty json string
// Because the colorjson package only interprets the type map[string]interface{} correctly
// we need to first pass it through the json marshalling process to turn the object given
// into the correct type
func (pp *plainPrinter) print(obj interface{}) {
	rawJSON, _ := json.Marshal(obj)

	var tmp map[string]interface{}
	_ = json.Unmarshal(rawJSON, &tmp)

	fmtter := colorjson.NewFormatter()
	fmtter.Indent = 2
	s, _ := fmtter.Marshal(tmp)

	fmt.Println(string(s))
}

func (pp *plainPrinter) printMsg(msg string) {
	obj := map[string]interface{}{"msg": msg}
	s, _ := colorjson.Marshal(obj)
	fmt.Println(string(s))
}

func (pp *plainPrinter) printErr(msg string) {
	obj := map[string]interface{}{"msg": msg}
	s, _ := colorjson.Marshal(obj)
	fmt.Println(string(s))
}
