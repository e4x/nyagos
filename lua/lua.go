package lua

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

var luaDLL = syscall.NewLazyDLL("lua53")

type Integer int64

const LUAINT_PER_UINTPTR = unsafe.Sizeof(Integer(0)) / unsafe.Sizeof(uintptr(0))

func (value Integer) Expand(list []uintptr) []uintptr {
	for i := uintptr(0); i < LUAINT_PER_UINTPTR; i++ {
		list = append(list, uintptr(value))
		value >>= (8 * unsafe.Sizeof(uintptr(1)))
	}
	return list
}

func CGoBytes(p, length uintptr) []byte {
	if length <= 0 || p == 0 {
		return []byte{}
	}
	buffer := make([]byte, length)
	for i := uintptr(0); i < length; i++ {
		buffer[i] = *(*byte)(unsafe.Pointer(p))
		p++
	}
	return buffer
}

func CGoStringN(p, length uintptr) string {
	if length <= 0 || p == 0 {
		return ""
	}
	return string(CGoBytes(p, length))
}

type Lua struct {
	lua uintptr
}

var luaL_newstate = luaDLL.NewProc("luaL_newstate")

func New() *Lua {
	lua, _, _ := luaL_newstate.Call()
	return &Lua{lua}
}

func (this *Lua) State() uintptr {
	return this.lua
}

var luaL_openlibs = luaDLL.NewProc("luaL_openlibs")

func (this *Lua) OpenLibs() {
	luaL_openlibs.Call(this.State())
}

var lua_close = luaDLL.NewProc("lua_close")

func (this *Lua) Close() {
	lua_close.Call(this.State())
}

var lua_isstring = luaDLL.NewProc("lua_isstring")

func (this *Lua) IsString(index int) bool {
	rc, _, _ := lua_isstring.Call(this.State(), uintptr(index))
	return rc != 0
}

func (this *Lua) IsFunction(index int) bool {
	return this.GetType(index) == LUA_TFUNCTION
}

var lua_type = luaDLL.NewProc("lua_type")

func (this *Lua) GetType(index int) int {
	rv, _, _ := lua_type.Call(this.State(), uintptr(index))
	return int(rv)
}

var lua_pushvalue = luaDLL.NewProc("lua_pushvalue")

func (this *Lua) PushValue(index int) {
	lua_pushvalue.Call(this.State(), uintptr(index))
}

func (this *Lua) Source(fname string) error {
	if err := this.Load(fname); err != nil {
		return err
	}
	return this.Call(0, 0)
}

var lua_pushlstring = luaDLL.NewProc("lua_pushlstring")

func (this *Lua) PushAnsiString(data []byte) {
	if data != nil && len(data) > 0 {
		lua_pushlstring.Call(this.State(),
			uintptr(unsafe.Pointer(&data[0])),
			uintptr(len(data)))
	} else {
		this.PushString("")
	}
}

var lua_pushinteger = luaDLL.NewProc("lua_pushinteger")

func (this *Lua) PushInteger(value Integer) {
	params := make([]uintptr, 0, 4)
	params = append(params, this.State())
	params = value.Expand(params)
	lua_pushinteger.Call(params...)
}

var lua_tointegerx = luaDLL.NewProc("lua_tointegerx")

func (this *Lua) ToInteger(index int) (int, error) {
	var issucceeded uintptr
	value, _, _ := lua_tointegerx.Call(this.State(), uintptr(index),
		uintptr(unsafe.Pointer(&issucceeded)))
	if issucceeded != 0 {
		return int(value), nil
	} else {
		return 0, errors.New("ToInteger: the value in not integer on the stack")
	}
}

var lua_settable = luaDLL.NewProc("lua_settable")

func (this *Lua) SetTable(index int) {
	lua_settable.Call(this.State(), uintptr(index))
}

var lua_gettable = luaDLL.NewProc("lua_gettable")

func (this *Lua) GetTable(index int) {
	lua_gettable.Call(this.State(), uintptr(index))
}

