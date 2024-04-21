package dic

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/sarulabs/di/v2"
	"github.com/sarulabs/dingo/v4"

	providerPkg "github.com/sarulabs/dingo/v4/tests/app/services/provider"

	models "github.com/sarulabs/dingo/v4/tests/app/models"
)

// C retrieves a Container from an interface.
// The function panics if the Container can not be retrieved.
//
// The interface can be :
//   - a *Container
//   - an *http.Request containing a *Container in its context.Context
//     for the dingo.ContainerKey("dingo") key.
//
// The function can be changed to match the needs of your application.
var C = func(i interface{}) *Container {
	if c, ok := i.(*Container); ok {
		return c
	}
	r, ok := i.(*http.Request)
	if !ok {
		panic("could not get the container with dic.C()")
	}
	c, ok := r.Context().Value(dingo.ContainerKey("dingo")).(*Container)
	if !ok {
		panic("could not get the container from the given *http.Request in dic.C()")
	}
	return c
}

type builder struct {
	builder *di.Builder
}

// NewBuilder creates a builder that can be used to create a Container.
// You probably should use NewContainer to create the container directly.
// But using NewBuilder allows you to redefine some di services.
// This can be used for testing.
// But this behavior is not safe, so be sure to know what you are doing.
func NewBuilder(scopes ...string) (*builder, error) {
	if len(scopes) == 0 {
		scopes = []string{di.App, di.Request, di.SubRequest}
	}
	b, err := di.NewBuilder(scopes...)
	if err != nil {
		return nil, fmt.Errorf("could not create di.Builder: %v", err)
	}
	provider := &providerPkg.Provider{}
	if err := provider.Load(); err != nil {
		return nil, fmt.Errorf("could not load definitions with the Provider (Provider from github.com/sarulabs/dingo/v4/tests/app/services/provider): %v", err)
	}
	for _, d := range getDiDefs(provider) {
		if err := b.Add(d); err != nil {
			return nil, fmt.Errorf("could not add di.Def in di.Builder: %v", err)
		}
	}
	return &builder{builder: b}, nil
}

// Add adds one or more definitions in the Builder.
// It returns an error if a definition can not be added.
func (b *builder) Add(defs ...di.Def) error {
	return b.builder.Add(defs...)
}

// Set is a shortcut to add a definition for an already built object.
func (b *builder) Set(name string, obj interface{}) error {
	return b.builder.Set(name, obj)
}

// Build creates a Container in the most generic scope.
func (b *builder) Build() *Container {
	return &Container{ctn: b.builder.Build()}
}

// NewContainer creates a new Container.
// If no scope is provided, di.App, di.Request and di.SubRequest are used.
// The returned Container has the most generic scope (di.App).
// The SubContainer() method should be called to get a Container in a more specific scope.
func NewContainer(scopes ...string) (*Container, error) {
	b, err := NewBuilder(scopes...)
	if err != nil {
		return nil, err
	}
	return b.Build(), nil
}

// Container represents a generated dependency injection container.
// It is a wrapper around a di.Container.
//
// A Container has a scope and may have a parent in a more generic scope
// and children in a more specific scope.
// Objects can be retrieved from the Container.
// If the requested object does not already exist in the Container,
// it is built thanks to the object definition.
// The following attempts to get this object will return the same object.
type Container struct {
	ctn di.Container
}

// Scope returns the Container scope.
func (c *Container) Scope() string {
	return c.ctn.Scope()
}

// Scopes returns the list of available scopes.
func (c *Container) Scopes() []string {
	return c.ctn.Scopes()
}

// ParentScopes returns the list of scopes wider than the Container scope.
func (c *Container) ParentScopes() []string {
	return c.ctn.ParentScopes()
}

// SubScopes returns the list of scopes that are more specific than the Container scope.
func (c *Container) SubScopes() []string {
	return c.ctn.SubScopes()
}

// Parent returns the parent Container.
func (c *Container) Parent() *Container {
	if p := c.ctn.Parent(); p != nil {
		return &Container{ctn: p}
	}
	return nil
}

// SubContainer creates a new Container in the next sub-scope
// that will have this Container as parent.
func (c *Container) SubContainer() (*Container, error) {
	sub, err := c.ctn.SubContainer()
	if err != nil {
		return nil, err
	}
	return &Container{ctn: sub}, nil
}

// SafeGet retrieves an object from the Container.
// The object has to belong to this scope or a more generic one.
// If the object does not already exist, it is created and saved in the Container.
// If the object can not be created, it returns an error.
func (c *Container) SafeGet(name string) (interface{}, error) {
	return c.ctn.SafeGet(name)
}

// Get is similar to SafeGet but it does not return the error.
// Instead it panics.
func (c *Container) Get(name string) interface{} {
	return c.ctn.Get(name)
}

// Fill is similar to SafeGet but it does not return the object.
// Instead it fills the provided object with the value returned by SafeGet.
// The provided object must be a pointer to the value returned by SafeGet.
func (c *Container) Fill(name string, dst interface{}) error {
	return c.ctn.Fill(name, dst)
}

// UnscopedSafeGet retrieves an object from the Container, like SafeGet.
// The difference is that the object can be retrieved
// even if it belongs to a more specific scope.
// To do so, UnscopedSafeGet creates a sub-container.
// When the created object is no longer needed,
// it is important to use the Clean method to delete this sub-container.
func (c *Container) UnscopedSafeGet(name string) (interface{}, error) {
	return c.ctn.UnscopedSafeGet(name)
}

// UnscopedGet is similar to UnscopedSafeGet but it does not return the error.
// Instead it panics.
func (c *Container) UnscopedGet(name string) interface{} {
	return c.ctn.UnscopedGet(name)
}

// UnscopedFill is similar to UnscopedSafeGet but copies the object in dst instead of returning it.
func (c *Container) UnscopedFill(name string, dst interface{}) error {
	return c.ctn.UnscopedFill(name, dst)
}

// Clean deletes the sub-container created by UnscopedSafeGet, UnscopedGet or UnscopedFill.
func (c *Container) Clean() error {
	return c.ctn.Clean()
}

// DeleteWithSubContainers takes all the objects saved in this Container
// and calls the Close function of their Definition on them.
// It will also call DeleteWithSubContainers on each child and remove its reference in the parent Container.
// After deletion, the Container can no longer be used.
// The sub-containers are deleted even if they are still used in other goroutines.
// It can cause errors. You may want to use the Delete method instead.
func (c *Container) DeleteWithSubContainers() error {
	return c.ctn.DeleteWithSubContainers()
}

// Delete works like DeleteWithSubContainers if the Container does not have any child.
// But if the Container has sub-containers, it will not be deleted right away.
// The deletion only occurs when all the sub-containers have been deleted manually.
// So you have to call Delete or DeleteWithSubContainers on all the sub-containers.
func (c *Container) Delete() error {
	return c.ctn.Delete()
}

// IsClosed returns true if the Container has been deleted.
func (c *Container) IsClosed() bool {
	return c.ctn.IsClosed()
}

// SafeGetTestAutofill1 retrieves the "test_autofill_1" object from the main scope.
//
// Test description.
//
// Even on multiple lines.
//
// ---------------------------------------------
//
//	name: "test_autofill_1"
//	type: *models.AutofillTestA
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestAutofill1() (*models.AutofillTestA, error) {
	i, err := c.ctn.SafeGet("test_autofill_1")
	if err != nil {
		var eo *models.AutofillTestA
		return eo, err
	}
	o, ok := i.(*models.AutofillTestA)
	if !ok {
		return o, errors.New("could get 'test_autofill_1' because the object could not be cast to *models.AutofillTestA")
	}
	return o, nil
}

