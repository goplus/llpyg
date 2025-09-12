package pysig

import "strings"

type Arg struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	DefVal   string `json:"defVal"`
	Optional bool   `json:"optional"`
}

func Parse(sig string) (args []*Arg) {
	// get signature between ()
	end := findMatchingBracket(sig, '(', ')')
	if end == -1 {
		return
	}
	sig = sig[:end+1]
	sig = strings.TrimPrefix(sig, "(")

	for {
		sig = strings.TrimSpace(sig)
		if len(sig) == 0 {
			return
		}
		if sig[0] == '[' {		// optional args
			var optArgs []*Arg
			optArgs, sig = parseOptArgs(sig)
			for _, arg := range optArgs {
				arg.Optional = true
			}
			args = append(args, optArgs...)
			continue
		}
		if sig[0] == '(' {     // (a1, a2, ...)
			pos := findMatchingBracket(sig, '(', ')')
			if pos > 0 {
				name := strings.TrimSpace(sig[:pos+1])
				arg := &Arg{Name: name}
				args = append(args, arg)
				sig = sig[pos+1:]
			}
			continue
		}
		// (a) (, a, ) (a:int) (a=1) (a[, b])
		pos := strings.IndexAny(sig, ",:=[)")
		if pos < 0 || (pos == 0 && sig[0] == ')') {
			return
		}
		name := strings.TrimSpace(sig[:pos])
		if name == "" {
			sig = sig[1:]
			continue
		}
		if name == "..." {
			length := len(args)
			if length > 0 && args[length-1].DefVal != "" {
				name = "**kwargs"
			}else {
				name = "**args"
			}
		}
		arg := &Arg{Name: name}
		args = append(args, arg)
		split := sig[pos]    // , : = ) [
		switch split {
		case ',':
			sig = sig[pos+1:]
			continue
		case ':':
			arg.Type, sig = parseType(sig[pos+1:])
			if sig[0] == '=' {
				arg.DefVal, sig = parseDefVal(sig[1:])
			}
		case '=':
			arg.DefVal, sig = parseDefVal(sig[pos+1:])
		case '[':
			var optArgs []*Arg
			optArgs, sig = parseOptArgs(sig[pos:])
			for _, arg := range optArgs {
				arg.Optional = true
			}
			args = append(args, optArgs...)
			continue
		case ')':
			return
		}
		sig = strings.TrimPrefix(sig, ",")
	}
}


func parseOptArgs(sig string) (optArgs []*Arg, newSig string) {
	end := findMatchingBracket(sig, '[', ']')
	if end == -1 {
		return
	}
	optArgs = Parse("(" + sig[1:end] + ")")
	newSig = sig[end+1:]
	return
}




// default value pairs
var pairs = map[byte]byte {
	'(': ')',
	'[': ']',
	'{': '}',
}


func parseDefVal(sig string) (defVal string, newSig string) {
	sig = strings.TrimSpace(sig)
	// list, tuple, dict
	if close, exists := pairs[sig[0]]; exists {
		idx := findMatchingBracket(sig, sig[0], close)
		if idx > 0 {
			defVal = strings.TrimSpace(sig[0:idx+1])
			newSig = sig[idx+1:]
			return
		}
	}
	pos := strings.IndexAny(sig, "[,)")
	if pos > 0 {
		defVal = strings.TrimSpace(sig[:pos])
		newSig = sig[pos:]
		return
	}
	return
}


func parseType(sig string) (typeStr string, newSig string) {
	right := strings.IndexAny(sig, "=,)[")
	if right < 0 {
		return "", sig
	}
	if sig[right] == '[' { 		// 'Union[int, float]'
		tmp := strings.TrimSpace(sig[right+1:])
		if len(tmp) > 0 && tmp[0] != ',' {
			length := findMatchingBracket(sig[right:], '[', ']')
			right = right + length
			idx := strings.IndexAny(sig[right:], "=,)")
			right = right + idx
		}
	}
	typeStr = strings.TrimSpace(sig[:right])
	newSig = sig[right:]
	return
}


// start=0, return end index, else return -1
func findMatchingBracket(sig string, open, close byte) int {
	if len(sig) == 0 || sig[0] != open {
		return -1
	}

	count := 1
	for i := 1; i < len(sig); i++ {
		switch sig[i] {
		case open:
			count++
		case close:
			count--
			if count == 0 {
				return i
			}
		}
	}
	return -1
}
