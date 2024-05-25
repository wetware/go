using Go = import "/go.capnp";

@0xb73f1b42636e2285;

$Go.package("ww");
$Go.import("github.com/wetware/go");

interface Signer {
    sign @0 (nonce :Data) -> (envelope :Data);
}

interface Terminal {
    login @0 (account :Signer) -> ();
}
