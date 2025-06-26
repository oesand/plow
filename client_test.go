package giglet

import (
	"context"
	"crypto/tls"
	"github.com/oesand/giglet/specs"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestClient_MakeContext(t *testing.T) {
	t.Errorf("Cover client & server")
	t.Errorf("Add more encodings support!!")
	type fields struct {
		ReadLineMaxLength   int64
		HeadMaxLength       int64
		MaxBodySize         int64
		MaxRedirectCount    int
		ReadTimeout         time.Duration
		WriteTimeout        time.Duration
		TLSHandshakeTimeout time.Duration
		Header              *specs.Header
		Jar                 *specs.CookieJar
		TLSConfig           *tls.Config
		TLSHandshakeContext func(ctx context.Context, conn net.Conn, host string) (net.Conn, error)
		DialContext         func(ctx context.Context, network, addr string) (net.Conn, error)
		mu                  sync.RWMutex
	}
	type args struct {
		ctx     context.Context
		request ClientRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ClientResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cln := &Client{
				ReadLineMaxLength:   tt.fields.ReadLineMaxLength,
				HeadMaxLength:       tt.fields.HeadMaxLength,
				MaxBodySize:         tt.fields.MaxBodySize,
				MaxRedirectCount:    tt.fields.MaxRedirectCount,
				ReadTimeout:         tt.fields.ReadTimeout,
				WriteTimeout:        tt.fields.WriteTimeout,
				TLSHandshakeTimeout: tt.fields.TLSHandshakeTimeout,
				Header:              tt.fields.Header,
				Jar:                 tt.fields.Jar,
				TLSConfig:           tt.fields.TLSConfig,
				TLSHandshakeContext: tt.fields.TLSHandshakeContext,
				DialContext:         tt.fields.DialContext,
				mu:                  tt.fields.mu,
			}
			got, err := cln.MakeContext(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeContext() got = %v, want %v", got, tt.want)
			}
		})
	}
}
