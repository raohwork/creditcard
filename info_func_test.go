/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package creditcard

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	cases := map[string]error{
		"0000000000000000": nil,
		"0000000000000001": ErrValidate,
		"00000000000000*0": ErrValidateMasked,
		"0000000000000019": nil,
		"0000000000000108": nil,
	}

	for pan, expect := range cases {
		t.Run(pan, func(t *testing.T) {
			info, err := FromRaw(pan)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			err = info.Validate()
			if err != expect {
				t.Log("expect:", expect)
				t.Log("actual:", err)
				t.Fatal("unexpected result")
			}
		})
	}
}

func TestPAN(t *testing.T) {
	info, err := FromRaw("1234567890123456")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	cases := map[string]func() string{
		"1234-5678-9012-3456": info.PAN,
		"1234567890123456":    info.RawPAN,
		"1234-56**-****-3456": info.Masked,
		"123456******3456":    info.RawMasked,
		"123456":              info.First6,
		"3456":                info.Last4,
		"1234-56**-****-****": info.FullFirst6,
		"****-****-****-3456": info.FullLast4,
		"6":                   info.Checksum,
	}

	for expect, f := range cases {
		val := reflect.ValueOf(f)
		name := val.Type().Name()
		t.Run(name, func(t *testing.T) {
			actual := f()
			if expect != actual {
				t.Log("expect:", expect)
				t.Log("actual:", actual)
				t.Fatal("unexpected result")
			}
		})
	}
}

func TestCardType(t *testing.T) {
	const digits = 4
	max := 10 ^ (digits + 1) - 1
	type tc struct {
		name   string
		prefix string
		expect CardType
	}
	cases := append(
		make([]tc, 0, max),
		[]tc{
			{
				name:   "visa",
				prefix: "4",
				expect: VISACard,
			},
			{
				name:   "master51",
				prefix: "51",
				expect: MasterCard,
			},
			{
				name:   "master52",
				prefix: "52",
				expect: MasterCard,
			},
			{
				name:   "master53",
				prefix: "53",
				expect: MasterCard,
			},
			{
				name:   "master54",
				prefix: "54",
				expect: MasterCard,
			},
			{
				name:   "master55",
				prefix: "55",
				expect: MasterCard,
			},
			{
				name:   "americanexpress34",
				prefix: "34",
				expect: AmericanExpress,
			},
			{
				name:   "americanexpress37",
				prefix: "37",
				expect: AmericanExpress,
			},
			{
				name:   "unionpay62",
				prefix: "62",
				expect: UnionPay,
			},
			{
				name:   "unionpay81",
				prefix: "81",
				expect: UnionPay,
			},
		}...,
	)

	// generate JCB test cases
	for i := 3528; i <= 3589; i++ {
		str := strconv.Itoa(i)
		cases = append(cases, tc{
			name:   "jcb" + str,
			prefix: str,
			expect: JCBCard,
		})
	}

	// generate rest cases
	tmpl := "%0" + strconv.Itoa(digits) + "d"
	for i := 0; i <= max; i++ {
		str := fmt.Sprintf(tmpl, i)
		dupe := false
		for _, c := range cases {
			if !strings.HasPrefix(str, c.prefix) {
				continue
			}

			dupe = true
			break
		}

		if dupe {
			continue
		}

		cases = append(cases, tc{
			name:   "unknown" + str,
			prefix: str,
			expect: UnknownCardType,
		})
	}

	f := func(padding, prefix string, expect CardType) func(t *testing.T) {
		return func(t *testing.T) {
			pan := prefix + strings.Repeat(padding, 16-len(prefix))
			info, err := FromRaw(pan)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if actual := info.CardType(); expect != actual {
				t.Log("expect:", expect)
				t.Log("actual:", actual)
				t.Fatal("unexpected result")
			}
		}
	}

	for _, c := range cases {
		t.Run(c.name+"_begin", f("0", c.prefix, c.expect))
		t.Run(c.name+"_mid", f("5", c.prefix, c.expect))
		t.Run(c.name+"_end", f("9", c.prefix, c.expect))
	}
}

type creationTestCase struct {
	name   string
	slice  []string
	err    error
	expect string // raw
}

func testCreation(t *testing.T, f func(creationTestCase) (Info, error), c creationTestCase, prefix string) {
	name := c.name
	if prefix != "" {
		name = prefix + "_" + c.name
	}
	t.Run(name, func(t *testing.T) {
		info, err := f(c)
		if c.err != err {
			t.Log("expect (err):", c.err)
			t.Log("actual (err):", err)
			t.Fatal("unexpected error")
		}
		if c.err != nil {
			return
		}

		actual := info.RawPAN()
		if c.expect != actual {
			t.Log("expect:", c.expect)
			t.Log("actual:", actual)
			t.Fatal("unexpected result")
		}
	})
}

