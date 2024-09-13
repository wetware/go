//go:generate mockgen -source=auth.go -destination=auth_mock_test.go -package=auth_test AuthProvider
//go:generate capnp compile -I$GOPATH/src/capnproto.org/go/capnp/std -ogo auth.capnp

package auth
