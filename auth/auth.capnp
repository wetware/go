using Go = import "/go.capnp";

@0xe82706a772b0927b;

$Go.package("auth");
$Go.import("github.com/wetware/go/auth");


# Signer identifies an accound.  It is a capability that can be
# used to sign arbitrary nonces.
interface Signer {
    sign @0 (data :Data) -> (rawEnvelope :Data);
}


interface Terminal {
    login @0 (account :Signer) -> (
        stdio :Socket,
    );
}

struct Socket {
    reader @0 :ReadPipe;
    writer @1 :WritePipe;
    error  @2 :WritePipe;
}


interface ReadPipe {
    read @0 (size :Int64) -> (data :Data, eof :Bool);
}

interface WritePipe {
    write @0 (data :Data) -> (n :Int64);
}