var lua_setmetatable = luaDLL.NewProc("lua_setmetatable")

func (this *Lua) SetMetaTable(index int) {
	lua_setmetatable.Call(this.State(), uintptr(index))
}

var lua_gettop = luaDLL.NewProc("lua_gettop")

func (this *Lua) GetTop() int {
	rv, _, _ := lua_gettop.Call(this.State())
	return int(rv)
}

var lua_settop = luaDLL.NewProc("lua_settop")

func (this *Lua) SetTop(index int) {
	lua_settop.Call(this.State(), uintptr(index))
}

func (this *Lua) Pop(n int) {
	this.SetTop(-n - 1)
}

var lua_pushlightuserdata = luaDLL.NewProc("lua_pushlightuserdata")

func (this *Lua) PushLightUserData(p unsafe.Pointer) {
	lua_pushlightuserdata.Call(this.State(), uintptr(p))
}

var lua_newuserdata = luaDLL.NewProc("lua_newuserdata")

func (this *Lua) NewUserData(size uintptr) unsafe.Pointer {
	area, _, _ := lua_newuserdata.Call(this.State(), size)
	return unsafe.Pointer(area)
}

var lua_touserdata = luaDLL.NewProc("lua_touserdata")

func (this *Lua) ToUserData(index int) unsafe.Pointer {
	rv, _, _ := lua_touserdata.Call(this.State(), uintptr(index))
	return unsafe.Pointer(rv)
}

var lua_pushnil = luaDLL.NewProc("lua_pushnil")

func (this *Lua) PushNil() {
	lua_pushnil.Call(this.State())
}

var lua_pushboolean = luaDLL.NewProc("lua_pushboolean")

func (this *Lua) PushBool(value bool) {
	if value {
		lua_pushboolean.Call(this.State(), 1)
	} else {
		lua_pushboolean.Call(this.State(), 0)
	}
}

var lua_toboolean = luaDLL.NewProc("lua_toboolean")

func (this *Lua) ToBool(index int) bool {
	rv, _, _ := lua_toboolean.Call(this.State(), uintptr(index))
	if rv != 0 {
		return true
	} else {
		return false
	}
}

var lua_rawseti = luaDLL.NewProc("lua_rawseti")

func (this *Lua) RawSetI(index int, value Integer) {
	params := make([]uintptr, 0, 4)
	params = append(params, this.State(), uintptr(index))
	params = value.Expand(params)
	lua_rawseti.Call(params...)
}

// 5.2
// var lua_remove = luaDLL.NewProc("lua_remove")
// 5.3
var lua_rotate = luaDLL.NewProc("lua_rotate")

func lua_remove_Call(state uintptr, index int) {
	lua_rotate.Call(state, uintptr(index), ^uintptr(0))
	lua_settop.Call(state, ^uintptr(1)) // ^1 == -2
}

func (this *Lua) Remove(index int) {
	// 5.2
	// lua_remove.Call(this.State(), uintptr(index))
	// 5.3
	lua_remove_Call(this.State(), index)
}

var lua_replace = luaDLL.NewProc("lua_replace")

func (this *Lua) Replace(index int) {
	lua_replace.Call(this.State(), uintptr(index))
}

var lua_setglobal = luaDLL.NewProc("lua_setglobal")

func (this *Lua) SetGlobal(str string) {
	cstr, err := syscall.BytePtrFromString(str)
	if err != nil {
		panic(err.Error())
	}
	lua_setglobal.Call(this.State(), uintptr(unsafe.Pointer(cstr)))
}

var lua_setfield = luaDLL.NewProc("lua_setfield")

func (this *Lua) SetField(index int, str string) {
	cstr, err := syscall.BytePtrFromString(str)
	if err != nil {
		panic(err.Error())
	}
	lua_setfield.Call(this.State(), uintptr(index), uintptr(unsafe.Pointer(cstr)))
}

var lua_getfield = luaDLL.NewProc("lua_getfield")

