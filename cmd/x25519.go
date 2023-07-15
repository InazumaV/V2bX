package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Yuzuki616/V2bX/common/crypt"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/curve25519"
)

var x25519Command = cobra.Command{
	Use:   "x25519",
	Short: "Generate key pair for x25519 key exchange",
	Run: func(cmd *cobra.Command, args []string) {
		executeX25519()
	},
}

func init() {
	command.AddCommand(&x25519Command)
}

func executeX25519() {
	var yes, key string
	fmt.Println("要对私钥进行加密吗?(Y/n)")
	fmt.Scan(&yes)
	if strings.ToLower(yes) == "y" {
		var temp string
		fmt.Println("请输入Api接口地址:")
		fmt.Scan(&temp)
		key = temp
		fmt.Println("请输入Api认证Token:")
		fmt.Scan(&temp)
		key += temp
		key = crypt.GenShaHash([]byte(key))
	}
	var output string
	var err error
	defer func() {
		fmt.Println(output)
	}()
	var privateKey []byte
	var publicKey []byte
	privateKey = make([]byte, curve25519.ScalarSize)
	if _, err = rand.Read(privateKey); err != nil {
		output = Err("read rand error: ", err)
		return
	}

	// Modify random bytes using algorithm described at:
	// https://cr.yp.to/ecdh.html.
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	if publicKey, err = curve25519.X25519(privateKey, curve25519.Basepoint); err != nil {
		output = Err("gen X25519 error: ", err)
		return
	}
	p := base64.RawURLEncoding.EncodeToString(privateKey)
	output = fmt.Sprint("Private key: ",
		p,
		"\nPublic key: ",
		base64.RawURLEncoding.EncodeToString(publicKey))
	if strings.ToLower(yes) == "y" {
		key, err = crypt.AesEncrypt([]byte(p), []byte(key[:32]))
		if err != nil {
			output = Err("encrypt private key error: ", err)
		}
		output += "\n加密后的Private key：" + key
	}
}
