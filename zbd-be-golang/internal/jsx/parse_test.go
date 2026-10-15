package jsx

import (
	"strings"
	"testing"
)

// JSON.parse → JSON.stringify 왕복이 Node 와 같은지(키 순서·숫자 표기·문자열 escape)
func TestRoundTripLikeNode(t *testing.T) {
	cases := map[string]string{
		`{"b":1,"2":0,"a":{"10":1,"9":2,"x":3},"1":true}`: `{"1":true,"2":0,"b":1,"a":{"9":2,"10":1,"x":3}}`,
		`{"01":1,"-1":2,"1.5":3,"4294967295":4,"7":5}`:    `{"7":5,"01":1,"-1":2,"1.5":3,"4294967295":4}`,
		`[1.0, 2.50, 1e21, 123456789012, -0, "<&>"]`:      `[1,2.5,1e+21,123456789012,0,"<&>"]`,
		`{"a":1,"a":2,"b":3}`:                             `{"a":2,"b":3}`,
	}
	for in, want := range cases {
		v, err := ParseString(in)
		if err != nil {
			t.Fatal(in, err)
		}
		if got := Stringify(v); got != want {
			t.Errorf("%s\n got  %s\n want %s", in, got, want)
		}
	}
}

func TestNumStrLikeJS(t *testing.T) {
	// node -e 'console.log([...].map(String).join(" "))' 결과
	in := []float64{0.1, 0.00001, 0.000001, 1e-7, 123.456, 1e21, 1.5e21, 1e20, 123456789012345680000, 5e-324, 1.7976931348623157e308, -2.5, 100, sum(0.1, 0.2)}
	want := "0.1 0.00001 0.000001 1e-7 123.456 1e+21 1.5e+21 100000000000000000000 123456789012345680000 5e-324 1.7976931348623157e+308 -2.5 100 0.30000000000000004"
	var got []string
	for _, f := range in {
		got = append(got, NumStr(f))
	}
	if g := strings.Join(got, " "); g != want {
		t.Errorf("\n got  %s\n want %s", g, want)
	}
}

func sum(a, b float64) float64 { return a + b } // 상수 접기 피하기
