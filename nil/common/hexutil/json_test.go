package hexutil

import (
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
	"testing"

	"github.com/NilFoundation/nil/nil/common/check"
	"github.com/stretchr/testify/require"
)

func checkError(t *testing.T, input string, got, want error) {
	t.Helper()
	if want == nil {
		require.NoErrorf(t, got, "input %s", input)
		return
	}
	if got == nil {
		require.NoError(t, want, "input %s", input)
		return
	}
	require.Equal(t, want.Error(), got.Error(), "input %s", input)
}

func bigFromString(s string) *big.Int {
	b, ok := new(big.Int).SetString(s, 16)
	check.PanicIfNot(ok)
	return b
}

var errJSONEOF = errors.New("unexpected end of JSON input")

var unmarshalBigTests = []unmarshalTest{
	// invalid encoding
	{input: "", wantErr: errJSONEOF},
	{input: "null", wantErr: errNonString(bigT)},
	{input: "10", wantErr: errNonString(bigT)},
	{input: `"0"`, wantErr: wrapTypeError(ErrMissingPrefix, bigT)},
	{input: `"0x"`, wantErr: wrapTypeError(ErrEmptyNumber, bigT)},
	{input: `"0x01"`, wantErr: wrapTypeError(ErrLeadingZero, bigT)},
	{input: `"0xx"`, wantErr: wrapTypeError(ErrSyntax, bigT)},
	{input: `"0x1zz01"`, wantErr: wrapTypeError(ErrSyntax, bigT)},
	{
		input:   `"0x10000000000000000000000000000000000000000000000000000000000000000"`,
		wantErr: wrapTypeError(ErrBig256Range, bigT),
	},
	// valid encoding
	{input: `""`, want: big.NewInt(0)},
	{input: `"0x0"`, want: big.NewInt(0)},
	{input: `"0x2"`, want: big.NewInt(0x2)},
	{input: `"0x2F2"`, want: big.NewInt(0x2f2)},
	{input: `"0X2F2"`, want: big.NewInt(0x2f2)},
	{input: `"0x1122aaff"`, want: big.NewInt(0x1122aaff)},
	{input: `"0xbBb"`, want: big.NewInt(0xbbb)},
	{input: `"0xfffffffff"`, want: big.NewInt(0xfffffffff)},
	{
		input: `"0x112233445566778899aabbccddeeff"`,
		want:  bigFromString("112233445566778899aabbccddeeff"),
	},
	{
		input: `"0xffffffffffffffffffffffffffffffffffff"`,
		want:  bigFromString("ffffffffffffffffffffffffffffffffffff"),
	},
	{
		input: `"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"`,
		want:  bigFromString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
	},
}

func TestUnmarshalBig(t *testing.T) {
	t.Parallel()

	for idx, test := range unmarshalBigTests {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			t.Parallel()

			var v Big
			err := json.Unmarshal([]byte(test.input), &v)
			checkError(t, test.input, err, test.wantErr)
			if test.want != nil {
				expected, ok := test.want.(*big.Int)
				require.True(t, ok)
				require.Zero(t, expected.Cmp(v.ToInt()))
			}
		})
	}
}

func BenchmarkUnmarshalBig(b *testing.B) {
	input := []byte(`"0x123456789abcdef123456789abcdef"`)
	for b.Loop() {
		var v Big
		if err := v.UnmarshalJSON(input); err != nil {
			b.Fatal(err)
		}
	}
}

func TestMarshalBig(t *testing.T) {
	t.Parallel()

	for idx, test := range encodeBigTests {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			t.Parallel()

			in, ok := test.input.(*big.Int)
			require.True(t, ok)
			out, err := json.Marshal((*Big)(in))
			require.NoError(t, err)
			want := `"` + test.want + `"`
			require.Equal(t, want, string(out))
			require.Equal(t, test.want, (*Big)(in).String())
		})
	}
}

