package lnurl

import (
	"fmt"
	"strings"
	"testing"
)

// Starting tests with validation

// TestLNURLDecodeStrict will test if LNURLDecodeStrict returns the expected string
func TestLNURLDecodeStrict(t *testing.T) {
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
		{desc: "LUD17_CHANGE_SCHEMA",
			args: args{code: "lnurlp://lnurl.fiatjaf.com"}, // change scheme
			want: "https://lnurl.fiatjaf.com"},
		{desc: "ONION_SCHEMA_CHANGE_SCHEMA_ERROR",
			args: args{code: "onion://lnurl.fiatjaf.onion"}, // change scheme
			want: "http://lnurl.fiatjaf.onion", wantErr: true},
		{desc: "LUD17_BECH32_ENCODED_CHANGE_SCHEMA_ERROR",
			args: args{code: "lnurl1d3h82unvwqaz7tmpwp5juenfv96x5ctx9e3k7mf0wccj7mrww4exctmsv9usxr8j98"}, // lnurlp://api.fiatjaf.com/v1/lnurl/pay
			want: "https://api.fiatjaf.com/v1/lnurl/pay", wantErr: true},
		{desc: "LUD17_BECH32_ENCODED_CHANGE_SCHEMA_IP_ERROR",
			args: args{code: "lnurl1dp68gup69uhnzt339ccjuvf0wccj7mrww4exctmsv9usux0rql"}, // lnurlp://api.fiatjaf.com/v1/lnurl/pay
			want: "https://1.1.1.1/v1/lnurl/pay", wantErr: true},
		{desc: "LUD1_ENCODED_CHANGE_SCHEMA",
			args: args{code: "lnurl1dp68gurn8ghj7vfwxyhrzt339amrztmvde6hymp0wpshjr50pch"}, // https://1.1.1.1/v1/lnurl/pay
			want: "https://1.1.1.1/v1/lnurl/pay"},
		{desc: "LUD1_PUNY_SAME_SCHEMA",
			args: args{code: "lnurl1dp68gurn8ghj77rw95khx7r3wc6kwv3nv3uhyvmpxserscfwvdhk6tmkxyhkcmn4wfkz7urp0y6jmn9j"}, // https://xn--sxqv5g23dyr3a428a.com/v1/lnurl/pay
			want: "https://xn--sxqv5g23dyr3a428a.com/v1/lnurl/pay"},                                                    // check puny code
		{desc: "LUD1_IDN_SAME_SCHEMA",
			args: args{code: "lnurl1dp68gurn8ghjle443052l909sxr7t8ulukgg6tnrdakj7a339akxuatjdshhqctea4v3jm"},
			want: "https://测试假域名.com/v1/lnurl/pay"}, // check puny code
		{desc: "LUD17_BECH32_IDN_CHANGE_SCHEMA_ERROR",
			args: args{code: "lnurl1d3h82unvwuaz7tlxkk973tu4ukqc0evlnljeprfwvdhk6tmkxyhkcmn4wfkz7urp0y93ltxj"},
			want: "https://测试假域名.com/v1/lnurl/pay", wantErr: true}, // check puny code
		{desc: "LUD1_BECH32_ONION_SAME_SCHEMA",
			args: args{code: "lnurl1dp68gup69uh7ddvtazhetevpsljel8l9jzxjummwd9hkutmkxyhkcmn4wfkz7urp0ygc52rr"}, // http://测试假域名.onion/v1/lnurl/pay
			want: "http://测试假域名.onion/v1/lnurl/pay"},                                                           // check puny onion (not sure know if this is possible)
		{desc: "LUD1_BECH32_ONION_CHANGE_SCHEMA_ERROR",
			args: args{code: "lnurl1dp68gurn8ghj7er0d4skjm3wdahxjmmw9amrztmvde6hymp0wpshjcw5kvw"},
			want: "http://domain.onion/v1/lnurl/pay", wantErr: true}, // check puny onion (not sure know if this is possible)
		{desc: "RANDOM_STRING_ERROR",
			args: args{code: "lnurl1d3h82unvjhypn2"},
			want: "lnurl", wantErr: true}, // invalid domain name. returns error and decoded input
		{desc: "RANDOM_STRING_UPPERCASE_ERROR",
			args: args{code: strings.ToUpper("lnurl1d3h82unvjhypn2")},
			want: "lnurl", wantErr: true},
		{desc: "HTTPS",
			args: args{code: "https://lnurl.fiatjaf.com"}, // do noting
			want: "https://lnurl.fiatjaf.com"},
		{desc: "ONION_SCHEMA_CHANGE_SCHEMA_ERROR",
			args: args{code: "lnurl1dahxjmmw8ghj7mrww4exctnxd9shg6npvchx7mnfdahquxueex"}, // change scheme
			want: "http://lnurl.fiatjaf.onion", wantErr: true},
		{desc: "ONION_HTTPS_CHANGE_SCHEMA_ERROR",
			args: args{code: "lnurl1dp68gurn8ghj7mrww4exctnxd9shg6npvchx7mnfdahq874q6e"}, // change scheme
			want: "http://lnurl.fiatjaf.onion", wantErr: true},
	}
	for _, tt := range tests {
		tt.name = tt.args.code
		t.Run(fmt.Sprintf("%s", tt.desc), func(t *testing.T) {
			got, err := LNURLDecodeStrict(tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("LNURLDecodeStrict() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LNURLDecodeStrict() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLNURLEncodeStrict will test if LNURLEncodeStrict returns the expected string
func TestLNURLEncodeStrict(t *testing.T) {
	type args struct {
		actualurl string
	}
	tests := []struct {
		name    string
		desc    string
		args    args
		want    string
		wantErr bool
	}{{desc: "INVALID_IP_ERROR",
		args: args{actualurl: "onion://10.2.0.10:8888"},
		want: strings.ToUpper("LNURL1DP68GURN8GHJ7VFS9CEZUVPWXYCR5WPC8QUQAPPZPR"), wantErr: false},
		{desc: "INVALID_IP_ERROR",
			args: args{actualurl: "onion://111.2.999.2:88"},
			want: strings.ToUpper("lnurl1dahxjmmw8ghj7vf3xyhryt3e8yujuv368quq39nwhr"), wantErr: true},
		{desc: "ONION_SCHEME_ERROR",
			args: args{actualurl: "onion://lnurl.com"},
			want: strings.ToUpper("LNURL1DP68GURN8GHJ7MRWW4EXCTNRDAKS6NPC70"), wantErr: true},
		{desc: "INVALID_DOMAIN_NAME_2_ERROR",
			args: args{actualurl: "lnur%%%l"},
			want: strings.ToUpper("lnurl1d3h82u39y5jkc45g4zl"), wantErr: true},
		{desc: "IDN_CHANGE_SCHEMA_ERROR",
			args: args{actualurl: "http://看.com"},
			want: strings.ToUpper("lnurl1dp68gurn8ghjleuu3vhxxmmdrmldr8"), wantErr: true}, // https://看.com
		{desc: "NO_SCHEMA_INVALID_TLD_ERROR",
			args: args{actualurl: "lnurl.com.btc"},                                     //invalid tld
			want: strings.ToUpper("LNURL1D3H82UNV9E3K7MFWVF6XXSVAUH5"), wantErr: true}, // lnurl.com.btc
		{desc: "INVALID_DOMAIN_NAME_ERROR",
			args: args{actualurl: "lnurl"},
			want: strings.ToUpper("lnurl1d3h82unvjhypn2"), wantErr: true},
		{desc: "INVALID_TLD_ERROR",
			args: args{actualurl: "lnurl.msfmsdkfns&dfandfasdf"},                                              // invalid tld
			want: strings.ToUpper("LNURL1D3H82UNV9EKHXENDWDJXKENWWVNXGENPDEJXVCTNV3NQXM32ED"), wantErr: true}, // lnurl.msfmsdkfns&dfandfasdf
		{desc: "ADD_SCHEMA_COM_ORG",
			args: args{actualurl: "lnurl.com.org"},
			want: strings.ToUpper("LNURL1DP68GURN8GHJ7MRWW4EXCTNRDAKJUMMJVUK5QJ46"), wantErr: true}, // https://lnurl.com.org,
		{desc: "ADD_SCHEMA_COM",
			args: args{actualurl: "lnurl.com"},
			want: "LNURL1DP68GURN8GHJ7MRWW4EXCTNRDAKS6NPC70", wantErr: true}, // https://lnurl.com
		{desc: "COM_ORG_ADD_SCHEMA",
			args: args{actualurl: "lnurl.com.org"},
			want: strings.ToUpper("LNURL1DP68GURN8GHJ7MRWW4EXCTNRDAKJUMMJVUK5QJ46"), wantErr: true}, // https://lnurl.com.org,
		{desc: "LUD17_P_ONION",
			args: args{actualurl: "lnurlp://api.fiatjaf.onion/v2/lnurl/pay"},
			want: "lnurlp://api.fiatjaf.onion/v2/lnurl/pay"}, // http://api.fiatjaf.onion/v2/lnurl/pay
		{desc: "LUD17_A_ONION",
			args: args{actualurl: "lnurla://api.fiatjaf.onion/v2/lnurl/pay"},
			want: "lnurla://api.fiatjaf.onion/v2/lnurl/pay"}, // http://api.fiatjaf.onion/v2/lnurl/pay
		{desc: "LUD17_W_ONION",
			args: args{actualurl: "lnurlw://api.fiatjaf.onion/v2/lnurl/pay"},
			want: "lnurlw://api.fiatjaf.onion/v2/lnurl/pay"}, // https://api.fiatjaf.com/v2/lnurl/pay

		{desc: "LUD17_IP_AND_PORT",
			args: args{actualurl: "lnurlp://1.1.1.1:8080/v1/lnurl/pay"},
			want: "lnurlp://1.1.1.1:8080/v1/lnurl/pay"}, // https://1.1.1.1:8080/v1/lnurl/pay
		{desc: "LUD17_IP",
			args: args{actualurl: "lnurlp://1.1.1.1/v1/lnurl/pay"},
			want: "lnurlp://1.1.1.1/v1/lnurl/pay"}, // https://1.1.1.1/v1/lnurl/pay

		{desc: "LUD1_ONION",
			args: args{actualurl: "http://juhaavcaxhxlp77nkq76byazcldy2hlmovfu2epvl5ankdibsot4csyd.onion/"},
			want: strings.ToUpper("lnurl1dp68gup69uhk5atgv9shvcmp0p58smrsxumku6m3xumxy7tp0f3kcerexf5xcmt0wen82vn9wpmxcdtpde4kg6tzwdhhgdrrwdukgtn0de5k7m30p4k848"), wantErr: false},
		{desc: "HTTPS",
			args: args{actualurl: "https://api.fiatjaf.com/v2/lnurl/pay"},
			want: strings.ToUpper("lnurl1dp68gurn8ghj7ctsdyhxv6tpw34xze3wvdhk6tmkxghkcmn4wfkz7urp0y0q3peg")}, // https://api.fiatjaf.com/v2/lnurl/pay

	}
	for _, tt := range tests {
		tt.name = tt.args.actualurl
		t.Run(fmt.Sprintf("%s", tt.desc), func(t *testing.T) {
			got, err := LNURLEncodeStrict(tt.args.actualurl)
			if (err != nil) != tt.wantErr {
				t.Errorf("LNURLDecodeStrict() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LNURLDecodeStrict() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// Start test without validation

// TestLNURLDecodeStrict will test if LNURLDecodeStrict returns the expected string
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
	}{{desc: "INVALID_IP_NO_ERROR",
		args: args{actualurl: "onion://111.2.999.2:88"},
		want: strings.ToUpper("lnurl1dahxjmmw8ghj7vf3xyhryt3e8yujuv368quq39nwhr"), wantErr: false},
		{desc: "ONION_SCHEME_NO_CHANGE",
			args: args{actualurl: "onion://lnurl.com"},
			want: strings.ToUpper("LNURL1DAHXJMMW8GHJ7MRWW4EXCTNRDAKSU44R9H"), wantErr: false},
		{desc: "INVALID_DOMAIN_NAME_2_ERROR",
			args: args{actualurl: "lnur%%%l"},
			want: strings.ToUpper("lnurl1d3h82u39y5jkc45g4zl"), wantErr: false},
		{desc: "IDN_NO_CHANGE_SCHEMA_ERROR",
			args: args{actualurl: "http://看.com"},
			want: strings.ToUpper("LNURL1DP68GUP69UH708YT9E3K7MGL572TH"), wantErr: false}, // http://看.com
		{desc: "NO_SCHEMA_INVALID_TLD_NO_ERROR",
			args: args{actualurl: "lnurl.com.btc"},                                      //invalid tld
			want: strings.ToUpper("LNURL1D3H82UNV9E3K7MFWVF6XXSVAUH5"), wantErr: false}, // lnurl.com.btc
		{desc: "NO_SCHEMA_INVALID_TLD_NO_ERROR_2",
			args: args{actualurl: "https://lnurl.com.btc"},                                           //invalid tld
			want: strings.ToUpper("LNURL1DP68GURN8GHJ7MRWW4EXCTNRDAKJUCN5VVXPFXPM"), wantErr: false}, // lnurl.com.btc
		{desc: "INVALID_DOMAIN_NAME_ERROR",
			args: args{actualurl: "lnurl"},
			want: strings.ToUpper("lnurl1d3h82unvjhypn2"), wantErr: false},
		{desc: "INVALID_TLD_ERROR",
			args: args{actualurl: "lnurl.msfmsdkfns&dfandfasdf"},                                               // invalid tld
			want: strings.ToUpper("LNURL1D3H82UNV9EKHXENDWDJXKENWWVNXGENPDEJXVCTNV3NQXM32ED"), wantErr: false}, // lnurl.msfmsdkfns&dfandfasdf
		{desc: "NO_SCHEMA_ERROR",
			args: args{actualurl: "lnurl.com.org"},
			want: strings.ToUpper("LNURL1D3H82UNV9E3K7MFWDAEXWSE0TLM"), wantErr: false}, // https://lnurl.com.org,
		{desc: "NO_SCHEMA_TLD_COM",
			args: args{actualurl: "lnurl.com"},
			want: "LNURL1D3H82UNV9E3K7MG347503", wantErr: false}, // https://lnurl.com
		{desc: "COM_ORG_NO_SCHEMA",
			args: args{actualurl: "lnurl.com.org"},
			want: strings.ToUpper("LNURL1D3H82UNV9E3K7MFWDAEXWSE0TLM"), wantErr: false}, // https://lnurl.com.org,
		{desc: "LUD17_P_ONION_BECH",
			args: args{actualurl: "lnurlp://api.fiatjaf.onion/v2/lnurl/pay"},
			want: "LNURL1D3H82UNVWQAZ7TMPWP5JUENFV96X5CTX9EHKU6T0DCHHVV30D3H82UNV9ACXZ7GLKMARH"}, // http://api.fiatjaf.onion/v2/lnurl/pay
		{desc: "LUD17_A_ONION_BECH",
			args: args{actualurl: "lnurla://api.fiatjaf.onion/v2/lnurl/pay"},
			want: "LNURL1D3H82UNVVYAZ7TMPWP5JUENFV96X5CTX9EHKU6T0DCHHVV30D3H82UNV9ACXZ7GEDZQYL"}, // http://api.fiatjaf.onion/v2/lnurl/pay
		{desc: "LUD17_W_ONION_BECH",
			args: args{actualurl: "lnurlw://api.fiatjaf.onion/v2/lnurl/pay"},
			want: "LNURL1D3H82UNVWUAZ7TMPWP5JUENFV96X5CTX9EHKU6T0DCHHVV30D3H82UNV9ACXZ7GRWJJDV"}, // lnurlp://api.fiatjaf.onion/v2/lnurl/pay
		{desc: "LUD17_IP_AND_PORT_BECH",
			args: args{actualurl: "lnurlp://1.1.1.1:8080/v1/lnurl/pay"},
			want: "LNURL1D3H82UNVWQAZ7TE39CCJUVFWXYARSVPCXQHHVVF0D3H82UNV9ACXZ7G5A0M2Q"}, // https://1.1.1.1:8080/v1/lnurl/pay
		{desc: "LUD17_IP_BECH",
			args: args{actualurl: "lnurlp://1.1.1.1/v1/lnurl/pay"},
			want: "LNURL1D3H82UNVWQAZ7TE39CCJUVFWXYHHVVF0D3H82UNV9ACXZ7GTCNKJU"}, // https://1.1.1.1/v1/lnurl/pay
		{desc: "LUD1_ONION",
			args: args{actualurl: "http://juhaavcaxhxlp77nkq76byazcldy2hlmovfu2epvl5ankdibsot4csyd.onion/"},
			want: strings.ToUpper("lnurl1dp68gup69uhk5atgv9shvcmp0p58smrsxumku6m3xumxy7tp0f3kcerexf5xcmt0wen82vn9wpmxcdtpde4kg6tzwdhhgdrrwdukgtn0de5k7m30p4k848"), wantErr: false},
		{desc: "HTTPS",
			args: args{actualurl: "https://api.fiatjaf.com/v2/lnurl/pay"},
			want: strings.ToUpper("lnurl1dp68gurn8ghj7ctsdyhxv6tpw34xze3wvdhk6tmkxghkcmn4wfkz7urp0y0q3peg")}}
	for _, tt := range tests {
		tt.name = tt.args.actualurl
		t.Run(fmt.Sprintf("%s", tt.desc), func(t *testing.T) {
			got, err := LNURLEncode(tt.args.actualurl)
			if (err != nil) != tt.wantErr {
				t.Errorf("LNURLDecodeStrict() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LNURLDecodeStrict() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLNURLDecodeStrict will test if LNURLDecodeStrict returns the expected string
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
	}{{desc: "INVALID_IP_NO_ERROR",
		args: args{code: "lnurl1dahxjmmw8ghj7vf3xyhryt3e8yujuv368quq39nwhr"},
		want: "onion://111.2.999.2:88", wantErr: false},
		{desc: "LUD17_CHANGE_SCHEMA",
			args: args{code: "lnurlp://lnurl.fiatjaf.com"}, // change scheme
			want: "https://lnurl.fiatjaf.com"},
		{desc: "ONION_SCHEMA_CHANGE_SCHEMA_ERROR",
			args: args{code: "onion://lnurl.fiatjaf.onion"}, // change scheme
			want: "", wantErr: true},
		{desc: "LUD17_BECH32_ENCODED_CHANGE_SCHEMA_ERROR",
			args: args{code: "lnurl1d3h82unvwqaz7tmpwp5juenfv96x5ctx9e3k7mf0wccj7mrww4exctmsv9usxr8j98"}, // lnurlp://api.fiatjaf.com/v1/lnurl/pay
			want: "lnurlp://api.fiatjaf.com/v1/lnurl/pay", wantErr: false},
		{desc: "LUD17_BECH32_ENCODED_CHANGE_SCHEMA_IP_ERROR",
			args: args{code: "lnurl1dp68gup69uhnzt339ccjuvf0wccj7mrww4exctmsv9usux0rql"}, // lnurlp://api.fiatjaf.com/v1/lnurl/pay
			want: "http://1.1.1.1/v1/lnurl/pay", wantErr: false},
		{desc: "LUD1_ENCODED_CHANGE_SCHEMA",
			args: args{code: "lnurl1dp68gurn8ghj7vfwxyhrzt339amrztmvde6hymp0wpshjr50pch"}, // https://1.1.1.1/v1/lnurl/pay
			want: "https://1.1.1.1/v1/lnurl/pay"},
		{desc: "LUD1_PUNY_SAME_SCHEMA",
			args: args{code: "lnurl1dp68gurn8ghj77rw95khx7r3wc6kwv3nv3uhyvmpxserscfwvdhk6tmkxyhkcmn4wfkz7urp0y6jmn9j"}, // https://xn--sxqv5g23dyr3a428a.com/v1/lnurl/pay
			want: "https://xn--sxqv5g23dyr3a428a.com/v1/lnurl/pay"},                                                    // check puny code
		{desc: "LUD1_IDN_SAME_SCHEMA",
			args: args{code: "lnurl1dp68gurn8ghjle443052l909sxr7t8ulukgg6tnrdakj7a339akxuatjdshhqctea4v3jm"},
			want: "https://测试假域名.com/v1/lnurl/pay"}, // check puny code
		{desc: "LUD17_BECH32_IDN_CHANGE_SCHEMA_ERROR",
			args: args{code: "lnurl1d3h82unvwuaz7tlxkk973tu4ukqc0evlnljeprfwvdhk6tmkxyhkcmn4wfkz7urp0y93ltxj"},
			want: "lnurlw://测试假域名.com/v1/lnurl/pay", wantErr: false}, // check puny code
		{desc: "LUD1_BECH32_ONION_SAME_SCHEMA",
			args: args{code: "lnurl1dp68gup69uh7ddvtazhetevpsljel8l9jzxjummwd9hkutmkxyhkcmn4wfkz7urp0ygc52rr"}, // http://测试假域名.onion/v1/lnurl/pay
			want: "http://测试假域名.onion/v1/lnurl/pay"},                                                           // check puny onion (not sure know if this is possible)
		{desc: "LUD1_BECH32_ONION_CHANGE_SCHEMA_ERROR",
			args: args{code: "lnurl1dp68gurn8ghj7er0d4skjm3wdahxjmmw9amrztmvde6hymp0wpshjcw5kvw"},
			want: "https://domain.onion/v1/lnurl/pay", wantErr: false}, // check puny onion (not sure know if this is possible)
		{desc: "RANDOM_STRING_ERROR",
			args: args{code: "lnurl1d3h82unvjhypn2"},
			want: "lnurl", wantErr: false}, // invalid domain name. returns error and decoded input
		{desc: "RANDOM_STRING_UPPERCASE_ERROR",
			args: args{code: strings.ToUpper("lnurl1d3h82unvjhypn2")},
			want: "lnurl", wantErr: false},
		{desc: "HTTPS",
			args: args{code: "https://lnurl.fiatjaf.com"}, // do noting
			want: "https://lnurl.fiatjaf.com"},
		{desc: "ONION_SCHEMA_CHANGE_SCHEMA_ERROR",
			args: args{code: "lnurl1dahxjmmw8ghj7mrww4exctnxd9shg6npvchx7mnfdahquxueex"}, // change scheme
			want: "onion://lnurl.fiatjaf.onion", wantErr: false},
		{desc: "ONION_HTTPS_CHANGE_SCHEMA_ERROR",
			args: args{code: "lnurl1dp68gurn8ghj7mrww4exctnxd9shg6npvchx7mnfdahq874q6e"}, // change scheme
			want: "https://lnurl.fiatjaf.onion", wantErr: false},
	}
	for _, tt := range tests {
		tt.name = tt.args.code
		t.Run(fmt.Sprintf("%s", tt.desc), func(t *testing.T) {
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

// TestLNURLEncodeDecode will test if encoding and decoding multiple times returns the expected string
// this test will fail without validation
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
			args: args{input: "lnurl.fiatjaf.com"}, want: "https://lnurl.fiatjaf.com", wantErr: true},
		{desc: "no_domain_string",
			args: args{input: "lnurl"}, want: "lnurl", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.desc), func(t *testing.T) {
			// encode string first
			enc, err := LNURLEncodeStrict(tt.args.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LNURLDecodeStrict() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			// decode string
			dec, err := LNURLDecodeStrict(enc)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LNURLDecodeStrict() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			// decode string
			dec, err = LNURLDecodeStrict(enc)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LNURLDecodeStrict() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			enc, err = LNURLEncodeStrict(tt.args.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LNURLDecodeStrict() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			// decode string
			dec, err = LNURLDecodeStrict(enc)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LNURLDecodeStrict() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			// check if decoded string == actualurl
			if dec != tt.want {
				t.Errorf("LNURLDecodeStrict() enc = %v, dec %v", enc, dec)
			}
		})
	}
}

// TestLNURLEncodeDecode will test if encoding strict and decoding default returns the expected string
func TestLNURLEncodeStrictDecode(t *testing.T) {
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
		{desc: "ONION_ADD_SCHEME",
			args: args{input: "hellosir.onion"}, wantErr: true, want: "http://hellosir.onion"}, // will be encoded to http://
		{desc: "HTTPS",
			args: args{input: "https://api.fiatjaf.com/v1/lnurl/pay"}, want: "https://api.fiatjaf.com/v1/lnurl/pay"},
		{desc: "ONION_SCHEME",
			args: args{input: "onion://lnurl.fiatjaf.com"}, want: "https://lnurl.fiatjaf.com", wantErr: true},
		{desc: "lnurlp",
			args: args{input: "lnurlp://lnurl.fiatjaf.com"}, want: "https://lnurl.fiatjaf.com", wantErr: false}, // will be encoded to https://
		{desc: "NO_SCHEME",
			args: args{input: "lnurl.fiatjaf.com"}, want: "https://lnurl.fiatjaf.com", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.desc), func(t *testing.T) {
			// encode string first
			enc, err := LNURLEncodeStrict(tt.args.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LNURLDecodeStrict() error = %v, wantErr %v", err, tt.wantErr)
					return

				}
			}
			// decode string
			dec, err := LNURLDecode(enc)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LNURLDecodeStrict() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

			}
			// check if decoded string == actualurl
			if dec != tt.want {
				t.Errorf("LNURLDecodeStrict() enc = %v, dec %v", enc, dec)
			}
		})
	}
}

// TestLNURLEncodeDecodeStrict will test if encoding default and decoding strict returns the expected string
func TestLNURLEncodeDecodeStrict(t *testing.T) {
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
		{desc: "ONION_ADD_SCHEME",
			args: args{input: "hellosir.onion"}, wantErr: true, want: "http://hellosir.onion"}, // will be encoded to http://
		{desc: "HTTPS",
			args: args{input: "https://api.fiatjaf.com/v1/lnurl/pay"}, want: "https://api.fiatjaf.com/v1/lnurl/pay"},
		{desc: "ONION_SCHEME",
			args: args{input: "onion://lnurl.fiatjaf.com"}, want: "https://lnurl.fiatjaf.com", wantErr: true},
		{desc: "lnurlp",
			args: args{input: "lnurlp://lnurl.fiatjaf.com"}, want: "https://lnurl.fiatjaf.com", wantErr: true}, // will be encoded to https://
		{desc: "NO_SCHEME",
			args: args{input: "lnurl.fiatjaf.com"}, want: "https://lnurl.fiatjaf.com", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.desc), func(t *testing.T) {
			// encode string first
			enc, err := LNURLEncode(tt.args.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LNURLDecodeStrict() error = %v, wantErr %v", err, tt.wantErr)
					return

				}
			}
			// decode string
			dec, err := LNURLDecodeStrict(enc)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LNURLDecodeStrict() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

			}
			// check if decoded string == actualurl
			if dec != tt.want {
				t.Errorf("LNURLDecodeStrict() enc = %v, dec %v", enc, dec)
			}
		})
	}
}