// GetTestAutofill1 retrieves the "test_autofill_1" object from the main scope.
//
// Test description.
//
// Even on multiple lines.
//
// ---------------------------------------------
//
//	name: "test_autofill_1"
//	type: *models.AutofillTestA
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestAutofill1() *models.AutofillTestA {
	o, err := c.SafeGetTestAutofill1()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestAutofill1 retrieves the "test_autofill_1" object from the main scope.
//
// Test description.
//
// Even on multiple lines.
//
// ---------------------------------------------
//
//	name: "test_autofill_1"
//	type: *models.AutofillTestA
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestAutofill1() (*models.AutofillTestA, error) {
	i, err := c.ctn.UnscopedSafeGet("test_autofill_1")
	if err != nil {
		var eo *models.AutofillTestA
		return eo, err
	}
	o, ok := i.(*models.AutofillTestA)
	if !ok {
		return o, errors.New("could get 'test_autofill_1' because the object could not be cast to *models.AutofillTestA")
	}
	return o, nil
}

// UnscopedGetTestAutofill1 retrieves the "test_autofill_1" object from the main scope.
//
// Test description.
//
// Even on multiple lines.
//
// ---------------------------------------------
//
//	name: "test_autofill_1"
//	type: *models.AutofillTestA
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestAutofill1() *models.AutofillTestA {
	o, err := c.UnscopedSafeGetTestAutofill1()
	if err != nil {
		panic(err)
	}
	return o
}

// TestAutofill1 retrieves the "test_autofill_1" object from the main scope.
//
// Test description.
//
// Even on multiple lines.
//
// ---------------------------------------------
//
//	name: "test_autofill_1"
//	type: *models.AutofillTestA
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestAutofill1 method.
// If the container can not be retrieved, it panics.
func TestAutofill1(i interface{}) *models.AutofillTestA {
	return C(i).GetTestAutofill1()
}

// SafeGetTestAutofill2 retrieves the "test_autofill_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_autofill_2"
//	type: *models.AutofillTestA
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestAutofill2() (*models.AutofillTestA, error) {
	i, err := c.ctn.SafeGet("test_autofill_2")
	if err != nil {
		var eo *models.AutofillTestA
		return eo, err
	}
	o, ok := i.(*models.AutofillTestA)
	if !ok {
		return o, errors.New("could get 'test_autofill_2' because the object could not be cast to *models.AutofillTestA")
	}
	return o, nil
}

// GetTestAutofill2 retrieves the "test_autofill_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_autofill_2"
//	type: *models.AutofillTestA
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestAutofill2() *models.AutofillTestA {
	o, err := c.SafeGetTestAutofill2()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestAutofill2 retrieves the "test_autofill_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_autofill_2"
//	type: *models.AutofillTestA
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestAutofill2() (*models.AutofillTestA, error) {
	i, err := c.ctn.UnscopedSafeGet("test_autofill_2")
	if err != nil {
		var eo *models.AutofillTestA
		return eo, err
	}
	o, ok := i.(*models.AutofillTestA)
	if !ok {
		return o, errors.New("could get 'test_autofill_2' because the object could not be cast to *models.AutofillTestA")
	}
	return o, nil
}

// UnscopedGetTestAutofill2 retrieves the "test_autofill_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_autofill_2"
//	type: *models.AutofillTestA
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestAutofill2() *models.AutofillTestA {
	o, err := c.UnscopedSafeGetTestAutofill2()
	if err != nil {
		panic(err)
	}
	return o
}

// TestAutofill2 retrieves the "test_autofill_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_autofill_2"
//	type: *models.AutofillTestA
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestAutofill2 method.
// If the container can not be retrieved, it panics.
func TestAutofill2(i interface{}) *models.AutofillTestA {
	return C(i).GetTestAutofill2()
}

// SafeGetTestAutofill3 retrieves the "test_autofill_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_autofill_3"
//	type: *models.AutofillTestB
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Service(*models.AutofillTestA) ["test_autofill_2"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestAutofill3() (*models.AutofillTestB, error) {
	i, err := c.ctn.SafeGet("test_autofill_3")
	if err != nil {
		var eo *models.AutofillTestB
		return eo, err
	}
	o, ok := i.(*models.AutofillTestB)
	if !ok {
		return o, errors.New("could get 'test_autofill_3' because the object could not be cast to *models.AutofillTestB")
	}
	return o, nil
}

// GetTestAutofill3 retrieves the "test_autofill_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_autofill_3"
//	type: *models.AutofillTestB
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Service(*models.AutofillTestA) ["test_autofill_2"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestAutofill3() *models.AutofillTestB {
	o, err := c.SafeGetTestAutofill3()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestAutofill3 retrieves the "test_autofill_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_autofill_3"
//	type: *models.AutofillTestB
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Service(*models.AutofillTestA) ["test_autofill_2"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestAutofill3() (*models.AutofillTestB, error) {
	i, err := c.ctn.UnscopedSafeGet("test_autofill_3")
	if err != nil {
		var eo *models.AutofillTestB
		return eo, err
	}
	o, ok := i.(*models.AutofillTestB)
	if !ok {
		return o, errors.New("could get 'test_autofill_3' because the object could not be cast to *models.AutofillTestB")
	}
	return o, nil
}

// UnscopedGetTestAutofill3 retrieves the "test_autofill_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_autofill_3"
//	type: *models.AutofillTestB
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Service(*models.AutofillTestA) ["test_autofill_2"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestAutofill3() *models.AutofillTestB {
	o, err := c.UnscopedSafeGetTestAutofill3()
	if err != nil {
		panic(err)
	}
	return o
}

// TestAutofill3 retrieves the "test_autofill_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_autofill_3"
//	type: *models.AutofillTestB
//	scope: "main"
//	build: struct
//	params:
//		- "Value": Service(*models.AutofillTestA) ["test_autofill_2"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestAutofill3 method.
// If the container can not be retrieved, it panics.
func TestAutofill3(i interface{}) *models.AutofillTestB {
	return C(i).GetTestAutofill3()
}

// SafeGetTestBuildFunc1 retrieves the "test_build_func_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_1"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Service(models.BuildFuncTestB) ["test_build_func_2"]
//		- "1": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestBuildFunc1() (*models.BuildFuncTestA, error) {
	i, err := c.ctn.SafeGet("test_build_func_1")
	if err != nil {
		var eo *models.BuildFuncTestA
		return eo, err
	}
	o, ok := i.(*models.BuildFuncTestA)
	if !ok {
		return o, errors.New("could get 'test_build_func_1' because the object could not be cast to *models.BuildFuncTestA")
	}
	return o, nil
}

// GetTestBuildFunc1 retrieves the "test_build_func_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_1"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Service(models.BuildFuncTestB) ["test_build_func_2"]
//		- "1": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestBuildFunc1() *models.BuildFuncTestA {
	o, err := c.SafeGetTestBuildFunc1()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestBuildFunc1 retrieves the "test_build_func_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_1"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Service(models.BuildFuncTestB) ["test_build_func_2"]
//		- "1": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestBuildFunc1() (*models.BuildFuncTestA, error) {
	i, err := c.ctn.UnscopedSafeGet("test_build_func_1")
	if err != nil {
		var eo *models.BuildFuncTestA
		return eo, err
	}
	o, ok := i.(*models.BuildFuncTestA)
	if !ok {
		return o, errors.New("could get 'test_build_func_1' because the object could not be cast to *models.BuildFuncTestA")
	}
	return o, nil
}

// UnscopedGetTestBuildFunc1 retrieves the "test_build_func_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_1"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Service(models.BuildFuncTestB) ["test_build_func_2"]
//		- "1": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestBuildFunc1() *models.BuildFuncTestA {
	o, err := c.UnscopedSafeGetTestBuildFunc1()
	if err != nil {
		panic(err)
	}
	return o
}

// TestBuildFunc1 retrieves the "test_build_func_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_1"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Service(models.BuildFuncTestB) ["test_build_func_2"]
//		- "1": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestBuildFunc1 method.
// If the container can not be retrieved, it panics.
func TestBuildFunc1(i interface{}) *models.BuildFuncTestA {
	return C(i).GetTestBuildFunc1()
}

// SafeGetTestBuildFunc2 retrieves the "test_build_func_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_2"
//	type: models.BuildFuncTestB
//	scope: "main"
//	build: func
//	params:
//		- "0": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestBuildFunc2() (models.BuildFuncTestB, error) {
	i, err := c.ctn.SafeGet("test_build_func_2")
	if err != nil {
		var eo models.BuildFuncTestB
		return eo, err
	}
	o, ok := i.(models.BuildFuncTestB)
	if !ok {
		return o, errors.New("could get 'test_build_func_2' because the object could not be cast to models.BuildFuncTestB")
	}
	return o, nil
}

// GetTestBuildFunc2 retrieves the "test_build_func_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_2"
//	type: models.BuildFuncTestB
//	scope: "main"
//	build: func
//	params:
//		- "0": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestBuildFunc2() models.BuildFuncTestB {
	o, err := c.SafeGetTestBuildFunc2()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestBuildFunc2 retrieves the "test_build_func_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_2"
//	type: models.BuildFuncTestB
//	scope: "main"
//	build: func
//	params:
//		- "0": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestBuildFunc2() (models.BuildFuncTestB, error) {
	i, err := c.ctn.UnscopedSafeGet("test_build_func_2")
	if err != nil {
		var eo models.BuildFuncTestB
		return eo, err
	}
	o, ok := i.(models.BuildFuncTestB)
	if !ok {
		return o, errors.New("could get 'test_build_func_2' because the object could not be cast to models.BuildFuncTestB")
	}
	return o, nil
}

// UnscopedGetTestBuildFunc2 retrieves the "test_build_func_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_2"
//	type: models.BuildFuncTestB
//	scope: "main"
//	build: func
//	params:
//		- "0": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestBuildFunc2() models.BuildFuncTestB {
	o, err := c.UnscopedSafeGetTestBuildFunc2()
	if err != nil {
		panic(err)
	}
	return o
}

// TestBuildFunc2 retrieves the "test_build_func_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_2"
//	type: models.BuildFuncTestB
//	scope: "main"
//	build: func
//	params:
//		- "0": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestBuildFunc2 method.
// If the container can not be retrieved, it panics.
func TestBuildFunc2(i interface{}) models.BuildFuncTestB {
	return C(i).GetTestBuildFunc2()
}

// SafeGetTestBuildFunc3 retrieves the "test_build_func_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_3"
//	type: *models.BuildFuncTestC
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestBuildFunc3() (*models.BuildFuncTestC, error) {
	i, err := c.ctn.SafeGet("test_build_func_3")
	if err != nil {
		var eo *models.BuildFuncTestC
		return eo, err
	}
	o, ok := i.(*models.BuildFuncTestC)
	if !ok {
		return o, errors.New("could get 'test_build_func_3' because the object could not be cast to *models.BuildFuncTestC")
	}
	return o, nil
}

// GetTestBuildFunc3 retrieves the "test_build_func_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_3"
//	type: *models.BuildFuncTestC
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestBuildFunc3() *models.BuildFuncTestC {
	o, err := c.SafeGetTestBuildFunc3()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestBuildFunc3 retrieves the "test_build_func_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_3"
//	type: *models.BuildFuncTestC
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestBuildFunc3() (*models.BuildFuncTestC, error) {
	i, err := c.ctn.UnscopedSafeGet("test_build_func_3")
	if err != nil {
		var eo *models.BuildFuncTestC
		return eo, err
	}
	o, ok := i.(*models.BuildFuncTestC)
	if !ok {
		return o, errors.New("could get 'test_build_func_3' because the object could not be cast to *models.BuildFuncTestC")
	}
	return o, nil
}

// UnscopedGetTestBuildFunc3 retrieves the "test_build_func_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_3"
//	type: *models.BuildFuncTestC
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestBuildFunc3() *models.BuildFuncTestC {
	o, err := c.UnscopedSafeGetTestBuildFunc3()
	if err != nil {
		panic(err)
	}
	return o
}

// TestBuildFunc3 retrieves the "test_build_func_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_3"
//	type: *models.BuildFuncTestC
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestBuildFunc3 method.
// If the container can not be retrieved, it panics.
func TestBuildFunc3(i interface{}) *models.BuildFuncTestC {
	return C(i).GetTestBuildFunc3()
}

// SafeGetTestBuildFunc4 retrieves the "test_build_func_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_4"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(int)
//		- "1": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//		- "2": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestBuildFunc4() (*models.BuildFuncTestA, error) {
	i, err := c.ctn.SafeGet("test_build_func_4")
	if err != nil {
		var eo *models.BuildFuncTestA
		return eo, err
	}
	o, ok := i.(*models.BuildFuncTestA)
	if !ok {
		return o, errors.New("could get 'test_build_func_4' because the object could not be cast to *models.BuildFuncTestA")
	}
	return o, nil
}

// GetTestBuildFunc4 retrieves the "test_build_func_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_4"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(int)
//		- "1": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//		- "2": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestBuildFunc4() *models.BuildFuncTestA {
	o, err := c.SafeGetTestBuildFunc4()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestBuildFunc4 retrieves the "test_build_func_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_4"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(int)
//		- "1": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//		- "2": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestBuildFunc4() (*models.BuildFuncTestA, error) {
	i, err := c.ctn.UnscopedSafeGet("test_build_func_4")
	if err != nil {
		var eo *models.BuildFuncTestA
		return eo, err
	}
	o, ok := i.(*models.BuildFuncTestA)
	if !ok {
		return o, errors.New("could get 'test_build_func_4' because the object could not be cast to *models.BuildFuncTestA")
	}
	return o, nil
}

