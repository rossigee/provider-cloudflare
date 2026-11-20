/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clients

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
)

const (
	testAccountA = "c319c02aa39a98bfaaad068b83c0b179"
	testAccountB = "c1b74f148aee28025816e104a92622c5"
)

func listAccounts(accounts []cloudflare.Account, err error) listAccountsFn {
	return func(context.Context, cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
		return accounts, cloudflare.ResultInfo{}, err
	}
}

func TestResolveAccountID(t *testing.T) {
	cases := map[string]struct {
		configured string
		list       listAccountsFn
		want       string
	}{
		"ConfiguredAlwaysWins": {
			configured: testAccountB,
			list: listAccounts([]cloudflare.Account{
				{ID: testAccountA}, {ID: testAccountB},
			}, nil),
			want: testAccountB,
		},
		"SingleAccountFromAPI": {
			list: listAccounts([]cloudflare.Account{{ID: testAccountB}}, nil),
			want: testAccountB,
		},
		"MultipleAccountsAreAmbiguous": {
			list: listAccounts([]cloudflare.Account{
				{ID: testAccountA}, {ID: testAccountB},
			}, nil),
			want: "",
		},
		"NoAccounts": {
			list: listAccounts(nil, nil),
			want: "",
		},
		"APIError": {
			list: listAccounts(nil, errors.New("boom")),
			want: "",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := resolveAccountID(context.Background(), tc.configured, tc.list); got != tc.want {
				t.Errorf("resolveAccountID() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestConfiguredAccountID(t *testing.T) {
	if got := ConfiguredAccountID(nil); got != "" {
		t.Errorf("ConfiguredAccountID(nil) = %q, want empty", got)
	}
	if got := ConfiguredAccountID(&Config{}); got != "" {
		t.Errorf("ConfiguredAccountID(empty) = %q, want empty", got)
	}
	id := testAccountB
	if got := ConfiguredAccountID(&Config{AccountID: &id}); got != id {
		t.Errorf("ConfiguredAccountID() = %q, want %q", got, id)
	}
}
