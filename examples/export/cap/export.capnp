using Go = import "/go.capnp";

@0xa0266946850e6061;

$Go.package("cap");
$Go.import("github.com/wetware/go/examples/export/cap");


interface Greeter {
    greet @0 (name :Text) -> (greeting :Text);
}