// UnscopedGetTestBuildFunc4 retrieves the "test_build_func_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_4"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(int)
//		- "1": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//		- "2": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestBuildFunc4() *models.BuildFuncTestA {
	o, err := c.UnscopedSafeGetTestBuildFunc4()
	if err != nil {
		panic(err)
	}
	return o
}

// TestBuildFunc4 retrieves the "test_build_func_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_4"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(int)
//		- "1": Service(*models.BuildFuncTestC) ["test_build_func_3"]
//		- "2": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestBuildFunc4 method.
// If the container can not be retrieved, it panics.
func TestBuildFunc4(i interface{}) *models.BuildFuncTestA {
	return C(i).GetTestBuildFunc4()
}

// SafeGetTestBuildFunc5 retrieves the "test_build_func_5" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_5"
//	type: models.TypeBasedOnBasicType
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestBuildFunc5() (models.TypeBasedOnBasicType, error) {
	i, err := c.ctn.SafeGet("test_build_func_5")
	if err != nil {
		var eo models.TypeBasedOnBasicType
		return eo, err
	}
	o, ok := i.(models.TypeBasedOnBasicType)
	if !ok {
		return o, errors.New("could get 'test_build_func_5' because the object could not be cast to models.TypeBasedOnBasicType")
	}
	return o, nil
}

// GetTestBuildFunc5 retrieves the "test_build_func_5" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_5"
//	type: models.TypeBasedOnBasicType
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestBuildFunc5() models.TypeBasedOnBasicType {
	o, err := c.SafeGetTestBuildFunc5()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestBuildFunc5 retrieves the "test_build_func_5" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_5"
//	type: models.TypeBasedOnBasicType
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestBuildFunc5() (models.TypeBasedOnBasicType, error) {
	i, err := c.ctn.UnscopedSafeGet("test_build_func_5")
	if err != nil {
		var eo models.TypeBasedOnBasicType
		return eo, err
	}
	o, ok := i.(models.TypeBasedOnBasicType)
	if !ok {
		return o, errors.New("could get 'test_build_func_5' because the object could not be cast to models.TypeBasedOnBasicType")
	}
	return o, nil
}

// UnscopedGetTestBuildFunc5 retrieves the "test_build_func_5" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_5"
//	type: models.TypeBasedOnBasicType
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestBuildFunc5() models.TypeBasedOnBasicType {
	o, err := c.UnscopedSafeGetTestBuildFunc5()
	if err != nil {
		panic(err)
	}
	return o
}

// TestBuildFunc5 retrieves the "test_build_func_5" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_5"
//	type: models.TypeBasedOnBasicType
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestBuildFunc5 method.
// If the container can not be retrieved, it panics.
func TestBuildFunc5(i interface{}) models.TypeBasedOnBasicType {
	return C(i).GetTestBuildFunc5()
}

// SafeGetTestBuildFunc6 retrieves the "test_build_func_6" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_6"
//	type: models.TypeBasedOnSliceOfBasicType
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestBuildFunc6() (models.TypeBasedOnSliceOfBasicType, error) {
	i, err := c.ctn.SafeGet("test_build_func_6")
	if err != nil {
		var eo models.TypeBasedOnSliceOfBasicType
		return eo, err
	}
	o, ok := i.(models.TypeBasedOnSliceOfBasicType)
	if !ok {
		return o, errors.New("could get 'test_build_func_6' because the object could not be cast to models.TypeBasedOnSliceOfBasicType")
	}
	return o, nil
}

// GetTestBuildFunc6 retrieves the "test_build_func_6" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_6"
//	type: models.TypeBasedOnSliceOfBasicType
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestBuildFunc6() models.TypeBasedOnSliceOfBasicType {
	o, err := c.SafeGetTestBuildFunc6()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestBuildFunc6 retrieves the "test_build_func_6" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_6"
//	type: models.TypeBasedOnSliceOfBasicType
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestBuildFunc6() (models.TypeBasedOnSliceOfBasicType, error) {
	i, err := c.ctn.UnscopedSafeGet("test_build_func_6")
	if err != nil {
		var eo models.TypeBasedOnSliceOfBasicType
		return eo, err
	}
	o, ok := i.(models.TypeBasedOnSliceOfBasicType)
	if !ok {
		return o, errors.New("could get 'test_build_func_6' because the object could not be cast to models.TypeBasedOnSliceOfBasicType")
	}
	return o, nil
}

// UnscopedGetTestBuildFunc6 retrieves the "test_build_func_6" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_6"
//	type: models.TypeBasedOnSliceOfBasicType
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestBuildFunc6() models.TypeBasedOnSliceOfBasicType {
	o, err := c.UnscopedSafeGetTestBuildFunc6()
	if err != nil {
		panic(err)
	}
	return o
}

// TestBuildFunc6 retrieves the "test_build_func_6" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_6"
//	type: models.TypeBasedOnSliceOfBasicType
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestBuildFunc6 method.
// If the container can not be retrieved, it panics.
func TestBuildFunc6(i interface{}) models.TypeBasedOnSliceOfBasicType {
	return C(i).GetTestBuildFunc6()
}

// SafeGetTestBuildFunc7 retrieves the "test_build_func_7" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_7"
//	type: struct{}
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestBuildFunc7() (struct{}, error) {
	i, err := c.ctn.SafeGet("test_build_func_7")
	if err != nil {
		var eo struct{}
		return eo, err
	}
	o, ok := i.(struct{})
	if !ok {
		return o, errors.New("could get 'test_build_func_7' because the object could not be cast to struct{}")
	}
	return o, nil
}

// GetTestBuildFunc7 retrieves the "test_build_func_7" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_7"
//	type: struct{}
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestBuildFunc7() struct{} {
	o, err := c.SafeGetTestBuildFunc7()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestBuildFunc7 retrieves the "test_build_func_7" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_7"
//	type: struct{}
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestBuildFunc7() (struct{}, error) {
	i, err := c.ctn.UnscopedSafeGet("test_build_func_7")
	if err != nil {
		var eo struct{}
		return eo, err
	}
	o, ok := i.(struct{})
	if !ok {
		return o, errors.New("could get 'test_build_func_7' because the object could not be cast to struct{}")
	}
	return o, nil
}

// UnscopedGetTestBuildFunc7 retrieves the "test_build_func_7" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_7"
//	type: struct{}
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestBuildFunc7() struct{} {
	o, err := c.UnscopedSafeGetTestBuildFunc7()
	if err != nil {
		panic(err)
	}
	return o
}

// TestBuildFunc7 retrieves the "test_build_func_7" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_7"
//	type: struct{}
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestBuildFunc7 method.
// If the container can not be retrieved, it panics.
func TestBuildFunc7(i interface{}) struct{} {
	return C(i).GetTestBuildFunc7()
}

// SafeGetTestBuildFunc8 retrieves the "test_build_func_8" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_8"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(int)
//		- "1": Value(*models.BuildFuncTestC)
//		- "2": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestBuildFunc8() (*models.BuildFuncTestA, error) {
	i, err := c.ctn.SafeGet("test_build_func_8")
	if err != nil {
		var eo *models.BuildFuncTestA
		return eo, err
	}
	o, ok := i.(*models.BuildFuncTestA)
	if !ok {
		return o, errors.New("could get 'test_build_func_8' because the object could not be cast to *models.BuildFuncTestA")
	}
	return o, nil
}

// GetTestBuildFunc8 retrieves the "test_build_func_8" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_8"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(int)
//		- "1": Value(*models.BuildFuncTestC)
//		- "2": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestBuildFunc8() *models.BuildFuncTestA {
	o, err := c.SafeGetTestBuildFunc8()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestBuildFunc8 retrieves the "test_build_func_8" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_8"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(int)
//		- "1": Value(*models.BuildFuncTestC)
//		- "2": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestBuildFunc8() (*models.BuildFuncTestA, error) {
	i, err := c.ctn.UnscopedSafeGet("test_build_func_8")
	if err != nil {
		var eo *models.BuildFuncTestA
		return eo, err
	}
	o, ok := i.(*models.BuildFuncTestA)
	if !ok {
		return o, errors.New("could get 'test_build_func_8' because the object could not be cast to *models.BuildFuncTestA")
	}
	return o, nil
}

// UnscopedGetTestBuildFunc8 retrieves the "test_build_func_8" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_8"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(int)
//		- "1": Value(*models.BuildFuncTestC)
//		- "2": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestBuildFunc8() *models.BuildFuncTestA {
	o, err := c.UnscopedSafeGetTestBuildFunc8()
	if err != nil {
		panic(err)
	}
	return o
}

