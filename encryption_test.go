package password

import (
	"testing"
)

func Test_encrypt_decrypt(t *testing.T) {
	type args struct {
		text   string
		secret string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"first", args{"foo", "123456"}, "foo", false},
		{"second", args{"foo", "789"}, "foo", false},
		{"third", args{"bar", "123456"}, "bar", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err1 := encrypt(tt.args.text, tt.args.secret)
			got, err2 := decrypt(got, tt.args.secret)
			if (err1 != nil) != tt.wantErr {
				t.Errorf("encrypt() error = %v, wantErr %v", err1, tt.wantErr)
				return
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("encrypt() error = %v, wantErr %v", err2, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("encrypt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_EncryptOTP_DecryptOTP(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
	}{
		{"first", args{"foo"}},
		{"second", args{"bar"}},
		{"third", args{"foobar"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipherBytes, secret := EncryptOTP(tt.args.text)
			text := DecryptOTP(cipherBytes, secret)
			if text != tt.args.text {
				t.Errorf("encrypt_decrypt() got = %v, want %v", text, tt.args.text)
			}
		})
	}
}

func Test_compareHashedPassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"success", args{"foo"}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := getHashedPassword(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("getHashedPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := compareHashedPassword(hashedPassword, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("compareHashedPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("compareHashedPassword() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_packData(t *testing.T) {
	type args struct {
		id   string
		data string
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{"normal", args{"foo", "bar"}, 94, false},
		{"no escape", args{"foo<>&", "bar<>&"}, 94, false},
		{"empty", args{"", ""}, 94, false},
		{"short", args{"", "123456789012345"}, 94, false},
		{"long", args{"", "1234567890123456"}, 110, false},
		{"also long", args{"", "12345678901234567"}, 110, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := packData(tt.args.id, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("packData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (len(got) - len(tt.args.id)) != tt.wantLen {
				t.Errorf("len(packData()) got = %v, want %v", len(got), tt.wantLen+len(tt.args.id))
			}

			id, data, err := unpackData(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("unpackData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if id != tt.args.id {
				t.Errorf("packData() got = %v, want %v", id, tt.args.id)
			}
			if data != tt.args.data {
				t.Errorf("packData() got = %v, want %v", data, tt.args.data)
			}
		})
	}
}
