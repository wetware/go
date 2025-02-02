using Go = import "/go.capnp";

@0xf381800d6f8057ad;

$Go.package("glia");
$Go.import("github.com/wetware/go/glia");


struct Header {
    peer   @0 :Data;
    proc   @1 :Text;
    method @2 :Text;
    stack  @3 :List(UInt64);
}


struct Result {
    stack  @0 :List(UInt64);
    status @1 :Status;
    info   @2 :Text;

    enum Status {
        unset          @0;
        ok             @1;
        invalidRequest @2;
        routingError   @3;
        procNotFound   @4;
        invalidMethod  @5;
        methodNotFound @6;
        guestError     @7;
    }
}