var unmarshalUint64Tests = []unmarshalTest{
	// invalid encoding
	{input: "", wantErr: errJSONEOF},
	{input: "null", wantErr: errNonString(uint64T)},
	{input: "10", wantErr: errNonString(uint64T)},
	{input: `"0"`, wantErr: wrapTypeError(ErrMissingPrefix, uint64T)},
	{input: `"0x"`, wantErr: wrapTypeError(ErrEmptyNumber, uint64T)},
	{input: `"0x01"`, wantErr: wrapTypeError(ErrLeadingZero, uint64T)},
	{input: `"0xfffffffffffffffff"`, wantErr: wrapTypeError(ErrUint64Range, uint64T)},
	{input: `"0xx"`, wantErr: wrapTypeError(ErrSyntax, uint64T)},
	{input: `"0x1zz01"`, wantErr: wrapTypeError(ErrSyntax, uint64T)},

	// valid encoding
	{input: `""`, want: uint64(0)},
	{input: `"0x0"`, want: uint64(0)},
	{input: `"0x2"`, want: uint64(0x2)},
	{input: `"0x2F2"`, want: uint64(0x2f2)},
	{input: `"0X2F2"`, want: uint64(0x2f2)},
	{input: `"0x1122aaff"`, want: uint64(0x1122aaff)},
	{input: `"0xbbb"`, want: uint64(0xbbb)},
	{input: `"0xffffffffffffffff"`, want: uint64(0xffffffffffffffff)},
}

func TestUnmarshalUint64(t *testing.T) {
	t.Parallel()

	for idx, test := range unmarshalUint64Tests {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			t.Parallel()

			var v Uint64
			err := json.Unmarshal([]byte(test.input), &v)
			checkError(t, test.input, err, test.wantErr)
			if test.want != nil {
				require.EqualValues(t, test.want, v)
			}
		})
	}
}

func BenchmarkUnmarshalUint64(b *testing.B) {
	input := []byte(`"0x123456789abcdf"`)
	for b.Loop() {
		var v Uint64
		_ = v.UnmarshalJSON(input)
	}
}

func TestMarshalUint64(t *testing.T) {
	t.Parallel()

	for idx, test := range encodeUint64Tests {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			t.Parallel()

			in, ok := test.input.(uint64)
			require.True(t, ok)
			out, err := json.Marshal(Uint64(in))
			require.NoError(t, err)
			want := `"` + test.want + `"`
			require.Equal(t, want, string(out))
			require.Equal(t, test.want, (Uint64)(in).String())
		})
	}
}

func TestMarshalUint(t *testing.T) {
	t.Parallel()

	for idx, test := range encodeUintTests {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			t.Parallel()

			in, ok := test.input.(uint)
			require.True(t, ok)
			out, err := json.Marshal(Uint(in))
			require.NoError(t, err)
			want := `"` + test.want + `"`
			require.Equal(t, want, string(out))
			require.Equal(t, test.want, (Uint)(in).String())
		})
	}
}

var (
	maxUint33bits = uint64(^uint32(0)) + 1
	maxUint64bits = ^uint64(0)
)

var unmarshalUintTests = []unmarshalTest{
	// invalid encoding
	{input: "", wantErr: errJSONEOF},
	{input: "null", wantErr: errNonString(uintT)},
	{input: "10", wantErr: errNonString(uintT)},
	{input: `"0"`, wantErr: wrapTypeError(ErrMissingPrefix, uintT)},
	{input: `"0x"`, wantErr: wrapTypeError(ErrEmptyNumber, uintT)},
	{input: `"0x01"`, wantErr: wrapTypeError(ErrLeadingZero, uintT)},
	{input: `"0x100000000"`, want: uint(maxUint33bits), wantErr32bit: wrapTypeError(ErrUintRange, uintT)},
	{input: `"0xfffffffffffffffff"`, wantErr: wrapTypeError(ErrUintRange, uintT)},
	{input: `"0xx"`, wantErr: wrapTypeError(ErrSyntax, uintT)},
	{input: `"0x1zz01"`, wantErr: wrapTypeError(ErrSyntax, uintT)},

	// valid encoding
	{input: `""`, want: uint(0)},
	{input: `"0x0"`, want: uint(0)},
	{input: `"0x2"`, want: uint(0x2)},
	{input: `"0x2F2"`, want: uint(0x2f2)},
	{input: `"0X2F2"`, want: uint(0x2f2)},
	{input: `"0x1122aaff"`, want: uint(0x1122aaff)},
	{input: `"0xbbb"`, want: uint(0xbbb)},
	{input: `"0xffffffff"`, want: uint(0xffffffff)},
	{input: `"0xffffffffffffffff"`, want: uint(maxUint64bits), wantErr32bit: wrapTypeError(ErrUintRange, uintT)},
}

func TestUnmarshalUint(t *testing.T) {
	t.Parallel()

	for idx, test := range unmarshalUintTests {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			t.Parallel()

			var v Uint
			err := json.Unmarshal([]byte(test.input), &v)
			if uintBits == 32 && test.wantErr32bit != nil {
				checkError(t, test.input, err, test.wantErr32bit)
				return
			}
			checkError(t, test.input, err, test.wantErr)
			if test.want != nil {
				require.EqualValues(t, test.want, v)
			}
		})
	}
}
