package lnurl

import (
	"strings"
	"testing"
)

// TestLNURLDecode will test if LNURLDecode returns the expected string
func TestLNURLDecode(t *testing.T) {
	type args struct {
		code string
	}
	tests := []struct {
		name    string
		desc    string
		args    args
		want    string
		wantErr bool
	}{
		{desc: "change_scheme",
			args: args{code: "lnurl1d3h82unvwqaz7tmpwp5juenfv96x5ctx9e3k7mf0wccj7mrww4exctmsv9usxr8j98"}, // lnurlp://api.fiatjaf.com/v1/lnurl/pay
			want: "https://api.fiatjaf.com/v1/lnurl/pay"},
		{desc: "puny",
			args: args{code: "lnurl1dp68gurn8ghj77rw95khx7r3wc6kwv3nv3uhyvmpxserscfwvdhk6tmkxyhkcmn4wfkz7urp0y6jmn9j"}, // https://xn--sxqv5g23dyr3a428a.com/v1/lnurl/pay
			want: "https://xn--sxqv5g23dyr3a428a.com/v1/lnurl/pay"},                                                    // check puny code
		{desc: "puny2",
			args: args{code: "lnurl1dp68gurn8ghjle443052l909sxr7t8ulukgg6tnrdakj7a339akxuatjdshhqctea4v3jm"},
			want: "https://测试假域名.com/v1/lnurl/pay"}, // check puny code
		{desc: "puny_onion",
			args: args{code: "lnurl1dp68gup69uh7ddvtazhetevpsljel8l9jzxjummwd9hkutmkxyhkcmn4wfkz7urp0ygc52rr"},
			want: "http://测试假域名.onion/v1/lnurl/pay"}, // check puny onion (not sure know if this is possible)
		{desc: "puny_onion",
			args: args{code: "lnurl1dp68gurn8ghj7er0d4skjm3wdahxjmmw9amrztmvde6hymp0wpshjcw5kvw"},
			want: "http://domain.onion/v1/lnurl/pay"}, // check puny onion (not sure know if this is possible)
		{desc: "string",
			args: args{code: "lnurl1d3h82unvjhypn2"},
			want: "lnurl", wantErr: true}, // invalid domain name. returns error and decoded input
		{desc: "string_uppercase",
			args: args{code: strings.ToUpper("lnurl1d3h82unvjhypn2")},
			want: "lnurl", wantErr: true},
		{desc: "httpsScheme",
			args: args{code: "https://lnurl.fiatjaf.com"}, // do noting
			want: "https://lnurl.fiatjaf.com"},
		{desc: "lnurlp",
			args: args{code: "lnurlp://lnurl.fiatjaf.com"}, // change scheme
			want: "https://lnurl.fiatjaf.com"},
		{desc: "onion",
			args: args{code: "onion://lnurl.fiatjaf.onion"}, // change scheme
			want: "http://lnurl.fiatjaf.onion"},
		{desc: "encoded_onion",
			args: args{code: "lnurl1dahxjmmw8ghj7mrww4exctnxd9shg6npvchx7mnfdahquxueex"}, // change scheme
			want: "http://lnurl.fiatjaf.onion"},
		{desc: "encoded_https_onion",
			args: args{code: "lnurl1dp68gurn8ghj7mrww4exctnxd9shg6npvchx7mnfdahq874q6e"}, // change scheme
			want: "http://lnurl.fiatjaf.onion"},
	}
	for _, tt := range tests {
		tt.name = tt.args.code
		t.Run(tt.name, func(t *testing.T) {
			got, err := LNURLDecode(tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("LNURLDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LNURLDecode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLNURLEncodeWithValidation(t *testing.T) {
	type args struct {
		actualurl string
	}
	tests := []struct {
		name    string
		desc    string
		args    args
		want    string
		wantErr bool
	}{
		{desc: "com.org_invalid",
			args: args{actualurl: "lnurl.com.orgs"},                                      //invalid tld
			want: strings.ToUpper("LNURL1D3H82UNV9E3K7MFWDAEXWUCYVANJJ"), wantErr: true}, // lnurl.com.orgs
		{desc: "onion",
			args: args{actualurl: "http://juhaavcaxhxlp77nkq76byazcldy2hlmovfu2epvl5ankdibsot4csyd.onion/"},                                                                        //invalid tld
			want: strings.ToUpper("lnurl1dp68gup69uhk5atgv9shvcmp0p58smrsxumku6m3xumxy7tp0f3kcerexf5xcmt0wen82vn9wpmxcdtpde4kg6tzwdhhgdrrwdukgtn0de5k7m30p4k848"), wantErr: false}, // lnurl.com.orgs
		{desc: "invalid",
			args: args{actualurl: "lnurl.msfmsdkfns&dfandfasdf"},                                              // invalid TLD
			want: strings.ToUpper("LNURL1D3H82UNV9EKHXENDWDJXKENWWVNXGENPDEJXVCTNV3NQXM32ED"), wantErr: true}, // lnurl.msfmsdkfns&dfandfasdf
		{desc: "com.org",
			args: args{actualurl: "lnurl.com.org"},
			want: strings.ToUpper("LNURL1DP68GURN8GHJ7MRWW4EXCTNRDAKJUMMJVUK5QJ46"), wantErr: false}, // https://lnurl.com.org,
		{desc: "invalid_domain_name",
			args: args{actualurl: "lnurl"},
			want: strings.ToUpper("lnurl1d3h82unvjhypn2"), wantErr: true},
		{desc: "invalid_domain_name",
			args: args{actualurl: "lnur%%%l"},
			want: strings.ToUpper("lnurl1d3h82u39y5jkc45g4zl"), wantErr: true},
		{desc: "com",
			args: args{actualurl: "lnurl.com"},
			want: "LNURL1DP68GURN8GHJ7MRWW4EXCTNRDAKS6NPC70", wantErr: false}, // https://lnurl.com
	}
	for _, tt := range tests {
		tt.name = tt.args.actualurl
		t.Run(tt.name, func(t *testing.T) {
			got, err := LNURLEncodeWithValidation(tt.args.actualurl)
			if (err != nil) != tt.wantErr {
				t.Errorf("LNURLDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LNURLDecode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLNURLDecode will test if LNURLDecode returns the expected string
func TestLNURLEncode(t *testing.T) {
	type args struct {
		actualurl string
	}
	tests := []struct {
		name    string
		desc    string
		args    args
		want    string
		wantErr bool
	}{
		{desc: "com.org",
			args: args{actualurl: "lnurl.com.org"},
			want: strings.ToUpper("LNURL1DP68GURN8GHJ7MRWW4EXCTNRDAKJUMMJVUK5QJ46"), wantErr: false}, // https://lnurl.com.org,
		{desc: "invalid_domain_name",
			args: args{actualurl: "lnurl"},
			want: strings.ToUpper("lnurl1d3h82unvjhypn2"), wantErr: false}, // no domain validation
		{desc: "invalid_domain_name",
			args: args{actualurl: "lnur%%%l"},
			want: strings.ToUpper("lnurl1d3h82u39y5jkc45g4zl"), wantErr: false}, // no domain validation
		{desc: "com",
			args: args{actualurl: "lnurl.com"},
			want: "LNURL1DP68GURN8GHJ7MRWW4EXCTNRDAKS6NPC70", wantErr: false}, // https://lnurl.com
		{desc: "lnurlp",
			args: args{actualurl: "lnurlp://api.fiatjaf.com/v2/lnurl/pay"},
			want: "LNURL1DP68GURN8GHJ7CTSDYHXV6TPW34XZE3WVDHK6TMKXGHKCMN4WFKZ7URP0Y0Q3PEG"}, // https://api.fiatjaf.com/v2/lnurl/pay
		{desc: "onion_lnurlp",
			args: args{actualurl: "lnurlp://api.fiatjaf.onion/v2/lnurl/pay"},
			want: "LNURL1DP68GUP69UHKZURF9ENXJCT5DFSKVTN0DE5K7M30WCEZ7MRWW4EXCTMSV9USWEVHQG"}, // http://api.fiatjaf.onion/v2/lnurl/pay
		{desc: "https_onion",
			args: args{actualurl: "https://api.fiatjaf.onion/v2/lnurl/pay"},
			want: "LNURL1DP68GUP69UHKZURF9ENXJCT5DFSKVTN0DE5K7M30WCEZ7MRWW4EXCTMSV9USWEVHQG"}, // http://api.fiatjaf.onion/v2/lnurl/pay
		{desc: "http",
			args: args{actualurl: "http://api.fiatjaf.com/v2/lnurl/pay"},
			want: "LNURL1DP68GURN8GHJ7CTSDYHXV6TPW34XZE3WVDHK6TMKXGHKCMN4WFKZ7URP0Y0Q3PEG"}, // https://api.fiatjaf.com/v2/lnurl/pay
	}
	for _, tt := range tests {
		tt.name = tt.args.actualurl
		t.Run(tt.name, func(t *testing.T) {
			got, err := LNURLEncode(tt.args.actualurl)
			if (err != nil) != tt.wantErr {
				t.Errorf("LNURLDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LNURLDecode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLNURLEncodeDecode will test if encoding and decoding a lnurl will result in the same string
func TestLNURLEncodeDecode(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		desc    string
		name    string
		args    args
		wantErr bool
		want    string
	}{
		{desc: "no_scheme_onion",
			args: args{input: "hellosir.onion"}, wantErr: true, want: "http://hellosir.onion"}, // will be encoded to http://
		{desc: "https",
			args: args{input: "https://api.fiatjaf.com/v1/lnurl/pay"}, want: "https://api.fiatjaf.com/v1/lnurl/pay"},
		{desc: "https_scheme",
			args: args{input: "https://lnurl.fiatjaf.com"}, want: "https://lnurl.fiatjaf.com"},
		{desc: "lnurlp",
			args: args{input: "lnurlp://lnurl.fiatjaf.com"}, want: "https://lnurl.fiatjaf.com"}, // will be encoded to https://
		{desc: "no_scheme_com",
			args: args{input: "lnurl.fiatjaf.com"}, want: "https://lnurl.fiatjaf.com"},
		{desc: "no_domain_string",
			args: args{input: "lnurl"}, want: "lnurl", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// encode string first
			enc, err := LNURLEncode(tt.args.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LNURLDecode() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			// decode string
			dec, err := LNURLDecode(enc)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LNURLDecode() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			// decode string
			dec, err = LNURLDecode(enc)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LNURLDecode() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			enc, err = LNURLEncode(tt.args.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LNURLDecode() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			// decode string
			dec, err = LNURLDecode(enc)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LNURLDecode() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			// check if decoded string == actualurl
			if dec != tt.want {
				t.Errorf("LNURLDecode() enc = %v, dec %v", enc, dec)
			}
		})
	}
}
