/*
 * Copyright (c) 2024 The XGo Authors (xgo.dev). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"github.com/goplus/lib/c"
	"github.com/llarhub/cjson"
	"github.com/llarhub/py"
	"github.com/llarhub/py/inspect"
)

func main() {
	if c.Argc < 2 {
		c.Fprintf(c.Stderr, c.Str("Usage: pydump <pythonLibPath>\n"))
		return
	}
	pyLib := c.Index(c.Argv, 1)

	py.Initialize()

	root := cjson.CreateObject()
	root.AddItemToObject(c.Str("name"), cjson.CreateString(pyLib))

	items := cjson.CreateArray()
	mod := py.ImportModule(pyLib)
	keys := mod.Dict().Keys()
	for i, n := py.SsizeT(0), keys.Size(); i < n; i++ {
		key := keys.Item(i)
		val := mod.Attr(key)
		doc := val.AttrString(c.Str("__doc__"))
		sym := cjson.CreateObject()
		sym.AddItemToObject(c.Str("type"), cjson.CreateString(val.Type().Name().CStr()))
		sym.AddItemToObject(c.Str("name"), cjson.CreateString(key.CStr()))
		if doc != nil {
			sym.AddItemToObject(c.Str("doc"), cjson.CreateString(doc.CStr()))
		}
		if val.Callable() != 0 {
			sig := inspect.Signature(val)
			sym.AddItemToObject(c.Str("sig"), cjson.CreateString(sig.Str().CStr()))
		}
		cjson.AddItemToArray(items, sym)
	}
	root.AddItemToObject(c.Str("items"), items)

	c.Printf(c.Str("%s\n"), root.Print())
}
