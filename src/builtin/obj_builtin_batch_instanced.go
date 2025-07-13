package builtin

import (
	"glox/src/core"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// for drawing batches of textured cubes

// ---------------------------------------------------------------------------------------------
// shader singleton
var shaderInstanced *rl.Shader

// ---------------------------------------------------------------------------------------------
type BatchInstancedObject struct {
	core.BuiltInObject
	Methods map[int]*core.BuiltInObject
	model   *Model
	batch   *Batch
}

// ---------------------------------------------------------------------------------------------
// Instance represents a single drawn instance of a mesh with its transformation matrices
type Instance struct {
	translation rl.Matrix
	rotation    rl.Matrix
}

// ---------------------------------------------------------------------------------------------
// Model represents a 3D model with its mesh and material
type Model struct {
	mesh     rl.Mesh
	material rl.Material
}

// ---------------------------------------------------------------------------------------------
// Instances holds a list of Instance objects
type Instances struct {
	list []*Instance
}

// ---------------------------------------------------------------------------------------------
// Batch holds a model, a collection of draw instances and their transforms for rendering
type Batch struct {
	instances  *Instances
	transforms []rl.Matrix // Transform matrices for instancing

}

func BatchInstancedBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {

	if argCount != 3 {
		vm.RunTimeError("BatchInstancedBuiltIn: expected 3 arguments, got %d", argCount)
		return core.NIL_VALUE
	}
	textureVal := vm.Stack(arg_stackptr)
	if !textureVal.IsObj() {
		vm.RunTimeError("BatchInstancedBuiltIn: expected texture2D, got %s", textureVal.String())
		return core.NIL_VALUE
	}
	to, ok := textureVal.Obj.(*TextureObject)
	if !ok {
		vm.RunTimeError("BatchInstancedBuiltIn: expected texture2D, got %s", textureVal.String())
		return core.NIL_VALUE
	}
	cubeSizeVal := vm.Stack(arg_stackptr + 1)
	if !cubeSizeVal.IsFloat() {
		vm.RunTimeError("BatchInstancedBuiltIn: expected float for cubeSize, got %s", cubeSizeVal.String())
		return core.NIL_VALUE
	}
	cubeSize := cubeSizeVal.Float

	maxInstancesVal := vm.Stack(arg_stackptr + 2)
	if !maxInstancesVal.IsInt() {
		vm.RunTimeError("BatchInstancedBuiltIn: expected int for maxInstances, got %s", maxInstancesVal.String())
		return core.NIL_VALUE
	}
	maxInstances := maxInstancesVal.Int

	batchObj := MakeInstancedBatchInstancedObject(to.Data.Texture, float32(cubeSize), maxInstances)
	RegisterAllBatchInstancedMethods(batchObj)
	return core.MakeObjectValue(batchObj, true)
}

func MakeInstancedBatchInstancedObject(texture rl.Texture2D, cubeSize float32, maxInstances int) *BatchInstancedObject {
	if shaderInstanced == nil {
		shaderInstanced = InitShader()
	}
	fs := float32(cubeSize)
	rv := &BatchInstancedObject{
		BuiltInObject: core.BuiltInObject{},
		Methods:       make(map[int]*core.BuiltInObject),
		model:         MakeModel(rl.GenMeshCube(fs, fs, fs), texture),
		batch:         MakeBatch(maxInstances),
	}
	return rv
}

// Standard interface implementations
func (o *BatchInstancedObject) String() string {

	return "BatchInstancedObject"
}

func (o *BatchInstancedObject) GetType() core.ObjectType {
	return core.OBJECT_NATIVE
}

func (o *BatchInstancedObject) GetNativeType() core.NativeType {
	return core.NATIVE_BATCH_INSTANCED
}

func (o *BatchInstancedObject) GetMethod(stringId int) *core.BuiltInObject {
	return o.Methods[stringId]
}

func (o *BatchInstancedObject) RegisterMethod(name string, method *core.BuiltInObject) {
	if o.Methods == nil {
		o.Methods = make(map[int]*core.BuiltInObject)
	}
	o.Methods[core.InternName(name)] = method
}

func (o *BatchInstancedObject) IsBuiltIn() bool {
	return true
}

// Utility functions
func IsBatchInstancedObject(v core.Value) bool {
	_, ok := v.Obj.(*BatchInstancedObject)
	return ok
}

func AsBatchInstanced(v core.Value) *BatchInstancedObject {
	return v.Obj.(*BatchInstancedObject)
}

func InitShader() *rl.Shader {

	vspath := "src/shaders/instanced/base_lighting_instanced.vs"
	if _, err := os.Stat(vspath); os.IsNotExist(err) {
		core.LogFmtLn(core.ERROR, "Shader vertex file not found: %s", vspath)
		return nil
	}
	fspath := "src/shaders/instanced/lighting.fs"
	if _, err := os.Stat(fspath); os.IsNotExist(err) {
		core.LogFmtLn(core.ERROR, "Shader fragment file not found: %s", fspath)
		return nil
	}
	rv := rl.LoadShader(vspath, fspath)
	if rv.ID == 0 {
		panic("Failed to load instanced shader")
	}

	core.LogFmtLn(core.INFO, "Shader loaded successfully with ID: %d", rv.ID)
	var mvp = rl.GetShaderLocation(rv, "mvp")
	var viewPos = rl.GetShaderLocation(rv, "viewPos")
	var transform = rl.GetShaderLocationAttrib(rv, "instanceTransform")
	rv.UpdateLocation(rl.ShaderLocMatrixMvp, mvp)
	rv.UpdateLocation(rl.ShaderLocVectorView, viewPos)
	rv.UpdateLocation(rl.ShaderLocMatrixModel, transform)

	// ambient light level
	ambientLoc := rl.GetShaderLocation(rv, "ambient")
	rl.SetShaderValue(rv, ambientLoc, []float32{10.0, 10.0, 10.0, 10.0}, rl.ShaderUniformVec4)
	return &rv
}

func MakeModel(mesh rl.Mesh, texture rl.Texture2D) *Model {
	rv := &Model{
		mesh:     mesh,
		material: rl.LoadMaterialDefault(),
	}
	rv.material.Shader = *shaderInstanced
	rv.material.Maps.Texture = texture
	rv.material.Maps.Color = rl.White
	rv.material.Maps.Value = 1.0
	mmap := rv.material.GetMap(rl.MapDiffuse)
	mmap.Color = rl.White
	return rv
}

//---------------------------------------------------------------------------------------------

func MakeInstance(x, y, z, axisX, axisY, axisZ, angle float64) Instance {

	x32 := float32(x)
	y32 := float32(y)
	z32 := float32(z)
	axisX32 := float32(axisX)
	axisY32 := float32(axisY)
	axisZ32 := float32(axisZ)
	angle32 := float32(angle)
	translation := rl.MatrixTranslate(x32, y32, z32)
	axis := rl.Vector3Normalize(rl.NewVector3(axisX32, axisY32, axisZ32))
	rotation := rl.MatrixRotate(axis, angle32*rl.Deg2rad)
	return Instance{translation: translation, rotation: rotation}
}

//---------------------------------------------------------------------------------------------

func MakeInstances(max int) *Instances {
	return &Instances{list: make([]*Instance, 0, max)}
}

func (i *Instances) AddInstance(x, y, z, axisX, axisY, axisZ, angle float64) {
	instance := MakeInstance(x, y, z, axisX, axisY, axisZ, angle)
	i.list = append(i.list, &instance)
}

func (i *Instances) GetInstance(index int) *Instance {
	if index < 0 || index >= len(i.list) {
		return nil // Handle out of bounds
	}
	return i.list[index]
}

func MakeBatch(maxInstances int) *Batch {
	return &Batch{
		instances:  MakeInstances(maxInstances),
		transforms: make([]rl.Matrix, maxInstances), // Initialize transform matrices
	}
}

func (b *Batch) AddInstance(x, y, z, axisX, axisY, axisZ, angle float64) bool {
	if len(b.instances.list) >= cap(b.instances.list) {
		return false
	}
	b.instances.AddInstance(x, y, z, axisX, axisY, axisZ, angle)
	return true
}

func (b *Batch) MakeTransforms() {
	for i := 0; i < len(b.instances.list); i++ {
		instance := b.instances.GetInstance(i)
		if instance != nil {
			b.transforms[i] = rl.MatrixMultiply(instance.rotation, instance.translation)
		} else {
			b.transforms[i] = rl.MatrixIdentity() // Default identity matrix if instance is nil
		}
	}
}

func (b *BatchInstancedObject) Draw(camera *CameraObject) {
	if len(b.batch.instances.list) == 0 {
		core.LogFmtLn(core.WARN, "BatchInstancedObject.Draw: No instances to draw")
		return // No instances to draw
	}
	count := int32(len(b.batch.instances.list))

	rl.DrawMeshInstanced(b.model.mesh, b.model.material, b.batch.transforms[:count], count)
}
