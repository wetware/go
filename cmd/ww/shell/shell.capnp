using Go = import "/go.capnp";

@0xead48f650c32a806;

$Go.package("main");
$Go.import("github.com/wetware/go/shell");


interface EventLoop {
    addActor @0 (handler :Text) -> (mailbox :Mailbox);
}

interface Mailbox {
    newBuffer @0 () -> (buffer :Buffer);
}

interface Buffer {
    write       @0 (input :Data) -> (status :Status);
    writeString @1 (input :Text) -> (status :Status);
    read        @2 (count :UInt64) -> (output :Data, status :Status);
    flush       @3 ();
    struct Status {
        union {
            ok    @0 :Void;
            eof   @1 :Void;
            error @2 :Text;
        }
    }
}
