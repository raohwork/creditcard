/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package creditcard

// CardType denotes a card issuer, only few are supported.
type CardType int

// Supported card issuers
const (
	UnknownCardType CardType = -1

	beginKnownCardType CardType = iota
	VISACard                    // VISA
	MasterCard                  // MAstercard
	JCBCard                     // JCB
	AmericanExpress             // American Express
	UnionPay                    // China UnionPay
	endKnownCardType
)

func asCardType(t CardType) (ret CardType) {
	switch {
	case t <= beginKnownCardType:
		return UnknownCardType
	case t > endKnownCardType:
		return UnknownCardType
	}

	return t
}

// ErrPANFormat indicates there's something wrong with PAN numbers
type ErrPANFormat string

func (e ErrPANFormat) Error() (ret string) {
	return "creditcard: incorrect pan format: " + string(e)
}

// Possible errors returned by this package
const (
	ErrSection        ErrPANFormat = "there must be 4 sections in PAN, each section must be 4 digits"
	ErrRaw            ErrPANFormat = "raw pan must be 16 digits or asterisks"
	ErrMasked         ErrPANFormat = "masked pan must be first 6 digits and last 4 digits"
	ErrValidateMasked ErrPANFormat = "masked pan cannot be validated"
	ErrValidate       ErrPANFormat = "invalid pan"
)

// Info is the main interface to acces helpers in this package
type Info interface {
	CardType() (ret CardType)
	Checksum() (ret string)   // returns last digit
	Last4() (ret string)      // returns "3456"
	First6() (ret string)     // returns "123456"
	FullLast4() (ret string)  // returns "****-****-****-3456"
	FullFirst6() (ret string) // returns "1234-56**-****-****"
	RawMasked() (ret string)  // returns "123456******1234"
	Masked() (ret string)     // returns "1234-56**-****-1234"
	RawPAN() (ret string)     // returns "1234567890123456"
	PAN() (ret string)        // returns "1234-5678-9012-3456"
	// returns ErrValidateMasked if pan is masked, ErrValidate if pan is
	// invalid, or nil if pan is valid
	Validate() (err error)
}

type info struct {
	pan [4]string
	typ CardType
}
