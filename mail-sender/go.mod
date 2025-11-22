module github.com/isimtekin/go-packages/mail-sender

go 1.24.0

toolchain go1.24.4

require (
	github.com/isimtekin/go-packages/env-util v0.0.2
	github.com/sendgrid/sendgrid-go v3.14.0+incompatible
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	golang.org/x/net v0.47.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/isimtekin/go-packages/env-util => ../env-util
