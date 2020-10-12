// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package fake

import (
	"strings"

	v1 "github.com/marmotedu/api/apiserver/v1"
	"github.com/marmotedu/component-base/pkg/fields"
	metav1 "github.com/marmotedu/component-base/pkg/meta/v1"
	"github.com/marmotedu/component-base/pkg/util/stringutil"
	"github.com/marmotedu/errors"
	"github.com/marmotedu/iam/internal/pkg/code"
	"github.com/marmotedu/iam/internal/pkg/util/gormutil"
	reflectutil "github.com/marmotedu/iam/internal/pkg/util/reflect"
)

type secrets struct {
	ds *datastore
}

func newSecrets(ds *datastore) *secrets {
	return &secrets{ds}
}

// Create creates a new secret account.
func (s *secrets) Create(secret *v1.Secret, opts metav1.CreateOptions) error {
	s.ds.Lock()
	defer s.ds.Unlock()

	for _, sec := range s.ds.secrets {
		if sec.Username == secret.Username && sec.Name == secret.Name {
			return errors.New("record already exist")
		}
	}

	if len(s.ds.secrets) > 0 {
		secret.ID = s.ds.secrets[len(s.ds.secrets)-1].ID + 1
	}
	s.ds.secrets = append(s.ds.secrets, secret)

	return nil
}

// Update updates an secret information by the secret identifier.
func (s *secrets) Update(secret *v1.Secret, opts metav1.UpdateOptions) error {
	s.ds.Lock()
	defer s.ds.Unlock()

	for _, sec := range s.ds.secrets {
		if sec.Username == secret.Username && sec.Name == secret.Name {
			if _, err := reflectutil.CopyObj(secret, sec, nil); err != nil {
				return err
			}
		}
	}

	return nil
}

// Delete deletes the secret by the secret identifier.
func (s *secrets) Delete(username, name string, opts metav1.DeleteOptions) error {
	s.ds.Lock()
	defer s.ds.Unlock()

	secrets := s.ds.secrets
	s.ds.secrets = make([]*v1.Secret, 0)
	for _, sec := range secrets {
		if sec.Username == username && sec.Name == name {
			continue
		}

		s.ds.secrets = append(s.ds.secrets, sec)
	}

	return nil
}

// DeleteCollection batch deletes the secrets.
func (s *secrets) DeleteCollection(username string, names []string, opts metav1.DeleteOptions) error {
	s.ds.Lock()
	defer s.ds.Unlock()

	secrets := s.ds.secrets
	s.ds.secrets = make([]*v1.Secret, 0)
	for _, sec := range secrets {
		if sec.Username == username && stringutil.StringIn(sec.Name, names) {
			continue
		}

		s.ds.secrets = append(s.ds.secrets, sec)
	}

	return nil
}

// Get return an secret by the secret identifier.
func (s *secrets) Get(username, name string, opts metav1.GetOptions) (*v1.Secret, error) {
	s.ds.RLock()
	defer s.ds.RUnlock()

	for _, sec := range s.ds.secrets {
		if sec.Username == username && sec.Name == name {
			return sec, nil
		}
	}

	return nil, errors.WithCode(code.ErrSecretNotFound, "record not found")
}

// List return all secrets.
func (s *secrets) List(username string, opts metav1.ListOptions) (*v1.SecretList, error) {
	s.ds.RLock()
	defer s.ds.RUnlock()

	ol := gormutil.Unpointer(opts.Offset, opts.Limit)
	selector, _ := fields.ParseSelector(opts.FieldSelector)
	name, _ := selector.RequiresExactMatch("name")

	secrets := make([]*v1.Secret, 0)
	var i int64 = 0
	for _, sec := range s.ds.secrets {
		if i == ol.Limit {
			break
		}

		if sec.Username != username {
			continue
		}

		if !strings.Contains(sec.Name, name) {
			continue
		}

		secrets = append(secrets, sec)
		i++
	}

	return &v1.SecretList{
		ListMeta: metav1.ListMeta{
			TotalCount: uint64(len(s.ds.secrets)),
		},
		Items: secrets,
	}, nil
}