func TestGeneralCreation(t *testing.T) {
	cases := []creationTestCase{
		{
			name:   "normal",
			slice:  []string{"1234", "5678", "9012", "3456"},
			expect: "1234567890123456",
		},
		{
			name:   "asterisk",
			slice:  []string{"1234", "56**", "****", "3456"},
			expect: "123456******3456",
		},
		{
			name:   "less",
			slice:  []string{"1234", "56", "", "3456"},
			expect: "123456******3456",
		},
		{
			name:  "more1",
			slice: []string{"12341", "5678", "9012", "3456"},
			err:   ErrSection,
		},
		{
			name:  "more2",
			slice: []string{"1234", "56781", "9012", "3456"},
			err:   ErrSection,
		},
		{
			name:  "more3",
			slice: []string{"1234", "5678", "90121", "3456"},
			err:   ErrSection,
		},
		{
			name:  "more4",
			slice: []string{"1234", "5678", "9012", "34561"},
			err:   ErrSection,
		},
		{
			name:  "alpahbet1",
			slice: []string{"134a", "5678", "9012", "3456"},
			err:   ErrSection,
		},
		{
			name:  "alpahbet2",
			slice: []string{"1234", "578a", "9012", "3456"},
			err:   ErrSection,
		},
		{
			name:  "alpahbet3",
			slice: []string{"1234", "5678", "912a", "3456"},
			err:   ErrSection,
		},
		{
			name:  "alpahbet4",
			slice: []string{"1234", "5678", "9012", "356a"},
			err:   ErrSection,
		},
	}

	types := map[string]func(creationTestCase) (Info, error){
		"slice": func(c creationTestCase) (ret Info, err error) {
			return FromSlice(c.slice)
		},
		"part": func(c creationTestCase) (ret Info, err error) {
			return FromPart(c.slice...)
		},
		"dashed": func(c creationTestCase) (ret Info, err error) {
			return FromDashed(strings.Join(c.slice, "-"))
		},
	}

	for prefix, f := range types {
		for _, c := range cases {
			testCreation(t, f, c, prefix)
		}
	}
}

func TestRawCreation(t *testing.T) {
	cases := []creationTestCase{
		{
			name:   "normal",
			slice:  []string{"1234", "5678", "9012", "3456"},
			expect: "1234567890123456",
		},
		{
			name:   "asterisk",
			slice:  []string{"1234", "56**", "****", "3456"},
			expect: "123456******3456",
		},
		{
			name:  "less",
			slice: []string{"1234", "56", "", "3456"},
			err:   ErrRaw,
		},
		{
			name:  "more1",
			slice: []string{"12341", "5678", "9012", "3456"},
			err:   ErrRaw,
		},
		{
			name:  "more2",
			slice: []string{"1234", "56781", "9012", "3456"},
			err:   ErrRaw,
		},
		{
			name:  "more3",
			slice: []string{"1234", "5678", "90121", "3456"},
			err:   ErrRaw,
		},
		{
			name:  "more4",
			slice: []string{"1234", "5678", "9012", "34561"},
			err:   ErrRaw,
		},
		{
			name:  "alpahbet1",
			slice: []string{"124a", "5678", "9012", "3456"},
			err:   ErrSection,
		},
		{
			name:  "alpahbet2",
			slice: []string{"1234", "578a", "9012", "3456"},
			err:   ErrSection,
		},
		{
			name:  "alpahbet3",
			slice: []string{"1234", "5678", "902a", "3456"},
			err:   ErrSection,
		},
		{
			name:  "alpahbet4",
			slice: []string{"1234", "5678", "9012", "346a"},
			err:   ErrSection,
		},
	}

	for _, c := range cases {
		testCreation(t, func(c creationTestCase) (ret Info, err error) {
			return FromRaw(strings.Join(c.slice, ""))
		}, c, "")
	}
}

func TestMaskedCreation(t *testing.T) {
	cases := []creationTestCase{
		{
			name:   "normal",
			slice:  []string{"123456", "3456"},
			expect: "123456******3456",
		},
		{
			name:   "first_mask",
			slice:  []string{"1234**", "3456"},
			expect: "1234********3456",
		},
		{
			name:   "first_mask",
			slice:  []string{"123456", "34**"},
			expect: "123456******34**",
		},
		{
			name:  "first_less",
			slice: []string{"12345", "3456"},
			err:   ErrMasked,
		},
		{
			name:  "last_less",
			slice: []string{"123456", "345"},
			err:   ErrMasked,
		},
	}

	for _, c := range cases {
		testCreation(t, func(c creationTestCase) (ret Info, err error) {
			return FromMasked(c.slice[0], c.slice[1])
		}, c, "")
	}
}
