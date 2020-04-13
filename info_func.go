/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package creditcard

import (
	"regexp"
	"strings"
)

var (
	reVISA            *regexp.Regexp
	reMaster          *regexp.Regexp
	reAmericanExpress *regexp.Regexp
	reJCB             *regexp.Regexp
	reUnionPay        *regexp.Regexp
)

func init() {
	// 4xxx
	reVISA = regexp.MustCompile("^4")
	// 51-55, 2221-2720
	reMaster = regexp.MustCompile("^(5[1-5]|222[1-9]|22[3-9][0-9]|27[01][0-9]|2720)")
	// 34, 37
	reAmericanExpress = regexp.MustCompile("^3[47]")
	// 3528-3589
	reJCB = regexp.MustCompile("^35(2[89]|[3-8][0-9])")
	reUnionPay = regexp.MustCompile("^(62|81)")
}

func cardType(pan [4]string) (ret CardType) {
	switch {
	case reVISA.MatchString(pan[0]):
		return VISACard
	case reMaster.MatchString(pan[0]):
		return MasterCard
	case reAmericanExpress.MatchString(pan[0]):
		return AmericanExpress
	case reJCB.MatchString(pan[0]):
		return JCBCard
	case reUnionPay.MatchString(pan[0]):
		return UnionPay
	}

	return UnknownCardType
}

func (i *info) Validate() (err error) {
	pan := i.RawPAN()
	if strings.Index(pan, "*") != -1 {
		return ErrValidateMasked
	}

	l := len(pan)
	checksum := pan[l-1] - '0'
	sum := byte(0)
	for idx, c := range []byte(pan[:l-1]) {
		x := byte(2)
		if idx%2 == 0 {
			x = byte(1)
		}
		c -= '0'
		c *= x
		if c > 9 {
			c -= 9
		}
		sum += c
	}
	if m := (checksum + sum) % 10; m != 0 {
		return ErrValidate
	}

	return
}

func (i *info) CardType() (ret CardType) {
	return i.typ
}

func (i *info) Checksum() (ret string) {
	return i.pan[3][3:]
}

func (i *info) Last4() (ret string) {
	return i.pan[3]
}

func (i *info) First6() (ret string) {
	return i.pan[0] + i.pan[1][:2]
}

func (i *info) FullLast4() (ret string) {
	return "****-****-****-" + i.pan[3]
}

func (i *info) FullFirst6() (ret string) {
	return i.pan[0] + "-" + i.pan[1][:2] + "**-****-****"
}

func (i *info) RawMasked() (ret string) {
	return i.First6() + "******" + i.Last4()
}

func (i *info) Masked() (ret string) {
	return i.pan[0] + "-" + i.pan[1][:2] + "**-****-" + i.pan[3]
}

func (i *info) RawPAN() (ret string) {
	return strings.Join(i.pan[:], "")
}

func (i *info) PAN() (ret string) {
	return strings.Join(i.pan[:], "-")
}

var reSlicedPAN *regexp.Regexp

func init() {
	reSlicedPAN = regexp.MustCompile("^[0-9*]{0,4}$")
}

// FromSlice creates Info instance from slice of string
//
// The slice is limited with following restrictions:
//
//   - len(arr) <= 4
//   - len(each element) <= 4
//   - each elemment is composed by digits of asterisk (/[0-9*]/)
//
// Missing digits are padded by asterisks ("*"). For example,
// FromSlice(nil).PAN() == "****-****-****-****"
func FromSlice(arr []string) (ret Info, err error) {
	l := len(arr)
	if l > 4 {
		err = ErrSection
		return
	}

	for l < 4 {
		arr = append(arr, "")
		l = len(arr)
	}

	for idx, v := range arr {
		if !reSlicedPAN.MatchString(v) {
			err = ErrSection
			return
		}
		if l := len(v); l < 4 {
			arr[idx] = v + strings.Repeat("*", 4-l)
		}
	}

	pan := [4]string{arr[0], arr[1], arr[2], arr[3]}
	typ := cardType(pan)
	ret = &info{pan: pan, typ: typ}
	return
}

// FromDashed creates Info instance by dashed PAN (xxxx-xxxx-xxxx-xxxx)
//
// It's nothing but FromSlice(strings.Split(pan, "-")), so everything about
// FromSlice applies to it.
func FromDashed(str string) (ret Info, err error) {
	arr := strings.Split(str, "-")
	return FromSlice(arr)
}

// FromRaw creates Info instance by raw PAN (xxxxxxxxxxxxxxxx)
//
// It checks if len(pan) is 16, and FromSlice is called to create Info instance.
func FromRaw(str string) (ret Info, err error) {
	if len(str) != 16 {
		err = ErrRaw
		return
	}

	return FromPart(
		str[:4],
		str[4:8],
		str[8:12],
		str[12:],
	)
}

// FromPart wraps FromSlice, so everything about FromSlice applies to it
func FromPart(parts ...string) (ret Info, err error) {
	return FromSlice(parts)
}

// FromMasked creates Info instance with first 6 digits and last 4 digits
//
// You can omit any of first6/last4, asterisks are padded to it. But passing more
// than 6/4 digits is not allowed.
func FromMasked(first6, last4 string) (ret Info, err error) {
	if len(first6) != 6 || len(last4) != 4 {
		err = ErrMasked
		return
	}
	return FromPart(first6[:4], first6[4:]+"**", "****", last4)
}