// TestBuildFunc8 retrieves the "test_build_func_8" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_func_8"
//	type: *models.BuildFuncTestA
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(int)
//		- "1": Value(*models.BuildFuncTestC)
//		- "2": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestBuildFunc8 method.
// If the container can not be retrieved, it panics.
func TestBuildFunc8(i interface{}) *models.BuildFuncTestA {
	return C(i).GetTestBuildFunc8()
}

// SafeGetTestBuildStruct1 retrieves the "test_build_struct_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_1"
//	type: *models.BuildStructTestA
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestB) ["test_build_struct_2"]
//		- "P3": Service(*models.BuildStructTestC) ["test_build_struct_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestBuildStruct1() (*models.BuildStructTestA, error) {
	i, err := c.ctn.SafeGet("test_build_struct_1")
	if err != nil {
		var eo *models.BuildStructTestA
		return eo, err
	}
	o, ok := i.(*models.BuildStructTestA)
	if !ok {
		return o, errors.New("could get 'test_build_struct_1' because the object could not be cast to *models.BuildStructTestA")
	}
	return o, nil
}

// GetTestBuildStruct1 retrieves the "test_build_struct_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_1"
//	type: *models.BuildStructTestA
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestB) ["test_build_struct_2"]
//		- "P3": Service(*models.BuildStructTestC) ["test_build_struct_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestBuildStruct1() *models.BuildStructTestA {
	o, err := c.SafeGetTestBuildStruct1()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestBuildStruct1 retrieves the "test_build_struct_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_1"
//	type: *models.BuildStructTestA
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestB) ["test_build_struct_2"]
//		- "P3": Service(*models.BuildStructTestC) ["test_build_struct_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestBuildStruct1() (*models.BuildStructTestA, error) {
	i, err := c.ctn.UnscopedSafeGet("test_build_struct_1")
	if err != nil {
		var eo *models.BuildStructTestA
		return eo, err
	}
	o, ok := i.(*models.BuildStructTestA)
	if !ok {
		return o, errors.New("could get 'test_build_struct_1' because the object could not be cast to *models.BuildStructTestA")
	}
	return o, nil
}

// UnscopedGetTestBuildStruct1 retrieves the "test_build_struct_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_1"
//	type: *models.BuildStructTestA
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestB) ["test_build_struct_2"]
//		- "P3": Service(*models.BuildStructTestC) ["test_build_struct_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestBuildStruct1() *models.BuildStructTestA {
	o, err := c.UnscopedSafeGetTestBuildStruct1()
	if err != nil {
		panic(err)
	}
	return o
}

// TestBuildStruct1 retrieves the "test_build_struct_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_1"
//	type: *models.BuildStructTestA
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestB) ["test_build_struct_2"]
//		- "P3": Service(*models.BuildStructTestC) ["test_build_struct_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestBuildStruct1 method.
// If the container can not be retrieved, it panics.
func TestBuildStruct1(i interface{}) *models.BuildStructTestA {
	return C(i).GetTestBuildStruct1()
}

// SafeGetTestBuildStruct2 retrieves the "test_build_struct_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_2"
//	type: *models.BuildStructTestB
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestC) ["test_build_struct_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestBuildStruct2() (*models.BuildStructTestB, error) {
	i, err := c.ctn.SafeGet("test_build_struct_2")
	if err != nil {
		var eo *models.BuildStructTestB
		return eo, err
	}
	o, ok := i.(*models.BuildStructTestB)
	if !ok {
		return o, errors.New("could get 'test_build_struct_2' because the object could not be cast to *models.BuildStructTestB")
	}
	return o, nil
}

// GetTestBuildStruct2 retrieves the "test_build_struct_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_2"
//	type: *models.BuildStructTestB
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestC) ["test_build_struct_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestBuildStruct2() *models.BuildStructTestB {
	o, err := c.SafeGetTestBuildStruct2()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestBuildStruct2 retrieves the "test_build_struct_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_2"
//	type: *models.BuildStructTestB
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestC) ["test_build_struct_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestBuildStruct2() (*models.BuildStructTestB, error) {
	i, err := c.ctn.UnscopedSafeGet("test_build_struct_2")
	if err != nil {
		var eo *models.BuildStructTestB
		return eo, err
	}
	o, ok := i.(*models.BuildStructTestB)
	if !ok {
		return o, errors.New("could get 'test_build_struct_2' because the object could not be cast to *models.BuildStructTestB")
	}
	return o, nil
}

// UnscopedGetTestBuildStruct2 retrieves the "test_build_struct_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_2"
//	type: *models.BuildStructTestB
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestC) ["test_build_struct_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestBuildStruct2() *models.BuildStructTestB {
	o, err := c.UnscopedSafeGetTestBuildStruct2()
	if err != nil {
		panic(err)
	}
	return o
}

// TestBuildStruct2 retrieves the "test_build_struct_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_2"
//	type: *models.BuildStructTestB
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestC) ["test_build_struct_3"]
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestBuildStruct2 method.
// If the container can not be retrieved, it panics.
func TestBuildStruct2(i interface{}) *models.BuildStructTestB {
	return C(i).GetTestBuildStruct2()
}

// SafeGetTestBuildStruct3 retrieves the "test_build_struct_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_3"
//	type: *models.BuildStructTestC
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestBuildStruct3() (*models.BuildStructTestC, error) {
	i, err := c.ctn.SafeGet("test_build_struct_3")
	if err != nil {
		var eo *models.BuildStructTestC
		return eo, err
	}
	o, ok := i.(*models.BuildStructTestC)
	if !ok {
		return o, errors.New("could get 'test_build_struct_3' because the object could not be cast to *models.BuildStructTestC")
	}
	return o, nil
}

// GetTestBuildStruct3 retrieves the "test_build_struct_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_3"
//	type: *models.BuildStructTestC
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestBuildStruct3() *models.BuildStructTestC {
	o, err := c.SafeGetTestBuildStruct3()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestBuildStruct3 retrieves the "test_build_struct_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_3"
//	type: *models.BuildStructTestC
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestBuildStruct3() (*models.BuildStructTestC, error) {
	i, err := c.ctn.UnscopedSafeGet("test_build_struct_3")
	if err != nil {
		var eo *models.BuildStructTestC
		return eo, err
	}
	o, ok := i.(*models.BuildStructTestC)
	if !ok {
		return o, errors.New("could get 'test_build_struct_3' because the object could not be cast to *models.BuildStructTestC")
	}
	return o, nil
}

// UnscopedGetTestBuildStruct3 retrieves the "test_build_struct_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_3"
//	type: *models.BuildStructTestC
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestBuildStruct3() *models.BuildStructTestC {
	o, err := c.UnscopedSafeGetTestBuildStruct3()
	if err != nil {
		panic(err)
	}
	return o
}

// TestBuildStruct3 retrieves the "test_build_struct_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_3"
//	type: *models.BuildStructTestC
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestBuildStruct3 method.
// If the container can not be retrieved, it panics.
func TestBuildStruct3(i interface{}) *models.BuildStructTestC {
	return C(i).GetTestBuildStruct3()
}

// SafeGetTestBuildStruct4 retrieves the "test_build_struct_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_4"
//	type: *models.BuildStructTestA
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestB) ["test_build_struct_2"]
//		- "P3": Value(*models.BuildStructTestC)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestBuildStruct4() (*models.BuildStructTestA, error) {
	i, err := c.ctn.SafeGet("test_build_struct_4")
	if err != nil {
		var eo *models.BuildStructTestA
		return eo, err
	}
	o, ok := i.(*models.BuildStructTestA)
	if !ok {
		return o, errors.New("could get 'test_build_struct_4' because the object could not be cast to *models.BuildStructTestA")
	}
	return o, nil
}

// GetTestBuildStruct4 retrieves the "test_build_struct_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_4"
//	type: *models.BuildStructTestA
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestB) ["test_build_struct_2"]
//		- "P3": Value(*models.BuildStructTestC)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestBuildStruct4() *models.BuildStructTestA {
	o, err := c.SafeGetTestBuildStruct4()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestBuildStruct4 retrieves the "test_build_struct_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_4"
