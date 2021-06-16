package main

import (
	"github.com/umputun/reproxy/lib"
	"net/textproto"
	"testing"
)

func TestHandler_getIP(t *testing.T) {
	type fields struct {
		ipSource string
	}
	type args struct {
		req lib.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:   "empty remote addr",
			fields: fields{ipSource: ""},
			args: args{
				req: lib.Request{},
			},
			want:    "",
			wantErr: true,
		},
		{
			name:   "empty remote addr, with empty value of ipSource",
			fields: fields{ipSource: "XFF"},
			args: args{
				req: lib.Request{},
			},
			want:    "",
			wantErr: true,
		},
		{
			name:   "empty remote addr, with ipSource",
			fields: fields{ipSource: "xff"},
			args: args{
				req: lib.Request{Header: map[string][]string{textproto.CanonicalMIMEHeaderKey("xff"): {"10.20.30.40"}}},
			},
			want:    "10.20.30.40",
			wantErr: false,
		},
		{
			name:   "with remote addr, no port",
			fields: fields{ipSource: ""},
			args: args{
				req: lib.Request{RemoteAddr: "10.20.30.40"},
			},
			want:    "",
			wantErr: true,
		},
		{
			name:   "with remote addr, with port",
			fields: fields{ipSource: ""},
			args: args{
				req: lib.Request{RemoteAddr: "10.20.30.40:1234"},
			},
			want:    "10.20.30.40",
			wantErr: false,
		},
		{
			name:   "with remote addr, ipv6",
			fields: fields{ipSource: ""},
			args: args{
				req: lib.Request{RemoteAddr: "1fff:0:a88:85a3::ac1f"},
			},
			want:    "",
			wantErr: true,
		},
		{
			name:   "with remote addr, ipv6",
			fields: fields{ipSource: ""},
			args: args{
				req: lib.Request{RemoteAddr: "[1fff:0:a88:85a3::ac1f]:1234"},
			},
			want:    "1fff:0:a88:85a3::ac1f",
			wantErr: false,
		},
		{
			name:   "wrong ip",
			fields: fields{ipSource: ""},
			args: args{
				req: lib.Request{RemoteAddr: "abc"},
			},
			want:    "",
			wantErr: true,
		},
		{
			name:   "wrong ip in header",
			fields: fields{ipSource: "xff"},
			args: args{
				req: lib.Request{Header: map[string][]string{textproto.CanonicalMIMEHeaderKey("xff"): {"abc"}}},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				ipSource: tt.fields.ipSource,
			}
			got, err := h.getIP(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getIP() got = %v, want %v", got, tt.want)
			}
		})
	}
}
