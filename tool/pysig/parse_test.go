package pysig

import (
	"testing"
)

func TestParse(t *testing.T) {
	type testCase struct {
		sig  string
		args []*Arg
	}
	cases := []testCase{
		{"()", nil},
		{"() -> int", nil},
		{"(a) -> int", []*Arg{
			{Name: "a"},
		}},
		{"(a: int)", []*Arg{
			{Name: "a", Type: "int"},
		}},
		{"(a: int = 1, b: float)", []*Arg{
			{Name: "a", Type: "int", DefVal: "1"},
			{Name: "b", Type: "float"},
		}},
		{"(a = <1>, b = 2.0)", []*Arg{
			{Name: "a", DefVal: "<1>"},
			{Name: "b", DefVal: "2.0"},
		}},
		{"(a: 'Suffixes' = ('_x', '_y'))", []*Arg{
			{Name: "a", Type: "'Suffixes'", DefVal: "('_x', '_y')"},
		}},
		{"(start=None, *, unit: 'str | None' = None) -> 'TimedeltaIndex'", []*Arg{
			{Name: "start", DefVal: "None"},
			{Name: "*"},
			{Name: "unit", Type: "'str | None'", DefVal: "None"},
		}},
		{"([start,] stop[, step,], dtype=None, *, device=None, like=None)", []*Arg{
			{Name: "start", Optional: true},
			{Name: "stop"},
			{Name: "step", Optional: true},
			{Name: "dtype", DefVal: "None"},
			{Name: "*"},
			{Name: "device", DefVal: "None"},
			{Name: "like", DefVal: "None"},
		}},
		{"([start, ]stop, [step, ]dtype=None, *, device=None, like=None)", []*Arg{
			{Name: "start", Optional: true},
			{Name: "stop"},
			{Name: "step", Optional: true},
			{Name: "dtype", DefVal: "None"},
			{Name: "*"},
			{Name: "device", DefVal: "None"},
			{Name: "like", DefVal: "None"},
		}},
		{"( (a1, a2, ...), axis=0, out=None, dtype=None, casting=\"same_kind\" )", []*Arg{
			{Name: "(a1, a2, ...)"},
			{Name: "axis", DefVal: "0"},
			{Name: "out", DefVal: "None"},
			{Name: "dtype", DefVal: "None"},
			{Name: "casting", DefVal: "\"same_kind\""},
		}},
		{"(x1, x2, /, out=None, *, where=True, casting='same_kind', order='K', dtype=None, subok=True[, signature])", []*Arg{
			{Name: "x1"},
			{Name: "x2"},
			{Name: "/"},
			{Name: "out", DefVal: "None"},
			{Name: "*"},
			{Name: "where", DefVal: "True"},
			{Name: "casting", DefVal: "'same_kind'"},
			{Name: "order", DefVal: "'K'"},
			{Name: "dtype", DefVal: "None"},
			{Name: "subok", DefVal: "True"},
			{Name: "signature", Optional: true},
		}},
		{"(x[, out1, out2], / [, out=(None, None)], *, where=True, casting='same_kind', order='K', dtype=None, subok=True[, signature, extobj])", []*Arg{
			{Name: "x"},
			{Name: "out1", Optional: true},
			{Name: "out2", Optional: true},
			{Name: "/"},
			{Name: "out", DefVal: "(None, None)", Optional: true},
			{Name: "*"},
			{Name: "where", DefVal: "True"},
			{Name: "casting", DefVal: "'same_kind'"},
			{Name: "order", DefVal: "'K'"},
			{Name: "dtype", DefVal: "None"},
			{Name: "subok", DefVal: "True"},
			{Name: "signature", Optional: true},
			{Name: "extobj", Optional: true},
		}},
		{"(op1=func1, op2=func2, ...)", []*Arg{
			{Name: "op1", DefVal: "func1"},
			{Name: "op2", DefVal: "func2"},
			{Name: "**kwargs"},
		}},
		{"(*args, **kwargs)", []*Arg{
			{Name: "*args"},
			{Name: "**kwargs"},
		}},
		{"(start: 'Union[int, float]', stop: 'Union[int, float]', /, num: 'int', *, dtype: 'Optional[Dtype]' = None, device: 'Optional[Device]' = None, endpoint: 'bool' = True) -> 'Array'", []*Arg{
			{Name: "start", Type: "'Union[int, float]'"},
			{Name: "stop", Type: "'Union[int, float]'"},
			{Name: "/"},
			{Name: "num", Type: "'int'"},
			{Name: "*"},
			{Name: "dtype", Type: "'Optional[Dtype]'", DefVal: "None"},
			{Name: "device", Type: "'Optional[Device]'", DefVal: "None"},
			{Name: "endpoint", Type: "'bool'", DefVal: "True"},
		}},
		{"(input, k=1, dims=[0,1]) -> Tensor", []*Arg{
			{Name: "input"},
			{Name: "k", DefVal: "1"},
			{Name: "dims", DefVal: "[0,1]"},
		}},
		{"(start: int[, step], ...)", []*Arg{
			{Name: "start", Type: "int"},
			{Name: "step", Optional: true},
			{Name: "**args"},
		}},
	}
	for _, c := range cases {
		args := Parse(c.sig)
		if len(args) != len(c.args) {
			t.Fatalf("%s: len(args) = %v, want %v", c.sig, len(args), len(c.args))
		}
		for i, arg := range args {
			want := c.args[i]
			if arg.Name != want.Name || arg.Type != want.Type || arg.DefVal != want.DefVal || arg.Optional != want.Optional {
				t.Fatalf("%s: args[%v] = %v, want %v", c.sig, i, arg, want)
			}
		}
	}
}