//	type: *models.BuildStructTestA
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestB) ["test_build_struct_2"]
//		- "P3": Value(*models.BuildStructTestC)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestBuildStruct4() (*models.BuildStructTestA, error) {
	i, err := c.ctn.UnscopedSafeGet("test_build_struct_4")
	if err != nil {
		var eo *models.BuildStructTestA
		return eo, err
	}
	o, ok := i.(*models.BuildStructTestA)
	if !ok {
		return o, errors.New("could get 'test_build_struct_4' because the object could not be cast to *models.BuildStructTestA")
	}
	return o, nil
}

// UnscopedGetTestBuildStruct4 retrieves the "test_build_struct_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_4"
//	type: *models.BuildStructTestA
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestB) ["test_build_struct_2"]
//		- "P3": Value(*models.BuildStructTestC)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestBuildStruct4() *models.BuildStructTestA {
	o, err := c.UnscopedSafeGetTestBuildStruct4()
	if err != nil {
		panic(err)
	}
	return o
}

// TestBuildStruct4 retrieves the "test_build_struct_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_build_struct_4"
//	type: *models.BuildStructTestA
//	scope: "main"
//	build: struct
//	params:
//		- "P1": Value(string)
//		- "P2": Service(*models.BuildStructTestB) ["test_build_struct_2"]
//		- "P3": Value(*models.BuildStructTestC)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestBuildStruct4 method.
// If the container can not be retrieved, it panics.
func TestBuildStruct4(i interface{}) *models.BuildStructTestA {
	return C(i).GetTestBuildStruct4()
}

// SafeGetTestClose1 retrieves the "test_close_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_close_1"
//	type: *models.CloseTest
//	scope: "main"
//	build: struct
//	params:
//		- "Closed": Value(bool)
//	unshared: false
//	close: true
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestClose1() (*models.CloseTest, error) {
	i, err := c.ctn.SafeGet("test_close_1")
	if err != nil {
		var eo *models.CloseTest
		return eo, err
	}
	o, ok := i.(*models.CloseTest)
	if !ok {
		return o, errors.New("could get 'test_close_1' because the object could not be cast to *models.CloseTest")
	}
	return o, nil
}

// GetTestClose1 retrieves the "test_close_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_close_1"
//	type: *models.CloseTest
//	scope: "main"
//	build: struct
//	params:
//		- "Closed": Value(bool)
//	unshared: false
//	close: true
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestClose1() *models.CloseTest {
	o, err := c.SafeGetTestClose1()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestClose1 retrieves the "test_close_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_close_1"
//	type: *models.CloseTest
//	scope: "main"
//	build: struct
//	params:
//		- "Closed": Value(bool)
//	unshared: false
//	close: true
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestClose1() (*models.CloseTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_close_1")
	if err != nil {
		var eo *models.CloseTest
		return eo, err
	}
	o, ok := i.(*models.CloseTest)
	if !ok {
		return o, errors.New("could get 'test_close_1' because the object could not be cast to *models.CloseTest")
	}
	return o, nil
}

// UnscopedGetTestClose1 retrieves the "test_close_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_close_1"
//	type: *models.CloseTest
//	scope: "main"
//	build: struct
//	params:
//		- "Closed": Value(bool)
//	unshared: false
//	close: true
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestClose1() *models.CloseTest {
	o, err := c.UnscopedSafeGetTestClose1()
	if err != nil {
		panic(err)
	}
	return o
}

// TestClose1 retrieves the "test_close_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_close_1"
//	type: *models.CloseTest
//	scope: "main"
//	build: struct
//	params:
//		- "Closed": Value(bool)
//	unshared: false
//	close: true
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestClose1 method.
// If the container can not be retrieved, it panics.
func TestClose1(i interface{}) *models.CloseTest {
	return C(i).GetTestClose1()
}

// SafeGetTestDeclType0 retrieves the "test_decl_type_0" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_0"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDeclType0() (*models.DeclTypeTest, error) {
	i, err := c.ctn.SafeGet("test_decl_type_0")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_0' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// GetTestDeclType0 retrieves the "test_decl_type_0" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_0"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDeclType0() *models.DeclTypeTest {
	o, err := c.SafeGetTestDeclType0()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDeclType0 retrieves the "test_decl_type_0" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_0"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDeclType0() (*models.DeclTypeTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_decl_type_0")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_0' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// UnscopedGetTestDeclType0 retrieves the "test_decl_type_0" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_0"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDeclType0() *models.DeclTypeTest {
	o, err := c.UnscopedSafeGetTestDeclType0()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDeclType0 retrieves the "test_decl_type_0" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_0"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDeclType0 method.
// If the container can not be retrieved, it panics.
func TestDeclType0(i interface{}) *models.DeclTypeTest {
	return C(i).GetTestDeclType0()
}

// SafeGetTestDeclType1 retrieves the "test_decl_type_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_1"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDeclType1() (*models.DeclTypeTest, error) {
	i, err := c.ctn.SafeGet("test_decl_type_1")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_1' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// GetTestDeclType1 retrieves the "test_decl_type_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_1"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDeclType1() *models.DeclTypeTest {
	o, err := c.SafeGetTestDeclType1()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDeclType1 retrieves the "test_decl_type_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_1"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDeclType1() (*models.DeclTypeTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_decl_type_1")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_1' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// UnscopedGetTestDeclType1 retrieves the "test_decl_type_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_1"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDeclType1() *models.DeclTypeTest {
	o, err := c.UnscopedSafeGetTestDeclType1()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDeclType1 retrieves the "test_decl_type_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_1"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDeclType1 method.
// If the container can not be retrieved, it panics.
func TestDeclType1(i interface{}) *models.DeclTypeTest {
	return C(i).GetTestDeclType1()
}

// SafeGetTestDeclType10 retrieves the "test_decl_type_10" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_10"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDeclType10() (*models.DeclTypeTest, error) {
	i, err := c.ctn.SafeGet("test_decl_type_10")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_10' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// GetTestDeclType10 retrieves the "test_decl_type_10" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_10"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDeclType10() *models.DeclTypeTest {
	o, err := c.SafeGetTestDeclType10()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDeclType10 retrieves the "test_decl_type_10" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_10"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDeclType10() (*models.DeclTypeTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_decl_type_10")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_10' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// UnscopedGetTestDeclType10 retrieves the "test_decl_type_10" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_10"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDeclType10() *models.DeclTypeTest {
	o, err := c.UnscopedSafeGetTestDeclType10()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDeclType10 retrieves the "test_decl_type_10" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_10"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDeclType10 method.
// If the container can not be retrieved, it panics.
func TestDeclType10(i interface{}) *models.DeclTypeTest {
	return C(i).GetTestDeclType10()
}

// SafeGetTestDeclType11 retrieves the "test_decl_type_11" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_11"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDeclType11() (*models.DeclTypeTest, error) {
	i, err := c.ctn.SafeGet("test_decl_type_11")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_11' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// GetTestDeclType11 retrieves the "test_decl_type_11" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_11"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDeclType11() *models.DeclTypeTest {
	o, err := c.SafeGetTestDeclType11()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDeclType11 retrieves the "test_decl_type_11" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_11"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDeclType11() (*models.DeclTypeTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_decl_type_11")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_11' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// UnscopedGetTestDeclType11 retrieves the "test_decl_type_11" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_11"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDeclType11() *models.DeclTypeTest {
	o, err := c.UnscopedSafeGetTestDeclType11()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDeclType11 retrieves the "test_decl_type_11" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_11"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDeclType11 method.
// If the container can not be retrieved, it panics.
func TestDeclType11(i interface{}) *models.DeclTypeTest {
	return C(i).GetTestDeclType11()
}

