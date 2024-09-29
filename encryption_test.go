package password

import "testing"

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
