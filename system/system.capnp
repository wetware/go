using Go = import "/go.capnp";

@0xda965b22da734daf;

$Go.package("system");
$Go.import("github.com/wetware/go/system");


interface Terminal {
    login @0 () -> (
        exec :Executor,
    );
}

interface Executor {
    exec @0 (bytecode :Data) -> (protocol :Text);
}

struct MethodCall {
    name @0 :Text;
    stack  @1 :List(UInt64);
}