// SafeGetTestDeclType2 retrieves the "test_decl_type_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_2"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDeclType2() (*models.DeclTypeTest, error) {
	i, err := c.ctn.SafeGet("test_decl_type_2")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_2' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// GetTestDeclType2 retrieves the "test_decl_type_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_2"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDeclType2() *models.DeclTypeTest {
	o, err := c.SafeGetTestDeclType2()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDeclType2 retrieves the "test_decl_type_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_2"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDeclType2() (*models.DeclTypeTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_decl_type_2")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_2' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// UnscopedGetTestDeclType2 retrieves the "test_decl_type_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_2"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDeclType2() *models.DeclTypeTest {
	o, err := c.UnscopedSafeGetTestDeclType2()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDeclType2 retrieves the "test_decl_type_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_2"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDeclType2 method.
// If the container can not be retrieved, it panics.
func TestDeclType2(i interface{}) *models.DeclTypeTest {
	return C(i).GetTestDeclType2()
}

// SafeGetTestDeclType3 retrieves the "test_decl_type_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_3"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDeclType3() (*models.DeclTypeTest, error) {
	i, err := c.ctn.SafeGet("test_decl_type_3")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_3' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// GetTestDeclType3 retrieves the "test_decl_type_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_3"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDeclType3() *models.DeclTypeTest {
	o, err := c.SafeGetTestDeclType3()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDeclType3 retrieves the "test_decl_type_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_3"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDeclType3() (*models.DeclTypeTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_decl_type_3")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_3' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// UnscopedGetTestDeclType3 retrieves the "test_decl_type_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_3"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDeclType3() *models.DeclTypeTest {
	o, err := c.UnscopedSafeGetTestDeclType3()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDeclType3 retrieves the "test_decl_type_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_3"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDeclType3 method.
// If the container can not be retrieved, it panics.
func TestDeclType3(i interface{}) *models.DeclTypeTest {
	return C(i).GetTestDeclType3()
}

// SafeGetTestDeclType4 retrieves the "test_decl_type_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_4"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDeclType4() (*models.DeclTypeTest, error) {
	i, err := c.ctn.SafeGet("test_decl_type_4")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_4' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// GetTestDeclType4 retrieves the "test_decl_type_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_4"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDeclType4() *models.DeclTypeTest {
	o, err := c.SafeGetTestDeclType4()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDeclType4 retrieves the "test_decl_type_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_4"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDeclType4() (*models.DeclTypeTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_decl_type_4")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_4' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// UnscopedGetTestDeclType4 retrieves the "test_decl_type_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_4"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDeclType4() *models.DeclTypeTest {
	o, err := c.UnscopedSafeGetTestDeclType4()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDeclType4 retrieves the "test_decl_type_4" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_4"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDeclType4 method.
// If the container can not be retrieved, it panics.
func TestDeclType4(i interface{}) *models.DeclTypeTest {
	return C(i).GetTestDeclType4()
}

// SafeGetTestDeclType5 retrieves the "test_decl_type_5" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_5"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDeclType5() (*models.DeclTypeTest, error) {
	i, err := c.ctn.SafeGet("test_decl_type_5")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_5' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// GetTestDeclType5 retrieves the "test_decl_type_5" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_5"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDeclType5() *models.DeclTypeTest {
	o, err := c.SafeGetTestDeclType5()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDeclType5 retrieves the "test_decl_type_5" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_5"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDeclType5() (*models.DeclTypeTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_decl_type_5")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_5' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// UnscopedGetTestDeclType5 retrieves the "test_decl_type_5" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_5"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDeclType5() *models.DeclTypeTest {
	o, err := c.UnscopedSafeGetTestDeclType5()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDeclType5 retrieves the "test_decl_type_5" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_5"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDeclType5 method.
// If the container can not be retrieved, it panics.
func TestDeclType5(i interface{}) *models.DeclTypeTest {
	return C(i).GetTestDeclType5()
}

// SafeGetTestDeclType6 retrieves the "test_decl_type_6" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_6"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDeclType6() (*models.DeclTypeTest, error) {
	i, err := c.ctn.SafeGet("test_decl_type_6")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_6' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// GetTestDeclType6 retrieves the "test_decl_type_6" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_6"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDeclType6() *models.DeclTypeTest {
	o, err := c.SafeGetTestDeclType6()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDeclType6 retrieves the "test_decl_type_6" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_6"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDeclType6() (*models.DeclTypeTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_decl_type_6")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_6' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// UnscopedGetTestDeclType6 retrieves the "test_decl_type_6" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_6"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDeclType6() *models.DeclTypeTest {
	o, err := c.UnscopedSafeGetTestDeclType6()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDeclType6 retrieves the "test_decl_type_6" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_6"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDeclType6 method.
// If the container can not be retrieved, it panics.
func TestDeclType6(i interface{}) *models.DeclTypeTest {
	return C(i).GetTestDeclType6()
}

// SafeGetTestDeclType7 retrieves the "test_decl_type_7" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_7"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDeclType7() (*models.DeclTypeTest, error) {
	i, err := c.ctn.SafeGet("test_decl_type_7")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_7' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// GetTestDeclType7 retrieves the "test_decl_type_7" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_7"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDeclType7() *models.DeclTypeTest {
	o, err := c.SafeGetTestDeclType7()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDeclType7 retrieves the "test_decl_type_7" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_7"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDeclType7() (*models.DeclTypeTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_decl_type_7")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_7' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// UnscopedGetTestDeclType7 retrieves the "test_decl_type_7" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_7"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDeclType7() *models.DeclTypeTest {
	o, err := c.UnscopedSafeGetTestDeclType7()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDeclType7 retrieves the "test_decl_type_7" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_7"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDeclType7 method.
// If the container can not be retrieved, it panics.
func TestDeclType7(i interface{}) *models.DeclTypeTest {
	return C(i).GetTestDeclType7()
}

// SafeGetTestDeclType8 retrieves the "test_decl_type_8" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_8"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDeclType8() (*models.DeclTypeTest, error) {
	i, err := c.ctn.SafeGet("test_decl_type_8")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_8' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// GetTestDeclType8 retrieves the "test_decl_type_8" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_8"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDeclType8() *models.DeclTypeTest {
	o, err := c.SafeGetTestDeclType8()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDeclType8 retrieves the "test_decl_type_8" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_8"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDeclType8() (*models.DeclTypeTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_decl_type_8")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_8' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// UnscopedGetTestDeclType8 retrieves the "test_decl_type_8" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_8"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDeclType8() *models.DeclTypeTest {
	o, err := c.UnscopedSafeGetTestDeclType8()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDeclType8 retrieves the "test_decl_type_8" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_8"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDeclType8 method.
// If the container can not be retrieved, it panics.
func TestDeclType8(i interface{}) *models.DeclTypeTest {
	return C(i).GetTestDeclType8()
}

// SafeGetTestDeclType9 retrieves the "test_decl_type_9" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_9"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDeclType9() (*models.DeclTypeTest, error) {
	i, err := c.ctn.SafeGet("test_decl_type_9")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_9' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// GetTestDeclType9 retrieves the "test_decl_type_9" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_9"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDeclType9() *models.DeclTypeTest {
	o, err := c.SafeGetTestDeclType9()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDeclType9 retrieves the "test_decl_type_9" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_9"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDeclType9() (*models.DeclTypeTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_decl_type_9")
	if err != nil {
		var eo *models.DeclTypeTest
		return eo, err
	}
	o, ok := i.(*models.DeclTypeTest)
	if !ok {
		return o, errors.New("could get 'test_decl_type_9' because the object could not be cast to *models.DeclTypeTest")
	}
	return o, nil
}

// UnscopedGetTestDeclType9 retrieves the "test_decl_type_9" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_9"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDeclType9() *models.DeclTypeTest {
	o, err := c.UnscopedSafeGetTestDeclType9()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDeclType9 retrieves the "test_decl_type_9" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_decl_type_9"
//	type: *models.DeclTypeTest
//	scope: "main"
//	build: struct
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDeclType9 method.
// If the container can not be retrieved, it panics.
func TestDeclType9(i interface{}) *models.DeclTypeTest {
	return C(i).GetTestDeclType9()
}

// SafeGetTestDi1 retrieves the "test_di_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_1"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDi1() (models.DiTest, error) {
	i, err := c.ctn.SafeGet("test_di_1")
	if err != nil {
		var eo models.DiTest
		return eo, err
	}
	o, ok := i.(models.DiTest)
	if !ok {
		return o, errors.New("could get 'test_di_1' because the object could not be cast to models.DiTest")
	}
	return o, nil
}

// GetTestDi1 retrieves the "test_di_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_1"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDi1() models.DiTest {
	o, err := c.SafeGetTestDi1()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDi1 retrieves the "test_di_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_1"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDi1() (models.DiTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_di_1")
	if err != nil {
		var eo models.DiTest
		return eo, err
	}
	o, ok := i.(models.DiTest)
	if !ok {
		return o, errors.New("could get 'test_di_1' because the object could not be cast to models.DiTest")
	}
	return o, nil
}

// UnscopedGetTestDi1 retrieves the "test_di_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_1"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDi1() models.DiTest {
	o, err := c.UnscopedSafeGetTestDi1()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDi1 retrieves the "test_di_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_1"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDi1 method.
// If the container can not be retrieved, it panics.
func TestDi1(i interface{}) models.DiTest {
	return C(i).GetTestDi1()
}

// SafeGetTestDi2 retrieves the "test_di_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_2"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDi2() (models.DiTest, error) {
	i, err := c.ctn.SafeGet("test_di_2")
	if err != nil {
		var eo models.DiTest
		return eo, err
	}
	o, ok := i.(models.DiTest)
	if !ok {
		return o, errors.New("could get 'test_di_2' because the object could not be cast to models.DiTest")
	}
	return o, nil
}

// GetTestDi2 retrieves the "test_di_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_2"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDi2() models.DiTest {
	o, err := c.SafeGetTestDi2()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDi2 retrieves the "test_di_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_2"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDi2() (models.DiTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_di_2")
	if err != nil {
		var eo models.DiTest
		return eo, err
	}
	o, ok := i.(models.DiTest)
	if !ok {
		return o, errors.New("could get 'test_di_2' because the object could not be cast to models.DiTest")
	}
	return o, nil
}

// UnscopedGetTestDi2 retrieves the "test_di_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_2"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDi2() models.DiTest {
	o, err := c.UnscopedSafeGetTestDi2()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDi2 retrieves the "test_di_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_2"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDi2 method.
// If the container can not be retrieved, it panics.
func TestDi2(i interface{}) models.DiTest {
	return C(i).GetTestDi2()
}

// SafeGetTestDi3 retrieves the "test_di_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_3"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestDi3() (models.DiTest, error) {
	i, err := c.ctn.SafeGet("test_di_3")
	if err != nil {
		var eo models.DiTest
		return eo, err
	}
	o, ok := i.(models.DiTest)
	if !ok {
		return o, errors.New("could get 'test_di_3' because the object could not be cast to models.DiTest")
	}
	return o, nil
}

