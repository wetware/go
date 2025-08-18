using Go = import "/go.capnp";

@0xda965b22da734daf;

$Go.package("system");
$Go.import("github.com/wetware/go/system");


interface Importer {
    import @0 (envelope :Data) -> (service :Capability);
}