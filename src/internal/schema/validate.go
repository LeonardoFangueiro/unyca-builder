package schema

import (
	"fmt"
	"github.com/xeipuuv/gojsonschema"
)

func ValidateConfig(path, schemaPath string) error {
	sl := gojsonschema.NewReferenceLoader("file://" + schemaPath)
	dl := gojsonschema.NewReferenceLoader("file://" + path)
	res, err := gojsonschema.Validate(sl, dl)
	if err != nil { return err }
	if !res.Valid() { return fmt.Errorf("config invalid: %v", res.Errors()) }
	return nil
}
