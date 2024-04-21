package dic

import (
	"errors"

	"github.com/sarulabs/di/v2"
	"github.com/sarulabs/dingo/v4"

	models "github.com/sarulabs/dingo/v4/tests/app/models"
	testinterfaces "github.com/sarulabs/dingo/v4/tests/app/models/testinterfaces"
)

func getDiDefs(provider dingo.Provider) []di.Def {
	return []di.Def{
		{
			Name:  "test_autofill_1",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_autofill_1")
				if err != nil {
					var eo *models.AutofillTestA
					return eo, err
				}
				pi0, ok := d.Params["Value"]
				if !ok {
					var eo *models.AutofillTestA
					return eo, errors.New("could not find parameter Value")
				}
				p0, ok := pi0.(string)
				if !ok {
					var eo *models.AutofillTestA
					return eo, errors.New("could not cast parameter Value to string")
				}
				return &models.AutofillTestA{
					Value: p0,
				}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_autofill_2",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_autofill_2")
				if err != nil {
					var eo *models.AutofillTestA
					return eo, err
				}
				pi0, ok := d.Params["Value"]
				if !ok {
					var eo *models.AutofillTestA
					return eo, errors.New("could not find parameter Value")
				}
				p0, ok := pi0.(string)
				if !ok {
					var eo *models.AutofillTestA
					return eo, errors.New("could not cast parameter Value to string")
				}
				return &models.AutofillTestA{
					Value: p0,
				}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_autofill_3",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				pi0, err := ctn.SafeGet("test_autofill_2")
				if err != nil {
					var eo *models.AutofillTestB
					return eo, err
				}
				p0, ok := pi0.(*models.AutofillTestA)
				if !ok {
					var eo *models.AutofillTestB
					return eo, errors.New("could not cast parameter Value to *models.AutofillTestA")
				}
				return &models.AutofillTestB{
					Value: p0,
				}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_build_func_1",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_build_func_1")
				if err != nil {
					var eo *models.BuildFuncTestA
					return eo, err
				}
				pi0, err := ctn.SafeGet("test_build_func_2")
				if err != nil {
					var eo *models.BuildFuncTestA
					return eo, err
				}
				p0, ok := pi0.(models.BuildFuncTestB)
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not cast parameter 0 to models.BuildFuncTestB")
				}
				pi1, err := ctn.SafeGet("test_build_func_3")
				if err != nil {
					var eo *models.BuildFuncTestA
					return eo, err
				}
				p1, ok := pi1.(*models.BuildFuncTestC)
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not cast parameter 1 to *models.BuildFuncTestC")
				}
				b, ok := d.Build.(func(models.BuildFuncTestB, *models.BuildFuncTestC) (*models.BuildFuncTestA, error))
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not cast build function to func(models.BuildFuncTestB, *models.BuildFuncTestC) (*models.BuildFuncTestA, error)")
				}
				return b(p0, p1)
			},
			Unshared: false,
		},
		{
			Name:  "test_build_func_2",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_build_func_2")
				if err != nil {
					var eo models.BuildFuncTestB
					return eo, err
				}
				pi0, err := ctn.SafeGet("test_build_func_3")
				if err != nil {
					var eo models.BuildFuncTestB
					return eo, err
				}
				p0, ok := pi0.(*models.BuildFuncTestC)
				if !ok {
					var eo models.BuildFuncTestB
					return eo, errors.New("could not cast parameter 0 to *models.BuildFuncTestC")
				}
				b, ok := d.Build.(func(*models.BuildFuncTestC) (models.BuildFuncTestB, error))
				if !ok {
					var eo models.BuildFuncTestB
					return eo, errors.New("could not cast build function to func(*models.BuildFuncTestC) (models.BuildFuncTestB, error)")
				}
				return b(p0)
			},
			Unshared: false,
		},
		{
			Name:  "test_build_func_3",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_build_func_3")
				if err != nil {
					var eo *models.BuildFuncTestC
					return eo, err
				}
				b, ok := d.Build.(func() (*models.BuildFuncTestC, error))
				if !ok {
					var eo *models.BuildFuncTestC
					return eo, errors.New("could not cast build function to func() (*models.BuildFuncTestC, error)")
				}
				return b()
			},
			Unshared: false,
		},
		{
			Name:  "test_build_func_4",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_build_func_4")
				if err != nil {
					var eo *models.BuildFuncTestA
					return eo, err
				}
				pi0, ok := d.Params["0"]
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not find parameter 0")
				}
				p0, ok := pi0.(int)
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not cast parameter 0 to int")
				}
				pi1, err := ctn.SafeGet("test_build_func_3")
				if err != nil {
					var eo *models.BuildFuncTestA
					return eo, err
				}
				p1, ok := pi1.(*models.BuildFuncTestC)
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not cast parameter 1 to *models.BuildFuncTestC")
				}
				pi2, ok := d.Params["2"]
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not find parameter 2")
				}
				p2, ok := pi2.(string)
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not cast parameter 2 to string")
				}
				b, ok := d.Build.(func(int, *models.BuildFuncTestC, string) (*models.BuildFuncTestA, error))
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not cast build function to func(int, *models.BuildFuncTestC, string) (*models.BuildFuncTestA, error)")
				}
				return b(p0, p1, p2)
			},
			Unshared: false,
		},
		{
			Name:  "test_build_func_5",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_build_func_5")
				if err != nil {
					var eo models.TypeBasedOnBasicType
					return eo, err
				}
				b, ok := d.Build.(func() (models.TypeBasedOnBasicType, error))
				if !ok {
					var eo models.TypeBasedOnBasicType
					return eo, errors.New("could not cast build function to func() (models.TypeBasedOnBasicType, error)")
				}
				return b()
			},
			Unshared: false,
		},
		{
			Name:  "test_build_func_6",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_build_func_6")
				if err != nil {
					var eo models.TypeBasedOnSliceOfBasicType
					return eo, err
				}
				b, ok := d.Build.(func() (models.TypeBasedOnSliceOfBasicType, error))
				if !ok {
					var eo models.TypeBasedOnSliceOfBasicType
					return eo, errors.New("could not cast build function to func() (models.TypeBasedOnSliceOfBasicType, error)")
				}
				return b()
			},
			Unshared: false,
		},
		{
			Name:  "test_build_func_7",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_build_func_7")
				if err != nil {
					var eo struct{}
					return eo, err
				}
				b, ok := d.Build.(func() (struct{}, error))
				if !ok {
					var eo struct{}
					return eo, errors.New("could not cast build function to func() (struct{}, error)")
				}
				return b()
			},
			Unshared: false,
		},
		{
			Name:  "test_build_func_8",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_build_func_8")
				if err != nil {
					var eo *models.BuildFuncTestA
					return eo, err
				}
				pi0, ok := d.Params["0"]
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not find parameter 0")
				}
				p0, ok := pi0.(int)
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not cast parameter 0 to int")
				}
				pi1, ok := d.Params["1"]
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not find parameter 1")
				}
				p1, ok := pi1.(*models.BuildFuncTestC)
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not cast parameter 1 to *models.BuildFuncTestC")
				}
				pi2, ok := d.Params["2"]
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not find parameter 2")
				}
				p2, ok := pi2.(string)
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not cast parameter 2 to string")
				}
				b, ok := d.Build.(func(int, *models.BuildFuncTestC, string) (*models.BuildFuncTestA, error))
				if !ok {
					var eo *models.BuildFuncTestA
					return eo, errors.New("could not cast build function to func(int, *models.BuildFuncTestC, string) (*models.BuildFuncTestA, error)")
				}
				return b(p0, p1, p2)
			},
			Unshared: false,
		},
		{
			Name:  "test_build_struct_1",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_build_struct_1")
				if err != nil {
					var eo *models.BuildStructTestA
					return eo, err
				}
				pi0, ok := d.Params["P1"]
				if !ok {
					var eo *models.BuildStructTestA
					return eo, errors.New("could not find parameter P1")
				}
				p0, ok := pi0.(string)
				if !ok {
					var eo *models.BuildStructTestA
					return eo, errors.New("could not cast parameter P1 to string")
				}
				pi1, err := ctn.SafeGet("test_build_struct_2")
				if err != nil {
					var eo *models.BuildStructTestA
					return eo, err
				}
				p1, ok := pi1.(*models.BuildStructTestB)
				if !ok {
					var eo *models.BuildStructTestA
					return eo, errors.New("could not cast parameter P2 to *models.BuildStructTestB")
				}
				pi2, err := ctn.SafeGet("test_build_struct_3")
				if err != nil {
					var eo *models.BuildStructTestA
					return eo, err
				}
				p2, ok := pi2.(*models.BuildStructTestC)
				if !ok {
					var eo *models.BuildStructTestA
					return eo, errors.New("could not cast parameter P3 to *models.BuildStructTestC")
				}
				return &models.BuildStructTestA{
					P1: p0,
					P2: p1,
					P3: p2,
				}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_build_struct_2",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_build_struct_2")
				if err != nil {
					var eo *models.BuildStructTestB
					return eo, err
				}
				pi0, ok := d.Params["P1"]
				if !ok {
					var eo *models.BuildStructTestB
					return eo, errors.New("could not find parameter P1")
				}
				p0, ok := pi0.(string)
				if !ok {
					var eo *models.BuildStructTestB
					return eo, errors.New("could not cast parameter P1 to string")
				}
				pi1, err := ctn.SafeGet("test_build_struct_3")
				if err != nil {
					var eo *models.BuildStructTestB
					return eo, err
				}
				p1, ok := pi1.(*models.BuildStructTestC)
				if !ok {
					var eo *models.BuildStructTestB
					return eo, errors.New("could not cast parameter P2 to *models.BuildStructTestC")
				}
				return &models.BuildStructTestB{
					P1: p0,
					P2: p1,
				}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_build_struct_3",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_build_struct_3")
				if err != nil {
					var eo *models.BuildStructTestC
					return eo, err
				}
				pi0, ok := d.Params["P1"]
				if !ok {
					var eo *models.BuildStructTestC
					return eo, errors.New("could not find parameter P1")
				}
				p0, ok := pi0.(string)
				if !ok {
					var eo *models.BuildStructTestC
					return eo, errors.New("could not cast parameter P1 to string")
				}
				return &models.BuildStructTestC{
					P1: p0,
				}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_build_struct_4",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_build_struct_4")
				if err != nil {
					var eo *models.BuildStructTestA
					return eo, err
				}
				pi0, ok := d.Params["P1"]
				if !ok {
					var eo *models.BuildStructTestA
					return eo, errors.New("could not find parameter P1")
				}
				p0, ok := pi0.(string)
				if !ok {
					var eo *models.BuildStructTestA
					return eo, errors.New("could not cast parameter P1 to string")
				}
				pi1, err := ctn.SafeGet("test_build_struct_2")
				if err != nil {
					var eo *models.BuildStructTestA
					return eo, err
				}
				p1, ok := pi1.(*models.BuildStructTestB)
				if !ok {
					var eo *models.BuildStructTestA
					return eo, errors.New("could not cast parameter P2 to *models.BuildStructTestB")
				}
				pi2, ok := d.Params["P3"]
				if !ok {
					var eo *models.BuildStructTestA
					return eo, errors.New("could not find parameter P3")
				}
				p2, ok := pi2.(*models.BuildStructTestC)
				if !ok {
					var eo *models.BuildStructTestA
					return eo, errors.New("could not cast parameter P3 to *models.BuildStructTestC")
				}
				return &models.BuildStructTestA{
					P1: p0,
					P2: p1,
					P3: p2,
				}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_close_1",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				var p0 bool
				return &models.CloseTest{
					Closed: p0,
				}, nil
			},
			Close: func(obj interface{}) error {
				d, err := provider.Get("test_close_1")
				if err != nil {
					return err
				}
				c, ok := d.Close.(func(*models.CloseTest) error)
				if !ok {
					return errors.New("could not cast close function to 'func(*models.CloseTest) error'")
				}
				o, ok := obj.(*models.CloseTest)
				if !ok {
					return errors.New("could not cast object to '*models.CloseTest'")
				}
				return c(o)
			},
			Unshared: false,
		},
		{
			Name:  "test_decl_type_0",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				return &models.DeclTypeTest{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_decl_type_1",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				return &models.DeclTypeTest{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_decl_type_10",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				return &models.DeclTypeTest{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_decl_type_11",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				return &models.DeclTypeTest{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_decl_type_2",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				return &models.DeclTypeTest{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_decl_type_3",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				return &models.DeclTypeTest{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_decl_type_4",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				return &models.DeclTypeTest{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_decl_type_5",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				return &models.DeclTypeTest{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_decl_type_6",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				return &models.DeclTypeTest{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_decl_type_7",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				return &models.DeclTypeTest{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_decl_type_8",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				return &models.DeclTypeTest{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_decl_type_9",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				return &models.DeclTypeTest{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_di_1",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_di_1")
				if err != nil {
					var eo models.DiTest
					return eo, err
				}
				b, ok := d.Build.(func() (models.DiTest, error))
				if !ok {
					var eo models.DiTest
					return eo, errors.New("could not cast build function to func() (models.DiTest, error)")
				}
				return b()
			},
			Unshared: false,
		},
		{
			Name:  "test_di_2",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_di_2")
				if err != nil {
					var eo models.DiTest
					return eo, err
				}
				b, ok := d.Build.(func() (models.DiTest, error))
				if !ok {
					var eo models.DiTest
					return eo, errors.New("could not cast build function to func() (models.DiTest, error)")
				}
				return b()
			},
			Unshared: false,
		},
		{
			Name:  "test_di_3",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_di_3")
				if err != nil {
					var eo models.DiTest
					return eo, err
				}
				b, ok := d.Build.(func() (models.DiTest, error))
				if !ok {
					var eo models.DiTest
					return eo, errors.New("could not cast build function to func() (models.DiTest, error)")
				}
				return b()
			},
			Unshared: false,
		},
		{
			Name:  "test_interfaces_1",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_interfaces_1")
				if err != nil {
					var eo *models.InterfacesTestB
					return eo, err
				}
				pi0, ok := d.Params["0"]
				if !ok {
					var eo *models.InterfacesTestB
					return eo, errors.New("could not find parameter 0")
				}
				p0, ok := pi0.(testinterfaces.InterfacesTestInterface)
				if !ok {
					var eo *models.InterfacesTestB
					return eo, errors.New("could not cast parameter 0 to testinterfaces.InterfacesTestInterface")
				}
				b, ok := d.Build.(func(testinterfaces.InterfacesTestInterface) (*models.InterfacesTestB, error))
				if !ok {
					var eo *models.InterfacesTestB
					return eo, errors.New("could not cast build function to func(testinterfaces.InterfacesTestInterface) (*models.InterfacesTestB, error)")
				}
				return b(p0)
			},
			Unshared: false,
		},
		{
			Name:  "test_interfaces_2",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_interfaces_2")
				if err != nil {
					var eo *models.InterfacesTestB
					return eo, err
				}
				pi0, ok := d.Params["InterfacesTestInterface"]
				if !ok {
					var eo *models.InterfacesTestB
					return eo, errors.New("could not find parameter InterfacesTestInterface")
				}
				p0, ok := pi0.(testinterfaces.InterfacesTestInterface)
				if !ok {
					var eo *models.InterfacesTestB
					return eo, errors.New("could not cast parameter InterfacesTestInterface to testinterfaces.InterfacesTestInterface")
				}
				return &models.InterfacesTestB{
					InterfacesTestInterface: p0,
				}, nil
			},
			Unshared: false,
		},
		{
			Name:  "test_retrieval_1",
			Scope: "app",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_retrieval_1")
				if err != nil {
					var eo *models.RetrievalTest
					return eo, err
				}
				b, ok := d.Build.(func() (*models.RetrievalTest, error))
				if !ok {
					var eo *models.RetrievalTest
					return eo, errors.New("could not cast build function to func() (*models.RetrievalTest, error)")
				}
				return b()
			},
			Unshared: false,
		},
		{
			Name:  "test_retrieval_2",
			Scope: "request",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_retrieval_2")
				if err != nil {
					var eo *models.RetrievalTest
					return eo, err
				}
				b, ok := d.Build.(func() (*models.RetrievalTest, error))
				if !ok {
					var eo *models.RetrievalTest
					return eo, errors.New("could not cast build function to func() (*models.RetrievalTest, error)")
				}
				return b()
			},
			Unshared: false,
		},
		{
			Name:  "test_scope_1",
			Scope: "app",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_scope_1")
				if err != nil {
					var eo *models.ScopeTest
					return eo, err
				}
				b, ok := d.Build.(func() (*models.ScopeTest, error))
				if !ok {
					var eo *models.ScopeTest
					return eo, errors.New("could not cast build function to func() (*models.ScopeTest, error)")
				}
				return b()
			},
			Unshared: false,
		},
		{
			Name:  "test_scope_2",
			Scope: "request",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_scope_2")
				if err != nil {
					var eo *models.ScopeTest
					return eo, err
				}
				b, ok := d.Build.(func() (*models.ScopeTest, error))
				if !ok {
					var eo *models.ScopeTest
					return eo, errors.New("could not cast build function to func() (*models.ScopeTest, error)")
				}
				return b()
			},
			Unshared: false,
		},
		{
			Name:  "test_unshared_1",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_unshared_1")
				if err != nil {
					var eo *models.UnsharedTest
					return eo, err
				}
				b, ok := d.Build.(func() (*models.UnsharedTest, error))
				if !ok {
					var eo *models.UnsharedTest
					return eo, errors.New("could not cast build function to func() (*models.UnsharedTest, error)")
				}
				return b()
			},
			Unshared: true,
		},
		{
			Name:  "test_unshared_2",
			Scope: "",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("test_unshared_2")
				if err != nil {
					var eo *models.UnsharedTest
					return eo, err
				}
				b, ok := d.Build.(func() (*models.UnsharedTest, error))
				if !ok {
					var eo *models.UnsharedTest
					return eo, errors.New("could not cast build function to func() (*models.UnsharedTest, error)")
				}
				return b()
			},
			Unshared: false,
		},
	}
}