// GetTestDi3 retrieves the "test_di_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_3"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestDi3() models.DiTest {
	o, err := c.SafeGetTestDi3()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestDi3 retrieves the "test_di_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_3"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestDi3() (models.DiTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_di_3")
	if err != nil {
		var eo models.DiTest
		return eo, err
	}
	o, ok := i.(models.DiTest)
	if !ok {
		return o, errors.New("could get 'test_di_3' because the object could not be cast to models.DiTest")
	}
	return o, nil
}

// UnscopedGetTestDi3 retrieves the "test_di_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_3"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestDi3() models.DiTest {
	o, err := c.UnscopedSafeGetTestDi3()
	if err != nil {
		panic(err)
	}
	return o
}

// TestDi3 retrieves the "test_di_3" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_di_3"
//	type: models.DiTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestDi3 method.
// If the container can not be retrieved, it panics.
func TestDi3(i interface{}) models.DiTest {
	return C(i).GetTestDi3()
}

// SafeGetTestInterfaces1 retrieves the "test_interfaces_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_interfaces_1"
//	type: *models.InterfacesTestB
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(testinterfaces.InterfacesTestInterface)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestInterfaces1() (*models.InterfacesTestB, error) {
	i, err := c.ctn.SafeGet("test_interfaces_1")
	if err != nil {
		var eo *models.InterfacesTestB
		return eo, err
	}
	o, ok := i.(*models.InterfacesTestB)
	if !ok {
		return o, errors.New("could get 'test_interfaces_1' because the object could not be cast to *models.InterfacesTestB")
	}
	return o, nil
}

// GetTestInterfaces1 retrieves the "test_interfaces_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_interfaces_1"
//	type: *models.InterfacesTestB
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(testinterfaces.InterfacesTestInterface)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestInterfaces1() *models.InterfacesTestB {
	o, err := c.SafeGetTestInterfaces1()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestInterfaces1 retrieves the "test_interfaces_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_interfaces_1"
//	type: *models.InterfacesTestB
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(testinterfaces.InterfacesTestInterface)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestInterfaces1() (*models.InterfacesTestB, error) {
	i, err := c.ctn.UnscopedSafeGet("test_interfaces_1")
	if err != nil {
		var eo *models.InterfacesTestB
		return eo, err
	}
	o, ok := i.(*models.InterfacesTestB)
	if !ok {
		return o, errors.New("could get 'test_interfaces_1' because the object could not be cast to *models.InterfacesTestB")
	}
	return o, nil
}

// UnscopedGetTestInterfaces1 retrieves the "test_interfaces_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_interfaces_1"
//	type: *models.InterfacesTestB
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(testinterfaces.InterfacesTestInterface)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestInterfaces1() *models.InterfacesTestB {
	o, err := c.UnscopedSafeGetTestInterfaces1()
	if err != nil {
		panic(err)
	}
	return o
}

// TestInterfaces1 retrieves the "test_interfaces_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_interfaces_1"
//	type: *models.InterfacesTestB
//	scope: "main"
//	build: func
//	params:
//		- "0": Value(testinterfaces.InterfacesTestInterface)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestInterfaces1 method.
// If the container can not be retrieved, it panics.
func TestInterfaces1(i interface{}) *models.InterfacesTestB {
	return C(i).GetTestInterfaces1()
}

// SafeGetTestInterfaces2 retrieves the "test_interfaces_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_interfaces_2"
//	type: *models.InterfacesTestB
//	scope: "main"
//	build: struct
//	params:
//		- "InterfacesTestInterface": Value(testinterfaces.InterfacesTestInterface)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestInterfaces2() (*models.InterfacesTestB, error) {
	i, err := c.ctn.SafeGet("test_interfaces_2")
	if err != nil {
		var eo *models.InterfacesTestB
		return eo, err
	}
	o, ok := i.(*models.InterfacesTestB)
	if !ok {
		return o, errors.New("could get 'test_interfaces_2' because the object could not be cast to *models.InterfacesTestB")
	}
	return o, nil
}

// GetTestInterfaces2 retrieves the "test_interfaces_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_interfaces_2"
//	type: *models.InterfacesTestB
//	scope: "main"
//	build: struct
//	params:
//		- "InterfacesTestInterface": Value(testinterfaces.InterfacesTestInterface)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestInterfaces2() *models.InterfacesTestB {
	o, err := c.SafeGetTestInterfaces2()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestInterfaces2 retrieves the "test_interfaces_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_interfaces_2"
//	type: *models.InterfacesTestB
//	scope: "main"
//	build: struct
//	params:
//		- "InterfacesTestInterface": Value(testinterfaces.InterfacesTestInterface)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestInterfaces2() (*models.InterfacesTestB, error) {
	i, err := c.ctn.UnscopedSafeGet("test_interfaces_2")
	if err != nil {
		var eo *models.InterfacesTestB
		return eo, err
	}
	o, ok := i.(*models.InterfacesTestB)
	if !ok {
		return o, errors.New("could get 'test_interfaces_2' because the object could not be cast to *models.InterfacesTestB")
	}
	return o, nil
}

// UnscopedGetTestInterfaces2 retrieves the "test_interfaces_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_interfaces_2"
//	type: *models.InterfacesTestB
//	scope: "main"
//	build: struct
//	params:
//		- "InterfacesTestInterface": Value(testinterfaces.InterfacesTestInterface)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestInterfaces2() *models.InterfacesTestB {
	o, err := c.UnscopedSafeGetTestInterfaces2()
	if err != nil {
		panic(err)
	}
	return o
}

// TestInterfaces2 retrieves the "test_interfaces_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_interfaces_2"
//	type: *models.InterfacesTestB
//	scope: "main"
//	build: struct
//	params:
//		- "InterfacesTestInterface": Value(testinterfaces.InterfacesTestInterface)
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestInterfaces2 method.
// If the container can not be retrieved, it panics.
func TestInterfaces2(i interface{}) *models.InterfacesTestB {
	return C(i).GetTestInterfaces2()
}

// SafeGetTestRetrieval1 retrieves the "test_retrieval_1" object from the app scope.
//
// ---------------------------------------------
//
//	name: "test_retrieval_1"
//	type: *models.RetrievalTest
//	scope: "app"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestRetrieval1() (*models.RetrievalTest, error) {
	i, err := c.ctn.SafeGet("test_retrieval_1")
	if err != nil {
		var eo *models.RetrievalTest
		return eo, err
	}
	o, ok := i.(*models.RetrievalTest)
	if !ok {
		return o, errors.New("could get 'test_retrieval_1' because the object could not be cast to *models.RetrievalTest")
	}
	return o, nil
}

// GetTestRetrieval1 retrieves the "test_retrieval_1" object from the app scope.
//
// ---------------------------------------------
//
//	name: "test_retrieval_1"
//	type: *models.RetrievalTest
//	scope: "app"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestRetrieval1() *models.RetrievalTest {
	o, err := c.SafeGetTestRetrieval1()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestRetrieval1 retrieves the "test_retrieval_1" object from the app scope.
//
// ---------------------------------------------
//
//	name: "test_retrieval_1"
//	type: *models.RetrievalTest
//	scope: "app"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if app is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestRetrieval1() (*models.RetrievalTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_retrieval_1")
	if err != nil {
		var eo *models.RetrievalTest
		return eo, err
	}
	o, ok := i.(*models.RetrievalTest)
	if !ok {
		return o, errors.New("could get 'test_retrieval_1' because the object could not be cast to *models.RetrievalTest")
	}
	return o, nil
}

// UnscopedGetTestRetrieval1 retrieves the "test_retrieval_1" object from the app scope.
//
// ---------------------------------------------
//
//	name: "test_retrieval_1"
//	type: *models.RetrievalTest
//	scope: "app"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if app is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestRetrieval1() *models.RetrievalTest {
	o, err := c.UnscopedSafeGetTestRetrieval1()
	if err != nil {
		panic(err)
	}
	return o
}

// TestRetrieval1 retrieves the "test_retrieval_1" object from the app scope.
//
// ---------------------------------------------
//
//	name: "test_retrieval_1"
//	type: *models.RetrievalTest
//	scope: "app"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestRetrieval1 method.
// If the container can not be retrieved, it panics.
func TestRetrieval1(i interface{}) *models.RetrievalTest {
	return C(i).GetTestRetrieval1()
}

// SafeGetTestRetrieval2 retrieves the "test_retrieval_2" object from the request scope.
//
// ---------------------------------------------
//
//	name: "test_retrieval_2"
//	type: *models.RetrievalTest
//	scope: "request"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestRetrieval2() (*models.RetrievalTest, error) {
	i, err := c.ctn.SafeGet("test_retrieval_2")
	if err != nil {
		var eo *models.RetrievalTest
		return eo, err
	}
	o, ok := i.(*models.RetrievalTest)
	if !ok {
		return o, errors.New("could get 'test_retrieval_2' because the object could not be cast to *models.RetrievalTest")
	}
	return o, nil
}

// GetTestRetrieval2 retrieves the "test_retrieval_2" object from the request scope.
//
// ---------------------------------------------
//
//	name: "test_retrieval_2"
//	type: *models.RetrievalTest
//	scope: "request"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestRetrieval2() *models.RetrievalTest {
	o, err := c.SafeGetTestRetrieval2()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestRetrieval2 retrieves the "test_retrieval_2" object from the request scope.
