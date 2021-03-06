package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"../conio"
	"../lua"
)

type KeyLuaFuncT struct {
	L            *lua.Lua
	registoryKey string
}

func getBufferForCallBack(L *lua.Lua) (*conio.Buffer, int) {
	if L.GetType(1) != lua.LUA_TTABLE {
		L.PushNil()
		L.PushString("bindKeyExec: call with : not .")
		return nil, 2
	}
	L.GetField(1, "buffer")
	if L.GetType(-1) != lua.LUA_TLIGHTUSERDATA {
		L.PushNil()
		L.PushString("bindKey.Call: invalid object")
		return nil, 2
	}
	buffer := (*conio.Buffer)(L.ToUserData(-1))
	if buffer == nil {
		L.PushNil()
		L.PushString("bindKey.Call: invalid member")
		return nil, 2
	}
	return buffer, 0
}

func callInsert(L *lua.Lua) int {
	buffer, stackRc := getBufferForCallBack(L)
	if buffer == nil {
		return stackRc
	}
	text, textErr := L.ToString(2)
	if textErr != nil {
		L.PushNil()
		L.PushString(textErr.Error())
		return 2
	}
	buffer.InsertAndRepaint(text)
	L.PushBool(true)
	return 1
}

func callKeyFunc(L *lua.Lua) int {
	buffer, stackRc := getBufferForCallBack(L)
	if buffer == nil {
		return stackRc
	}
	key, keyErr := L.ToString(2)
	if keyErr != nil {
		L.PushNil()
		L.PushString(keyErr.Error())
		return 2
	}
	function, funcErr := conio.GetFunc(key)
	if funcErr != nil {
		L.PushNil()
		L.PushString(funcErr.Error())
		return 2
	}
	rc := function.Call(buffer)
	L.PushBool(true)
	switch rc {
	case conio.ENTER:
		L.PushBool(true)
		return 2
	case conio.ABORT:
		L.PushBool(false)
		return 2
	}
	return 1
}

func (this *KeyLuaFuncT) Call(buffer *conio.Buffer) conio.Result {
	this.L.GetField(lua.LUA_REGISTRYINDEX, this.registoryKey)
	this.L.NewTable()
	pos := -1
	var text bytes.Buffer
	for i, c := range buffer.Buffer {
		if i >= buffer.Length {
			break
		}
		if i == buffer.Cursor {
			pos = text.Len() + 1
		}
		text.WriteRune(c)
	}
	if pos < 0 {
		pos = text.Len() + 1
	}
	this.L.PushInteger(lua.Integer(pos))
	this.L.SetField(-2, "pos")
	this.L.PushString(text.String())
	this.L.SetField(-2, "text")
	this.L.PushLightUserData(unsafe.Pointer(buffer))
	this.L.SetField(-2, "buffer")
	this.L.PushGoFunction(callKeyFunc)
	this.L.SetField(-2, "call")
	this.L.PushGoFunction(callInsert)
	this.L.SetField(-2, "insert")
	if err := this.L.Call(1, 1); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	switch this.L.GetType(-1) {
	case lua.LUA_TSTRING:
		str, strErr := this.L.ToString(-1)
		if strErr == nil {
			buffer.InsertAndRepaint(str)
		}
	case lua.LUA_TBOOLEAN:
		if !this.L.ToBool(-1) {
			buffer.Buffer = []rune{}
			buffer.Length = 0
		}
		return conio.ENTER
	}
	return conio.CONTINUE
}

func cmdBindKey(L *lua.Lua) int {
	key, keyErr := L.ToString(-2)
	if keyErr != nil {
		L.PushString(keyErr.Error())
		return 1
	}
	key = strings.Replace(strings.ToUpper(key), "-", "_", -1)
	switch L.GetType(-1) {
	case lua.LUA_TFUNCTION:
		regkey := "nyagos.bind." + key
		L.SetField(lua.LUA_REGISTRYINDEX, regkey)
		if err := conio.BindKeyFunc(key, &KeyLuaFuncT{L, regkey}); err != nil {
			L.PushNil()
			L.PushString(err.Error())
			return 2
		} else {
			L.PushBool(true)
			return 1
		}
	default:
		val, valErr := L.ToString(-1)
		if valErr != nil {
			L.PushNil()
			L.PushString(valErr.Error())
			return 2
		}
		err := conio.BindKeySymbol(key, val)
		if err != nil {
			L.PushNil()
			L.PushString(err.Error())
			return 2
		} else {
			L.PushBool(true)
			return 1
		}
	}
}
