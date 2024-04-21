package validation

import "github.com/dembygenesis/platform_engineer_exam/src/utils/date_handling"

type CustomValidation struct {
	Name     string
	Logic    func(i interface{}) bool
	Response string
}

var customValidations = []CustomValidation{
	{
		Name: "date_format",
		Logic: func(i interface{}) bool {
			val, ok := i.(string)
			if !ok {
				return false
			}
			return date_handling.ValidDate(val)
		},
		Response: "must be a valid date format of YYYY-MM-DD",
	},
}
