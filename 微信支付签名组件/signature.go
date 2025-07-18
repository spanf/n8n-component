package wechatpay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
)

func GenerateSignature(method string, url string, timestamp string, nonce string, body string, privateKey *rsa.PrivateKey) (string, error) {
	if privateKey == nil {
		return "", errors.New("private key is nil")
	}

	signStr := method + "\n" + url + "\n" + timestamp + "\n" + nonce + "\n" + body + "\n"
	hashed := sha256.Sum256([]byte(signStr))

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func VerifySignature(signature string, timestamp string, nonce string, body string, cert *x509.Certificate) (bool, error) {
	if cert == nil {
		return false, errors.New("certificate is nil")
	}

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return false, errors.New("certificate does not contain RSA public key")
	}

	signBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}

	signStr := timestamp + "\n" + nonce + "\n" + body + "\n"
	hashed := sha256.Sum256([]byte(signStr))

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signBytes)
	if err != nil {
		if err == rsa.ErrVerification {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
