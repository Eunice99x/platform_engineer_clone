package validation

import (
	"fmt"
	"github.com/dembygenesis/platform_engineer_exam/src/utils/data"
	strings2 "github.com/dembygenesis/platform_engineer_exam/src/utils/strings"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/ssoroka/slice"
	"reflect"
	"strings"
)

type Validator interface {
	ValidateStruct(i interface{})
}

var (
	uni      *ut.UniversalTranslator
	validate *validator.Validate
	trans    ut.Translator
)

// configValidate adds custom validator rules, and custom tag name returns
func configValidate() {
	uni = ut.New(en.New(), en.New())
	trans, _ = uni.GetTranslator("en")
	validate = validator.New()

	setInitialConfiguration(validate)
	addCustomValidations(validate, &customValidations)
}

func SliceIsUnique(_slice interface{}) (bool, error) {
	var err error
	var sliceIsUnique bool
	var arrVals []string

	// We convert the _slice to an array of strings to guarantee compatibility with
	// our SortedUnique helper function
	arrVals, err = data.GetSliceValuesAsSliceOfStrings(_slice)
	if err != nil {
		return false, errors.Wrap(err, "error trying to convert slice as a string slice: "+strings2.GetJSON(_slice))
	}

	arrValsUnique := slice.SortedUnique(arrVals)
	lenOfUniqueVals := 0
	for index, _ := range arrValsUnique {
		lenOfUniqueVals = index + 1
	}
	sliceIsUnique = lenOfUniqueVals == len(arrVals)
	return sliceIsUnique, nil
}

// ValidateStructParams validates the struct validation rules
// provided in the tag using the "validator" library
func ValidateStructParams(p interface{}) ([]string, error) {
	var missingParams []string

	_structVal, err := data.GetStructAsValue(p)
	if err != nil {
		return nil, err
	}

	err = validate.Struct(_structVal)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		for _, err := range validationErrors {
			missingParams = append(missingParams, err.Translate(trans))
		}
	}
	return missingParams, nil
}

func init() {
	configValidate()
}

// registerCustomTranslations adds a custom response for the specific validator "name"
func registerCustomTranslations(v *validator.Validate, name, response string) {
	_ = v.RegisterTranslation(name, trans, func(ut ut.Translator) error {
		return ut.Add(name, fmt.Sprintf("{0} %v", response), true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(name, fe.Field())
		return t
	})
}

// registerCustomValidations adds a custom logic
func registerCustomValidations(v *validator.Validate, name string, logic func(i interface{}) bool) {
	_ = v.RegisterValidation(name, func(fl validator.FieldLevel) bool {
		return logic(fl.Field().Interface())
	})
}

func addCustomValidations(v *validator.Validate, customValidations *[]CustomValidation) {
	for _, customValidation := range *customValidations {
		registerCustomValidations(v, customValidation.Name, customValidation.Logic)
		registerCustomTranslations(v, customValidation.Name, customValidation.Response)
	}
}

func setInitialConfiguration(v *validator.Validate) {
	// Display the "json" tag instead of the struct tag  when there are errors
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Set the "required" response error to something more user-friendly
	_ = v.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} must have a value", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	})
}
