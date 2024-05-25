using Go = import "/go.capnp";

@0x910958683ac350d8;

$Go.package("system");
$Go.import("github.com/wetware/go/system");

interface Proc {
    handle @0 (event :Data) -> ();
}
