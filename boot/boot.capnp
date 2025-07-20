using Go = import "/go.capnp";

@0xe82706a772b0927b;

$Go.package("boot");
$Go.import("github.com/wetware/go/boot");


interface Env {
    type  @0 () -> (schema :import "/schema.capnp".Node);
    node @1 () -> (value :AnyPointer);
}
