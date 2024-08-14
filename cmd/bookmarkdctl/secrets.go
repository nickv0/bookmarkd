package main

import (
	"aidanwoods.dev/go-paseto"
)

func secret() {
	secretKey := paseto.NewV4AsymmetricSecretKey()

	println(secretKey.ExportHex())
}
