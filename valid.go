package apollo

import (
	"github.com/astaxie/beego/validation"
)

func Valid(args validation.ValidFormer) error {
	valid := validation.Validation{}

	if b, err := valid.Valid(args); err != nil {
		return err
	} else if !b {
		l := len(valid.Errors)
		if l != 0 {
			return valid.Errors[l-1]
		}
	}
	return nil
}