//
// ---------------------------------------------
//
//	name: "test_retrieval_2"
//	type: *models.RetrievalTest
//	scope: "request"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if request is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestRetrieval2() (*models.RetrievalTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_retrieval_2")
	if err != nil {
		var eo *models.RetrievalTest
		return eo, err
	}
	o, ok := i.(*models.RetrievalTest)
	if !ok {
		return o, errors.New("could get 'test_retrieval_2' because the object could not be cast to *models.RetrievalTest")
	}
	return o, nil
}

// UnscopedGetTestRetrieval2 retrieves the "test_retrieval_2" object from the request scope.
//
// ---------------------------------------------
//
//	name: "test_retrieval_2"
//	type: *models.RetrievalTest
//	scope: "request"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if request is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestRetrieval2() *models.RetrievalTest {
	o, err := c.UnscopedSafeGetTestRetrieval2()
	if err != nil {
		panic(err)
	}
	return o
}

// TestRetrieval2 retrieves the "test_retrieval_2" object from the request scope.
//
// ---------------------------------------------
//
//	name: "test_retrieval_2"
//	type: *models.RetrievalTest
//	scope: "request"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestRetrieval2 method.
// If the container can not be retrieved, it panics.
func TestRetrieval2(i interface{}) *models.RetrievalTest {
	return C(i).GetTestRetrieval2()
}

// SafeGetTestScope1 retrieves the "test_scope_1" object from the app scope.
//
// ---------------------------------------------
//
//	name: "test_scope_1"
//	type: *models.ScopeTest
//	scope: "app"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestScope1() (*models.ScopeTest, error) {
	i, err := c.ctn.SafeGet("test_scope_1")
	if err != nil {
		var eo *models.ScopeTest
		return eo, err
	}
	o, ok := i.(*models.ScopeTest)
	if !ok {
		return o, errors.New("could get 'test_scope_1' because the object could not be cast to *models.ScopeTest")
	}
	return o, nil
}

// GetTestScope1 retrieves the "test_scope_1" object from the app scope.
//
// ---------------------------------------------
//
//	name: "test_scope_1"
//	type: *models.ScopeTest
//	scope: "app"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestScope1() *models.ScopeTest {
	o, err := c.SafeGetTestScope1()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestScope1 retrieves the "test_scope_1" object from the app scope.
//
// ---------------------------------------------
//
//	name: "test_scope_1"
//	type: *models.ScopeTest
//	scope: "app"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if app is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestScope1() (*models.ScopeTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_scope_1")
	if err != nil {
		var eo *models.ScopeTest
		return eo, err
	}
	o, ok := i.(*models.ScopeTest)
	if !ok {
		return o, errors.New("could get 'test_scope_1' because the object could not be cast to *models.ScopeTest")
	}
	return o, nil
}

// UnscopedGetTestScope1 retrieves the "test_scope_1" object from the app scope.
//
// ---------------------------------------------
//
//	name: "test_scope_1"
//	type: *models.ScopeTest
//	scope: "app"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if app is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestScope1() *models.ScopeTest {
	o, err := c.UnscopedSafeGetTestScope1()
	if err != nil {
		panic(err)
	}
	return o
}

// TestScope1 retrieves the "test_scope_1" object from the app scope.
//
// ---------------------------------------------
//
//	name: "test_scope_1"
//	type: *models.ScopeTest
//	scope: "app"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestScope1 method.
// If the container can not be retrieved, it panics.
func TestScope1(i interface{}) *models.ScopeTest {
	return C(i).GetTestScope1()
}

// SafeGetTestScope2 retrieves the "test_scope_2" object from the request scope.
//
// ---------------------------------------------
//
//	name: "test_scope_2"
//	type: *models.ScopeTest
//	scope: "request"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestScope2() (*models.ScopeTest, error) {
	i, err := c.ctn.SafeGet("test_scope_2")
	if err != nil {
		var eo *models.ScopeTest
		return eo, err
	}
	o, ok := i.(*models.ScopeTest)
	if !ok {
		return o, errors.New("could get 'test_scope_2' because the object could not be cast to *models.ScopeTest")
	}
	return o, nil
}

// GetTestScope2 retrieves the "test_scope_2" object from the request scope.
//
// ---------------------------------------------
//
//	name: "test_scope_2"
//	type: *models.ScopeTest
//	scope: "request"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestScope2() *models.ScopeTest {
	o, err := c.SafeGetTestScope2()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestScope2 retrieves the "test_scope_2" object from the request scope.
//
// ---------------------------------------------
//
//	name: "test_scope_2"
//	type: *models.ScopeTest
//	scope: "request"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if request is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestScope2() (*models.ScopeTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_scope_2")
	if err != nil {
		var eo *models.ScopeTest
		return eo, err
	}
	o, ok := i.(*models.ScopeTest)
	if !ok {
		return o, errors.New("could get 'test_scope_2' because the object could not be cast to *models.ScopeTest")
	}
	return o, nil
}

// UnscopedGetTestScope2 retrieves the "test_scope_2" object from the request scope.
//
// ---------------------------------------------
//
//	name: "test_scope_2"
//	type: *models.ScopeTest
//	scope: "request"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if request is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestScope2() *models.ScopeTest {
	o, err := c.UnscopedSafeGetTestScope2()
	if err != nil {
		panic(err)
	}
	return o
}

// TestScope2 retrieves the "test_scope_2" object from the request scope.
//
// ---------------------------------------------
//
//	name: "test_scope_2"
//	type: *models.ScopeTest
//	scope: "request"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestScope2 method.
// If the container can not be retrieved, it panics.
func TestScope2(i interface{}) *models.ScopeTest {
	return C(i).GetTestScope2()
}

// SafeGetTestUnshared1 retrieves the "test_unshared_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_unshared_1"
//	type: *models.UnsharedTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: true
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestUnshared1() (*models.UnsharedTest, error) {
	i, err := c.ctn.SafeGet("test_unshared_1")
	if err != nil {
		var eo *models.UnsharedTest
		return eo, err
	}
	o, ok := i.(*models.UnsharedTest)
	if !ok {
		return o, errors.New("could get 'test_unshared_1' because the object could not be cast to *models.UnsharedTest")
	}
	return o, nil
}

// GetTestUnshared1 retrieves the "test_unshared_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_unshared_1"
//	type: *models.UnsharedTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: true
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestUnshared1() *models.UnsharedTest {
	o, err := c.SafeGetTestUnshared1()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestUnshared1 retrieves the "test_unshared_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_unshared_1"
//	type: *models.UnsharedTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: true
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestUnshared1() (*models.UnsharedTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_unshared_1")
	if err != nil {
		var eo *models.UnsharedTest
		return eo, err
	}
	o, ok := i.(*models.UnsharedTest)
	if !ok {
		return o, errors.New("could get 'test_unshared_1' because the object could not be cast to *models.UnsharedTest")
	}
	return o, nil
}

// UnscopedGetTestUnshared1 retrieves the "test_unshared_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_unshared_1"
//	type: *models.UnsharedTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: true
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestUnshared1() *models.UnsharedTest {
	o, err := c.UnscopedSafeGetTestUnshared1()
	if err != nil {
		panic(err)
	}
	return o
}

// TestUnshared1 retrieves the "test_unshared_1" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_unshared_1"
//	type: *models.UnsharedTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: true
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestUnshared1 method.
// If the container can not be retrieved, it panics.
func TestUnshared1(i interface{}) *models.UnsharedTest {
	return C(i).GetTestUnshared1()
}

// SafeGetTestUnshared2 retrieves the "test_unshared_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_unshared_2"
//	type: *models.UnsharedTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetTestUnshared2() (*models.UnsharedTest, error) {
	i, err := c.ctn.SafeGet("test_unshared_2")
	if err != nil {
		var eo *models.UnsharedTest
		return eo, err
	}
	o, ok := i.(*models.UnsharedTest)
	if !ok {
		return o, errors.New("could get 'test_unshared_2' because the object could not be cast to *models.UnsharedTest")
	}
	return o, nil
}

// GetTestUnshared2 retrieves the "test_unshared_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_unshared_2"
//	type: *models.UnsharedTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetTestUnshared2() *models.UnsharedTest {
	o, err := c.SafeGetTestUnshared2()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetTestUnshared2 retrieves the "test_unshared_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_unshared_2"
//	type: *models.UnsharedTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetTestUnshared2() (*models.UnsharedTest, error) {
	i, err := c.ctn.UnscopedSafeGet("test_unshared_2")
	if err != nil {
		var eo *models.UnsharedTest
		return eo, err
	}
	o, ok := i.(*models.UnsharedTest)
	if !ok {
		return o, errors.New("could get 'test_unshared_2' because the object could not be cast to *models.UnsharedTest")
	}
	return o, nil
}

// UnscopedGetTestUnshared2 retrieves the "test_unshared_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_unshared_2"
//	type: *models.UnsharedTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// This method can be called even if main is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetTestUnshared2() *models.UnsharedTest {
	o, err := c.UnscopedSafeGetTestUnshared2()
	if err != nil {
		panic(err)
	}
	return o
}

// TestUnshared2 retrieves the "test_unshared_2" object from the main scope.
//
// ---------------------------------------------
//
//	name: "test_unshared_2"
//	type: *models.UnsharedTest
//	scope: "main"
//	build: func
//	params: nil
//	unshared: false
//	close: false
//
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetTestUnshared2 method.
// If the container can not be retrieved, it panics.
func TestUnshared2(i interface{}) *models.UnsharedTest {
	return C(i).GetTestUnshared2()
}