func (this *Lua) GetField(index int, str string) {
	cstr, err := syscall.BytePtrFromString(str)
	if err != nil {
		panic(err.Error())
	}
	lua_getfield.Call(this.State(), uintptr(index), uintptr(unsafe.Pointer(cstr)))
}

var lua_getglobal = luaDLL.NewProc("lua_getglobal")

func (this *Lua) GetGlobal(str string) {
	cstr, err := syscall.BytePtrFromString(str)
	if err != nil {
		panic(err.Error())
	}
	lua_getglobal.Call(this.State(), uintptr(unsafe.Pointer(cstr)))
}

var lua_createtable = luaDLL.NewProc("lua_createtable")

func (this *Lua) NewTable() {
	lua_createtable.Call(this.State(), 0, 0)
}

var lua_pushstring = luaDLL.NewProc("lua_pushstring")

func (this *Lua) PushString(str string) {
	cstr, err := syscall.BytePtrFromString(str)
	if err != nil {
		panic(err.Error())
	}
	lua_pushstring.Call(this.State(), uintptr(unsafe.Pointer(cstr)))
}

var lua_tolstring = luaDLL.NewProc("lua_tolstring")

func (this *Lua) ToAnsiString(index int) []byte {
	var length uintptr
	p, _, _ := lua_tolstring.Call(this.State(),
		uintptr(index),
		uintptr(unsafe.Pointer(&length)))
	if length <= 0 {
		return []byte{}
	} else {
		return CGoBytes(p, length)
	}
}

func (this *Lua) ToString(index int) (string, error) {
	if !this.IsString(index) {
		return "", fmt.Errorf("Lua.ToString(%d): Not String", index)
	}
	var length uintptr
	p, _, _ := lua_tolstring.Call(this.State(),
		uintptr(index),
		uintptr(unsafe.Pointer(&length)))
	return CGoStringN(p, length), nil
}

var luaL_loadfilex = luaDLL.NewProc("luaL_loadfilex")

func (this *Lua) Load(fname string) error {
	cfname, err := syscall.BytePtrFromString(fname)
	if err != nil {
		return err
	}
	rc, _, _ := luaL_loadfilex.Call(this.State(),
		uintptr(unsafe.Pointer(cfname)),
		uintptr(0))
	if rc != 0 {
		msg, err := this.ToString(-1)
		if err == nil {
			return fmt.Errorf("%s: %s..", fname, msg)
		} else {
			return err
		}
	}
	return nil
}

var lua_pcallk = luaDLL.NewProc("lua_pcallk")

func (this *Lua) Call(nargs, nresult int) error {
	rc, _, _ := lua_pcallk.Call(this.State(),
		uintptr(nargs),
		uintptr(nresult), 0, 0, 0)
	if rc != 0 {
		if this.IsString(-1) {
			msg, err := this.ToString(-1)
			if err == nil {
				return errors.New(msg)
			} else {
				return err
			}
		} else {
			return errors.New("<Lua Error>")
		}
	}
	return nil
}

func luaToGoBridge(lua uintptr) int {
	f, _, _ := lua_touserdata.Call(lua, 1)
	f_ := *(*goFunctionT)(unsafe.Pointer(f))
	lua_remove_Call(lua, 1)
	L := Lua{lua}
	return int(f_.function(&L))
}

type goFunctionT struct {
	function func(*Lua) int
}

var lua_pushcclosure = luaDLL.NewProc("lua_pushcclosure")

func (this *Lua) PushGoFunction(f func(L *Lua) int) {
	f_ := goFunctionT{f}
	voidptr := this.NewUserData(unsafe.Sizeof(f_))
	*(*goFunctionT)(voidptr) = f_
	this.NewTable()
	lua_pushcclosure.Call(this.State(),
		syscall.NewCallbackCDecl(luaToGoBridge),
		0)
	this.SetField(-2, "__call")
	this.SetMetaTable(-2)
}
