using Go = import "/go.capnp";
@0xc13229b64d08e68a;
$Go.package("proc");
$Go.import("proc");


struct MethodCall {
    name @0 :Text;
    # Method name

    stack @1 :List(UInt64);
    # WASM call stack

    callData @2 :Data;
    # Method call data is written to stdin
}
