package lego

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/Yuzuki616/V2bX/common/file"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/goccy/go-json"
	"os"
)

type User struct {
	Email        string                 `json:"Email"`
	Registration *registration.Resource `json:"Registration"`
	key          crypto.PrivateKey
	KeyEncoded   string `json:"Key"`
}

func (u *User) GetEmail() string {
	return u.Email
}
func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func NewUser(path string, email string) (*User, error) {
	var user User
	if file.IsExist(path) {
		err := user.Load(path)
		if err != nil {
			return nil, err
		}
		if user.Email != email {
			user.Registration = nil
			user.Email = email
			err := registerUser(&user, path)
			if err != nil {
				return nil, err
			}
		}
	} else {
		user.Email = email
		err := registerUser(&user, path)
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}

func registerUser(user *User, path string) error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key error: %s", err)
	}
	user.key = privateKey
	c := lego.NewConfig(user)
	client, err := lego.NewClient(c)
	if err != nil {
		return fmt.Errorf("create lego client error: %s", err)
	}
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}
	user.Registration = reg
	err = user.Save(path)
	if err != nil {
		return fmt.Errorf("save user error: %s", err)
	}
	return nil
}

func EncodePrivate(privKey *ecdsa.PrivateKey) (string, error) {
	encoded, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return "", err
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: encoded})
	return string(pemEncoded), nil
}
func (u *User) Save(path string) error {
	err := checkPath(path)
	if err != nil {
		return fmt.Errorf("check path error: %s", err)
	}
	u.KeyEncoded, _ = EncodePrivate(u.key.(*ecdsa.PrivateKey))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(u)
	if err != nil {
		return fmt.Errorf("marshal json error: %s", err)
	}
	u.KeyEncoded = ""
	return nil
}

func (u *User) DecodePrivate(pemEncodedPriv string) (*ecdsa.PrivateKey, error) {
	blockPriv, _ := pem.Decode([]byte(pemEncodedPriv))
	x509EncodedPriv := blockPriv.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509EncodedPriv)
	return privateKey, err
}
func (u *User) Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file error: %s", err)
	}
	err = json.NewDecoder(f).Decode(u)
	if err != nil {
		return fmt.Errorf("unmarshal json error: %s", err)
	}
	u.key, err = u.DecodePrivate(u.KeyEncoded)
	if err != nil {
		return fmt.Errorf("decode private key error: %s", err)
	}
	return nil
}
