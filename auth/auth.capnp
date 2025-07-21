using Go = import "/go.capnp";

@0xe82706a772b0927b;

$Go.package("auth");
$Go.import("github.com/wetware/go/auth");

# Signer identifies an account.  It is a capability that can be
# used to sign arbitrary nonces.
#
# The signature domain is "ww.auth"
interface Signer {
    sign @0 (src :Data) -> (rawEnvelope :Data);
}


interface Terminal {
    login @0 (account :Signer) -> (session :Env);
}


struct Env {
    schema @0 :import "/schema.capnp".Node;
}

