using Go = import "/go.capnp";

@0xf381800d6f8057ad;

$Go.package("glia");
$Go.import("github.com/wetware/go/glia");


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


struct CallData {
    stack  @0 :List(UInt64);
    method @1 :Text;
}


struct Result {
    stack  @0 :List(UInt64);
    status @1 :Status;
    info   @2 :Text;
}
